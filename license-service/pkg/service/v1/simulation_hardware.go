// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"fmt"
	"math"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) LicensesForEquipAndMetric(ctx context.Context, req *v1.LicensesForEquipAndMetricRequest) (*v1.LicensesForEquipAndMetricResponse, error) {
	// Retrieving Claims
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	// checking if the claims are outside of the user scope
	scopes := userClaims.Socpes
	if req.Scopes != nil {
		scps, err := compareScopes(req.Scopes, userClaims.Socpes)
		if err != nil {
			logger.Log.Error("Service/v1 - LicensesForEquipAndMetric - compareScopes - ", zap.String("scopes allowed", strings.Join(userClaims.Socpes, ",")), zap.String("Scopes asked", strings.Join(req.Scopes, ",")))
			return nil, status.Error(codes.PermissionDenied, "requested scopes are outside the scope of user")
		}
		scopes = scps
	}

	// Fetching equipment types
	equipTypes, err := s.licenseRepo.EquipmentTypes(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - LicensesForEquipAndMetric - EquipmentTypes", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	// Checking if given equipment type exist
	index := equipmentTypeExistsByType(req.EquipType, equipTypes)
	if index == -1 {
		return nil, status.Error(codes.NotFound, "equipment type does not exist")
	}

	// Working according to requested metric
	switch req.MetricType {
	case repo.MetricOPSOracleProcessorStandard.String():
		// Fetching all the OPS metrics
		metrics, err := s.licenseRepo.ListMetricOPS(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ListMetricOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch OPS metrics")
		}
		// Is the given metrics exists or not.
		metric, index := metricExistbyNameOPS(req.MetricName, metrics)
		if index == -1 {
			return nil, status.Error(codes.NotFound, "metric does not exist")
		}
		// fetching computed metric
		computedMetric, err := computedMetricFromMetricOPSWithName(metric, equipTypes, metric.Name)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedMetricFromMetricOPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch computed metric")
		}

		if req.EquipType != computedMetric.BaseType.Type {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricOPS", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "cannot simulate OPS metric for types other than base type")
		}
		//finding the position of the base equipment in eqTypeTree
		baseIndex := baseIndexInMetricEqTypeTreeOPS(computedMetric)

		// Finding the depth
		equipmentRecursionDepth := len(computedMetric.EqTypeTree) - baseIndex

		// Find the parent heirarchy of the equipment
		equipment, err := s.licenseRepo.ParentsHirerachyForEquipment(ctx, req.EquipId, req.EquipType, uint8(equipmentRecursionDepth), scopes)
		if err != nil && err == repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ParentsHirerachyForEquipment", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "equipment does not exist")
		} else if err != nil && err != repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ParentsHirerachyForEquipment", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "can not fetch equipment")
		}

		//Finding the top equipment
		topEquipment := topEquipmentInEquipmentLinkList(equipment)
		//Index of top equipment in EqTypeTree
		indexTopEquipment := topEquipmentInEquipmentTypeTreeOPS(computedMetric, topEquipment)

		//finding the products for the top equipment
		products, err := s.licenseRepo.ProductsForEquipmentForMetricOracleProcessorStandard(ctx, topEquipment.EquipID, topEquipment.Type, uint8(indexTopEquipment+1), computedMetric, scopes)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ProductsForEquipmentForMetricOracleProcessorStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch products for equipment")
		} else if err == repo.ErrNoData {
			return &v1.LicensesForEquipAndMetricResponse{}, nil
		}
		// Finding old licenses
		oldLicenses, err := s.licenseRepo.ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, topEquipment.EquipID, topEquipment.Type, computedMetric, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedLicensesForEquipmentForMetricOracleProcessorStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch old licenses for OPS metric")
		}
		aggEquipment := aggEquipmentInEquipmentLinkList(equipment, computedMetric.AggregateLevel.Type)
		oldLicensesAgg, unceiledLicensesAgg, err := s.licenseRepo.ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, aggEquipment.EquipID, aggEquipment.Type, computedMetric, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedLicensesForEquipmentForMetricOracleProcessorStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch old licenses for OPS metric")
		}

		computedMetric = getMetricWithNewValuesOPS(computedMetric, req.Attributes)
		servLicNew, servLicOld := computedMetric.Licenses()
		unceiledLicensesAgg = unceiledLicensesAgg + servLicNew - servLicOld

		newLicenses := oldLicenses - oldLicensesAgg + int64(math.Ceil(unceiledLicensesAgg))
		delta := newLicenses - oldLicenses

		var licenses []*v1.ProductLicenseForEquipAndMetric

		for _, product := range products {
			licenses = append(licenses, &v1.ProductLicenseForEquipAndMetric{
				MetricName:  req.MetricName,
				OldLicences: oldLicenses,
				NewLicenses: newLicenses,
				Delta:       delta,
				Product:     repoProductToServProduct(product),
			})
		}

		return &v1.LicensesForEquipAndMetricResponse{
			Licenses: licenses,
		}, nil
	case repo.MetricOracleNUPStandard.String():
		metrics, err := s.licenseRepo.ListMetricNUP(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ListMetricNUP", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch NUP metric")
		}
		// Is the given metrics exists or not.
		metric, index := metricExistbyNameNUP(req.MetricName, metrics)
		if index == -1 {
			return nil, status.Error(codes.NotFound, "metric does not exist")
		}
		// fetching computed metric
		computedMetric, err := computedMetricFromMetricNUPWithName(metric, equipTypes, metric.Name)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedMetricFromMetricNUP", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch computed metric")
		}
		if req.EquipType != computedMetric.BaseType.Type {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricNUP", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "cannot simulate NUP metric for types other than base type")
		}
		baseIndex := baseIndexInMetricEqTypeTreeNUP(computedMetric)

		// Finding the depth
		equipmentRecursionDepth := len(computedMetric.EqTypeTree) - baseIndex

		// Find the parent heirarchy of the equipment
		equipment, err := s.licenseRepo.ParentsHirerachyForEquipment(ctx, req.EquipId, req.EquipType, uint8(equipmentRecursionDepth), scopes)
		if err != nil && err == repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ParentsHirerachyForEquipment", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "equipment does not exist")
		} else if err != nil && err != repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ParentsHirerachyForEquipment", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "can not fetch equipment")
		}

		//Finding the top equipment
		topEquipment := topEquipmentInEquipmentLinkList(equipment)

		//Index of top equipment in EqTypeTree
		indexTopEquipment := topEquipmentInEquipmentTypeTreeNUP(computedMetric, topEquipment)

		//finding the products for the top equipment
		products, err := s.licenseRepo.ProductsForEquipmentForMetricOracleNUPStandard(ctx, topEquipment.EquipID, topEquipment.Type, uint8(indexTopEquipment+1), computedMetric, scopes)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ProductsForEquipmentForMetricOracleNUPStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch products for equipment")
		} else if err == repo.ErrNoData {
			return &v1.LicensesForEquipAndMetricResponse{}, nil
		}
		// Finding old licenses
		oldLicenses, err := s.licenseRepo.ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx, topEquipment.EquipID, topEquipment.Type, computedMetric.MetricOPSComputed(), scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedLicensesForEquipmentForMetricOracleProcessorStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch old licenses for OPS metric")
		}

		aggEquipment := aggEquipmentInEquipmentLinkList(equipment, computedMetric.AggregateLevel.Type)
		oldLicensesAgg, unceiledLicensesAgg, err := s.licenseRepo.ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, aggEquipment.EquipID, aggEquipment.Type, computedMetric.MetricOPSComputed(), scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ComputedLicensesForEquipmentForMetricOracleProcessorStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch old licenses for OPS metric")
		}
		computedMetric = getMetricWithNewValuesNUP(computedMetric, req.Attributes)
		servLicNew, servLicOld := computedMetric.MetricOPSComputed().Licenses()
		unceiledLicensesAgg = unceiledLicensesAgg + servLicNew - servLicOld

		newLicenses := oldLicenses - oldLicensesAgg + int64(math.Ceil(unceiledLicensesAgg))

		oldLicenses = oldLicenses * int64(computedMetric.NumOfUsers)
		newLicenses = newLicenses * int64(computedMetric.NumOfUsers)
		var licenses []*v1.ProductLicenseForEquipAndMetric
		for _, product := range products {
			// get user nodes in the system
			fmt.Printf("Metric - num Cores Attr :%+v\n", computedMetric.NumCoresAttr)
			users, err := s.licenseRepo.UsersForEquipmentForMetricOracleNUP(ctx, topEquipment.EquipID, topEquipment.Type, product.Swidtag, uint8(indexTopEquipment+1), computedMetric, req.Scopes)
			if err != nil {
				if err == repo.ErrNoData {
					logger.Log.Info("service/v1 - LicensesForEquipAndMetric - user nodes not found assuming 1 node with 0 users", zap.String("product-swidtag", ""))
					licenses = append(licenses, &v1.ProductLicenseForEquipAndMetric{
						MetricName:  req.MetricName,
						OldLicences: oldLicenses,
						NewLicenses: newLicenses,
						Delta:       newLicenses - oldLicenses,
						Product:     repoProductToServProduct(product),
					})
					continue
				}

				logger.Log.Error("service/v1 - LicensesForEquipAndMetric - UsersForEquipmentForMetricOracleNUP", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch new licenses for OPS metric")
			}
			var ol, nl int64
			for _, user := range users {
				ol += max(oldLicenses, user.UserCount)
				nl += max(newLicenses, user.UserCount)
			}

			licenses = append(licenses, &v1.ProductLicenseForEquipAndMetric{
				MetricName:  req.MetricName,
				OldLicences: ol,
				NewLicenses: nl,
				Delta:       nl - ol,
				Product:     repoProductToServProduct(product),
			})
		}
		return &v1.LicensesForEquipAndMetricResponse{
			Licenses: licenses,
		}, nil
	case repo.MetricIPSIbmPvuStandard.String():
		metrics, err := s.licenseRepo.ListMetricIPS(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ListMetricIPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch IPS metric")
		}
		// Is the given metrics exists or not.
		index := metricNameExistsIPS(metrics, req.MetricName)
		if index == -1 {
			return nil, status.Error(codes.NotFound, "metric does not exist")
		}

		metric, err := computedMetricIPSWithName(metrics[index], equipTypes, metrics[index].Name)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricIPS", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot compute IPS metric")
		}
		if req.EquipType != metric.BaseType.Type {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricIPS", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "cannot simulate IPS metric for types other than base type")
		}

		metric = getMetricWithNewValuesIPS(metric, req.Attributes)

		//finding the products for the equipment
		products, err := s.licenseRepo.ProductsForEquipmentForMetricIPSStandard(ctx, req.EquipId, req.EquipType, uint8(1), metric, scopes)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ProductsForEquipmentForMetricIPSStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch products for equipment")
		} else if err == repo.ErrNoData {
			return &v1.LicensesForEquipAndMetricResponse{}, nil
		}

		oldLicenses := int64(getAttributeValues(metric.CoreFactorAttr, false) * getAttributeValues(metric.NumCoresAttr, false))
		newLicenses := int64(getAttributeValues(metric.CoreFactorAttr, true) * getAttributeValues(metric.NumCoresAttr, true))
		delta := newLicenses - oldLicenses
		var licenses []*v1.ProductLicenseForEquipAndMetric
		for _, product := range products {
			licenses = append(licenses, &v1.ProductLicenseForEquipAndMetric{
				MetricName:  req.MetricName,
				OldLicences: oldLicenses,
				NewLicenses: newLicenses,
				Delta:       delta,
				Product:     repoProductToServProduct(product),
			})
		}
		return &v1.LicensesForEquipAndMetricResponse{
			Licenses: licenses,
		}, nil
	case repo.MetricSPSSagProcessorStandard.String():
		metrics, err := s.licenseRepo.ListMetricSPS(ctx, scopes)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ListMetricSPS", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch SPS metric")
		}
		// Is the given metrics exists or not.
		index := metricNameExistsSPS(metrics, req.MetricName)
		if index == -1 {
			return nil, status.Error(codes.NotFound, "metric does not exist")
		}

		metric, err := computedMetricSPSWithName(metrics[index], equipTypes, metrics[index].Name)
		if err != nil {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricSPS", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot compute SPS metric")
		}

		if req.EquipType != metric.BaseType.Type {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - computedMetricSPS", zap.Error(err))
			return nil, status.Error(codes.InvalidArgument, "cannot simulate SPS metric for types other than base type")
		}

		metric = getMetricWithNewValuesSPS(metric, req.Attributes)

		//finding the products for the equipment
		products, err := s.licenseRepo.ProductsForEquipmentForMetricSAGStandard(ctx, req.EquipId, req.EquipType, uint8(1), metric, scopes)
		if err != nil && err != repo.ErrNoData {
			logger.Log.Error("service/v1 - LicensesForEquipAndMetric - ProductsForEquipmentForMetricSAGStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot fetch products for equipment")
		} else if err == repo.ErrNoData {
			return &v1.LicensesForEquipAndMetricResponse{}, nil
		}

		oldLicenses := int64(getAttributeValues(metric.CoreFactorAttr, false) * getAttributeValues(metric.NumCoresAttr, false))
		newLicenses := int64(getAttributeValues(metric.CoreFactorAttr, true) * getAttributeValues(metric.NumCoresAttr, true))
		delta := newLicenses - oldLicenses
		var licenses []*v1.ProductLicenseForEquipAndMetric
		for _, product := range products {
			licenses = append(licenses, &v1.ProductLicenseForEquipAndMetric{
				MetricName:  req.MetricName,
				OldLicences: oldLicenses,
				NewLicenses: newLicenses,
				Delta:       delta,
				Product:     repoProductToServProduct(product),
			})
		}
		return &v1.LicensesForEquipAndMetricResponse{
			Licenses: licenses,
		}, nil
	default:
		return nil, status.Error(codes.Unimplemented, "Metric is not supported for simulation")
	}

}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func userOwnedScopes(reqScopes []string, claimsScopes []string) error {
	_, err := compareScopes(reqScopes, claimsScopes)
	return err
}

// TODO: we should not use this fuction anymore if we want to check request scopes are owned by user or not
// remove this in future
func compareScopes(reqScopes []string, claimsScopes []string) ([]string, error) {
	//	fmt.Println(reqScopes.Scopes, claimsScopes)
	scopes := claimsScopes
	if !scopesIsSubSlice(reqScopes, claimsScopes) {
		return nil, status.Error(codes.InvalidArgument, "scopes are not owned by user")

	}
	scopes = reqScopes
	return scopes, nil
}

func metricExistbyNameOPS(metricName string, metrics []*repo.MetricOPS) (*repo.MetricOPS, int) {
	for i := 0; i < len(metrics); i++ {
		if metrics[i].Name == metricName {
			return metrics[i], i
		}
	}
	return nil, -1
}

func topEquipmentInEquipmentLinkList(equipment *repo.Equipment) *repo.Equipment {
	currentEquipment := equipment
	for currentEquipment.Parent != nil {
		currentEquipment = currentEquipment.Parent
	}

	return currentEquipment
}

func aggEquipmentInEquipmentLinkList(equipment *repo.Equipment, aggregateType string) *repo.Equipment {
	currentEquipment := equipment
	for currentEquipment.Parent != nil && currentEquipment.Type != aggregateType {
		currentEquipment = currentEquipment.Parent
	}

	return currentEquipment
}

func getMetricWithNewValuesOPS(computedMetric *repo.MetricOPSComputed, attributes []*v1.Attribute) *repo.MetricOPSComputed {

	if index := containsAttribute(attributes, computedMetric.CoreFactorAttr.Name); index != -1 {
		computedMetric.CoreFactorAttr = servAttrToRepoAttr(attributes[index])
	}

	if index := containsAttribute(attributes, computedMetric.NumCoresAttr.Name); index != -1 {

		computedMetric.NumCoresAttr = servAttrToRepoAttr(attributes[index])

	}

	if index := containsAttribute(attributes, computedMetric.NumCPUAttr.Name); index != -1 {

		computedMetric.NumCPUAttr = servAttrToRepoAttr(attributes[index])
	}

	return computedMetric

}

func getMetricWithNewValuesNUP(computedMetric *repo.MetricNUPComputed, attributes []*v1.Attribute) *repo.MetricNUPComputed {

	if index := containsAttribute(attributes, computedMetric.CoreFactorAttr.Name); index != -1 {

		computedMetric.CoreFactorAttr = servAttrToRepoAttr(attributes[index])

	}

	if index := containsAttribute(attributes, computedMetric.NumCoresAttr.Name); index != -1 {

		computedMetric.NumCoresAttr = servAttrToRepoAttr(attributes[index])

	}

	if index := containsAttribute(attributes, computedMetric.NumCPUAttr.Name); index != -1 {

		computedMetric.NumCPUAttr = servAttrToRepoAttr(attributes[index])

	}

	return computedMetric

}

func getMetricWithNewValuesIPS(computedMetric *repo.MetricIPSComputed, attributes []*v1.Attribute) *repo.MetricIPSComputed {
	if index := containsAttribute(attributes, computedMetric.CoreFactorAttr.Name); index != -1 {
		computedMetric.CoreFactorAttr = servAttrToRepoAttr(attributes[index])
	}

	if index := containsAttribute(attributes, computedMetric.NumCoresAttr.Name); index != -1 {
		computedMetric.NumCoresAttr = servAttrToRepoAttr(attributes[index])
	}

	return computedMetric
}

func getMetricWithNewValuesSPS(computedMetric *repo.MetricSPSComputed, attributes []*v1.Attribute) *repo.MetricSPSComputed {
	if index := containsAttribute(attributes, computedMetric.CoreFactorAttr.Name); index != -1 {
		computedMetric.CoreFactorAttr = servAttrToRepoAttr(attributes[index])
	}

	if index := containsAttribute(attributes, computedMetric.NumCoresAttr.Name); index != -1 {
		computedMetric.NumCoresAttr = servAttrToRepoAttr(attributes[index])
	}

	return computedMetric
}

func containsAttribute(attributes []*v1.Attribute, attributeName string) int {
	for i := 0; i < len(attributes); i++ {
		if attributes[i].Name == attributeName {
			return i
		}
	}
	return -1
}

func baseIndexInMetricEqTypeTreeOPS(computedMetric *repo.MetricOPSComputed) int {
	for i := 0; i < len(computedMetric.EqTypeTree); i++ {
		if computedMetric.BaseType.Type == computedMetric.EqTypeTree[i].Type {
			return i
		}
	}
	return -1
}

func baseIndexInMetricEqTypeTreeNUP(computedMetric *repo.MetricNUPComputed) int {
	for i := 0; i < len(computedMetric.EqTypeTree); i++ {
		if computedMetric.BaseType.Type == computedMetric.EqTypeTree[i].Type {
			return i
		}
	}
	return -1
}

// TODO: rename this to index for equipment type
func topEquipmentInEquipmentTypeTreeOPS(computedMetric *repo.MetricOPSComputed, top *repo.Equipment) int {
	for i := range computedMetric.EqTypeTree {
		if computedMetric.EqTypeTree[i].Type == top.Type {
			return i
		}
	}
	return -1
}

func topEquipmentInEquipmentTypeTreeNUP(computedMetric *repo.MetricNUPComputed, top *repo.Equipment) int {
	for i := range computedMetric.EqTypeTree {
		if computedMetric.EqTypeTree[i].Type == top.Type {
			return i
		}
	}
	return -1
}

func metricExistbyNameNUP(metricName string, metrics []*repo.MetricNUPOracle) (*repo.MetricNUPOracle, int) {
	for i := 0; i < len(metrics); i++ {
		if metrics[i].Name == metricName {
			return metrics[i], i
		}
	}
	return nil, -1
}

func getAttributeValues(a *repo.Attribute, useSimulated bool) float64 {
	switch a.Type {
	case repo.DataTypeFloat:
		if useSimulated && a.IsSimulated {
			return float64(a.FloatVal)
		}
		return float64(a.FloatValOld)
	case repo.DataTypeInt:
		if useSimulated && a.IsSimulated {
			return float64(a.IntVal)
		}
		return float64(a.IntValOld)
	}
	return 0
}
