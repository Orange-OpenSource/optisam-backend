package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

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
	proID, err := s.licenseRepo.ProductIDForSwidtag(ctx, req.SwidTag, &repo.QueryProducts{}, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - ProductIDForSwidtag", zap.Error(err))
		return nil, status.Error(codes.NotFound, "cannot get product id for swid tag")
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
	input[ProdID] = proID
	input[MetricName] = metricInfo.Name
	input[SCOPES] = []string{req.GetScope()}
	input[IsAgg] = false
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
