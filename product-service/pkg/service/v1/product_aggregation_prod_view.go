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

func (s *productServiceServer) ListProductAggregationView(ctx context.Context, req *v1.ListProductAggregationViewRequest) (*v1.ListProductAggregationViewResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := s.productRepo.ListAggregationsView(ctx, db.ListAggregationsViewParams{
		Scope:                 req.Scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		AggregationName:       req.GetSearchParams().GetName().GetFilteringkey(),
		IsAggregationName:     req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkAggregationName:     !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		AggregationNameAsc:    strings.Contains(req.GetSortBy().String(), "application_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggregationNameDesc:   strings.Contains(req.GetSortBy().String(), "application_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy().String(), "application_owner") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy().String(), "application_owner") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:               strings.Contains(req.GetSortBy().String(), "cost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:              strings.Contains(req.GetSortBy().String(), "cost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregationView - db/ListAggregationsView", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductAggregationViewResponse{}
	apiresp.Aggregations = make([]*v1.ProductAggregationView, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Aggregations[i] = &v1.ProductAggregationView{}
		apiresp.Aggregations[i].ID = dbresp[i].AggregationID
		apiresp.Aggregations[i].Name = dbresp[i].AggregationName
		apiresp.Aggregations[i].Editor = dbresp[i].ProductEditor
		apiresp.Aggregations[i].NumApplications = dbresp[i].NumOfApplications
		apiresp.Aggregations[i].NumEquipments = dbresp[i].NumOfEquipments
		apiresp.Aggregations[i].TotalCost = dbresp[i].TotalCost
		apiresp.Aggregations[i].Swidtags = dbresp[i].Swidtags
	}
	return &apiresp, nil
}

func (s *productServiceServer) ListProductAggregationProductView(ctx context.Context, req *v1.ListProductAggregationProductViewRequest) (*v1.ListProductAggregationProductViewResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.ListAggregationProductsView(ctx, db.ListAggregationProductsViewParams{
		AggregationID: req.GetID(),
		Scope:         req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregationProductView - db/ListAggregationProductsView", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ListProductAggregationProductViewResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))
	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost
	}
	return &apiresp, nil

}

func (s *productServiceServer) ProductAggregationProductViewDetails(ctx context.Context, req *v1.ProductAggregationProductViewDetailsRequest) (*v1.ProductAggregationProductViewDetailsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := s.productRepo.ProductAggregationDetails(ctx, db.ProductAggregationDetailsParams{
		AggregationID: req.GetID(),
		Scope:         req.Scopes,
	})

	if err != nil {
		logger.Log.Error("service/v1 - ProductAggregationProductViewDetails - db/ProductAggregationDetails", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ProductAggregationProductViewDetailsResponse{
		ID:              dbresp.AggregationID,
		Name:            dbresp.AggregationName,
		Editor:          dbresp.ProductEditor,
		Products:        dbresp.Swidtags,
		Editions:        dbresp.Editions,
		NumApplications: dbresp.NumOfApplications,
		NumEquipments:   dbresp.NumOfEquipments,
	}

	return &apiresp, nil
}

func (s *productServiceServer) ProductAggregationProductViewOptions(ctx context.Context, req *v1.ProductAggregationProductViewOptionsRequest) (*v1.ProductAggregationProductViewOptionsResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := s.productRepo.ProductAggregationChildOptions(ctx, db.ProductAggregationChildOptionsParams{
		AggregationID: req.GetID(),
		Scope:         req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - ProductAggregationProductViewOptions - db/ProductAggregationChildOptions", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ProductAggregationProductViewOptionsResponse{}
	apiresp.Optioninfo = make([]*v1.OptionInfo, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.NumOfOptions = int32(len(dbresp))
	}
	for i := range dbresp {
		apiresp.Optioninfo[i] = &v1.OptionInfo{}
		apiresp.Optioninfo[i].SwidTag = dbresp[i].Swidtag
		apiresp.Optioninfo[i].Name = dbresp[i].ProductName
		apiresp.Optioninfo[i].Edition = dbresp[i].ProductEdition
		apiresp.Optioninfo[i].Editor = dbresp[i].ProductEditor
		apiresp.Optioninfo[i].Version = dbresp[i].ProductVersion
	}
	return &apiresp, nil
}
