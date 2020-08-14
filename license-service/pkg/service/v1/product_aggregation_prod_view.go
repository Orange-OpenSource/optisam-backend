// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListAcqRightsForProductAggregation(ctx context.Context, req *v1.ListAcqRightsForProductAggregationRequest) (*v1.ListAcqRightsForProductAggregationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	params := &repo.QueryProductAggregations{}
	prodAgg, err := s.licenseRepo.ProductAggregationDetails(ctx, req.ID, params, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Product Aggregation Details-> "+err.Error())
	}
	logger.Log.Info("Aggregation", zap.Any("agrregation_detail", prodAgg))
	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	logger.Log.Info("Metrices", zap.Any("metrics_detail", metrics))
	var totalUnitPrice, totalCost float64
	var acqLicenses int32
	skus := make([]string, len(prodAgg.AcqRightsFull))
	swidTags := make([]string, len(prodAgg.AcqRightsFull))
	for i, acqRight := range prodAgg.AcqRightsFull {
		skus[i] = acqRight.SKU
		swidTags[i] = acqRight.SwidTag
		acqLicenses += int32(acqRight.AcquiredLicensesNumber)
		totalUnitPrice += float64(acqRight.AvgLicenesUnitPrice)
		totalCost += float64(acqRight.TotalCost)
	}

	acqRight := &v1.ProductAcquiredRights{
		SKU:     strings.Join(skus, ","),
		SwidTag: strings.Join(swidTags, ","),
		Metric:  prodAgg.Metric,
		//NumCptLicences: int32(computedLicenses),
		NumAcqLicences: acqLicenses,
		TotalCost:      totalCost,
		//	DeltaNumber:    delta,
		//DeltaCost:      float64(delta) * avgUnitPrice,
	}

	if prodAgg.NumOfEquipments == 0 {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - no equipments linked with product")

		return &v1.ListAcqRightsForProductAggregationResponse{
			AcqRights: []*v1.ProductAcquiredRights{
				acqRight,
			},
		}, nil
	}

	avgUnitPrice := totalUnitPrice / float64(len(prodAgg.AcqRightsFull))

	ind := metricNameExistsAll(metrics, prodAgg.Metric)
	if ind == -1 {
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - metric name doesnt exist - " + prodAgg.Metric)
		return &v1.ListAcqRightsForProductAggregationResponse{
			AcqRights: []*v1.ProductAcquiredRights{
				acqRight,
			},
		}, nil
	}
	metricFull := metrics[ind]

	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	computedLicenses := uint64(0)
	switch metricFull.Type {
	case repo.MetricOPSOracleProcessorStandard:
		cal := func(mat *repo.MetricOPSComputed) (uint64, error) {
			return s.licenseRepo.MetricOPSComputedLicensesAgg(ctx, prodAgg.Name, prodAgg.Metric, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesOPS(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute OPS licenses")
		}
	case repo.MetricSPSSagProcessorStandard:
		cal := func(mat *repo.MetricSPSComputed) (uint64, uint64, error) {
			return s.licenseRepo.MetricSPSComputedLicensesAgg(ctx, prodAgg.Name, prodAgg.Metric, mat, userClaims.Socpes)
		}
		licensesProd, licensesNonProd, err := s.computedLicensesSPS(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - MetricSPSSagProcessorStandard ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute SPS licenses")
		}
		if licensesProd > licensesNonProd {
			computedLicenses = licensesProd
		} else {
			computedLicenses = licensesNonProd
		}
	case repo.MetricIPSIbmPvuStandard:
		cal := func(mat *repo.MetricIPSComputed) (uint64, error) {
			return s.licenseRepo.MetricIPSComputedLicensesAgg(ctx, prodAgg.Name, prodAgg.Metric, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesIPS(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductListAcqRightsForProductAggregation - MetricIPSIbmPvuStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute IPS licenses")
		}
	case repo.MetricOracleNUPStandard:
		cal := func(mat *repo.MetricNUPComputed) (uint64, error) {
			return s.licenseRepo.MetricNUPComputedLicensesAgg(ctx, prodAgg.Name, prodAgg.Metric, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesNUP(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute NUP licenses")
		}
	case repo.MetricAttrCounterStandard:
		cal := func(mat *repo.MetricACSComputed) (uint64, error) {
			return s.licenseRepo.MetricACSComputedLicensesAgg(ctx, prodAgg.Name, prodAgg.Metric, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesACS(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute ACS licenses")
		}
	default:
		logger.Log.Error("service/v1 - ListAcqRightsForProductAggregation - metric type doesnt match - " + string(metrics[ind].Type))
		return nil, status.Error(codes.Internal, "cannot find metric for computation")
	}
	delta := acqLicenses - int32(computedLicenses)
	acqRight.NumCptLicences = int32(computedLicenses)
	acqRight.DeltaNumber = delta
	acqRight.DeltaCost = float64(delta) * avgUnitPrice
	acqRight.AvgUnitPrice = avgUnitPrice

	return &v1.ListAcqRightsForProductAggregationResponse{
		AcqRights: []*v1.ProductAcquiredRights{
			acqRight,
		},
	}, nil
}

func convertRepoToSrvProductAll(prods []*repo.ProductData) []*v1.Product {
	products := make([]*v1.Product, len(prods))
	for i := range prods {
		products[i] = convertRepoToSrvProduct(prods[i])
	}
	return products
}

func convertRepoToSrvProduct(prod *repo.ProductData) *v1.Product {
	return &v1.Product{
		SwidTag:           prod.Swidtag,
		Name:              prod.Name,
		Version:           prod.Version,
		Category:          prod.Category,
		Editor:            prod.Editor,
		NumOfApplications: prod.NumOfApplications,
		NumofEquipments:   prod.NumOfEquipments,
		TotalCost:         float64(prod.TotalCost),
	}
}
