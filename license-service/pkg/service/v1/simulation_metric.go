package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ProductLicensesForMetric(ctx context.Context, req *v1.ProductLicensesForMetricRequest) (*v1.ProductLicensesForMetricResponse, error) {
	input := make(map[string]interface{})
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ProductLicensesForMetric", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}

	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	ind := metricNameExistsAll(metrics, req.MetricName)
	if ind == -1 {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - metricNameExistsAll - " + req.MetricName)
		return nil, status.Error(codes.Internal, "metric name does not exist")
	}
	metricInfo := metrics[ind]
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	if req.AggregationName == "" {
		proID, err := s.licenseRepo.ProductIDForSwidtag(ctx, req.SwidTag, &repo.QueryProducts{}, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ProductLicensesForMetric - ProductIDForSwidtag", zap.Error(err))
			return nil, status.Error(codes.NotFound, "cannot get product id for swid tag")
		}
		input[ProdID] = proID
		input[IsAgg] = false
		input[MetricName] = metricInfo.Name
		input[SCOPES] = []string{req.GetScope()}
	} else {
		repoAgg, _, err := s.licenseRepo.AggregationDetails(ctx, req.AggregationName, metrics, true, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/AggregationDetails - failed to get aggregation details", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "failed to get aggregation details")
		}
		if len(repoAgg.ProductIDs) == 0 {
			return &v1.ProductLicensesForMetricResponse{}, status.Error(codes.InvalidArgument, "Inventory Park is not present")
		}

		input[IsAgg] = true
		input[ProdAggName] = repoAgg.Name
		input[MetricName] = metricInfo.Name
		input[SCOPES] = []string{req.GetScope()}

	}
	if _, ok := MetricCalculation[metricInfo.Type]; !ok {
		logger.Log.Error("service/v1 - Failed ProductLicensesForMetric for  - ", zap.String("metric :", metricInfo.Name), zap.Any("metricType", metricInfo.Type))
		return nil, status.Error(codes.Internal, "this metricType is not supported")
	}
	resp, err := MetricCalculation[metricInfo.Type](ctx, s, eqTypes, input)
	if err != nil {
		logger.Log.Error("service/v1 - Failed ProductLicensesForMetric for  - ", zap.String("metric :", metricInfo.Name), zap.Any("metricType", metricInfo.Type), zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute licenses")
	}
	return &v1.ProductLicensesForMetricResponse{
		MetricName:     req.MetricName,
		NumCptLicences: resp[ComputedLicenses].(uint64),
		TotalCost:      float64(resp[ComputedLicenses].(uint64)) * req.UnitCost,
	}, nil
}
