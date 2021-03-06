// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesNUP(ctx context.Context, eqTypes []*repo.EquipmentType, met string, cal func(mat *repo.MetricNUPComputed) (uint64, error)) (uint64, error) {
	// TODO pass claims here
	userClaims, _ := ctxmanage.RetrieveClaims(ctx)
	metrics, err := s.licenseRepo.ListMetricNUP(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		return 0, status.Error(codes.Internal, "cannot fetch metric OPS")

	}
	ind := 0
	if ind = metricNameExistsNUP(metrics, met); ind == -1 {
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
	mat := &repo.MetricNUPComputed{
		EqTypeTree:     parTree[:endLevelIdx+1],
		BaseType:       parTree[baseLevelIdx],
		AggregateLevel: parTree[aggLevelIdx],
		NumCoresAttr:   numOfCores,
		NumCPUAttr:     numOfCPU,
		CoreFactorAttr: coreFactor,
		NumOfUsers:     metrics[ind].NumberOfUsers,
	}
	computedLicenses, err := cal(mat)
	if err != nil {
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric OPS")

	}

	return computedLicenses, nil
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
