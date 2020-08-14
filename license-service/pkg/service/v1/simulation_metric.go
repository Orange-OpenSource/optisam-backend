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

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ProductLicensesForMetric(ctx context.Context, req *v1.ProductLicensesForMetricRequest) (*v1.ProductLicensesForMetricResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	proID, err := s.licenseRepo.ProductIDForSwidtag(ctx, req.SwidTag, &repo.QueryProducts{}, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - ProductIDForSwidtag", zap.Error(err))
		return nil, status.Error(codes.NotFound, "cannot get product id for swid tag")
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	ind := metricNameExistsAll(metrics, req.MetricName)
	if ind == -1 {
		logger.Log.Error("service/v1 - ProductLicensesForMetric - metricNameExistsAll - " + req.MetricName)
		return nil, status.Error(codes.Internal, "metric name does not exist")
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
			return s.licenseRepo.MetricOPSComputedLicenses(ctx, proID, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesOPS(ctx, eqTypes, metricFull.Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ProductLicensesForMetric - ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute OPS licenses")
		}
	case repo.MetricSPSSagProcessorStandard:
		cal := func(mat *repo.MetricSPSComputed) (uint64, uint64, error) {
			return s.licenseRepo.MetricSPSComputedLicenses(ctx, proID, mat, userClaims.Socpes)
		}
		licensesProd, licensesNonProd, err := s.computedLicensesSPS(ctx, eqTypes, metrics[ind].Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ProductLicensesForMetric - MetricSPSSagProcessorStandard ", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute SPS licenses")
		}
		if licensesProd > licensesNonProd {
			computedLicenses = licensesProd
		} else {
			computedLicenses = licensesNonProd
		}
	case repo.MetricIPSIbmPvuStandard:
		cal := func(mat *repo.MetricIPSComputed) (uint64, error) {
			return s.licenseRepo.MetricIPSComputedLicenses(ctx, proID, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesIPS(ctx, eqTypes, metrics[ind].Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ProductLicensesForMetric - MetricIPSIbmPvuStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute IPS licenses")
		}
	case repo.MetricOracleNUPStandard:
		cal := func(mat *repo.MetricNUPComputed) (uint64, error) {
			return s.licenseRepo.MetricNUPComputedLicenses(ctx, proID, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesNUP(ctx, eqTypes, metrics[ind].Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ProductLicensesForMetric - MetricIPSIbmPvuStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute NUP licenses")
		}
	case repo.MetricAttrCounterStandard:
		cal := func(mat *repo.MetricACSComputed) (uint64, error) {
			return s.licenseRepo.MetricACSComputedLicenses(ctx, proID, mat, userClaims.Socpes)
		}
		computedLicenses, err = s.computedLicensesACS(ctx, eqTypes, metrics[ind].Name, cal)
		if err != nil {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricAttrCounterStandard", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "cannot compute ACS licenses")
		}
	default:
		logger.Log.Error("service/v1 - ProductLicensesForMetric - metric type doesnt match - " + string(metricFull.Type))
		return nil, status.Error(codes.Internal, "cannot find metric for computation")
	}
	return &v1.ProductLicensesForMetricResponse{
		MetricName:     req.MetricName,
		NumCptLicences: computedLicenses,
		TotalCost:      float64(computedLicenses) * req.UnitCost,
	}, nil
}
