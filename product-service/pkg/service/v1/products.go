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
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"optisam-backend/product-service/pkg/worker"
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
	logger.Log.Info("Service", zap.Any("UpsertProduct", req))
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
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
	e := worker.Envelope{Type: worker.UpsertProductRequest, Json: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobId, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Succesfully pushed job", zap.Int32("jobId", jobId))
	return &v1.UpsertProductResponse{Success: true}, nil
}

func (s *productServiceServer) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {

	var apiresp *v1.ListProductsResponse
	var err error
	if req.GetSearchParams().GetApplicationId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInApplication(ctx, req)
	} else if req.GetSearchParams().GetEquipmentId().GetFilteringkey() != "" {
		apiresp, err = s.listProductViewInEquipment(ctx, req)
	} else {
		apiresp, err = s.listProductView(ctx, req)
	}
	if err != nil {
		return nil, err
	}
	return apiresp, nil
}

func (s *productServiceServer) listProductView(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.productRepo.ListProductsView(ctx, db.ListProductsViewParams{
		Scope:                 userClaims.Socpes,
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
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
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

func (s *productServiceServer) listProductViewInApplication(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.productRepo.ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
		Scope:                 userClaims.Socpes,
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
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
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

func (s *productServiceServer) listProductViewInEquipment(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.productRepo.ListProductsViewRedirectedEquipment(ctx, db.ListProductsViewRedirectedEquipmentParams{
		Scope:                 userClaims.Socpes,
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
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
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
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.productRepo.GetProductInformation(ctx, db.GetProductInformationParams{
		Swidtag: req.GetSwidTag(),
		Scope:   userClaims.Socpes,
	})
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
	}

	apiresp := v1.ProductResponse{}
	apiresp.SwidTag = dbresp.Swidtag
	apiresp.Edition = dbresp.ProductEdition
	apiresp.Release = dbresp.ProductVersion
	apiresp.Editor = dbresp.ProductEditor
	return &apiresp, nil

}

func (s *productServiceServer) GetProductOptions(ctx context.Context, req *v1.ProductRequest) (*v1.ProductOptionsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := s.productRepo.GetProductOptions(ctx, db.GetProductOptionsParams{
		Swidtag: req.GetSwidTag(),
		Scope:   userClaims.Socpes,
	})
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products Options-> "+err.Error())
	}

	apiresp := v1.ProductOptionsResponse{}
	apiresp.Optioninfo = make([]*v1.OptionInfo, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.NumOfOptions = int32(len(dbresp))
	}
	for i := range dbresp {
		apiresp.Optioninfo[i] = &v1.OptionInfo{}
		apiresp.Optioninfo[i].SwidTag = dbresp[i].Swidtag
		apiresp.Optioninfo[i].Edition = dbresp[i].ProductEdition
		apiresp.Optioninfo[i].Version = dbresp[i].ProductVersion
		apiresp.Optioninfo[i].Editor = dbresp[i].ProductEditor
	}
	return &apiresp, nil
}

func (s *productServiceServer) UpsertProductAggregation(ctx context.Context, req *v1.UpsertAggregationRequest) (*v1.UpsertAggregationResponse, error) {
	/*userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}*/
	var err error
	switch strings.ToLower(req.ActionType) {
	case "add":
		err = s.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
			AggregationID:   req.GetAggregationId(),
			AggregationName: req.GetAggregationName(),
			Swidtags:        req.GetSwidtags(),
		})
	case "delete":

		err = s.productRepo.DeleteProductAggregation(ctx, db.DeleteProductAggregationParams{AggregationID_2: req.GetAggregationId()})

	case "upsert":

		dbresp, err := s.productRepo.GetProductAggregation(ctx, db.GetProductAggregationParams{
			AggregationID:   req.AggregationId,
			AggregationName: req.AggregationName})
		if err != nil {
			return nil, status.Error(codes.Unknown, "failed in upsert aggregation "+err.Error())
		}

		var delIds []string
		ids := make(map[string]int)
		for _, id := range dbresp {
			ids[id] = 0
		}
		for _, id := range req.Swidtags {
			ids[id] = 1
		}
		for id, isDel := range ids {
			if isDel == 0 {
				delIds = append(delIds, id)
			}
		}
		if len(delIds) > 0 {

			err = s.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
				AggregationID:   0,
				AggregationName: "",
				Swidtags:        delIds})

			if err != nil {
				return nil, status.Error(codes.Unknown, "failed in upsert aggregation "+err.Error())
			}
		}

		err = s.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
			AggregationID:   req.AggregationId,
			AggregationName: req.AggregationName,
			Swidtags:        req.Swidtags})

	default:
		return nil, status.Error(codes.Internal, "Undefined action requested ")
	}
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed in upsert aggregation "+err.Error())
	}
	return &v1.UpsertAggregationResponse{Success: true}, nil
}
