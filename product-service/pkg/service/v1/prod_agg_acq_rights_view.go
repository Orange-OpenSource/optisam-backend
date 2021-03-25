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
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (lr *productServiceServer) ListAcqRightsAggregation(ctx context.Context, req *v1.ListAcqRightsAggregationRequest) (*v1.ListAcqRightsAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	dbresp, err := lr.productRepo.ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
		Scope:               req.Scopes,
		Sku:                 req.GetSearchParams().GetSKU().GetFilteringkey(),
		IsSku:               req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		LkSku:               !req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		Swidtag:             req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:           req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:           !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		AggregationName:     req.GetSearchParams().GetName().GetFilteringkey(),
		IsAggregationName:   req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkAggregationName:   !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		Metric:              req.GetSearchParams().GetMetric().GetFilteringkey(),
		IsMetric:            req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		LkMetric:            !req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		ProductEditor:       req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:     req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:     !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		AggregationNameAsc:  strings.Contains(req.GetSortBy().String(), "NAME") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggregationNameDesc: strings.Contains(req.GetSortBy().String(), "NAME") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:    strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:   strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "desc"),
		MetricAsc:           strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		MetricDesc:          strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalCostAsc:        strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalCostDesc:       strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregationAcqRightsView - ListAcqRightsAggregation", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListAcqRightsAggregationResponse{}
	apiresp.Aggregations = make([]*v1.AcqRightsAggregation, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Aggregations[i] = &v1.AcqRightsAggregation{}
		apiresp.Aggregations[i].ID = dbresp[i].AggregationID
		apiresp.Aggregations[i].Name = dbresp[i].AggregationName
		apiresp.Aggregations[i].Skus = dbresp[i].Skus
		apiresp.Aggregations[i].Swidtags = dbresp[i].Swidtags
		apiresp.Aggregations[i].Metric = dbresp[i].Metric
		apiresp.Aggregations[i].Editor = dbresp[i].ProductEditor
		apiresp.Aggregations[i].TotalCost, _ = dbresp[i].TotalCost.Float64()
	}

	return &apiresp, nil

}

func (lr *productServiceServer) ListAcqRightsAggregationRecords(ctx context.Context, req *v1.ListAcqRightsAggregationRecordsRequest) (*v1.ListAcqRightsAggregationRecordsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	dbresp, err := lr.productRepo.ListAcqRightsAggregationIndividual(ctx, db.ListAcqRightsAggregationIndividualParams{
		Scope:         req.Scopes,
		AggregationID: req.AggregationId,
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsAggregationRecords - ListAcqRightsAggregationIndividual", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, err.Error())
	}

	apiresp := v1.ListAcqRightsAggregationRecordsResponse{}
	apiresp.AcquiredRights = make([]*v1.AcqRights, len(dbresp))

	for i := range dbresp {
		apiresp.AcquiredRights[i] = &v1.AcqRights{}
		apiresp.AcquiredRights[i].SwidTag = dbresp[i].Swidtag
		apiresp.AcquiredRights[i].ProductName = dbresp[i].ProductName
		apiresp.AcquiredRights[i].Metric = dbresp[i].Metric
		apiresp.AcquiredRights[i].Editor = dbresp[i].ProductEditor
		apiresp.AcquiredRights[i].Entity = dbresp[i].Entity
		apiresp.AcquiredRights[i].SKU = dbresp[i].Sku
		apiresp.AcquiredRights[i].AcquiredLicensesNumber = dbresp[i].NumLicensesAcquired
		apiresp.AcquiredRights[i].LicensesUnderMaintenanceNumber = dbresp[i].NumLicencesMaintainance
		apiresp.AcquiredRights[i].AvgLicenesUnitPrice, _ = dbresp[i].AvgUnitPrice.Float64()
		apiresp.AcquiredRights[i].AvgMaintenanceUnitPrice, _ = dbresp[i].AvgMaintenanceUnitPrice.Float64()
		apiresp.AcquiredRights[i].TotalPurchaseCost, _ = dbresp[i].TotalPurchaseCost.Float64()
		apiresp.AcquiredRights[i].TotalMaintenanceCost, _ = dbresp[i].TotalMaintenanceCost.Float64()
		apiresp.AcquiredRights[i].TotalCost, _ = dbresp[i].TotalCost.Float64()
		apiresp.AcquiredRights[i].Version = dbresp[i].Version
	}

	return &apiresp, nil
}
