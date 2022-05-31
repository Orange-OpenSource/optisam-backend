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
	ind := 0
	for i, acqRight := range prodRights {
		prodAcqRights[i] = &v1.ProductAcquiredRights{
			SKU:            acqRight.SKU,
			SwidTag:        req.ProdId,
			Metric:         acqRight.Metric,
			NumAcqLicences: int32(acqRight.AcqLicenses),
			TotalCost:      acqRight.TotalCost,
			AvgUnitPrice:   acqRight.TotalCost / float64(acqRight.AcqLicenses),
		}
		if ind = metricNameExistsAll(metrics, acqRight.Metric); ind == -1 {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - metric name does not exist", zap.String("metric name", acqRight.Metric))
			continue
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct - no equipments linked with product")
			continue
		}
		input := make(map[string]interface{})
		input[ProdID] = ID
		input[MetricName] = acqRight.Metric
		input[SCOPES] = []string{req.GetScope()}
		input[IsAgg] = false
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
		delta := int32(acqRight.AcqLicenses) - int32(computedLicenses)

		prodAcqRights[i].NumCptLicences = int32(computedLicenses)
		prodAcqRights[i].DeltaNumber = delta
		prodAcqRights[i].DeltaCost = prodAcqRights[i].TotalCost - prodAcqRights[i].AvgUnitPrice*float64(computedLicenses)
	}

	return &v1.ListAcqRightsForApplicationsProductResponse{
		AcqRights: prodAcqRights,
	}, nil

}
