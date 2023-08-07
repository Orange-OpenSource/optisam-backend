package v1

import (
	"context"
	"strconv"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesAttrSum(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricAttrSum(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesAttrSum", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric Attr sum")
	}
	ind := metricNameExistsAttrSum(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}
	mat, err := computedMetricAttrSum(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesAttrSum - computedMetricACS - ", zap.Error(err))
		return 0, "", err
	}
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricAttrSumComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricAttrSumComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesAttrSum - ", zap.String("reason", err.Error()))
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric attribute sum standard")

	}
	return computedLicenses, "Sum of values: " + strconv.FormatUint(computedDetails, 10), nil
}

func computedMetricAttrSum(met *repo.MetricAttrSumStand, eqTypes []*repo.EquipmentType) (*repo.MetricAttrSumStandComputed, error) {
	idx := equipmentTypeExistsByType(met.EqType, eqTypes)
	if idx == -1 {
		logger.Log.Error("service/v1 - equipmentTypeExistsByType")
		return nil, status.Error(codes.Internal, "cannot find equipment type")
	}
	attr, err := attributeExistsByName(eqTypes[idx].Attributes, met.AttributeName)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesAttrSum - attributeExistsByName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "attribute doesnt exits")

	}
	return &repo.MetricAttrSumStandComputed{
		Name:           met.Name,
		BaseType:       eqTypes[idx],
		Attribute:      attr,
		ReferenceValue: met.ReferenceValue,
	}, nil
}

func metricNameExistsAttrSum(metrics []*repo.MetricAttrSumStand, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
