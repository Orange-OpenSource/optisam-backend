package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *ProductServiceServer) ListAggregationProducts(ctx context.Context, req *v1.ListAggregationProductsRequest) (*v1.ListAggregationProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregationProducts", zap.String("reason", "ScopeError"))
		return &v1.ListAggregationProductsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	availProds, err := s.ProductRepo.ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
		Editor: req.GetEditor(),
		Scope:  req.GetScope(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregationProductsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregationProducts - ListProductsForAggregation", zap.String("reason", err.Error()))
		return &v1.ListAggregationProductsResponse{}, status.Error(codes.Internal, "DBError")
	}

	var selectedProds []db.ListSelectedProductsForAggregrationRow
	if req.ID != 0 {
		selectedProds, err = s.ProductRepo.ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
			ID:     req.ID,
			Scope:  req.Scope,
			Editor: req.Editor,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return &v1.ListAggregationProductsResponse{
					AggrightsProducts: dbAggProductsToSrvAggProductsAll(availProds),
				}, nil
			}
			logger.Log.Error("service/v1 - ListAggregationProducts - ListSelectedProductsForAggregration", zap.String("reason", err.Error()))
			return &v1.ListAggregationProductsResponse{}, status.Error(codes.Internal, "DBError")
		}
	}
	return &v1.ListAggregationProductsResponse{
		AggrightsProducts: dbAggProductsToSrvAggProductsAll(availProds),
		SelectedProducts:  dbSelectedProductsToSrvSelectedProductsAll(selectedProds),
	}, nil
}

func (s *ProductServiceServer) ListAggregationEditors(ctx context.Context, req *v1.ListAggregationEditorsRequest) (*v1.ListAggregationEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	scopes := strings.Split(req.Scope, ",")
	if !helper.Contains(userClaims.Socpes, scopes...) {
		logger.Log.Error("service/v1 - ListAggregationEditors", zap.String("reason", "ScopeError"))
		return &v1.ListAggregationEditorsResponse{}, status.Error(codes.Internal, "ScopeValidationError")
	}

	dbresp, err := s.ProductRepo.ListEditorsForAggregation(ctx, scopes)
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregationEditorsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregationEditors - ListEditorsForAggregation", zap.String("reason", err.Error()))
		return &v1.ListAggregationEditorsResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListAggregationEditorsResponse{
		Editor: dbresp,
	}, nil
}

func (s *ProductServiceServer) CreateAggregation(ctx context.Context, req *v1.Aggregation) (*v1.AggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		logger.Log.Error("service/v1 - CreateAggregation ", zap.String("reason", "ClaimsError"))
		return &v1.AggregationResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregationResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.ProductRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateAggregation - GetAggregationByName", zap.String("reason", err.Error()))
			return &v1.AggregationResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		return &v1.AggregationResponse{}, status.Error(codes.InvalidArgument, "aggregation name already exists")
	}
	err = s.validateAggregation(ctx, req)
	if err != nil {
		logger.Log.Error("service/v1 - CreateAggregation - validateAggregation", zap.String("reason", err.Error()))
		return nil, err
	}
	aggid, inerr := s.ProductRepo.InsertAggregation(ctx, db.InsertAggregationParams{
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
		ProductEditor:   req.ProductEditor,
		Products:        req.ProductNames,
		Swidtags:        req.Swidtags,
		CreatedBy:       userClaims.UserID,
	})
	if inerr != nil {
		logger.Log.Error("service/v1 - CreateAggregation - InsertAggregation", zap.String("reason", inerr.Error()))
		return &v1.AggregationResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightsWorkerJob(ctx, &dgworker.UpsertAggregationRequest{
		ID:            aggid,
		Name:          req.AggregationName,
		Scope:         req.Scope,
		ProductEditor: req.ProductEditor,
		Products:      req.ProductNames,
		Swidtags:      req.Swidtags,
	})
	return &v1.AggregationResponse{Success: true}, nil
}

func (s *ProductServiceServer) ListAggregations(ctx context.Context, req *v1.ListAggregationsRequest) (*v1.ListAggregationsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregation ", zap.String("reason", "ScopeError"))
		return &v1.ListAggregationsResponse{}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.ListAggregations(ctx, db.ListAggregationsParams{
		IsAggName:         !req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		LsAggName:         req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		AggregationName:   req.GetSearchParams().GetAggregationName().GetFilteringkey(),
		IsProductEditor:   !req.GetSearchParams().GetProductEditor().GetFilterType() && req.GetSearchParams().GetProductEditor().GetFilteringkey() != "",
		LkProductEditor:   req.GetSearchParams().GetProductEditor().GetFilterType() && req.GetSearchParams().GetProductEditor().GetFilteringkey() != "",
		ProductEditor:     req.GetSearchParams().GetProductEditor().GetFilteringkey(),
		AggNameAsc:        strings.Contains(req.GetSortBy().String(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggNameDesc:       strings.Contains(req.GetSortBy().String(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:  strings.Contains(req.GetSortBy().String(), "product_editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc: strings.Contains(req.GetSortBy().String(), "product_editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
		Scope:             req.Scope,
		PageNum:           req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:          req.GetPageSize(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregationsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregation - ListAggregation", zap.String("reason", err.Error()))
		return &v1.ListAggregationsResponse{}, status.Error(codes.Internal, "DBError")
	}
	var total int32
	if len(dbresp) > 0 {
		total = int32(dbresp[0].Totalrecords)
	}
	return &v1.ListAggregationsResponse{
		TotalRecords: total,
		Aggregations: dbAggregationsToSrvAggregationsAll(dbresp),
	}, nil
}

func (s *ProductServiceServer) UpdateAggregation(ctx context.Context, req *v1.Aggregation) (*v1.AggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregationResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregationResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
		ID:    req.ID,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpdateAggregation - GetAggregationByID", zap.String("reason", err.Error()))
			return &v1.AggregationResponse{}, status.Error(codes.Internal, "DBError")
		}
		return &v1.AggregationResponse{}, status.Error(codes.InvalidArgument, "aggregation does not exist")
	}
	err = s.validateAggregation(ctx, req)
	if err != nil {
		return &v1.AggregationResponse{}, err
	}
	uperr := s.ProductRepo.UpdateAggregation(ctx, db.UpdateAggregationParams{
		ID:              req.ID,
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
		ProductEditor:   req.ProductEditor,
		ProductNames:    req.ProductNames,
		Swidtags:        req.Swidtags,
		UpdatedBy:       sql.NullString{String: userClaims.UserID, Valid: true},
	})
	if uperr != nil {
		logger.Log.Error("service/v1 - UpdateAggregation - UpdateAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregationResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightsWorkerJob(ctx, &dgworker.UpsertAggregationRequest{
		ID:            req.ID,
		Name:          req.AggregationName,
		Scope:         req.Scope,
		ProductEditor: req.ProductEditor,
		Products:      req.ProductNames,
		Swidtags:      req.Swidtags,
	})
	return &v1.AggregationResponse{Success: true}, nil
}

func (s *ProductServiceServer) DeleteAggregation(ctx context.Context, req *v1.DeleteAggregationRequest) (*v1.AggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregationResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - DeleteAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregationResponse{Success: false}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, er := s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
		ID:    req.ID,
		Scope: req.Scope,
	})
	if er != nil {
		if er != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpdateAggregation - GetAggregationByID", zap.String("reason", er.Error()))
			return &v1.AggregationResponse{}, status.Error(codes.Internal, "DBError")
		}
		return &v1.AggregationResponse{}, status.Error(codes.InvalidArgument, "aggregation does not exist")
	}
	if err := s.ProductRepo.DeleteAggregation(ctx, db.DeleteAggregationParams{
		ID:    req.ID,
		Scope: req.Scope,
	}); err != nil {
		logger.Log.Error("service/v1 - DeleteAggregation - DeleteAggregation", zap.String("reason", err.Error()))
		return &v1.AggregationResponse{Success: false}, status.Error(codes.Internal, "DBError")
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

	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	return &v1.AggregationResponse{Success: true}, nil
}

// nolint: maligned, gocyclo, funlen
func (s *ProductServiceServer) validateAggregation(ctx context.Context, req *v1.Aggregation) error {
	availProds, err := s.ProductRepo.ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
		Editor: req.ProductEditor,
		Scope:  req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - validateAggregation - ListProductsForAggregation", zap.String("reason", err.Error()))
			return status.Error(codes.Internal, "DBError")
		}
	}
	if req.ID != 0 {
		selectedProds, err := s.ProductRepo.ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
			ID:     req.ID,
			Scope:  req.Scope,
			Editor: req.ProductEditor,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return status.Error(codes.Internal, "unable to get selected products")
			}
			logger.Log.Error("service/v1 - validateAggregation - ListSelectedProductsForAggregration", zap.String("reason", err.Error()))
			return status.Error(codes.Internal, "DBError")
		}
		if !selectedProductExists(availProds, selectedProds, req.Swidtags) {
			logger.Log.Error("service/v1 - validateAggregation", zap.String("reason", err.Error()))
			return status.Error(codes.InvalidArgument, "ProductNotAvailable")
		}
	} else if len(availProds) == 0 || !availableProductExists(availProds, req.Swidtags) {
		return status.Error(codes.InvalidArgument, "ProductNotAvailable")
	}
	return nil
}

func (s *ProductServiceServer) pushUpsertAggrightsWorkerJob(ctx context.Context, req *dgworker.UpsertAggregationRequest) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertAggregation, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	// log.Println(string(envolveData))
	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
}

func availableProductExists(products []db.ListProductsForAggregationRow, reqSwid []string) bool {
	for _, rs := range reqSwid {
		flag := false
		for _, prod := range products {
			if rs == prod.Swidtag {
				flag = true
			}
		}
		if !flag {
			return false
		}
	}
	return true
}

func selectedProductExists(availproducts []db.ListProductsForAggregationRow, selectproducts []db.ListSelectedProductsForAggregrationRow, reqSwid []string) bool {
	for _, rs := range reqSwid {
		flag := false
		for _, prod := range availproducts {
			if rs == prod.Swidtag {
				flag = true
			}
		}
		if !flag {
			for _, prod := range selectproducts {
				if rs == prod.Swidtag {
					flag = true
				}
			}
			if !flag {
				return false
			}
		}
	}
	return true
}

func dbAggregationsToSrvAggregationsAll(aggregations []db.ListAggregationsRow) []*v1.Aggregation {
	servAggregation := make([]*v1.Aggregation, 0, len(aggregations))
	for _, agg := range aggregations {
		servAggregation = append(servAggregation, dbAggregationToSrvAggregation(agg))
	}
	return servAggregation
}

func dbAggregationToSrvAggregation(aggregation db.ListAggregationsRow) *v1.Aggregation {
	var mapping = aggregation.Coalesce.([]byte)
	var tmp []*v1.Mapping
	json.Unmarshal(mapping, &tmp)
	resp := &v1.Aggregation{
		ID:              aggregation.ID,
		AggregationName: aggregation.AggregationName,
		ProductEditor:   aggregation.ProductEditor,
		ProductNames:    aggregation.Products,
		Swidtags:        aggregation.Swidtags,
		Scope:           aggregation.Scope,
		EditorId:        aggregation.EditorID.String,
		Mapping:         tmp,
	}
	return resp
}

func dbAggProductsToSrvAggProductsAll(aggprods []db.ListProductsForAggregationRow) []*v1.AggregationProducts {
	servAggProds := make([]*v1.AggregationProducts, 0, len(aggprods))
	aggp := make(map[string]bool, len(aggprods))
	for _, aggprod := range aggprods {
		if aggprod.ProductVersion != "" {
			aggp[aggprod.ProductName] = true // for not empty product version
		}
		servAggProds = append(servAggProds, dbAggProductsToSrvAggProducts(aggprod))
	}
	return servAggProds
}

func dbAggProductsToSrvAggProducts(aggprod db.ListProductsForAggregationRow) *v1.AggregationProducts {
	return &v1.AggregationProducts{
		Swidtag:        aggprod.Swidtag,
		ProductName:    aggprod.ProductName,
		Editor:         aggprod.ProductEditor,
		ProductVersion: fmt.Sprintf("%v", aggprod.ProductVersion),
	}
}

func dbSelectedProductsToSrvSelectedProductsAll(selectedProds []db.ListSelectedProductsForAggregrationRow) []*v1.AggregationProducts {
	servSelectProds := make([]*v1.AggregationProducts, 0, len(selectedProds))
	for _, selectedProd := range selectedProds {
		servSelectProds = append(servSelectProds, dbSelectedProductsToSrvSelectedProducts(selectedProd))
	}
	return servSelectProds
}

func dbSelectedProductsToSrvSelectedProducts(selectedProd db.ListSelectedProductsForAggregrationRow) *v1.AggregationProducts {
	return &v1.AggregationProducts{
		Swidtag:        selectedProd.Swidtag,
		ProductName:    selectedProd.ProductName,
		Editor:         selectedProd.ProductEditor,
		ProductVersion: selectedProd.ProductVersion,
	}
}

func (s *ProductServiceServer) GetAggregationById(ctx context.Context, req *v1.GetAggregationByIdRequest) (*v1.GetAggregationByIdResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregationById", zap.String("reason", "ScopeError"))
		return &v1.GetAggregationByIdResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	aggr, err := s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
		ID:    req.GetAggregationId(),
		Scope: req.GetScope(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.GetAggregationByIdResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregationById - GetAggregationByID", zap.String("reason", err.Error()))
		return &v1.GetAggregationByIdResponse{}, status.Error(codes.Internal, "DBError")
	}

	return &v1.GetAggregationByIdResponse{
		Id:              aggr.ID,
		AggregationName: aggr.AggregationName,
		Scope:           aggr.Scope,
		ProductEditor:   aggr.ProductEditor,
		Products:        aggr.Products,
		Swidtags:        aggr.Swidtags,
		CreatedOn:       timestamppb.New(aggr.CreatedOn),
		CreatedBy:       aggr.CreatedBy,
		UpdatedOn:       timestamppb.New(aggr.UpdatedOn.Time),
		UpdatedBy:       aggr.UpdatedBy.String,
	}, nil
}
