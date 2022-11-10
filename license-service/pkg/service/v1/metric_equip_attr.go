package v1

import (
	"context"
	"strings"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesEquipAttr(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricEquipAttr(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesEquipAttr", zap.Error(err))
		return 0, status.Error(codes.Internal, "cannot fetch metric equip attribute")
	}
	ind := metricNameExistsEquipAttr(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, status.Error(codes.NotFound, "cannot find metric name")
	}
	mat, err := computedMetricEquipAttr(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesEquipAttr - computedMetricACS - ", zap.Error(err))
		return 0, err
	}
	computedLicenses := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, err = s.licenseRepo.MetricEquipAttrComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricEquipAttrComputedLicenses(ctx, input[ProdID].(string), mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesEquipAttr - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric attribute sum standard")

	}
	return computedLicenses, nil
}

func computedMetricEquipAttr(met *repo.MetricEquipAttrStand, eqTypes []*repo.EquipmentType) (*repo.MetricEquipAttrStandComputed, error) {
	baseidx := equipmentTypeExistsByType(met.EqType, eqTypes)
	if baseidx == -1 {
		logger.Log.Error("service/v1 - equipmentTypeExistsByType")
		return nil, status.Error(codes.Internal, "cannot find equipment type")
	}
	attr, err := attributeExistsByName(eqTypes[baseidx].Attributes, met.AttributeName)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesEquipAttr - attributeExistsByName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "attribute doesnt exits")
	}
	parTree, err := parentHierarchy(eqTypes, eqTypes[baseidx].ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot make equipment parent Hierarchy")
	}
	idx := equipmentTypeExistsByType(met.EqType, parTree)
	if idx == -1 {
		logger.Log.Error("service/v1 - equipmentTypeExistsByType")
		return nil, status.Error(codes.Internal, "cannot find equipment type")
	}
	envEqTypeIdx := equipmentTypeForAttribute("environment", parTree)
	if envEqTypeIdx == -1 {
		logger.Log.Error("service/v1 - equipmentTypeForAttribute")
		return nil, status.Error(codes.Internal, "cannot find equipment type for attribute environment")
	}
	if envEqTypeIdx > idx {
		parTree = parTree[idx : envEqTypeIdx+1]
	} else {
		parTree = parTree[idx : idx+1]
	}
	return &repo.MetricEquipAttrStandComputed{
		Name:        met.Name,
		EqTypeTree:  parTree,
		BaseType:    eqTypes[baseidx],
		Environment: met.Environment,
		Attribute:   attr,
		Value:       met.Value,
	}, nil
}

func metricNameExistsEquipAttr(metrics []*repo.MetricEquipAttrStand, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func equipmentTypeForAttribute(attr string, eqTypes []*repo.EquipmentType) int {
	for i := range eqTypes {
		for j := range eqTypes[i].Attributes {
			if strings.EqualFold(eqTypes[i].Attributes[j].Name, strings.ToLower(attr)) {
				return i
			}
		}
	}
	return -1
}
