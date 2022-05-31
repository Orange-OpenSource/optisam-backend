package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListAcqRightsForAggregation(ctx context.Context, req *v1.ListAcqRightsForAggregationRequest) (*v1.ListAcqRightsForAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	repoAgg, aggRights, err := s.licenseRepo.AggregationDetails(ctx, req.Name, metrics, req.Simulation, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/AggregationDetails - failed to get aggregation details", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get aggregation details")
	}
	if len(repoAgg.ProductIDs) == 0 {
		return &v1.ListAcqRightsForAggregationResponse{}, status.Error(codes.InvalidArgument, "Inventory Park is not present")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	aggCompLicenses := make([]*v1.AggregationAcquiredRights, len(aggRights))
	ind := 0
	input := make(map[string]interface{})
	input[ProdAggName] = repoAgg.Name
	input[SCOPES] = []string{req.Scope}
	input[IsAgg] = true
	for i, aggRight := range aggRights {
		aggCompLicenses[i] = &v1.AggregationAcquiredRights{
			SKU:             aggRight.SKU,
			AggregationName: req.Name,
			SwidTags:        strings.Join(repoAgg.Swidtags, ","),
			ProductNames:    strings.Join(repoAgg.ProductNames, ","),
			Metric:          aggRight.Metric,
			NumAcqLicences:  int32(aggRight.AcqLicenses),
			TotalCost:       aggRight.TotalCost,
			PurchaseCost:    aggRight.TotalPurchaseCost,
			AvgUnitPrice:    aggRight.AvgUnitPrice,
		}
		if repoAgg.NumOfEquipments == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			aggCompLicenses[i].DeltaNumber = int32(aggRight.AcqLicenses)
			aggCompLicenses[i].DeltaCost = aggCompLicenses[i].TotalCost
			aggCompLicenses[i].NotDeployed = true
			continue
		}
		rightsMetrics := strings.Split(aggRight.Metric, ",")
		indvRights, err := s.licenseRepo.AggregationIndividualRights(ctx, repoAgg.ProductIDs, rightsMetrics, req.Scope)
		if err != nil && err != repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/AggregationIndividualRights - failed to get aggregation individual details", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "failed to get aggregation individual rights")
		}
		for _, indacq := range indvRights {
			aggCompLicenses[i].NumAcqLicences += indacq.Licenses
			aggCompLicenses[i].TotalCost += indacq.TotalCost
		}
		// fmt.Printf("acq[%d]: %v", i, aggCompLicenses[i])
		var maxComputed uint64
		var computedDetails string
		metricExists := false
		for _, met := range rightsMetrics {
			if ind = metricNameExistsAll(metrics, met); ind == -1 {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric name doesnt exist - " + met)
				continue
			}
			input[MetricName] = metrics[ind].Name
			input[SCOPES] = []string{req.GetScope()}
			if _, ok := MetricCalculation[metrics[ind].Type]; !ok {
				return nil, status.Error(codes.Internal, "this metricType is not supported")
			}
			resp, err := MetricCalculation[metrics[ind].Type](ctx, s, eqTypes, input)
			if err != nil {
				logger.Log.Error("service/v1 - Failed ListAcqRightsForProduct  ", zap.String("metric name", metrics[ind].Name), zap.Any("metric type", metrics[ind].Type), zap.String("reason", err.Error()))
				continue
			}
			computedLicenses := resp[ComputedLicenses].(uint64)
			if computedLicenses >= maxComputed {
				metricExists = true
				maxComputed = computedLicenses
				if _, ok := resp[ComputedDetails]; ok {
					computedDetails = resp[ComputedDetails].(string)
				}
			}
		}
		if metricExists {
			aggCompLicenses[i].NumCptLicences = int32(maxComputed)
			aggCompLicenses[i].DeltaNumber = aggCompLicenses[i].NumAcqLicences - int32(maxComputed)
			aggCompLicenses[i].DeltaCost = aggCompLicenses[i].TotalCost - aggRight.AvgUnitPrice*float64(int32(maxComputed))
			aggCompLicenses[i].ComputedDetails = computedDetails
			aggCompLicenses[i].ComputedCost = aggRight.AvgUnitPrice * float64(int32(maxComputed))
		} else {
			aggCompLicenses[i].MetricNotDefined = true
		}
	}
	return &v1.ListAcqRightsForAggregationResponse{
		AcqRights: aggCompLicenses,
	}, nil
}
