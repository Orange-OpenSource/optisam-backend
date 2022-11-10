package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parentHierarchy(eqTypes []*repo.EquipmentType, startID string) ([]*repo.EquipmentType, error) {
	equip, err := equipmentTypeExistsByID(startID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}
	ancestors := []*repo.EquipmentType{}
	ancestors = append(ancestors, equip)
	parID := equip.ParentID
	for parID != "" {
		equipAnc, err := equipmentTypeExistsByID(parID, eqTypes)
		if err != nil {
			logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "parent hierarchy not found")
		}
		ancestors = append(ancestors, equipAnc)
		parID = equipAnc.ParentID
	}
	return ancestors, nil
}

func attributeExists(attributes []*repo.Attribute, attrID string) (*repo.Attribute, error) {
	for _, attr := range attributes {
		if attr.ID == attrID {
			return attr, nil
		}
	}
	return nil, status.Errorf(codes.Unknown, "attribute not exists")
}

func metricNameExistsOPS(metrics []*repo.MetricOPS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func validateLevelsNew(levels []*repo.EquipmentType, startIdx int, base string) (int, error) {
	for i := startIdx; i < len(levels); i++ {
		if levels[i].ID == base {
			return i, nil
		}
	}
	return -1, errors.New("not found")
}

func (s *licenseServiceServer) computedLicensesOPS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricOPS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		return 0, status.Error(codes.Internal, "cannot fetch metric OPS")

	}
	ind := 0
	if ind = metricNameExistsOPS(metrics, input[MetricName].(string)); ind == -1 {
		return 0, status.Error(codes.Internal, "metric name doesnot exists")
	}
	parTree, err := parentHierarchy(eqTypes, metrics[ind].StartEqTypeID)
	if err != nil {
		return 0, status.Error(codes.Internal, "cannot fetch equipment types")

	}

	baseLevelIdx, err := validateLevelsNew(parTree, 0, metrics[ind].BaseEqTypeID)
	if err != nil {
		return 0, status.Error(codes.Internal, "cannot find base level equipment type in parent hierarchy")
	}

	aggLevelIdx, err := validateLevelsNew(parTree, baseLevelIdx, metrics[ind].AggerateLevelEqTypeID)
	if err != nil {
		return 0, status.Error(codes.Internal, "cannot find aggregate level equipment type in parent hierarchy")
	}

	endLevelIdx, err := validateLevelsNew(parTree, aggLevelIdx, metrics[ind].EndEqTypeID)
	if err != nil {
		return 0, status.Error(codes.Internal, "cannot find end level equipment type in parent hierarchy")
	}

	numOfCores, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].NumCoreAttrID)
	if err != nil {
		return 0, status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	numOfCPU, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].NumCPUAttrID)
	if err != nil {
		return 0, status.Error(codes.Internal, "numofcpu attribute doesnt exits")

	}
	coreFactor, err := attributeExists(parTree[baseLevelIdx].Attributes, metrics[ind].CoreFactorAttrID)
	if err != nil {
		return 0, status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}
	mat := &repo.MetricOPSComputed{
		EqTypeTree:     parTree[:endLevelIdx+1],
		BaseType:       parTree[baseLevelIdx],
		AggregateLevel: parTree[aggLevelIdx],
		NumCoresAttr:   numOfCores,
		NumCPUAttr:     numOfCPU,
		CoreFactorAttr: coreFactor,
	}

	computedLicenses := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, err = s.licenseRepo.MetricOPSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		mat.Name = input[MetricName].(string)
		computedLicenses, err = s.licenseRepo.MetricOPSComputedLicenses(ctx, input[ProdID].(string), mat, scope...)
	}
	if err != nil {
		logger.Log.Sugar().Infow("cannot compute licenses for metric OPS", "err", err.Error())
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric OPS")
	}

	return computedLicenses, nil
}

func computedMetricFromMetricOPSWithName(met *repo.MetricOPS, eqTypes []*repo.EquipmentType, name string) (*repo.MetricOPSComputed, error) {
	metric, err := computedMetricFromMetricOPS(met, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricFromMetricOPSWithName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot compute OPS metric")
	}
	metric.Name = name
	return metric, nil
}

func computedMetricFromMetricOPS(metric *repo.MetricOPS, eqTypes []*repo.EquipmentType) (*repo.MetricOPSComputed, error) {
	parTree, err := parentHierarchy(eqTypes, metric.StartEqTypeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")

	}
	baseLevelIdx, err := validateLevelsNew(parTree, 0, metric.BaseEqTypeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find base level equipment type in parent hierarchy")
	}

	aggLevelIdx, err := validateLevelsNew(parTree, baseLevelIdx, metric.AggerateLevelEqTypeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find aggregate level equipment type in parent hierarchy")
	}

	endLevelIdx, err := validateLevelsNew(parTree, aggLevelIdx, metric.EndEqTypeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot find end level equipment type in parent hierarchy")
	}

	numOfCores, err := attributeExists(parTree[baseLevelIdx].Attributes, metric.NumCoreAttrID)
	if err != nil {
		return nil, status.Error(codes.Internal, "numofcores attribute doesnt exits")

	}
	numOfCPU, err := attributeExists(parTree[baseLevelIdx].Attributes, metric.NumCPUAttrID)
	if err != nil {
		return nil, status.Error(codes.Internal, "numofcpu attribute doesnt exits")

	}
	coreFactor, err := attributeExists(parTree[baseLevelIdx].Attributes, metric.CoreFactorAttrID)
	if err != nil {
		return nil, status.Error(codes.Internal, "coreFactor attribute doesnt exits")

	}
	mat := &repo.MetricOPSComputed{
		EqTypeTree:     parTree[:endLevelIdx+1],
		BaseType:       parTree[baseLevelIdx],
		AggregateLevel: parTree[aggLevelIdx],
		NumCoresAttr:   numOfCores,
		NumCPUAttr:     numOfCPU,
		CoreFactorAttr: coreFactor,
	}

	return mat, nil
}
