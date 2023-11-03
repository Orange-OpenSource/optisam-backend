package v1

import (
	"context"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func metricNameExistsSPS(metrics []*repo.MetricSPS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func (s *licenseServiceServer) computedLicensesSPS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricSPS(ctx, scope...)
	prodID, _ := input[ProdID].([]string)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot fetch metric SPS")

	}
	ind := 0
	if ind = metricNameExistsSPS(metrics, input[MetricName].(string)); ind == -1 {
		return 0, 0, status.Error(codes.Internal, "cannot find metric name")
	}

	mat, err := computedMetricSPS(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - computedMetricSPS - ", zap.Error(err))
		return 0, 0, err
	}
	computedLicensesProd, computedLicensesNonProd := uint64(0), uint64(0)
	if input[IsAgg].(bool) {
		computedLicensesProd, computedLicensesNonProd, err = s.licenseRepo.MetricSPSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicensesProd, computedLicensesNonProd, err = s.licenseRepo.MetricSPSComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesSPS - ", zap.String("reason", err.Error()))
		return 0, 0, status.Error(codes.Internal, "cannot compute licenses for metric SPS")

	}
	return computedLicensesProd, computedLicensesNonProd, nil

}

func computedMetricSPSWithName(met *repo.MetricSPS, eqTypes []*repo.EquipmentType, name string) (*repo.MetricSPSComputed, error) {
	metric, err := computedMetricSPS(met, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricSPSWithName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute SPS metric")
	}
	metric.Name = name
	return metric, nil
}

func computedMetricSPS(met *repo.MetricSPS, eqTypes []*repo.EquipmentType) (*repo.MetricSPSComputed, error) {
	equipBase, err := equipmentTypeExistsByID(met.BaseEqTypeID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - equipmentTypeExistsByID - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot find base level equipment type")
	}
	numOfCores, err := attributeExists(equipBase.Attributes, met.NumCoreAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - attributeExists - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	numOfCPU, err := attributeExists(equipBase.Attributes, met.NumCPUAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - attributeExists - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "numofcpu attribute doesnt exits")

	}
	coreFactor, err := attributeExists(equipBase.Attributes, met.CoreFactorAttrID)
	if err != nil {
		logger.Log.Error("service/v1 - attributeExists - corefactor", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}

	return &repo.MetricSPSComputed{
		BaseType:       equipBase,
		NumCoresAttr:   numOfCores,
		CoreFactorAttr: coreFactor,
		NumCPUAttr:     numOfCPU,
	}, nil
}
