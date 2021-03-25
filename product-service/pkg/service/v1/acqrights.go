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
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (lr *productServiceServer) UpsertAcqRights(ctx context.Context, req *v1.UpsertAcqRightsRequest) (*v1.UpsertAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	startOfMaintenance := sql.NullTime{Valid: false}
	endOfMaintenance := sql.NullTime{Valid: false}
	startTime, err1 := time.Parse(time.RFC3339Nano, req.StartOfMaintenance)
	endTime, err2 := time.Parse(time.RFC3339Nano, req.EndOfMaintenance)
	if err1 == nil {
		startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}
	}
	if err2 == nil {
		endOfMaintenance = sql.NullTime{Time: endTime, Valid: true}
	}

	if err1 == nil && err2 == nil && !endTime.After(startTime) {
		logger.Log.Error("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", "maintenance end time must be greater than maintenance start time"))
		return &v1.UpsertAcqRightsResponse{Success: false}, status.Error(codes.Unknown, "end time is less than start time")
	}
	err := lr.productRepo.UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
		Sku:                     req.GetSku(),
		Swidtag:                 req.GetSwidtag(),
		ProductName:             req.GetProductName(),
		ProductEditor:           req.GetProductEditor(),
		Metric:                  req.GetMetricType(),
		NumLicensesAcquired:     req.GetNumLicensesAcquired(),
		NumLicencesMaintainance: req.GetNumLicencesMaintainance(),
		AvgUnitPrice:            decimal.NewFromFloat(req.GetAvgUnitPrice()),
		AvgMaintenanceUnitPrice: decimal.NewFromFloat(req.GetAvgMaintenanceUnitPrice()),
		TotalPurchaseCost:       decimal.NewFromFloat(req.GetTotalPurchaseCost()),
		TotalMaintenanceCost:    decimal.NewFromFloat(req.GetTotalMaintenanceCost()),
		TotalCost:               decimal.NewFromFloat(req.GetTotalCost()),
		Entity:                  req.GetEntity(),
		Scope:                   req.GetScope(),
		StartOfMaintenance:      startOfMaintenance,
		EndOfMaintenance:        endOfMaintenance,
		Version:                 req.GetVersion(),
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
	e := dgworker.Envelope{Type: dgworker.UpsertAcqRightsRequest, JSON: jsonData}

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

func (lr *productServiceServer) ListAcqRights(ctx context.Context, req *v1.ListAcqRightsRequest) (*v1.ListAcqRightsResponse, error) {

	// ctx, span := trace.StartSpan(ctx, "Service Layer")
	// defer span.End()
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	//log.Println("SCOPES ", userClaims.Socpes)

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		logger.Log.Sugar().Infof("acrights-service - ListAcqRights - user don't have access to the scopes: %v, requested scopes: %v", userClaims.Socpes, req.Scopes)
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := lr.productRepo.ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
		Scope:                       req.Scopes,
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
		NumLicensesAcquiredAsc:      strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES_NUMBER") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicensesAcquiredDesc:     strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES_NUMBER") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumLicencesMaintainanceAsc:  strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE_NUMBER") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicencesMaintainanceDesc: strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE_NUMBER") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalPurchaseCostAsc:        strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASE_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalPurchaseCostDesc:       strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASE_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalMaintenanceCostAsc:     strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalMaintenanceCostDesc:    strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalCostAsc:                strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalCostDesc:               strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		StartOfMaintenanceAsc:       strings.Contains(req.GetSortBy().String(), "START_OF_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		StartOfMaintenanceDesc:      strings.Contains(req.GetSortBy().String(), "START_OF_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		EndOfMaintenanceAsc:         strings.Contains(req.GetSortBy().String(), "END_OF_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		EndOfMaintenanceDesc:        strings.Contains(req.GetSortBy().String(), "END_OF_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRights - ListAcqRightsIndividual", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.ListAcqRightsResponse{}
	apiresp.AcquiredRights = make([]*v1.AcqRights, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.AcquiredRights[i] = &v1.AcqRights{}
		apiresp.AcquiredRights[i].Version = dbresp[i].Version
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
		if dbresp[i].StartOfMaintenance.Valid {
			apiresp.AcquiredRights[i].StartOfMaintenance, _ = ptypes.TimestampProto(dbresp[i].StartOfMaintenance.Time)
		}
		apiresp.AcquiredRights[i].LicensesUnderMaintenance = "yes"
		if dbresp[i].EndOfMaintenance.Valid {
			apiresp.AcquiredRights[i].EndOfMaintenance, _ = ptypes.TimestampProto(dbresp[i].EndOfMaintenance.Time)
			if !dbresp[i].EndOfMaintenance.Time.After(time.Now()) {
				apiresp.AcquiredRights[i].LicensesUnderMaintenance = "no"
			}
		}
	}

	return &apiresp, nil
}

func (lr *productServiceServer) ListAcqRightsProducts(ctx context.Context, req *v1.ListAcqRightsProductsRequest) (*v1.ListAcqRightsProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsProducts", zap.String("reason", "ScopeError"))
		return &v1.ListAcqRightsProductsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.productRepo.ListAcqRightsProducts(ctx, db.ListAcqRightsProductsParams{
		Editor: req.GetEditor(),
		Metric: req.GetMetric(),
		Scope:  req.GetScope(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsProducts - ListAcqRightsProducts", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsProductsResponse{}, status.Error(codes.Internal, "DBError")
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
func (lr *productServiceServer) ListAcqRightsEditors(ctx context.Context, req *v1.ListAcqRightsEditorsRequest) (*v1.ListAcqRightsEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsEditors", zap.String("reason", "ScopeError"))
		return &v1.ListAcqRightsEditorsResponse{}, status.Error(codes.Internal, "ScopeValidationError")
	}
	dbresp, err := lr.productRepo.ListAcqRightsEditors(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsEditors - ListAcqRightsEditors", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsEditorsResponse{}, status.Error(codes.Internal, "DBError")
	}
	apiresp := &v1.ListAcqRightsEditorsResponse{}
	apiresp.Editor = make([]string, len(dbresp))
	for i := range dbresp {
		apiresp.Editor[i] = dbresp[i]
	}
	return apiresp, nil
}

func (lr *productServiceServer) ListAcqRightsMetrics(ctx context.Context, req *v1.ListAcqRightsMetricsRequest) (*v1.ListAcqRightsMetricsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAcqRightsMetrics", zap.String("reason", "ScopeValidationError"))
		return &v1.ListAcqRightsMetricsResponse{}, status.Error(codes.Internal, "ScopeValidationError")
	}
	dbresp, err := lr.productRepo.ListAcqRightsMetrics(ctx, req.GetScope())
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsMetrics - ListAcqRightsMetrics", zap.String("reason", err.Error()))
		return &v1.ListAcqRightsMetricsResponse{}, status.Error(codes.Internal, "DBError")
	}
	apiresp := &v1.ListAcqRightsMetricsResponse{}
	apiresp.Metric = make([]string, len(dbresp))
	for i := range dbresp {
		apiresp.Metric[i] = dbresp[i]
	}
	return apiresp, nil
}

func (lr *productServiceServer) CreateProductAggregation(ctx context.Context, req *v1.ProductAggregationMessage) (*v1.ProductAggregationMessage, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.productRepo.InsertAggregation(ctx, db.InsertAggregationParams{
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
		Scope:    dbresp.AggregationScope,
	}

	//For rpc worker Queue
	//lr.rpcCalls(ctx, "product", dbresp.AggregationID, dbresp.AggregationName, dbresp.Products, req.GetScope(), "add")

	// RPC to Method Change
	err = lr.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
		AggregationID:   dbresp.AggregationID,
		AggregationName: dbresp.AggregationName,
		Swidtags:        dbresp.Products,
		//SCOPE BASED CHANGE
		Scope: req.GetScope(),
	})

	if err != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - UpsertProductAggregation", zap.String("reason", err.Error()))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(apiresp)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertAggregation, JSON: jsonData}

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

func (lr *productServiceServer) ListProductAggregation(ctx context.Context, req *v1.ListProductAggregationRequest) (*v1.ListProductAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	dbresp, err := lr.productRepo.ListAggregation(ctx, req.Scopes)
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

func (lr *productServiceServer) UpdateProductAggregation(ctx context.Context, req *v1.ProductAggregationMessage) (*v1.ProductAggregationMessage, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	dbresp, err := lr.productRepo.UpdateAggregation(ctx, db.UpdateAggregationParams{
		Scope:           req.Scope,
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
	//lr.rpcCalls(ctx, "product", dbresp.AggregationID, dbresp.AggregationName, dbresp.Products, req.GetScope(), "upsert")

	// RPC TO METHOD
	aggregation, err := lr.productRepo.GetProductAggregation(ctx, db.GetProductAggregationParams{
		AggregationID:   dbresp.AggregationID,
		AggregationName: dbresp.AggregationName})
	if err != nil {
		logger.Log.Error("service/v1 - UpdateProductAggregation - GetProductAggregation", zap.String("reason", err.Error()))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
	}

	var delIds []string
	ids := make(map[string]int)
	for _, id := range aggregation {
		ids[id] = 0
	}
	for _, id := range dbresp.Products {
		ids[id] = 1
	}
	for id, isDel := range ids {
		if isDel == 0 {
			delIds = append(delIds, id)
		}
	}
	if len(delIds) > 0 {

		err = lr.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
			AggregationID:   0,
			AggregationName: "",
			Swidtags:        delIds})

		if err != nil {
			logger.Log.Error("service/v1 - UpdateProductAggregation - UpsertProductAggregation", zap.String("reason", err.Error()))
			return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
		}
	}

	err = lr.productRepo.UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
		AggregationID:   dbresp.AggregationID,
		AggregationName: dbresp.AggregationName,
		//SCOPE BASED CHANGE
		Scope:    req.Scope,
		Swidtags: dbresp.Products})

	if err != nil {
		logger.Log.Error("service/v1 - UpdateProductAggregation - UpsertProductAggregation", zap.String("reason", err.Error()))
		return &v1.ProductAggregationMessage{}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertAggregation, JSON: jsonData}

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

func (lr *productServiceServer) DeleteProductAggregation(ctx context.Context, req *v1.DeleteProductAggregationRequest) (*v1.DeleteProductAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	err := lr.productRepo.DeleteAggregation(ctx, db.DeleteAggregationParams{AggregationID: req.GetID(), Scope: userClaims.Socpes})
	if err != nil {
		logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteAggregation", zap.String("reason", err.Error()))
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}

	//For rpcWorker
	//lr.rpcCalls(ctx, "product", req.GetID(), "", []string{}, req.GetScope(), "delete")

	//RPC TO METHOD CALL
	err = lr.productRepo.DeleteProductAggregation(ctx, db.DeleteProductAggregationParams{AggregationID_2: req.GetID()})
	if err != nil {
		logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteProductAggregation", zap.String("reason", err.Error()))
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DeleteAggregation, JSON: jsonData}

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
