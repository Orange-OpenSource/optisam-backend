// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// productServiceServer is implementation of v1.authServiceServer proto interface
type productServiceServer struct {
	productRepo repo.Product
	queue       workerqueue.Workerqueue
}

// NewProductServiceServer creates Product service
func NewProductServiceServer(productRepo repo.Product, queue workerqueue.Workerqueue) v1.ProductServiceServer {
	return &productServiceServer{productRepo: productRepo, queue: queue}
}

func (s *productServiceServer) UpsertProduct(ctx context.Context, req *v1.UpsertProductRequest) (*v1.UpsertProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	err := s.productRepo.UpsertProductTx(ctx, req, userClaims.UserID)
	if err != nil {
		logger.Log.Error("UpsertProduct Failed", zap.Error(err))
		return &v1.UpsertProductResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertProductRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.UpsertProductResponse{Success: true}, nil
}

func (s *productServiceServer) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var apiresp *v1.ListProductsResponse
	var err error
	if req.GetSearchParams().GetApplicationId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInApplication(ctx, req, req.Scopes)
	} else if req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInEquipment(ctx, req, req.Scopes)
	} else {
		apiresp, err = s.listProductView(ctx, req, req.Scopes)
	}
	if err != nil {
		return nil, err
	}
	return apiresp, nil
}

func (s *productServiceServer) listProductView(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
	dbresp, err := s.productRepo.ListProductsView(ctx, db.ListProductsViewParams{
		Scope:                 scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:               strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:              strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductView - db/ListProductsView", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].Category = dbresp[i].ProductCategory
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost

	}
	return &apiresp, nil
}

func (s *productServiceServer) listProductViewInApplication(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {

	dbresp, err := s.productRepo.ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
		Scope:                 scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey(),
		IsApplicationID:       req.GetSearchParams().GetApplicationId().GetFilterType() && req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
		EquipmentID:           req.GetSearchParams().GetEquipmentId().GetFilteringkey(),
		IsEquipmentID:         req.GetSearchParams().GetEquipmentId().GetFilterType() && req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "",
		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:               strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:              strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductViewInApplication - db/ListProductsViewRedirectedApplication", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].Category = dbresp[i].ProductCategory
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost

	}
	return &apiresp, nil
}

func (s *productServiceServer) listProductViewInEquipment(ctx context.Context, req *v1.ListProductsRequest, scopes []string) (*v1.ListProductsResponse, error) {
	dbresp, err := s.productRepo.ListProductsViewRedirectedEquipment(ctx, db.ListProductsViewRedirectedEquipmentParams{
		Scope:                 scopes,
		Swidtag:               req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:             req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:             !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:           req.GetSearchParams().GetName().GetFilteringkey(),
		IsProductName:         req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkProductName:         !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:         req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:       req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:       !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		ApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey(),
		IsApplicationID:       req.GetSearchParams().GetApplicationId().GetFilterType() && req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
		EquipmentID:           req.GetSearchParams().GetEquipmentId().GetFilteringkey(),
		IsEquipmentID:         req.GetSearchParams().GetEquipmentId().GetFilterType() && req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "",
		ProductNameAsc:        strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:       strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:            strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:           strings.Contains(req.GetSortBy(), "swidtag") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:     strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:    strings.Contains(req.GetSortBy(), "version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditionAsc:     strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditionDesc:    strings.Contains(req.GetSortBy(), "edition") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductCategoryAsc:    strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductCategoryDesc:   strings.Contains(req.GetSortBy(), "category") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:      strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:     strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfApplicationsAsc:  strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfApplicationsDesc: strings.Contains(req.GetSortBy(), "numOfApplications") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy(), "numofEquipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:               strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:              strings.Contains(req.GetSortBy(), "totalCost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - listProductViewInEquipment - db/ListProductsViewRedirectedEquipment", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ListProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Edition = dbresp[i].ProductEdition
		apiresp.Products[i].Editor = dbresp[i].ProductEditor
		apiresp.Products[i].Version = dbresp[i].ProductVersion
		apiresp.Products[i].Category = dbresp[i].ProductCategory
		apiresp.Products[i].NumOfApplications = dbresp[i].NumOfApplications
		apiresp.Products[i].NumofEquipments = dbresp[i].NumOfEquipments
		apiresp.Products[i].TotalCost = dbresp[i].Cost

	}
	return &apiresp, nil
}

func (s *productServiceServer) GetProductDetail(ctx context.Context, req *v1.ProductRequest) (*v1.ProductResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetProductInformation(ctx, db.GetProductInformationParams{
		Swidtag: req.SwidTag,
		Scope:   req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetProductDetail - db/GetProductInformation", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ProductResponse{}
	apiresp.SwidTag = dbresp.Swidtag
	apiresp.Edition = dbresp.ProductEdition
	apiresp.Release = dbresp.ProductVersion
	apiresp.Editor = dbresp.ProductEditor
	return &apiresp, nil

}

func (s *productServiceServer) GetProductOptions(ctx context.Context, req *v1.ProductRequest) (*v1.ProductOptionsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetProductOptions(ctx, db.GetProductOptionsParams{
		Swidtag: req.GetSwidTag(),
		Scope:   req.Scopes,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetProductOptions - db/GetProductOptions", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	apiresp := v1.ProductOptionsResponse{}
	apiresp.Optioninfo = make([]*v1.OptionInfo, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.NumOfOptions = int32(len(dbresp))
	}
	for i := range dbresp {
		apiresp.Optioninfo[i] = &v1.OptionInfo{}
		apiresp.Optioninfo[i].SwidTag = dbresp[i].Swidtag
		apiresp.Optioninfo[i].Name = dbresp[i].ProductName
		apiresp.Optioninfo[i].Edition = dbresp[i].ProductEdition
		apiresp.Optioninfo[i].Version = dbresp[i].ProductVersion
		apiresp.Optioninfo[i].Editor = dbresp[i].ProductEditor
	}
	return &apiresp, nil
}

func (s *productServiceServer) DropProductData(ctx context.Context, req *v1.DropProductDataRequest) (*v1.DropProductDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.productRepo.DropProductDataTx(ctx, req.Scope); err != nil {
		return &v1.DropProductDataResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DropProductDataRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.DropProductDataResponse{Success: true}, nil
}
