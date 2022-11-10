package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/helper"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strconv"
	"strings"

	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: funlen
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
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		logger.Log.Debug("service/v1 - ListAcqRightsForProduct - ListMetrices - unable to fetch metrics:%v", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	if !req.Simulation {
		aggregationName, err := s.licenseRepo.IsProductPurchasedInAggregation(ctx, req.SwidTag, req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - failed to check is swidtag is part of aggregation", zap.Error(err))
			return nil, status.Error(codes.Internal, "couldn't check is swidtag part of aggregation")
		} else if aggregationName != "" {
			_, aggRights, err := s.licenseRepo.AggregationDetails(ctx, aggregationName, metrics, req.Simulation, req.GetScope())
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - repo/AggregationDetails - failed to get aggregation details", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "failed to get aggregation details")
			}
			if aggRights != nil {
				logger.Log.Info("service/v1 - ListAcqRightsForProduct - aggregation found", zap.String("swidtag", req.SwidTag), zap.String("aggName", aggregationName))
				return &v1.ListAcquiredRightsForProductResponse{AggregationName: aggregationName}, nil
			}
			logger.Log.Info("service/v1 - ListAcqRightsForProduct - aggregation found but no licenses bought for the aggregation", zap.String("swidtag", req.SwidTag), zap.String("aggName", aggregationName))
			// return &v1.ListAcquiredRightsForProductResponse{}, nil
		}
	}
	ID, prodname, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, metrics, req.Simulation, req.GetScope())
	if err != nil {
		if errors.Is(err, repo.ErrNodeNotFound) {
			return &v1.ListAcquiredRightsForProductResponse{}, nil
		}
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	var rgtsWithRepart, rgtsWithoutRepart []*repo.ProductAcquiredRight
	for _, prodacq := range prodRights {
		if prodacq.Repartition {
			rgtsWithRepart = append(rgtsWithRepart, prodacq)
		} else {
			rgtsWithoutRepart = append(rgtsWithoutRepart, prodacq)
		}
	}
	// log.Println("UID of Product : ", ID)
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get Products -> "+err.Error())
	}
	// if len(res.Products[0].ProdMetricAllocated) != 0 {
	// 	if res.Products[0].ProdMetricAllocated != prodRights[0].Metric {
	// 		return nil, nil
	// 	}
	// }
	// log.Printf("product info %+v", *res)
	numEquips := int32(0)
	if len(res.Products) != 0 {
		numEquips = res.Products[0].NumofEquipments
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	var repartedResProdAcqRights []*v1.ProductAcquiredRights
	if numEquips != 0 && len(rgtsWithRepart) > 1 {
		repartedResProdAcqRights, err = s.findRepartition(ctx, &repartition{
			ProductID:   ID,
			Swidtag:     req.SwidTag,
			ProductName: prodname,
			Rights:      rgtsWithRepart,
		}, eqTypes, metrics, req.Scope)
		if err != nil {
			logger.Log.Info("service/v1 - ListAcqRightsForProduct - findRepartition - error from repartition calculation", zap.String("swidtag", req.SwidTag), zap.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "unable to calculate repartition")
		}
	} else if numEquips != 0 && len(rgtsWithRepart) == 1 {
		rgtsWithoutRepart = append(rgtsWithoutRepart, rgtsWithRepart...)
	} else if numEquips == 0 && len(rgtsWithRepart) != 0 {
		rgtsWithoutRepart = append(rgtsWithoutRepart, rgtsWithRepart...)
	}
	prodAcqRights := make([]*v1.ProductAcquiredRights, len(rgtsWithoutRepart))
	ind := 0
	input := make(map[string]interface{})
	input[ProdID] = ID
	input[IsAgg] = false
	input[SWIDTAG] = req.SwidTag
	for i, acqRight := range rgtsWithoutRepart {
		// var avgUnitPrice float64
		// if acqRight.AcqLicenses != 0 {
		// 	avgUnitPrice = acqRight.TotalPurchaseCost / float64(acqRight.AcqLicenses)
		// } else {
		// 	avgUnitPrice = acqRight.TotalPurchaseCost / float64(len(strings.Split(acqRight.SKU, ",")))
		// }
		prodAcqRights[i] = &v1.ProductAcquiredRights{
			SKU:            acqRight.SKU,
			SwidTag:        req.SwidTag,
			ProductName:    prodname,
			Metric:         acqRight.Metric,
			NumAcqLicences: int32(acqRight.AcqLicenses),
			TotalCost:      acqRight.TotalCost,
			PurchaseCost:   acqRight.TotalPurchaseCost,
			AvgUnitPrice:   acqRight.AvgUnitPrice,
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			prodAcqRights[i].DeltaNumber = int32(acqRight.AcqLicenses)
			prodAcqRights[i].DeltaCost = prodAcqRights[i].PurchaseCost
			prodAcqRights[i].NotDeployed = true
			continue
		}
		var maxComputed uint64
		var computedDetails string
		if acqRight.TransformDetails != "" {
			computedDetails = acqRight.TransformDetails
		}
		metricExists := false
		for _, met := range strings.Split(acqRight.Metric, ",") {
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
				logger.Log.Sugar().Infow("Compalince for metric", "input", input, "compliance", err.Error())
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
			prodAcqRights[i].DeltaCost = prodAcqRights[i].PurchaseCost - acqRight.AvgUnitPrice*float64(int32(maxComputed))
			prodAcqRights[i].ComputedCost = acqRight.AvgUnitPrice * float64(int32(maxComputed))
			prodAcqRights[i].ComputedDetails = computedDetails
		} else {
			prodAcqRights[i].MetricNotDefined = true
		}
	}
	if repartedResProdAcqRights != nil {
		prodAcqRights = append(prodAcqRights, repartedResProdAcqRights...)
	}
	return &v1.ListAcquiredRightsForProductResponse{
		AcqRights: prodAcqRights,
	}, nil
}

// ListComputationDetails implements license service ListComputationDetails function
func (s *licenseServiceServer) ListComputationDetails(ctx context.Context, req *v1.ListComputationDetailsRequest) (*v1.ListComputationDetailsResponse, error) { // nolint
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
	var rights []*repo.ProductAcquiredRight
	var numEquips int32
	input := make(map[string]interface{})
	var repoAgg *repo.AggregationInfo
	if req.AggName != "" {
		repoAgg, rights, err = s.licenseRepo.AggregationDetails(ctx, req.AggName, metrics, false, req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - ListComputationDetails - repo/AggregationDetails - failed to get aggregation details", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "failed to get aggregation details")
		}
		numEquips = repoAgg.NumOfEquipments
		input[ProdAggName] = repoAgg.Name
		input[IsAgg] = true
	} else {
		ID, _, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, metrics, false, req.Scope)
		if err != nil {
			if errors.Is(err, repo.ErrNodeNotFound) {
				logger.Log.Error("service/v1 - ListComputationDetails - repo/ProductAcquiredRights - ", zap.String("reason", err.Error()))
				return nil, status.Error(codes.NotFound, "product acqruired rights does not exist")
			}
			logger.Log.Error("service/v1 - ListComputationDetails - repo/ProductAcquiredRights - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
		}
		res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ListComputationDetails - repo/GetProductInformation - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "failed to get product information( num of equipments)")
		}
		if len(res.Products) != 0 {
			numEquips = res.Products[0].NumofEquipments
		}
		rights = prodRights
		input[ProdID] = ID
		input[IsAgg] = false
	}
	idx := acqrightSKUexists(rights, req.Sku)
	if idx == -1 {
		logger.Log.Error("service/v1 - ListComputationDetails - acqrightSKUexists", zap.String("reason: sku does not exist", req.Sku))
		return nil, status.Error(codes.InvalidArgument, "SKU rquested is not correct")
	}
	acqRight := rights[idx]
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	acqRightMetrics := strings.Split(acqRight.Metric, ",")
	computedDetails := []*v1.ComputedDetails{}
	for _, met := range acqRightMetrics {
		totalcost := acqRight.TotalCost
		metricComputedDetails := &v1.ComputedDetails{
			MetricName:     met,
			NumAcqLicences: int32(acqRight.AcqLicenses),
		}
		if req.AggName != "" {
			indvRights, err := s.licenseRepo.AggregationIndividualRights(ctx, repoAgg.ProductIDs, []string{met}, req.Scope)
			if err != nil && err != repo.ErrNodeNotFound {
				logger.Log.Error("service/v1 - ListComputationDetails - repo/AggregationIndividualRights - failed to get aggregation individual details", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "failed to get aggregation individual rights")
			}
			for _, indacq := range indvRights {
				metricComputedDetails.NumAcqLicences += indacq.Licenses
				totalcost += indacq.TotalCost
			}
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			metricComputedDetails.DeltaNumber = metricComputedDetails.NumAcqLicences
			metricComputedDetails.DeltaCost = acqRight.TotalPurchaseCost
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
		metricComputedDetails.DeltaNumber = metricComputedDetails.NumAcqLicences - int32(computedLicenses)
		metricComputedDetails.DeltaCost = acqRight.TotalPurchaseCost - acqRight.AvgUnitPrice*float64(int32(computedLicenses))
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

// func repoProductToServProduct(repoProductData *repo.ProductData) *v1.Product {
// 	return &v1.Product{
// 		Name:              repoProductData.Name,
// 		Version:           repoProductData.Version,
// 		Category:          repoProductData.Category,
// 		Editor:            repoProductData.Editor,
// 		SwidTag:           repoProductData.Swidtag,
// 		NumofEquipments:   repoProductData.NumOfEquipments,
// 		NumOfApplications: repoProductData.NumOfApplications,
// 		TotalCost:         repoProductData.TotalCost,
// 	}
// }

func acqrightSKUexists(prodacq []*repo.ProductAcquiredRight, sku string) int {
	for i, acq := range prodacq {
		if strcomp.CompareStrings(acq.SKU, sku) {
			return i
		}
	}
	return -1
}

type repartition struct {
	ProductID   string
	Swidtag     string
	ProductName string
	AggName     string
	Rights      []*repo.ProductAcquiredRight
}

// nolint: funlen, gocyclo
func (s *licenseServiceServer) findRepartition(ctx context.Context, repart *repartition, eqTypes []*repo.EquipmentType, metrics []*repo.Metric, scope string) ([]*v1.ProductAcquiredRights, error) {
	attrSumMetExists := false
	insMetExists := false
	m := make(map[int]Contract, 2)
	var attrSumCompliance, insCompliance uint64
	sumOfValues := 0.0
	var numInstances int32
	for i, right := range repart.Rights {
		var memory float64
		var node float64
		mets := strings.Split(right.Metric, ",")
		for _, met := range mets {
			ind := metricNameExistsAll(metrics, met)
			if ind == -1 {
				logger.Log.Error("service/v1 - findRepartition - metric name doesnt exist - " + met)
				continue
			}
			switch metrics[ind].Type {
			case repo.MetricAttrSumStandard:
				sumMetrics, err := s.licenseRepo.ListMetricAttrSum(ctx, scope)
				if err != nil && err != repo.ErrNoData {
					logger.Log.Error("service/v1 - findRepartition -  computedLicensesAttrSum", zap.Error(err))
					return nil, status.Error(codes.Internal, "cannot fetch metric Attr sum")
				}
				idx := metricNameExistsAttrSum(sumMetrics, metrics[ind].Name)
				if ind == -1 {
					return nil, status.Error(codes.NotFound, "cannot find metric name")
				}
				memory = sumMetrics[idx].ReferenceValue
				sumOfValues += sumMetrics[idx].ReferenceValue * float64(right.AcqLicenses)
				if !attrSumMetExists {
					mat, err := computedMetricAttrSum(sumMetrics[idx], eqTypes)
					if err != nil {
						logger.Log.Error("service/v1 - findRepartition - computedMetricACS - ", zap.Error(err))
						return nil, err
					}
					if repart.AggName != "" {
						_, attrSumCompliance, err = s.licenseRepo.MetricAttrSumComputedLicensesAgg(ctx, repart.AggName, metrics[ind].Name, mat, scope)
					} else {
						_, attrSumCompliance, err = s.licenseRepo.MetricAttrSumComputedLicenses(ctx, repart.ProductID, mat, scope)
					}
					if err != nil {
						logger.Log.Error("service/v1 - findRepartition - ", zap.String("reason", err.Error()))
						return nil, status.Error(codes.Internal, "cannot compute licenses for metric attribute sum standard")

					}
					attrSumMetExists = true
				}
			case repo.MetricInstanceNumberStandard:
				insmetrics, err := s.licenseRepo.ListMetricINM(ctx, scope)
				if err != nil && err != repo.ErrNoData {
					logger.Log.Error("service/v1 - findRepartition - repo/ListMetricINM", zap.Error(err))
					return nil, status.Error(codes.Internal, "cannot fetch metric INM")
				}
				idx := metricNameExistsINM(insmetrics, metrics[ind].Name)
				if ind == -1 {
					return nil, status.Error(codes.NotFound, "cannot find metric name")
				}

				node = float64(insmetrics[idx].Coefficient)
				numInstances += insmetrics[idx].Coefficient * int32(right.AcqLicenses)
				if !insMetExists {
					mat := computedMetricINM(insmetrics[idx])
					if repart.AggName != "" {
						_, insCompliance, err = s.licenseRepo.MetricINMComputedLicensesAgg(ctx, repart.AggName, metrics[ind].Name, mat, scope)
					} else {
						_, insCompliance, err = s.licenseRepo.MetricINMComputedLicenses(ctx, repart.ProductID, mat, scope)
					}
					if err != nil {
						logger.Log.Error("service/v1 - findRepartition - computedLicensesINM - ", zap.String("reason", err.Error()))
						return nil, status.Error(codes.Internal, "cannot compute licenses for metric INM")
					}
					insMetExists = true
				}
			default:
				logger.Log.Sugar().Errorf("service/v1 - findRepartition - repartition not available for metric: %v of metric type: %v", metrics[ind].Name, metrics[ind].Type)
				return nil, status.Error(codes.InvalidArgument, "repartition not available for metric")
			}
		}
		m[i] = Contract{
			Memory: memory,
			Node:   node,
			Amount: right.AvgUnitPrice,
		}
	}
	computedDetails := "Sum of values: " + strconv.Itoa(int(attrSumCompliance)) + ",Total instances: " + strconv.Itoa(int(insCompliance))
	delta := [1][2]float64{}
	delta[0][0] = float64(attrSumCompliance)
	delta[0][1] = float64(insCompliance)
	setsForCostEffective := GetNumberOfAcquiredRightByOptimizingCost(m, delta)
	if len(setsForCostEffective) == 0 {
		logger.Log.Sugar().Error("service/v1 - findRepartition - GetNumberOfAcquiredRightByOptimizingCost - error in finding cost effective sets")
		return nil, status.Error(codes.Internal, "error in finding repartition sets")
	}
	repartRespAll := make([]*v1.ProductAcquiredRights, len(repart.Rights))
	for i, right := range repart.Rights {
		computedLicenses := setsForCostEffective[0].Value.Set[i]
		deltaNumber := int32(right.AcqLicenses) - int32(computedLicenses)
		repartRespAll[i] = &v1.ProductAcquiredRights{
			SKU:              right.SKU,
			SwidTag:          repart.Swidtag,
			Metric:           right.Metric,
			NumCptLicences:   int32(computedLicenses),
			NumAcqLicences:   int32(right.AcqLicenses),
			TotalCost:        right.TotalCost,
			DeltaNumber:      deltaNumber,
			DeltaCost:        float64(deltaNumber) * right.AvgUnitPrice,
			AvgUnitPrice:     right.AvgUnitPrice,
			ComputedDetails:  computedDetails,
			MetricNotDefined: false,
			NotDeployed:      false,
			ProductName:      repart.ProductName,
			PurchaseCost:     right.TotalPurchaseCost,
			ComputedCost:     float64(computedLicenses) * right.AvgUnitPrice,
			CostOptimization: true,
		}
	}
	return repartRespAll, nil
}
