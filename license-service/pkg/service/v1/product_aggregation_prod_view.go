package v1

import (
	"context"
	"math"
	"strings"

	prodv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: funlen, gocyclo
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
	var rgtsWithRepart, rgtsWithoutRepart []*repo.ProductAcquiredRight
	for _, prodacq := range aggRights {
		if prodacq.Repartition {
			rgtsWithRepart = append(rgtsWithRepart, prodacq)
		} else {
			rgtsWithoutRepart = append(rgtsWithoutRepart, prodacq)
		}
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	var repartedResAggAcqRights []*v1.AggregationAcquiredRights
	if repoAgg.NumOfEquipments != 0 && len(rgtsWithRepart) > 1 {
		repartedResProdAcqRights, err := s.findRepartition(ctx, &repartition{
			AggName:     req.Name,
			Swidtag:     strings.Join(repoAgg.Swidtags, ","),
			ProductName: strings.Join(repoAgg.ProductNames, ","),
			Rights:      rgtsWithRepart,
		}, eqTypes, metrics, req.Scope)
		if err != nil {
			logger.Log.Info("service/v1 - ListAcqRightsForProductAggregation - findRepartition - error from repartition calculation", zap.String("agg name", req.Name), zap.String("error", err.Error()))
			return nil, status.Error(codes.Internal, "unable to calculate repartition")
		}
		for _, prodRights := range repartedResProdAcqRights {
			licenses, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: prodRights.SKU, Scope: req.GetScope()})
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - GetAvailableLicenses", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "serviceError")
			}
			repartedResAggAcqRights = append(repartedResAggAcqRights, &v1.AggregationAcquiredRights{
				SKU:               prodRights.SKU,
				AggregationName:   req.Name,
				SwidTags:          prodRights.SwidTag,
				Metric:            prodRights.Metric,
				NumCptLicences:    prodRights.NumCptLicences,
				NumAcqLicences:    prodRights.NumAcqLicences,
				TotalCost:         prodRights.TotalCost,
				DeltaNumber:       prodRights.DeltaNumber,
				DeltaCost:         prodRights.DeltaCost,
				AvgUnitPrice:      prodRights.AvgUnitPrice,
				ComputedDetails:   prodRights.ComputedDetails,
				MetricNotDefined:  prodRights.MetricNotDefined,
				NotDeployed:       prodRights.NotDeployed,
				ProductNames:      prodRights.ProductName,
				PurchaseCost:      prodRights.PurchaseCost,
				ComputedCost:      prodRights.ComputedCost,
				CostOptimization:  prodRights.CostOptimization,
				AvailableLicences: licenses.AvailableLicenses,
				SharedLicences:    licenses.TotalSharedLicenses,
				RecievedLicences:  licenses.TotalRecievedLicenses,
			})
		}
	} else if repoAgg.NumOfEquipments != 0 && len(rgtsWithRepart) == 1 {
		rgtsWithoutRepart = append(rgtsWithoutRepart, rgtsWithRepart...)
	} else if repoAgg.NumOfEquipments == 0 && len(rgtsWithRepart) != 0 {
		rgtsWithoutRepart = append(rgtsWithoutRepart, rgtsWithRepart...)
	}
	aggCompLicenses := make([]*v1.AggregationAcquiredRights, len(rgtsWithoutRepart))
	ind := 0
	input := make(map[string]interface{})
	input[ProdAggName] = repoAgg.Name
	input[SCOPES] = []string{req.Scope}
	input[IsAgg] = true
	input[IsSa] = false
	var maintenance *prodv1.GetMaintenanceBySwidtagResponse
	for i, aggRight := range rgtsWithoutRepart {
		metricType := ""
		availbleLicences := 0
		sku := strings.Split(aggRight.SKU, ",")

		for i := range sku {
			metricName, err := s.productClient.GetMetric(ctx, &prodv1.GetMetricRequest{
				Sku:   sku[i],
				Scope: req.Scope,
			})
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - GetMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "serviceError")
			}
			for _, v := range metrics {
				if metricName.Metric == v.Name {
					metricType = string(v.Type)
				}
			}
			resp, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: sku[i], Scope: req.Scope})
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - GetAvailableLicenses", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "serviceError")
			}
			if metricType == "oracle.nup.standard" && aggRight.TransformDetails != "" {
				availbleLicences += int(math.Floor(float64(resp.AvailableLicenses) / 50))
			} else {
				availbleLicences += int(resp.AvailableLicenses)
			}
			if metricType == "microsoft.sql.standard" || metricType == "microsoft.sql.enterprise" || metricType == "windows.server.standard" || metricType == "windows.server.datacenter" {
				maintenance, err = s.productClient.GetMaintenanceBySwidtag(ctx, &prodv1.GetMaintenanceBySwidtagRequest{
					Scope:  req.GetScope(),
					Acqsku: sku[i],
				})
				if err != nil {
					logger.Log.Sugar().Errorf("service/v1 - ListAcqRightsForProduct - GetMaintenanceBySwidtag-acqRight.SKU", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				if maintenance != nil {
					sa := maintenance.Success
					if sa {
						input[IsSa] = sa
					}
				}

			}
		}
		aggCompLicenses[i] = &v1.AggregationAcquiredRights{
			SKU:               aggRight.SKU,
			AggregationName:   req.Name,
			SwidTags:          strings.Join(repoAgg.Swidtags, ","),
			ProductNames:      strings.Join(repoAgg.ProductNames, ","),
			Metric:            aggRight.Metric,
			NumAcqLicences:    int32(aggRight.AcqLicenses),
			TotalCost:         aggRight.TotalCost,
			PurchaseCost:      aggRight.TotalPurchaseCost,
			AvgUnitPrice:      aggRight.AvgUnitPrice,
			AvailableLicences: int32(availbleLicences),
		}
		for _, v := range metrics {
			if aggRight.Metric == v.Name {
				metricType = string(v.Type)
			}
		}
		if repoAgg.NumOfEquipments == 0 && (metricType != "user.nominative.standard" && metricType != "user.concurrent.standard") {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			aggCompLicenses[i].DeltaNumber = int32(availbleLicences)
			aggCompLicenses[i].DeltaCost = aggCompLicenses[i].PurchaseCost
			aggCompLicenses[i].NotDeployed = true
			continue
		}
		rightsMetrics := strings.Split(aggRight.Metric, ",")
		// indvRights, err := s.licenseRepo.AggregationIndividualRights(ctx, repoAgg.ProductIDs, rightsMetrics, req.Scope)
		// if err != nil && err != repo.ErrNodeNotFound {
		// 	logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - repo/AggregationIndividualRights - failed to get aggregation individual details", zap.String("reason", err.Error()))
		// 	return nil, status.Error(codes.Internal, "failed to get aggregation individual rights")
		// }
		// for _, indacq := range indvRights {
		// 	aggCompLicenses[i].NumAcqLicences += indacq.Licenses
		// 	aggCompLicenses[i].TotalCost += indacq.TotalCost
		// }
		// fmt.Printf("acq[%d]: %v", i, aggCompLicenses[i])
		var maxComputed uint64
		var computedDetails string
		if aggRight.TransformDetails != "" {
			computedDetails = aggRight.TransformDetails
		}
		var nupLicence, opsLicence int32
		var nupUprice, opsUprice float64
		sku1 := strings.Split(aggRight.SKU, ",")
		var maintenance *prodv1.GetMaintenanceBySwidtagResponse
		if aggRight.TransformDetails != "" {
			for i := range sku1 {
				metricName, err := s.productClient.GetMetric(ctx, &prodv1.GetMetricRequest{
					Sku:   sku1[i],
					Scope: req.Scope,
				})
				for _, v := range metrics {
					if metricName.Metric == v.Name {
						metricType = string(v.Type)
					}
				}
				maintenance, err = s.productClient.GetMaintenanceBySwidtag(ctx, &prodv1.GetMaintenanceBySwidtagRequest{
					Scope:  req.GetScope(),
					Acqsku: sku1[i],
				})
				if err != nil {
					logger.Log.Sugar().Errorf("service/v1 -  ListAcqRightsForProductAggregation - GetMaintenanceBySwidtag-sku1[i]", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				if metricType == "oracle.nup.standard" && aggRight.TransformDetails != "" {
					nupLicence = maintenance.AcqLicenses
					nupUprice = maintenance.UnitPrice
				} else if metricType == "oracle.processor.standard" && aggRight.TransformDetails != "" {
					opsLicence = maintenance.AcqLicenses
					opsUprice = maintenance.UnitPrice
				}
			}
		}
		metricExists := false
		for _, met := range rightsMetrics {
			if ind = metricNameExistsAll(metrics, met); ind == -1 {
				logger.Log.Error("service/v1 -  ListAcqRightsForProductAggregation - metric name doesnt exist - " + met)
				continue
			}
			input[MetricName] = metrics[ind].Name
			input[SCOPES] = []string{req.GetScope()}
			if _, ok := MetricCalculation[metrics[ind].Type]; !ok {
				return nil, status.Error(codes.Internal, "this metricType is not supported")
			}
			resp, err := MetricCalculation[metrics[ind].Type](ctx, s, eqTypes, input)
			if err != nil {
				logger.Log.Error("service/v1 - Failed  ListAcqRightsForProductAggregation  ", zap.String("metric name", metrics[ind].Name), zap.Any("metric type", metrics[ind].Type), zap.String("reason", err.Error()))
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
			aggCompLicenses[i].DeltaNumber = int32(availbleLicences) - int32(maxComputed)
			if aggRight.TransformDetails != "" {
				aggCompLicenses[i].DeltaCost = float64(float64(availbleLicences)*float64(opsUprice) - float64(opsUprice)*float64(maxComputed))
				aggCompLicenses[i].TotalCost = float64(float64(opsLicence)*float64(opsUprice) + float64(nupLicence)*float64(nupUprice))
			} else {
				//aggCompLicenses[i].DeltaCost = aggCompLicenses[i].PurchaseCost - aggRight.AvgUnitPrice*float64(int32(maxComputed))
				aggCompLicenses[i].DeltaCost = (float64(availbleLicences) * aggRight.AvgUnitPrice) - aggRight.AvgUnitPrice*float64(int32(maxComputed))
			}
			aggCompLicenses[i].ComputedDetails = computedDetails
			aggCompLicenses[i].ComputedCost = aggRight.AvgUnitPrice * float64(int32(maxComputed))
		} else {
			aggCompLicenses[i].MetricNotDefined = true
		}
	}
	if repartedResAggAcqRights != nil {
		aggCompLicenses = append(aggCompLicenses, repartedResAggAcqRights...)
	}
	return &v1.ListAcqRightsForAggregationResponse{
		AcqRights: aggCompLicenses,
	}, nil
}
