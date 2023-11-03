package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListAcqRightsForApplicationsProduct(ctx context.Context, req *v1.ListAcqRightsForApplicationsProductRequest) (*v1.ListAcqRightsForApplicationsProductResponse, error) {
	// Extract Claims
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct", zap.String("reason", "ScopeError"))
		return &v1.ListAcqRightsForApplicationsProductResponse{}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	// Check if the product is linked with application
	isProductExist, err := s.licenseRepo.ProductExistsForApplication(ctx, req.ProdId, req.AppId, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - ProductExistsForApplication", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	if !isProductExist {
		return nil, status.Errorf(codes.FailedPrecondition, "Application %s does not uses product %s", req.AppId, req.ProdId)
	}
	res, err := s.ListAcqRightsForProduct(ctx, &v1.ListAcquiredRightsForProductRequest{
		SwidTag:    req.ProdId,
		Scope:      req.Scope,
		Simulation: false,
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - ProductAcquiredRights", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}
	/*
		// Fetch all metric types
		metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - ListMetrices", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "Internal Server Error")
		} else if err == repo.ErrNoData {
			return nil, status.Error(codes.FailedPrecondition, "No metric type exists in the system")
		}

		// Fetch Product AcquiredRights
		ID, _, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.ProdId, metrics, false, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - ProductAcquiredRights", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "Internal Server Error")
		}

		// Fetch all equipment types
		eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - EquipmentTypes", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "Internal Server Error")

		}

		// Fetch Common Equipments
		res, err := s.licenseRepo.ProductApplicationEquipments(ctx, req.ProdId, req.AppId, req.GetScope())
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - ProductApplicationEquipments", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "Internal Server Error")
		}
		numEquips := len(res)
		prodAcqRights := make([]*v1.ProductAcquiredRights, len(prodRights))
		var maintenance *prodv1.GetMaintenanceBySwidtagResponse
		input := make(map[string]interface{})
		input[ProdID] = ID
		input[SCOPES] = []string{req.GetScope()}
		input[IsAgg] = false
		ind := 0
		for i, acqRight := range prodRights {
			metricType := ""
			availbleLicences := 0
			sku := strings.Split(acqRight.SKU, ",")
			input[MetricName] = acqRight.Metric
			for i := range sku {
				metricName, err := s.productClient.GetMetric(ctx, &prodv1.GetMetricRequest{
					Sku:   sku[i],
					Scope: req.Scope,
				})
				if err != nil {
					logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - GetMetric", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				for _, v := range metrics {
					if metricName.Metric == v.Name {
						metricType = string(v.Type)
					}
				}
				resp, err := s.productClient.GetAvailableLicenses(ctx, &prodv1.GetAvailableLicensesRequest{Sku: sku[i], Scope: req.Scope})
				if err != nil {
					logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - GetAvailableLicenses", zap.String("reason", err.Error()))
					return nil, status.Error(codes.Internal, "serviceError")
				}
				if metricType == "oracle.nup.standard" && acqRight.TransformDetails != "" {
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
			logger.Log.Sugar().Infow("Entered")
			prodAcqRights[i] = &v1.ProductAcquiredRights{
				SKU:               acqRight.SKU,
				SwidTag:           req.ProdId,
				Metric:            acqRight.Metric,
				NumAcqLicences:    int32(acqRight.AcqLicenses),
				TotalCost:         acqRight.TotalCost,
				AvgUnitPrice:      acqRight.TotalCost / float64(acqRight.AcqLicenses),
				AvailableLicences: int32(availbleLicences),
			}
			//for _, met := range strings.Split(acqRight.Metric, ",") {
			if ind = metricNameExistsAll(metrics, acqRight.Metric); ind == -1 {
				logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - metric name does not exist", zap.String("metric name", acqRight.Metric))
				continue
			}
			if numEquips == 0 && (metricType != "user.nominative.standard" && metricType != "user.concurrent.standard") {
				logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - no equipments linked with product")
				continue
			}
			var nupLicence, opsLicence int32
			var nupUprice, opsUprice float64
			sku1 := strings.Split(acqRight.SKU, ",")
			var maintenance *prodv1.GetMaintenanceBySwidtagResponse
			if acqRight.TransformDetails != "" {
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
						logger.Log.Sugar().Errorf("service/v1 - ListAcqRightsForProduct - GetMaintenanceBySwidtag-sku[i]", zap.String("reason", err.Error()))
						return nil, status.Error(codes.Internal, "serviceError")
					}
					if metricType == "oracle.nup.standard" && acqRight.TransformDetails != "" {
						nupLicence = maintenance.AcqLicenses
						nupUprice = maintenance.UnitPrice
					} else if metricType == "oracle.processor.standard" && acqRight.TransformDetails != "" {
						opsLicence = maintenance.AcqLicenses
						opsUprice = maintenance.UnitPrice
					}
				}
			}

			if _, ok := MetricCalculation[metrics[ind].Type]; !ok {
				logger.Log.Error("service/v1 - Failed ListAcqRightsForApplicationsProduct for  - ", zap.String("metric :", acqRight.Metric), zap.Any("metricType", metrics[ind].Type))
				return nil, status.Error(codes.Internal, "this metricType is not supported")
			}
			resp, err := MetricCalculation[metrics[ind].Type](ctx, s, eqTypes, input)
			if err != nil {
				logger.Log.Error("service/v1 - Failed ProductLicensesForMetric for  - ", zap.String("metric :", acqRight.Metric), zap.Any("metricType", metrics[ind].Type), zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot compute licenses")
			}
			computedLicenses := resp[ComputedLicenses].(uint64)
			delta := int32(availbleLicences) - int32(computedLicenses)

			prodAcqRights[i].NumCptLicences = int32(computedLicenses)
			prodAcqRights[i].DeltaNumber = delta
			if acqRight.TransformDetails != "" {
				prodAcqRights[i].DeltaCost = float64(float64(availbleLicences)*float64(opsUprice) - float64(opsUprice)*float64(computedLicenses))
				prodAcqRights[i].TotalCost = float64(float64(opsLicence)*float64(opsUprice) + float64(nupLicence)*float64(nupUprice))
			} else {
				prodAcqRights[i].DeltaCost = prodAcqRights[i].PurchaseCost - prodAcqRights[i].AvgUnitPrice*float64(computedLicenses)
			}
			//}
		}
	*/
	return &v1.ListAcqRightsForApplicationsProductResponse{
		AcqRights: res.AcqRights,
	}, nil

}
