package v1

import (
	"context"
	"errors"
	"math"
	"optisam-backend/common/optisam/helper"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	prodv1 "optisam-backend/product-service/pkg/api/v1"
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
			return &v1.ListAcquiredRightsForProductResponse{AggregationName: aggregationName}, nil
		}
	}

	var products, productsForUid []*repo.ProductDetail
	//return nil, status.Error(codes.Internal, "failed to get product details")
	ID, prodname, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, metrics, req.Simulation, req.GetScope())
	if err != nil {
		if errors.Is(err, repo.ErrNodeNotFound) {
			return &v1.ListAcquiredRightsForProductResponse{}, nil
		}
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, req.GetScope())
	if err != nil {
		logger.Log.Sugar().Errorw("unable to get product info from product",
			"err", err.Error(),
			"swidtag", req.SwidTag,
			"scope", req.Scope,
		)
		return nil, status.Error(codes.Internal, "failed to get Products -> "+err.Error())
	}
	var tmpProductDetail repo.ProductDetail
	if len(prodRights) > 0 {

		tmpProductDetail.ID = ID
		tmpProductDetail.SwidTag = req.SwidTag
		tmpProductDetail.Name = prodname
		tmpProductDetail.AcquiredRights = prodRights

	}

	if len(res.Products[0].Name) == 0 {
		res, err = s.licenseRepo.GetProductInformationFromAcqRight(ctx, req.SwidTag, req.GetScope())
		if err != nil {
			logger.Log.Sugar().Errorw("unable to get product info from acqRight",
				"err", err.Error(),
				"swidtag", req.SwidTag,
				"scope", req.Scope,
			)
			return nil, status.Error(codes.Internal, "failed to get Products -> "+err.Error())
		}
		tmpProductDetail.Name = res.Products[0].Name
		tmpProductDetail.Version = res.Products[0].Version
		products = append(products, &tmpProductDetail)
	} else {
		tmpProductDetail.NumOfEquipments = res.Products[0].NumofEquipments
		tmpProductDetail.Version = res.Products[0].Version
	}
	var isWithoutVersionProductExists bool

	if res.Products[0].Version == "" || strings.ToLower(res.Products[0].Version) == "all" {
		isWithoutVersionProductExists = true
	}
	var swittagWithoutVersion string
	// Without Verion Product not exists
	if res.Products[0].Name != "" {
		swittagWithoutVersion = strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{res.Products[0].Name, res.Products[0].Editor}, "_"), " ", "_"), "-", "_")
		if swittagWithoutVersion != req.SwidTag {
			// Get AcqRight without version
			ID, _, prodRights, _ := s.licenseRepo.ProductAcquiredRights(ctx, swittagWithoutVersion, metrics, req.Simulation, req.GetScope())
			if len(prodRights) > 0 {
				isWithoutVersionProductExists = true
				var tmpProduct repo.ProductDetail
				tmpProduct.ID = ID
				tmpProduct.SwidTag = swittagWithoutVersion
				tmpProduct.Name = res.Products[0].Name
				tmpProduct.Version = ""
				tmpProduct.AcquiredRights = prodRights
				products = append(products, &tmpProduct)

			} else {
				// when product without version acq have all value in version
				swittagWithoutVersion = strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{res.Products[0].Name, res.Products[0].Editor, "all"}, "_"), " ", "_"), "-", "_")
				if swittagWithoutVersion != req.SwidTag {
					ID, _, prodRights, _ := s.licenseRepo.ProductAcquiredRights(ctx, swittagWithoutVersion, metrics, req.Simulation, req.GetScope())
					if len(prodRights) > 0 {
						isWithoutVersionProductExists = true
						var tmpProduct repo.ProductDetail
						tmpProduct.ID = ID
						tmpProduct.SwidTag = swittagWithoutVersion
						tmpProduct.Name = res.Products[0].Name
						tmpProduct.Version = ""
						tmpProduct.AcquiredRights = prodRights
						products = append(products, &tmpProduct)

					}
				}
			}
		}
		productsForUid, err = s.licenseRepo.GetProductsByEditorProductName(ctx, metrics, req.GetScope(), res.Products[0].Editor, res.Products[0].Name)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get Products -> "+err.Error())
		}

	}

	// Get all acq rights & product for UID for without version product calculation
	productUIDs := []string{}
	for _, product := range productsForUid {
		aggregationName, _ := s.licenseRepo.IsProductPurchasedInAggregation(ctx, product.SwidTag, req.Scope)
		if aggregationName == "" {
			if len(product.AcquiredRights) == 0 || (product.Version == "" || strings.ToLower(product.Version) == "all") {
				productUIDs = append(productUIDs, product.ID)
			}

			if len(product.AcquiredRights) > 0 && product.SwidTag != swittagWithoutVersion && product.SwidTag != req.SwidTag {
				products = append(products, product)
			}
			if product.SwidTag == req.SwidTag {
				products = append(products, &tmpProductDetail)
			}
		}
	}
	var acqMetrics = []string{}
	var ProdAcqRights []*v1.ProductAcquiredRights
	for _, product := range products {
		prodRights := product.AcquiredRights
		ID := product.ID
		prodname := product.Name
		swidTag := product.SwidTag

		if len(prodRights) == 0 {
			continue
		}

		var rgtsWithRepart, rgtsWithoutRepart []*repo.ProductAcquiredRight
		for _, prodacq := range prodRights {
			if prodacq.Repartition {
				rgtsWithRepart = append(rgtsWithRepart, prodacq)
			} else {
				rgtsWithoutRepart = append(rgtsWithoutRepart, prodacq)
			}
		}
		ind := 0
		input := make(map[string]interface{})
		if isWithoutVersionProductExists && (product.Version == "" || strings.ToLower(product.Version) == "all") {
			input[ProdID] = productUIDs
		} else {
			input[ProdID] = []string{ID}
		}
		input[IsAgg] = false
		input[SWIDTAG] = swidTag
		numEquips := int32(0)
		numEquips = product.NumOfEquipments
		eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
		if err != nil {
			return nil, status.Error(codes.Internal, "cannot fetch equipment types")
		}
		var repartedResProdAcqRights []*v1.ProductAcquiredRights
		if numEquips != 0 && len(rgtsWithRepart) > 1 {
			repartedResProdAcqRights, err = s.findRepartition(ctx, &repartition{
				ProductID:   input[ProdID].([]string),
				Swidtag:     swidTag,
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
		//prodAcqRights := make([]*v1.ProductAcquiredRights, len(rgtsWithoutRepart))
		for _, acqRight := range rgtsWithoutRepart {
			// var avgUnitPrice float64
			// if acqRight.AcqLicenses != 0 {
			// 	avgUnitPrice = acqRight.TotalPurchaseCost / float64(acqRight.AcqLicenses)
			// } else {
			// 	avgUnitPrice = acqRight.TotalPurchaseCost / float64(len(strings.Split(acqRight.SKU, ",")))
			// }
			metricType := ""
			availbleLicences := 0
			sku := strings.Split(acqRight.SKU, ",")
			for i := range sku {
				metricName, err := s.productClient.GetMetric(ctx, &prodv1.GetMetricRequest{
					Sku:   sku[i],
					Scope: req.Scope,
				})
				if err != nil {
					logger.Log.Error("service/v1 - ListAcqRightsForProduct - GetMetric", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				for _, v := range metrics {
					if metricName.Metric == v.Name {
						metricType = string(v.Type)
					}
				}
				resp, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: sku[i], Scope: req.Scope})
				if err != nil {
					logger.Log.Error("service/v1 - ListAcqRightsForProduct - GetAvailableLicenses", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				if metricType == "oracle.nup.standard" && acqRight.TransformDetails != "" {
					availbleLicences += int(math.Ceil(float64(resp.AvailableLicenses) / 50))
				} else {
					availbleLicences += int(resp.AvailableLicenses)
				}
			}
			prodAcqRights := &v1.ProductAcquiredRights{
				SKU:               acqRight.SKU,
				SwidTag:           swidTag,
				ProductName:       prodname,
				Metric:            acqRight.Metric,
				NumAcqLicences:    int32(acqRight.AcqLicenses),
				TotalCost:         acqRight.TotalCost,
				PurchaseCost:      acqRight.TotalPurchaseCost,
				AvgUnitPrice:      acqRight.AvgUnitPrice,
				AvailableLicences: int32(availbleLicences),
			}
			if isWithoutVersionProductExists && (product.Version == "" || strings.ToLower(product.Version) == "all") {
				prodAcqRights.WithoutVerionAcq = true
			}
			for _, v := range metrics {
				if acqRight.Metric == v.Name {
					metricType = string(v.Type)
				}
			}
			if (!isWithoutVersionProductExists && (product.Version != "" || strings.ToLower(product.Version) != "all")) && numEquips == 0 && (metricType != "user.nominative.standard" && metricType != "user.concurrent.standard") {
				logger.Log.Sugar().Errorw("service/v1 - ListAcqRightsForProduct - no equipments linked with product",
					"product", product,
					"numEquips", numEquips,
					"metricType", metricType,
					"isWithoutVersionProductExists", isWithoutVersionProductExists,
				)
				prodAcqRights.DeltaNumber = int32(availbleLicences)
				prodAcqRights.DeltaCost = prodAcqRights.PurchaseCost
				prodAcqRights.NotDeployed = true
				ProdAcqRights = append(ProdAcqRights, prodAcqRights)
				continue
			}

			var maxComputed uint64
			var computedDetails string
			if acqRight.TransformDetails != "" {
				computedDetails = acqRight.TransformDetails
			}
			metricExists := false
			acqMetrics = append(acqMetrics, acqRight.Metric)

			for _, met := range strings.Split(acqRight.Metric, ",") {
				if ind = metricNameExistsAll(metrics, met); ind == -1 {
					logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric name doesnt exist - " + met)
					continue
				}
				input[MetricName] = metrics[ind].Name
				input[SCOPES] = []string{req.GetScope()}
				if len(input[ProdID].([]string)) > 0 {
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
				} else {
					metricExists = true
				}

			}
			if metricExists {
				prodAcqRights.NumCptLicences = int32(maxComputed)
				prodAcqRights.DeltaNumber = int32(availbleLicences) - int32(maxComputed)
				prodAcqRights.DeltaCost = prodAcqRights.PurchaseCost - acqRight.AvgUnitPrice*float64(int32(maxComputed))
				prodAcqRights.ComputedCost = acqRight.AvgUnitPrice * float64(int32(maxComputed))
				prodAcqRights.ComputedDetails = computedDetails
			} else {
				prodAcqRights.MetricNotDefined = true
			}

			ProdAcqRights = append(ProdAcqRights, prodAcqRights)
		}

		if repartedResProdAcqRights != nil {
			ProdAcqRights = append(ProdAcqRights, repartedResProdAcqRights...)
		}
	}

	// Change delta records if without version acqright exist for product
	if isWithoutVersionProductExists {

		for _, metricName := range acqMetrics {
			for _, met := range strings.Split(metricName, ",") {
				var remainDeltaWithoutVerAcq int32
				var withVerionAcqIndex int
				for index, acqRight := range ProdAcqRights {
					if checkProductExists(strings.Split(acqRight.Metric, ","), met) {
						if acqRight.DeltaNumber > 0 && acqRight.WithoutVerionAcq {
							remainDeltaWithoutVerAcq = acqRight.DeltaNumber
							ProdAcqRights[index].OldDeltaNumber = acqRight.DeltaNumber
							withVerionAcqIndex = index
							continue
						}

						if remainDeltaWithoutVerAcq > 0 && !acqRight.WithoutVerionAcq && acqRight.DeltaNumber < 0 {

							if remainDeltaWithoutVerAcq >= int32(math.Abs(float64(acqRight.DeltaNumber))) {
								remainDeltaWithoutVerAcq = remainDeltaWithoutVerAcq - int32(math.Abs(float64(acqRight.DeltaNumber)))
								ProdAcqRights[index].OldDeltaNumber = acqRight.DeltaNumber
								ProdAcqRights[index].DeltaNumber = 0
								ProdAcqRights[index].DeltaCost = 0

							} else if remainDeltaWithoutVerAcq < int32(math.Abs(float64(acqRight.DeltaNumber))) {
								ProdAcqRights[index].OldDeltaNumber = acqRight.DeltaNumber
								ProdAcqRights[index].DeltaNumber = remainDeltaWithoutVerAcq - int32(math.Abs(float64(acqRight.DeltaNumber)))
								ProdAcqRights[index].DeltaCost = acqRight.AvgUnitPrice * float64(ProdAcqRights[index].DeltaNumber)
								remainDeltaWithoutVerAcq = 0
							}
						}

						if remainDeltaWithoutVerAcq == 0 && index > 0 && !acqRight.WithoutVerionAcq {
							ProdAcqRights[withVerionAcqIndex].DeltaNumber = remainDeltaWithoutVerAcq
							ProdAcqRights[withVerionAcqIndex].DeltaCost = 0
							break
						}
					}
				}

				if remainDeltaWithoutVerAcq > 0 {
					ProdAcqRights[withVerionAcqIndex].DeltaNumber = remainDeltaWithoutVerAcq
					ProdAcqRights[withVerionAcqIndex].DeltaCost = (ProdAcqRights[withVerionAcqIndex].AvgUnitPrice * float64(remainDeltaWithoutVerAcq))
				}
			}
		}
	}
	return &v1.ListAcquiredRightsForProductResponse{
		AcqRights: ProdAcqRights,
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
		availLicences, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: acqRight.SKU, Scope: req.Scope})
		if err != nil {
			logger.Log.Error("service/v1 - ListComputationDetails - GetAvailableLicenses", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "serviceError")
		}
		metricComputedDetails := &v1.ComputedDetails{
			MetricName:       met,
			NumAcqLicences:   int32(acqRight.AcqLicenses),
			NumAvailLicences: int32(availLicences.AvailableLicenses),
		}
		if req.AggName != "" {
			indvRights, err := s.licenseRepo.AggregationIndividualRights(ctx, repoAgg.ProductIDs, []string{met}, req.Scope)
			if err != nil && err != repo.ErrNodeNotFound {
				logger.Log.Error("service/v1 - ListComputationDetails - repo/AggregationIndividualRights - failed to get aggregation individual details", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "failed to get aggregation individual rights")
			}
			for _, indacq := range indvRights {
				availLic, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: indacq.SKU, Scope: req.Scope})
				if err != nil {
					logger.Log.Error("service/v1 - ListComputationDetails - GetAvailableLicenses", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				// metricComputedDetails.NumAcqLicences += indacq.Licenses
				metricComputedDetails.NumAvailLicences += availLic.AvailableLicenses
				totalcost += indacq.TotalCost
			}
		}
		metricType := ""
		for _, v := range metrics {
			if met == v.Name {
				metricType = string(v.Type)
			}
		}
		if numEquips == 0 && (metricType != "user.nominative.standard" && metricType != "user.concurrent.standard") {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			metricComputedDetails.DeltaNumber = metricComputedDetails.NumAvailLicences
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
		metricComputedDetails.DeltaNumber = metricComputedDetails.NumAvailLicences - int32(computedLicenses)
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

func checkProductExists(pIDS []string, pid string) bool {
	for _, id := range pIDS {
		if strcomp.CompareStrings(id, pid) {
			return true
		}
	}
	return false
}

type repartition struct {
	ProductID   []string
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
				// acquiredLicenses has to be replaced with availableLicenses for metric calculation if needed.
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
				// acquiredLicenses has to be replaced with availableLicenses for metric calculation if needed.
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
		resp, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: right.SKU, Scope: scope})
		if err != nil {
			logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAvailableLicenses", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "serviceError")
		}
		computedLicenses := setsForCostEffective[0].Value.Set[i]
		deltaNumber := resp.AvailableLicenses - int32(computedLicenses)
		repartRespAll[i] = &v1.ProductAcquiredRights{
			SKU:               right.SKU,
			SwidTag:           repart.Swidtag,
			Metric:            right.Metric,
			NumCptLicences:    int32(computedLicenses),
			NumAcqLicences:    int32(right.AcqLicenses),
			TotalCost:         right.TotalCost,
			DeltaNumber:       deltaNumber,
			DeltaCost:         float64(deltaNumber) * right.AvgUnitPrice,
			AvgUnitPrice:      right.AvgUnitPrice,
			ComputedDetails:   computedDetails,
			MetricNotDefined:  false,
			NotDeployed:       false,
			ProductName:       repart.ProductName,
			PurchaseCost:      right.TotalPurchaseCost,
			ComputedCost:      float64(computedLicenses) * right.AvgUnitPrice,
			CostOptimization:  true,
			AvailableLicences: resp.AvailableLicenses,
			SharedLicences:    resp.TotalSharedLicenses,
			RecievedLicences:  resp.TotalRecievedLicenses,
		}
	}
	return repartRespAll, nil
}
