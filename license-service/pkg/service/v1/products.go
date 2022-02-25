package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/helper"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAcqRightsForProduct implements license service ListAcqRightsForProduct function
func (s *licenseServiceServer) ListAcqRightsForProduct(ctx context.Context, req *v1.ListAcquiredRightsForProductRequest) (*v1.ListAcquiredRightsForProductResponse, error) { // nolint
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsForProduct", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	aggregationName, err := s.licenseRepo.IsProductPurchasedInAggregation(ctx, req.SwidTag, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsForProduct - failed to check is swidtag is part of aggregation", zap.Error(err))
		return nil, status.Error(codes.Internal, "couldn't check is swidtag part of aggregation")
	} else if aggregationName != "" {
		logger.Log.Info("service/v1 - ListAcqRightsForProduct - aggregation found", zap.String("swidtag", req.SwidTag), zap.String("aggName", aggregationName))
		return &v1.ListAcquiredRightsForProductResponse{AggregationName: aggregationName}, nil
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		logger.Log.Debug("service/v1 - ListAcqRightsForProduct - ListMetrices - unable to fetch metrics:%v", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	ID, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, metrics, req.GetScope())
	if err != nil {
		if errors.Is(err, repo.ErrNodeNotFound) {
			return &v1.ListAcquiredRightsForProductResponse{}, nil
		}
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	// log.Println("UID of Product : ", ID)
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get Products -> "+err.Error())
	}
	// log.Printf("product info %+v", *res)
	numEquips := int32(0)
	if len(res.Products) != 0 {
		numEquips = res.Products[0].NumofEquipments
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")

	}
	prodAcqRights := make([]*v1.ProductAcquiredRights, len(prodRights))
	ind := 0
	input := make(map[string]interface{})
	input[ProdID] = ID
	for i, acqRight := range prodRights {
		var avgUnitPrice float64
		if acqRight.AcqLicenses != 0 {
			avgUnitPrice = acqRight.TotalPurchaseCost / float64(acqRight.AcqLicenses)
		} else {
			avgUnitPrice = acqRight.TotalPurchaseCost / float64(len(strings.Split(acqRight.SKU, ",")))
		}
		prodAcqRights[i] = &v1.ProductAcquiredRights{
			SKU:            acqRight.SKU,
			SwidTag:        req.SwidTag,
			Metric:         acqRight.Metric,
			NumAcqLicences: int32(acqRight.AcqLicenses),
			TotalCost:      acqRight.TotalCost,
			AvgUnitPrice:   avgUnitPrice,
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			prodAcqRights[i].DeltaNumber = int32(acqRight.AcqLicenses)
			prodAcqRights[i].DeltaCost = prodAcqRights[i].TotalCost
			continue
		}
		var maxComputed uint64
		var computedDetails string
		metricExists := false
		for _, met := range strings.Split(acqRight.Metric, ",") {
			if ind = metricNameExistsAll(metrics, met); ind == -1 {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric name doesnt exist - " + met)
				continue
			}

			input[MetricName] = metrics[ind].Name
			input[SCOPES] = []string{req.GetScope()}
			input[IsAgg] = false
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
			prodAcqRights[i].NumCptLicences = int32(maxComputed)
			prodAcqRights[i].DeltaNumber = int32(acqRight.AcqLicenses) - int32(maxComputed)
			prodAcqRights[i].DeltaCost = prodAcqRights[i].TotalCost - avgUnitPrice*float64(maxComputed)
			prodAcqRights[i].ComputedDetails = computedDetails
		} else {
			prodAcqRights[i].MetricNotDefined = true
		}
	}

	return &v1.ListAcquiredRightsForProductResponse{
		AcqRights: prodAcqRights,
	}, nil
}

// ListComputationDetails implements license service ListComputationDetails function
func (s *licenseServiceServer) ListComputationDetails(ctx context.Context, req *v1.ListComputationDetailsRequest) (*v1.ListComputationDetailsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("service/v1 - ListComputationDetails", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		logger.Log.Debug("service/v1 - ListAcqRightsForProduct - ListMetrices - unable to fetch metrics:%v", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	ID, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, metrics, req.Scope)
	if err != nil {
		if errors.Is(err, repo.ErrNodeNotFound) {
			logger.Log.Error("service/v1 - ListComputationDetails - repo/ProductAcquiredRights - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "product acqruired rights does not exist")
		}
		logger.Log.Error("service/v1 - ListComputationDetails - repo/ProductAcquiredRights - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	idx := acqrightSKUexists(prodRights, req.Sku)
	if idx == -1 {
		logger.Log.Error("service/v1 - ListComputationDetails - acqrightSKUexists", zap.String("reason: sku does not exist", req.Sku))
		return nil, status.Error(codes.InvalidArgument, "SKU rquested is not correct")
	}
	acqRight := prodRights[idx]
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListComputationDetails - repo/GetProductInformation - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "failed to get product information( num of equipments)")
	}
	numEquips := int32(0)
	if len(res.Products) != 0 {
		numEquips = res.Products[0].NumofEquipments
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	acqRightMetrics := strings.Split(prodRights[idx].Metric, ",")
	computedDetails := []*v1.ComputedDetails{}
	input := make(map[string]interface{})
	input[ProdID] = ID
	for _, met := range acqRightMetrics {
		var avgUnitPrice float64
		if acqRight.AcqLicenses != 0 {
			avgUnitPrice = prodRights[idx].TotalPurchaseCost / float64(acqRight.AcqLicenses)
		} else {
			avgUnitPrice = prodRights[idx].TotalPurchaseCost / float64(len(strings.Split(acqRight.SKU, ",")))
		}
		metricComputedDetails := &v1.ComputedDetails{
			MetricName:     met,
			NumAcqLicences: int32(acqRight.AcqLicenses),
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			metricComputedDetails.DeltaNumber = int32(acqRight.AcqLicenses)
			metricComputedDetails.DeltaCost = acqRight.TotalCost
			computedDetails = append(computedDetails, metricComputedDetails)
			continue
		}
		ind := metricNameExistsAll(metrics, met)
		if ind == -1 {
			logger.Log.Error("service/v1 - ListComputationDetails - metric name doesnt exist - " + met)
			continue
		}
		input[MetricName] = metrics[ind].Name
		input[SCOPES] = []string{req.GetScope()}
		input[IsAgg] = false
		if _, ok := MetricCalculation[metrics[ind].Type]; !ok {
			return nil, status.Error(codes.Internal, "this metricType is not supported")
		}
		resp, err := MetricCalculation[metrics[ind].Type](ctx, s, eqTypes, input)
		if err != nil {
			logger.Log.Error("service/v1 - ListComputationDetails - Failed ListAcqRightsForProduct  ", zap.String("metric name", metrics[ind].Name), zap.Any("metric type", metrics[ind].Type), zap.String("reason", err.Error()))
			continue
		}
		computedLicenses := resp[ComputedLicenses].(uint64)
		metricComputedDetails.NumCptLicences = int32(computedLicenses)
		metricComputedDetails.DeltaNumber = int32(acqRight.AcqLicenses) - int32(computedLicenses)
		metricComputedDetails.DeltaCost = acqRight.TotalCost - avgUnitPrice*float64(computedLicenses)
		if _, ok := resp[ComputedDetails]; ok {
			metricComputedDetails.ComputedDetails = resp[ComputedDetails].(string)
		}
		computedDetails = append(computedDetails, metricComputedDetails)
	}
	return &v1.ListComputationDetailsResponse{
		ComputedDetails: computedDetails,
	}, nil
}

// func productAcqRightFilter(notForMetric string) *repo.AggregateFilter {
// 	return &repo.AggregateFilter{
// 		Filters: []repo.Queryable{
// 			&repo.Filter{
// 				FilterMatchingType: repo.EqFilter,
// 				FilterKey:          repo.AcquiredRightsSearchKeyMetric.String(),
// 				FilterValue:        notForMetric,
// 			},
// 		},
// 	}
// }

// func stringToInterface(vals []string) []interface{} {
// 	interfaceSlice := make([]interface{}, len(vals))
// 	for i := range vals {
// 		interfaceSlice[i] = vals[i]
// 	}
// 	return interfaceSlice
// }

func repoProductToServProduct(repoProductData *repo.ProductData) *v1.Product {
	return &v1.Product{
		Name:              repoProductData.Name,
		Version:           repoProductData.Version,
		Category:          repoProductData.Category,
		Editor:            repoProductData.Editor,
		SwidTag:           repoProductData.Swidtag,
		NumofEquipments:   repoProductData.NumOfEquipments,
		NumOfApplications: repoProductData.NumOfApplications,
		TotalCost:         repoProductData.TotalCost,
	}
}

func acqrightSKUexists(prodacq []*repo.ProductAcquiredRight, sku string) int {
	for i, acq := range prodacq {
		if strcomp.CompareStrings(acq.SKU, sku) {
			return i
		}
	}
	return -1
}
