// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListAcqRightsForProductAggregation(ctx context.Context, req *v1.ListAcqRightsForProductAggregationRequest) (*v1.ListAcqRightsForProductAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsForApplicationsProduct", zap.String("reason", "ScopeError"))
		return nil, status.Error(codes.Unknown, "ScopeValidationError")
	}
	params := &repo.QueryProductAggregations{}
	prodAgg, err := s.licenseRepo.ProductAggregationDetails(ctx, req.ID, params, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Product Aggregation Details-> "+err.Error())
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, req.GetScope())
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
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
		SKU:            strings.Join(skus, ","),
		SwidTag:        strings.Join(swidTags, ","),
		Metric:         prodAgg.Metric,
		NumAcqLicences: acqLicenses,
		TotalCost:      totalCost,
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
	metricInfo := metrics[ind]

	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, req.GetScope())
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	input := make(map[string]interface{})
	input[PROD_AGG_NAME] = prodAgg.Name
	input[METRIC_NAME] = metricInfo.Name
	input[SCOPES] = []string{req.GetScope()}
	input[IS_AGG] = true
	if _, ok := MetricCalculation[metricInfo.Type]; !ok {
		logger.Log.Error("service/v1 -Failed ListAcqRightsForProductAggregation - ", zap.String("Agg name", prodAgg.Name), zap.String("metric name", prodAgg.Metric))
		return nil, status.Error(codes.Internal, "this metricType is not supported")
	}
	resp, err := MetricCalculation[metricInfo.Type](s, ctx, eqTypes, input)
	if err != nil {
		logger.Log.Error("service/v1 -Failed ListAcqRightsForProductAggregation - ", zap.String("Agg name", prodAgg.Name), zap.String("metric name", prodAgg.Metric), zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}
	computedLicenses := resp[COMPUTED_LICENCES].(uint64)
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
