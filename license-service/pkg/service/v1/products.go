// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"log"
	"optisam-backend/common/optisam/ctxmanage"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListAcqRightsForProduct implements license service ListAcqRightsForProduct function
func (s *licenseServiceServer) ListAcqRightsForProduct(ctx context.Context, req *v1.ListAcquiredRightsForProductRequest) (*v1.ListAcquiredRightsForProductResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	ID, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	log.Println("UID of Product : ", ID)
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products -> "+err.Error())
	}
	log.Printf("product info %+v", *res)
	numEquips := int32(0)
	if len(res.Products) != 0 {
		numEquips = res.Products[0].NumofEquipments
	}

	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metric OPS")

	}

	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")

	}
	prodAcqRights := make([]*v1.ProductAcquiredRights, len(prodRights))
	ind := 0
	for i, acqRight := range prodRights {
		prodAcqRights[i] = &v1.ProductAcquiredRights{
			SKU:            acqRight.SKU,
			SwidTag:        req.SwidTag,
			Metric:         acqRight.Metric,
			NumAcqLicences: int32(acqRight.AcqLicenses),
			TotalCost:      acqRight.TotalCost,
			AvgUnitPrice:   acqRight.AvgUnitPrice,
		}
		if ind = metricNameExistsAll(metrics, acqRight.Metric); ind == -1 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric name doesnt exist - " + acqRight.Metric)
			continue
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			continue
		}

		computedLicenses := uint64(0)
		switch metrics[ind].Type {
		case repo.MetricOPSOracleProcessorStandard:
			cal := func(mat *repo.MetricOPSComputed) (uint64, error) {
				return s.licenseRepo.MetricOPSComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			computedLicenses, err = s.computedLicensesOPS(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - ", zap.String("reason", err.Error()))
				continue
			}
		case repo.MetricSPSSagProcessorStandard:
			cal := func(mat *repo.MetricSPSComputed) (uint64, uint64, error) {
				return s.licenseRepo.MetricSPSComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			licensesProd, licensesNonProd, err := s.computedLicensesSPS(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricSPSSagProcessorStandard ", zap.String("reason", err.Error()))
				continue
			}
			if licensesProd > licensesNonProd {
				computedLicenses = licensesProd
			} else {
				computedLicenses = licensesNonProd
			}
		case repo.MetricIPSIbmPvuStandard:
			cal := func(mat *repo.MetricIPSComputed) (uint64, error) {
				return s.licenseRepo.MetricIPSComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			computedLicenses, err = s.computedLicensesIPS(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricIPSIbmPvuStandard", zap.String("reason", err.Error()))
				continue
			}
		case repo.MetricOracleNUPStandard:
			cal := func(mat *repo.MetricNUPComputed) (uint64, error) {
				return s.licenseRepo.MetricNUPComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			computedLicenses, err = s.computedLicensesNUP(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - ", zap.String("reason", err.Error()))
				continue
			}
		case repo.MetricAttrCounterStandard:
			cal := func(mat *repo.MetricACSComputed) (uint64, error) {
				return s.licenseRepo.MetricACSComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			computedLicenses, err = s.computedLicensesACS(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricAttrCounterStandard", zap.String("reason", err.Error()))
				continue
			}
		case repo.MetricInstanceNumberStandard:
			cal := func(mat *repo.MetricINMComputed) (uint64, error) {
				return s.licenseRepo.MetricINMComputedLicenses(ctx, ID, mat, userClaims.Socpes)
			}
			computedLicenses, err = s.computedLicensesINM(ctx, eqTypes, metrics[ind].Name, cal)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricInstanceNumberStandard", zap.String("reason", err.Error()))
				continue
			}
		default:
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric type doesnt match - " + string(metrics[ind].Type))
			continue
		}

		delta := int32(acqRight.AcqLicenses) - int32(computedLicenses)

		prodAcqRights[i].NumCptLicences = int32(computedLicenses)
		prodAcqRights[i].DeltaNumber = int32(delta)
		prodAcqRights[i].DeltaCost = acqRight.AvgUnitPrice * float64(delta)
	}

	return &v1.ListAcquiredRightsForProductResponse{
		AcqRights: prodAcqRights,
	}, nil
}

func productAcqRightFilter(notForMetric string) *repo.AggregateFilter {
	return &repo.AggregateFilter{
		Filters: []repo.Queryable{
			&repo.Filter{
				FilterMatchingType: repo.EqFilter,
				FilterKey:          repo.AcquiredRightsSearchKeyMetric.String(),
				FilterValue:        notForMetric,
			},
		},
	}
}

func scopesIsSubSlice(scopes []string, claimsScopes []string) bool {
	if len(scopes) > len(claimsScopes) {
		return false
	}
	for _, e := range scopes {
		if contains(claimsScopes, e) == -1 {
			return false
		}
	}
	return true
}
func contains(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}

func stringToInterface(vals []string) []interface{} {
	interfaceSlice := make([]interface{}, len(vals))
	for i := range vals {
		interfaceSlice[i] = vals[i]
	}
	return interfaceSlice
}

func repoProductToServProduct(repoProductData *repo.ProductData) *v1.Product {
	return &v1.Product{
		Name:              repoProductData.Name,
		Version:           repoProductData.Version,
		Category:          repoProductData.Category,
		Editor:            repoProductData.Editor,
		SwidTag:           repoProductData.Swidtag,
		NumofEquipments:   repoProductData.NumOfEquipments,
		NumOfApplications: repoProductData.NumOfApplications,
		TotalCost:         repoProductData.TotalCost,
	}
}
