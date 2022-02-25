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
	repoAgg, err := s.licenseRepo.GetAggregationDetails(ctx, req.Name, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/GetAggregationDetails - failed to get aggregation details", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get aggregation details")
	}
	// fmt.Println("rero agg:", repoAgg)
	aggAcqRights := &v1.AggregationAcquiredRights{
		SKU:             repoAgg.SKU,
		AggregationName: repoAgg.Name,
		SwidTags:        strings.Join(repoAgg.Swidtags, ","),
		Metric:          strings.Join(repoAgg.Metric, ","),
		NumAcqLicences:  repoAgg.Licenses,
		TotalCost:       repoAgg.TotalCost,
		AvgUnitPrice:    repoAgg.UnitPrice,
	}
	indvRights, err := s.licenseRepo.AggregationIndividualRights(ctx, repoAgg.ProductIDs, repoAgg.Metric, req.GetScope())
	if err != nil && err != repo.ErrNodeNotFound {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/AggregationIndividualRights - failed to get aggregation individual details", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get aggregation individual rights")
	}
	for _, indacq := range indvRights {
		aggAcqRights.NumAcqLicences += indacq.Licenses
	}
	if repoAgg.NumOfEquipments == 0 {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - no equipments linked with product")
		aggAcqRights.DeltaNumber = aggAcqRights.NumAcqLicences
		aggAcqRights.DeltaCost = aggAcqRights.TotalCost
		return &v1.ListAcqRightsForAggregationResponse{
			AcqRights: []*v1.AggregationAcquiredRights{
				aggAcqRights,
			},
		}, nil
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	ind := 0
	input := make(map[string]interface{})
	input[ProdAggName] = repoAgg.Name
	input[SCOPES] = []string{req.Scope}
	input[IsAgg] = true
	var maxComputed uint64
	var computedDetails string
	metricExists := false
	for _, met := range repoAgg.Metric {
		if ind = metricNameExistsAll(metrics, met); ind == -1 {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - metric name doesnt exist - " + met)
			continue
		}
		input[MetricName] = metrics[ind].Name
		if _, ok := MetricCalculation[metrics[ind].Type]; !ok {
			return nil, status.Error(codes.Internal, "this metricType is not supported")
		}
		resp, err := MetricCalculation[metrics[ind].Type](ctx, s, eqTypes, input)
		if err != nil {
			logger.Log.Error("service/v1 - Failed ListAcqRightsForProductAggregation  ", zap.String("metric name", metrics[ind].Name), zap.Any("metric type", metrics[ind].Type), zap.String("reason", err.Error()))
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
		aggAcqRights.NumCptLicences = int32(maxComputed)
		aggAcqRights.DeltaNumber = aggAcqRights.NumAcqLicences - int32(maxComputed)
		aggAcqRights.DeltaCost = aggAcqRights.TotalCost - aggAcqRights.AvgUnitPrice*float64(maxComputed)
		aggAcqRights.ComputedDetails = computedDetails
	} else {
		aggAcqRights.MetricNotDefined = true
	}
	return &v1.ListAcqRightsForAggregationResponse{
		AcqRights: []*v1.AggregationAcquiredRights{
			aggAcqRights,
		},
	}, nil
}

// func convertRepoToSrvProductAll(prods []*repo.ProductData) []*v1.Product {
// 	products := make([]*v1.Product, len(prods))
// 	for i := range prods {
// 		products[i] = convertRepoToSrvProduct(prods[i])
// 	}
// 	return products
// }

// func convertRepoToSrvProduct(prod *repo.ProductData) *v1.Product {
// 	return &v1.Product{
// 		SwidTag:           prod.Swidtag,
// 		Name:              prod.Name,
// 		Version:           prod.Version,
// 		Category:          prod.Category,
// 		Editor:            prod.Editor,
// 		NumOfApplications: prod.NumOfApplications,
// 		NumofEquipments:   prod.NumOfEquipments,
// 		TotalCost:         float64(prod.TotalCost),
// 	}
// }
