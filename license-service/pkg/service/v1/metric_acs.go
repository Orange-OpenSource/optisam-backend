// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesACS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricACS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesACS", zap.Error(err))
		return 0, status.Error(codes.Internal, "cannot fetch metric ACS")
	}
	ind := metricNameExistsACS(metrics, input[METRIC_NAME].(string))
	if ind == -1 {
		return 0, status.Error(codes.NotFound, "cannot find metric name")
	}
	mat, err := computedMetricACS(metrics[ind], eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesACS - computedMetricACS - ", zap.Error(err))
		return 0, err
	}
	computedLicenses := uint64(0)
	if input[IS_AGG].(bool) {
		computedLicenses, err = s.licenseRepo.MetricACSComputedLicensesAgg(ctx, input[PROD_AGG_NAME].(string), input[METRIC_NAME].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricACSComputedLicenses(ctx, input[PROD_ID].(string), mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesACS - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric ACS")

	}
	return computedLicenses, nil
}

func computedMetricACS(met *repo.MetricACS, eqTypes []*repo.EquipmentType) (*repo.MetricACSComputed, error) {
	idx := equipmentTypeExistsByType(met.EqType, eqTypes)
	if idx == -1 {
		logger.Log.Error("service/v1 - equipmentTypeExistsByType")
		return nil, status.Error(codes.Internal, "cannot find equipment type")
	}
	attr, err := attributeExistsByName(eqTypes[idx].Attributes, met.AttributeName)
	if err != nil {
		logger.Log.Error("service/v1 - computedMetricACS - attributeExistsByName - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "attribute doesnt exits")

	}
	return &repo.MetricACSComputed{
		Name:      met.Name,
		BaseType:  eqTypes[idx],
		Attribute: attr,
		Value:     met.Value,
	}, nil
}

func validateAttributeACSMetric(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	if attrName == "" {
		return nil, status.Error(codes.InvalidArgument, "attribute name is empty")
	}
	attr, err := attributeExistsByName(attributes, attrName)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

func attributeExistsByName(attributes []*repo.Attribute, attrName string) (*repo.Attribute, error) {
	for _, attr := range attributes {
		if attr.Name == attrName {
			return attr, nil
		}
	}
	return nil, status.Error(codes.InvalidArgument, "attribute does not exists")
}

func metricNameExistsACS(metrics []*repo.MetricACS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
