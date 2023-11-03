package v1

import (
	"context"
	"fmt"
	"strconv"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesNUP(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricNUP(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Sugar().Infow("computedLicensesNUP", "error", err.Error())
		return 0, "", status.Error(codes.Internal, "cannot fetch metric NUP")

	}
	ind := 0
	if ind = metricNameExistsNUP(metrics, input[MetricName].(string)); ind == -1 {
		return 0, "", status.Error(codes.Internal, "metric name doesnot exists")
	}
	parTree, err := parentHierarchy(eqTypes, metrics[ind].StartEqTypeID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "cannot fetch equipment types")

	}

	baseLevelIdx, err := validateLevelsNew(parTree, 0, metrics[ind].BaseEqTypeID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "cannot find base level equipment type in parent hierarchy")
	}

	aggLevelIdx, err := validateLevelsNew(parTree, baseLevelIdx, metrics[ind].AggerateLevelEqTypeID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "cannot find aggregate level equipment type in parent hierarchy")
	}

	endLevelIdx, err := validateLevelsNew(parTree, aggLevelIdx, metrics[ind].EndEqTypeID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "cannot find end level equipment type in parent hierarchy")
	}

	numOfCores, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].NumCoreAttrID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	numOfCPU, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].NumCPUAttrID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "numofcpu attribute doesnt exits")

	}
	coreFactor, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].CoreFactorAttrID)
	if err != nil {
		return 0, "", status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}
	mat := &repo.MetricNUPComputed{
		EqTypeTree:     parTree[:endLevelIdx+1],
		BaseType:       parTree[baseLevelIdx],
		AggregateLevel: parTree[aggLevelIdx],
		NumCoresAttr:   numOfCores,
		NumCPUAttr:     numOfCPU,
		CoreFactorAttr: coreFactor,
		NumOfUsers:     metrics[ind].NumberOfUsers,
	}
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricNUPComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		mat.Name = input[MetricName].(string)
		computedLicenses, computedDetails, err = s.licenseRepo.MetricNUPComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric OPS")

	}
	return computedLicenses, "Total users: " + strconv.FormatUint(computedDetails, 10), nil
}

func metricNameExistsNUP(metrics []*repo.MetricNUPOracle, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func computedMetricFromMetricNUPWithName(met *repo.MetricNUPOracle, eqTypes []*repo.EquipmentType, name string) (*repo.MetricNUPComputed, error) {
	metric, err := computedMetricFromMetricNUP(met, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricFromMetricNUPWithName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute NUP metric")
	}
	metric.Name = name
	return metric, nil
}

func computedMetricFromMetricNUP(metric *repo.MetricNUPOracle, eqTypes []*repo.EquipmentType) (*repo.MetricNUPComputed, error) {
	compMetOPS, err := computedMetricFromMetricOPS(metric.MetricOPS(), eqTypes)
	if err != nil {
		return nil, fmt.Errorf("computedMetricFromMetricOPS - Err: %v", err)
	}

	return repo.NewMetricNUPComputed(compMetOPS, metric.NumberOfUsers), nil
}
