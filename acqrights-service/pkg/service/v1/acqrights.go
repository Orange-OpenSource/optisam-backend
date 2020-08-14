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
	"log"
	v1 "optisam-backend/acqrights-service/pkg/api/v1"
	repo "optisam-backend/acqrights-service/pkg/repository/v1"
	"optisam-backend/acqrights-service/pkg/repository/v1/postgres/db"
	"optisam-backend/acqrights-service/pkg/rpc"
	"optisam-backend/acqrights-service/pkg/worker"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// acqRightsServiceServer is implementation of v1.authServiceServer proto interface
type acqRightsServiceServer struct {
	acqRightsRepo repo.AcqRights
	queue         workerqueue.Workerqueue
}

// NewAcquiredRightsServiceServer creates License service
func NewAcqRightsServiceServer(acqRightsRepo repo.AcqRights, queue workerqueue.Workerqueue) v1.AcqRightsServiceServer {
	return &acqRightsServiceServer{acqRightsRepo: acqRightsRepo, queue: queue}
}

func (lr *acqRightsServiceServer) UpsertAcqRights(ctx context.Context, req *v1.UpsertAcqRightsRequest) (*v1.UpsertAcqRightsResponse, error) {
	logger.Log.Info("Service", zap.Any("UpsertAcqRights", req))

	err := lr.acqRightsRepo.UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
		Sku:                     req.GetSku(),
		Swidtag:                 req.GetSwidtag(),
		ProductName:             req.GetProductName(),
		ProductEditor:           req.GetProductEditor(),
		Metric:                  req.GetMetricType(),
		NumLicensesAcquired:     req.GetNumLicensesAcquired(),
		NumLicencesMaintainance: req.GetNumLicencesMaintainance(),
		AvgUnitPrice:            req.GetAvgUnitPrice(),
		AvgMaintenanceUnitPrice: req.GetAvgMaintenanceUnitPrice(),
		TotalPurchaseCost:       req.GetTotalPurchaseCost(),
		TotalMaintenanceCost:    req.GetTotalMaintenanceCost(),
		TotalCost:               req.GetTotalCost(),
		Entity:                  req.GetEntity(),
		Scope:                   req.GetScope(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", err.Error()))
		return &v1.UpsertAcqRightsResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.UpsertAcqRightsRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := lr.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))

	return &v1.UpsertAcqRightsResponse{Success: true}, nil
}

func (lr *acqRightsServiceServer) ListAcqRights(ctx context.Context, req *v1.ListAcqRightsRequest) (*v1.ListAcqRightsResponse, error) {

	// ctx, span := trace.StartSpan(ctx, "Service Layer")
	// defer span.End()
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	dbresp, err := lr.acqRightsRepo.ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
		Scope:                       userClaims.Socpes,
		Sku:                         req.GetSearchParams().GetSKU().GetFilteringkey(),
		IsSku:                       req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		LkSku:                       !req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		Swidtag:                     req.GetSearchParams().GetSwidTag().GetFilteringkey(),
		IsSwidtag:                   req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		LkSwidtag:                   !req.GetSearchParams().GetSwidTag().GetFilterType() && req.GetSearchParams().GetSwidTag().GetFilteringkey() != "",
		ProductName:                 req.GetSearchParams().GetProductName().GetFilteringkey(),
		IsProductName:               req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		LkProductName:               !req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		Metric:                      req.GetSearchParams().GetMetric().GetFilteringkey(),
		IsMetric:                    req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		LkMetric:                    !req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		ProductEditor:               req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsProductEditor:             req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkProductEditor:             !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		EntityAsc:                   strings.Contains(req.GetSortBy().String(), "ENTITY") && strings.Contains(req.GetSortOrder().String(), "asc"),
		EntityDesc:                  strings.Contains(req.GetSortBy().String(), "ENTITY") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SkuAsc:                      strings.Contains(req.GetSortBy().String(), "SKU") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SkuDesc:                     strings.Contains(req.GetSortBy().String(), "SKU") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductNameAsc:              strings.Contains(req.GetSortBy().String(), "PRODUCT_NAME") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:             strings.Contains(req.GetSortBy().String(), "PRODUCT_NAME") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SwidtagAsc:                  strings.Contains(req.GetSortBy().String(), "SWID_TAG") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:                 strings.Contains(req.GetSortBy().String(), "SWID_TAG") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:            strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:           strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AvgUnitPriceAsc:             strings.Contains(req.GetSortBy().String(), "AVG_LICENSE_UNIT_PRICE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AvgUnitPriceDesc:            strings.Contains(req.GetSortBy().String(), "AVG_LICENSE_UNIT_PRICE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AvgMaintenanceUnitPriceAsc:  strings.Contains(req.GetSortBy().String(), "AVG_MAINTENANCE_UNIT_PRICE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AvgMaintenanceUnitPriceDesc: strings.Contains(req.GetSortBy().String(), "edAVG_MAINTENANCE_UNIT_PRICEitor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		MetricAsc:                   strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "asc"),
		MetricDesc:                  strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumLicencesMaintainanceAsc:  strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE_NUMBER") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicencesMaintainanceDesc: strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE_NUMBER") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumLicensesAcquiredAsc:      strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES_NUMBER") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicensesAcquiredDesc:     strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES_NUMBER") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalPurchaseCostAsc:        strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASE_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalPurchaseCostDesc:       strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASE_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalMaintenanceCostAsc:     strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalMaintenanceCostDesc:    strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalCostAsc:                strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalCostDesc:               strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRights - ListAcqRightsIndividual", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, err.Error())
	}

	apiresp := v1.ListAcqRightsResponse{}
	apiresp.AcquiredRights = make([]*v1.AcqRights, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

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
		apiresp.AcquiredRights[i].AvgLicenesUnitPrice = dbresp[i].AvgUnitPrice
		apiresp.AcquiredRights[i].AvgMaintenanceUnitPrice = dbresp[i].AvgMaintenanceUnitPrice
		apiresp.AcquiredRights[i].TotalPurchaseCost = dbresp[i].TotalPurchaseCost
		apiresp.AcquiredRights[i].TotalMaintenanceCost = dbresp[i].TotalMaintenanceCost
		apiresp.AcquiredRights[i].TotalCost = dbresp[i].TotalCost
	}

	return &apiresp, nil
}

func (lr *acqRightsServiceServer) ListAcqRightsProducts(ctx context.Context, req *v1.ListAcqRightsProductsRequest) (*v1.ListAcqRightsProductsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsProducts", zap.String("reason", "ScopeError"))
		return &v1.ListAcqRightsProductsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.acqRightsRepo.ListAcqRightsProducts(ctx, db.ListAcqRightsProductsParams{
		Editor: req.GetEditor(),
		Metric: req.GetMetric(),
		Scope:  req.GetScope(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsProducts - ListAcqRightsProducts", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsProductsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ListAcqRightsProductsResponse{}
	apiresp.AcqrightsProducts = make([]*v1.ListAcqRightsProductsResponse_AcqRightsProducts, len(dbresp))
	for i := range dbresp {
		apiresp.AcqrightsProducts[i] = &v1.ListAcqRightsProductsResponse_AcqRightsProducts{}
		apiresp.AcqrightsProducts[i].ProductName = dbresp[i].ProductName
		apiresp.AcqrightsProducts[i].Swidtag = dbresp[i].Swidtag
	}
	return apiresp, nil
}
func (lr *acqRightsServiceServer) ListAcqRightsEditors(ctx context.Context, req *v1.ListAcqRightsEditorsRequest) (*v1.ListAcqRightsEditorsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsEditors", zap.String("reason", "ScopeError"))
		return &v1.ListAcqRightsEditorsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.acqRightsRepo.ListAcqRightsEditors(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsEditors - ListAcqRightsEditors", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsEditorsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ListAcqRightsEditorsResponse{}
	apiresp.Editor = make([]string, len(dbresp))
	for i := range dbresp {
		apiresp.Editor[i] = dbresp[i]
	}
	return apiresp, nil
}

func (lr *acqRightsServiceServer) ListAcqRightsMetrics(ctx context.Context, req *v1.ListAcqRightsMetricsRequest) (*v1.ListAcqRightsMetricsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsMetrics", zap.String("reason", "ScopeValidationError"))
		return &v1.ListAcqRightsMetricsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.acqRightsRepo.ListAcqRightsMetrics(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsMetrics - ListAcqRightsMetrics", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsMetricsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ListAcqRightsMetricsResponse{}
	apiresp.Metric = make([]string, len(dbresp))
	for i := range dbresp {
		apiresp.Metric[i] = dbresp[i]
	}
	return apiresp, nil
}

func (lr *acqRightsServiceServer) CreateProductAggregation(ctx context.Context, req *v1.ProductAggregationMessage) (*v1.ProductAggregationMessage, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateProductAggregation", zap.String("reason", "ScopeError"))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "ScopeError")
	}
	dbresp, err := lr.acqRightsRepo.InsertAggregation(ctx, db.InsertAggregationParams{
		AggregationName:   req.GetName(),
		AggregationScope:  req.GetScope(),
		AggregationMetric: req.GetMetric(),
		Products:          req.GetProducts(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - InsertAggregation", zap.String("reason", err.Error()))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ProductAggregationMessage{
		ID:       dbresp.AggregationID,
		Name:     dbresp.AggregationName,
		Editor:   req.GetEditor(),
		Metric:   dbresp.AggregationMetric,
		Products: dbresp.Products,
	}

	//For rpc worker Queue
	lr.rpcCalls(ctx, "product", dbresp.AggregationID, dbresp.AggregationName, dbresp.Products, "add")

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.UpsertAggregation, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := lr.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))

	return apiresp, nil
}

func (lr *acqRightsServiceServer) ListProductAggregation(ctx context.Context, req *v1.ListProductAggregationRequest) (*v1.ListProductAggregationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	dbresp, err := lr.acqRightsRepo.ListAggregation(ctx, userClaims.Socpes)
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListProductAggregationResponse{}, nil
		} else {
			logger.Log.Error("service/v1 - ListProductAggregation - ListAggregation", zap.String("reason", err.Error()))
			return &v1.ListProductAggregationResponse{}, status.Error(codes.Unknown, "DBError")
		}
	}
	apiresp := &v1.ListProductAggregationResponse{}
	apiresp.Aggregations = make([]*v1.ProductAggregation, len(dbresp))
	for i := range dbresp {
		apiresp.Aggregations[i] = &v1.ProductAggregation{}
		apiresp.Aggregations[i].ID = dbresp[i].AggregationID
		apiresp.Aggregations[i].Name = dbresp[i].AggregationName
		apiresp.Aggregations[i].Metric = dbresp[i].AggregationMetric
		apiresp.Aggregations[i].Editor = dbresp[i].ProductEditor
		apiresp.Aggregations[i].ProductNames = dbresp[i].ProductNames
		apiresp.Aggregations[i].Products = dbresp[i].ProductSwidtags
		apiresp.Aggregations[i].Scope = dbresp[i].AggregationScope
	}
	return apiresp, nil
}

func (lr *acqRightsServiceServer) UpdateProductAggregation(ctx context.Context, req *v1.ProductAggregationMessage) (*v1.ProductAggregationMessage, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	dbresp, err := lr.acqRightsRepo.UpdateAggregation(ctx, db.UpdateAggregationParams{
		Scope:           userClaims.Socpes,
		AggregationID:   req.GetID(),
		AggregationName: req.GetName(),
		Products:        req.GetProducts(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - UpdateProductAggregation - UpdateAggregation", zap.String("reason", err.Error()))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
	}
	apiresp := &v1.ProductAggregationMessage{
		ID:       dbresp.AggregationID,
		Name:     dbresp.AggregationName,
		Editor:   req.GetEditor(),
		Metric:   dbresp.AggregationMetric,
		Products: dbresp.Products,
	}

	//For rpc worker
	lr.rpcCalls(ctx, "product", dbresp.AggregationID, dbresp.AggregationName, dbresp.Products, "upsert")

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.UpsertAggregation, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := lr.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	return apiresp, nil

}

func (lr *acqRightsServiceServer) DeleteProductAggregation(ctx context.Context, req *v1.DeleteProductAggregationRequest) (*v1.DeleteProductAggregationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	err := lr.acqRightsRepo.DeleteAggregation(ctx, db.DeleteAggregationParams{AggregationID: req.GetID(), Scope: userClaims.Socpes})
	if err != nil {
		logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteAggregation", zap.String("reason", err.Error()))
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}

	//For rpcWorker
	lr.rpcCalls(ctx, "product", req.GetID(), "", []string{}, "delete")

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.DeleteAggregation, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := lr.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	return &v1.DeleteProductAggregationResponse{Success: true}, nil

}

func (lr *acqRightsServiceServer) rpcCalls(ctx context.Context, service string, aggId int32, aggName string, swidtags []string, actionType string) {
	envData := rpc.Envelope{Id: aggId,
		Name:       aggName,
		Swidtags:   swidtags,
		ActionType: actionType}
	log.Println("Data sending tp product ", envData)
	dataToPush, err := json.Marshal(envData)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	jobID, err := lr.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "rpc"},
		Status: job.JobStatusPENDING,
		Data:   dataToPush,
	}, "rpc")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully rpc pushed job", zap.Int32("jobId", jobID))
}
