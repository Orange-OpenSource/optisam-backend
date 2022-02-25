package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue/job"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
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

func (s *productServiceServer) ListAggregatedRightsProducts(ctx context.Context, req *v1.ListAggregatedRightsProductsRequest) (*v1.ListAggregatedRightsProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregatedRightsProducts", zap.String("reason", "ScopeError"))
		return &v1.ListAggregatedRightsProductsResponse{}, status.Error(codes.Unknown, "ScopeValidationError")
	}
	availProds, err := s.productRepo.ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
		Editor: req.GetEditor(),
		Metric: req.GetMetric(),
		Scope:  req.GetScope(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregatedRightsProductsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregatedRightsProducts - ListProductsForAggregation", zap.String("reason", err.Error()))
		return &v1.ListAggregatedRightsProductsResponse{}, status.Error(codes.Internal, "DBError")
	}
	var selectedProds []db.ListSelectedProductsForAggregrationRow
	if req.ID != 0 {
		selectedProds, err = s.productRepo.ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
			ID:    req.ID,
			Scope: req.Scope,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return &v1.ListAggregatedRightsProductsResponse{
					AggrightsProducts: dbAggProductsToSrvAggProductsAll(availProds),
				}, nil
			}
			logger.Log.Error("service/v1 - ListAggregatedRightsProducts - ListSelectedProductsForAggregration", zap.String("reason", err.Error()))
			return &v1.ListAggregatedRightsProductsResponse{}, status.Error(codes.Internal, "DBError")
		}
	}
	return &v1.ListAggregatedRightsProductsResponse{
		AggrightsProducts: dbAggProductsToSrvAggProductsAll(availProds),
		SelectedProducts:  dbSelectedProductsToSrvSelectedProductsAll(selectedProds),
	}, nil
}

func (s *productServiceServer) ListAggregatedRightsEditors(ctx context.Context, req *v1.ListAggregatedRightsEditorsRequest) (*v1.ListAggregatedRightsEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregatedRightsEditors", zap.String("reason", "ScopeError"))
		return &v1.ListAggregatedRightsEditorsResponse{}, status.Error(codes.Internal, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.ListEditorsForAggregation(ctx, req.Scope)
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregatedRightsEditorsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregatedRightsEditors - ListEditorsForAggregation", zap.String("reason", err.Error()))
		return &v1.ListAggregatedRightsEditorsResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListAggregatedRightsEditorsResponse{
		Editor: dbresp,
	}, nil
}

func (s *productServiceServer) CreateAggregation(ctx context.Context, req *v1.AggregatedRights) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.productRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregationByName", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "aggregation name already exists")
	}
	_, err = s.productRepo.GetAggregationBySKU(ctx, db.GetAggregationBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregationBySKU", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku already exists")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRights(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	aggid, uperr := s.productRepo.UpsertAggregation(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	upsertAggRight.ID = aggid
	// For Worker Queue
	s.pushUpsertAggrightsWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

func (s *productServiceServer) ListAggregations(ctx context.Context, req *v1.ListAggregationsRequest) (*v1.ListAggregationsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - ListAggregations ", zap.String("reason", "ScopeError"))
		return &v1.ListAggregationsResponse{}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.ListAggregations(ctx, req.Scope)
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListAggregationsResponse{}, nil
		}
		logger.Log.Error("service/v1 - ListAggregations - ListAggregations", zap.String("reason", err.Error()))
		return &v1.ListAggregationsResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListAggregationsResponse{
		Aggregations: dbAggregationsToSrvAggregationsAll(dbresp),
	}, nil
}

func (s *productServiceServer) UpdateAggregation(ctx context.Context, req *v1.AggregatedRights) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	agg, err := s.productRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
		ID:    req.ID,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpdateAggregation - GetAggregationByID", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		}
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "aggregation does not exist")
	}
	if agg.AggregationName != req.AggregationName {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "aggregation name cannot be updated")
	}
	if agg.Sku != req.Sku {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku cannot be updated")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRights(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	aggid, uperr := s.productRepo.UpsertAggregation(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - UpdateAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	upsertAggRight.ID = aggid
	// For Worker Queue
	s.pushUpsertAggrightsWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

func (s *productServiceServer) DeleteProductAggregation(ctx context.Context, req *v1.DeleteProductAggregationRequest) (*v1.DeleteProductAggregationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - DeleteProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	if err := s.productRepo.DeleteAggregation(ctx, db.DeleteAggregationParams{
		AggregationID: req.GetID(),
		Scope:         req.Scope,
	}); err != nil {
		logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteAggregation", zap.String("reason", err.Error()))
		return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// if err := s.productRepo.UpdateAggregationForProduct(ctx, db.UpdateAggregationForProductParams{
	// 	OldAggregationID: req.ID,
	// 	Scope:            req.Scope,
	// }); err != nil {
	// 	logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteProductAggregation", zap.String("reason", err.Error()))
	// 	return &v1.DeleteProductAggregationResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	// }

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
	return &v1.DeleteProductAggregationResponse{Success: true}, nil

}

// nolint: maligned, gocyclo, funlen
func (s *productServiceServer) validateAggregatedRights(ctx context.Context, req *v1.AggregatedRights) (db.UpsertAggregationParams, *dgworker.UpsertAggregatedRightsRequest, error) {
	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - validateAggregatedRights - ListMetrices", zap.String("reason", err.Error()))
		return db.UpsertAggregationParams{}, nil, status.Error(codes.Internal, "ServiceError")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
	}
	for _, met := range strings.Split(req.MetricName, ",") {
		if idx := metricExists(metrics.Metrices, met); idx == -1 {
			logger.Log.Error("service/v1 - validateAggregatedRights - metric does not exist", zap.String("metric:", met))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
		}
	}
	availProds, err := s.productRepo.ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
		Editor: req.ProductEditor,
		Metric: req.MetricName,
		Scope:  req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - validateAggregatedRights - ListProductsForAggregation", zap.String("reason", err.Error()))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.Internal, "DBError")
		}
	}
	if len(availProds) == 0 {
		return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "ProductNotAvailable")
	}

	if req.ID != 0 {
		selectedProds, err := s.productRepo.ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
			ID:    req.ID,
			Scope: req.Scope,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				return db.UpsertAggregationParams{}, nil, status.Error(codes.Internal, "unable to get selected products")
			}
			logger.Log.Error("service/v1 - validateAggregatedRights - ListSelectedProductsForAggregration", zap.String("reason", err.Error()))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.Internal, "DBError")
		}
		if !selectedProductExists(availProds, selectedProds, req.Swidtags) {
			return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "ProductNotAvailable")
		}
	} else if !availableProductExists(availProds, req.Swidtags) {
		return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "ProductNotAvailable")
	}
	var totalPurchaseCost, totalMaintenanceCost float64
	totalPurchaseCost = req.AvgUnitPrice * float64(req.NumLicensesAcquired)
	var startOfMaintenance sql.NullTime
	var endOfMaintenance sql.NullTime
	var startTime, endTime time.Time
	var err1, err2 error
	if req.NumLicencesMaintainance == 0 && req.StartOfMaintenance == "" && req.EndOfMaintenance == "" {
		// do nothing
	} else if req.NumLicencesMaintainance != 0 && req.StartOfMaintenance != "" && req.EndOfMaintenance != "" {
		if strings.Contains(req.StartOfMaintenance, "/") && len(req.StartOfMaintenance) <= 8 {
			startTime, err1 = time.Parse("1/2/06", req.StartOfMaintenance)
		} else if len(req.StartOfMaintenance) == 10 {
			startTime, err1 = time.Parse("02-01-2006", req.StartOfMaintenance)
		} else {
			startTime, err1 = time.Parse(time.RFC3339Nano, req.StartOfMaintenance)
		}
		startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}
		if err1 != nil {
			logger.Log.Error("service/v1 - validateAcqRight - unable to parse start time", zap.String("reason", err1.Error()))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse start time")
		}
		if strings.Contains(req.EndOfMaintenance, "/") && len(req.EndOfMaintenance) <= 8 {
			endTime, err2 = time.Parse("1/2/06", req.EndOfMaintenance)
		} else if len(req.EndOfMaintenance) == 10 {
			endTime, err2 = time.Parse("02-01-2006", req.EndOfMaintenance)
		} else {
			endTime, err2 = time.Parse(time.RFC3339Nano, req.EndOfMaintenance)
		}
		endOfMaintenance = sql.NullTime{Time: endTime, Valid: true}
		if err2 != nil {
			logger.Log.Error("service/v1 - validateAcqRight - unable to parse end time", zap.String("reason", err2.Error()))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse end time")
		}
		if !endTime.After(startTime) {
			logger.Log.Error("service/v1 - validateAcqRight", zap.String("reason", "maintenance end time must be greater than maintenance start time"))
			return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "end time is less than start time")
		}
	} else {
		return db.UpsertAggregationParams{}, nil, status.Error(codes.InvalidArgument, "all or no fields should be present( maintenance licenses, start date, end date)")
	}
	totalMaintenanceCost = req.AvgMaintenanceUnitPrice * float64(req.NumLicencesMaintainance)
	return db.UpsertAggregationParams{
			AggregationName:         req.AggregationName,
			Sku:                     req.Sku,
			Swidtags:                req.Swidtags,
			Products:                req.ProductNames,
			ProductEditor:           req.ProductEditor,
			Scope:                   req.Scope,
			Metric:                  req.MetricName,
			NumLicensesAcquired:     req.NumLicensesAcquired,
			AvgUnitPrice:            decimal.NewFromFloat(req.AvgUnitPrice),
			AvgMaintenanceUnitPrice: decimal.NewFromFloat(req.AvgMaintenanceUnitPrice),
			TotalPurchaseCost:       decimal.NewFromFloat(totalPurchaseCost),
			TotalMaintenanceCost:    decimal.NewFromFloat(totalMaintenanceCost),
			TotalCost:               decimal.NewFromFloat(totalPurchaseCost + totalMaintenanceCost),
			StartOfMaintenance:      startOfMaintenance,
			EndOfMaintenance:        endOfMaintenance,
			NumLicencesMaintainance: req.NumLicencesMaintainance,
			Comment:                 sql.NullString{String: req.Comment, Valid: true},
		}, &dgworker.UpsertAggregatedRightsRequest{
			ID:                      req.ID,
			Name:                    req.AggregationName,
			Sku:                     req.Sku,
			Swidtags:                req.Swidtags,
			Products:                req.ProductNames,
			ProductEditor:           req.ProductEditor,
			Metric:                  req.MetricName,
			NumLicensesAcquired:     req.NumLicensesAcquired,
			AvgUnitPrice:            req.AvgUnitPrice,
			AvgMaintenanceUnitPrice: req.AvgMaintenanceUnitPrice,
			TotalPurchaseCost:       totalPurchaseCost,
			TotalMaintenanceCost:    totalMaintenanceCost,
			TotalCost:               (totalPurchaseCost + totalMaintenanceCost),
			Scope:                   req.Scope,
			StartOfMaintenance:      req.StartOfMaintenance,
			EndOfMaintenance:        req.EndOfMaintenance,
			NumLicencesMaintenance:  req.NumLicencesMaintainance,
		}, nil
}

func (s *productServiceServer) pushUpsertAggrightsWorkerJob(ctx context.Context, req dgworker.UpsertAggregatedRightsRequest) {
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

func dbAggregationsToSrvAggregationsAll(aggregations []db.ListAggregationsRow) []*v1.ListAggregatedRights {
	servAggregations := make([]*v1.ListAggregatedRights, 0, len(aggregations))
	for _, agg := range aggregations {
		servAggregations = append(servAggregations, dbAggregationsToSrvAggregation(agg))
	}
	return servAggregations
}

func dbAggregationsToSrvAggregation(aggregation db.ListAggregationsRow) *v1.ListAggregatedRights {
	avgUnitPrice, _ := aggregation.AvgUnitPrice.Float64()
	avgMaintenanceUnitPrice, _ := aggregation.AvgMaintenanceUnitPrice.Float64()
	resp := &v1.ListAggregatedRights{
		ID:                      aggregation.ID,
		AggregationName:         aggregation.AggregationName,
		Sku:                     aggregation.Sku,
		ProductEditor:           aggregation.ProductEditor,
		MetricName:              aggregation.Metric,
		ProductNames:            aggregation.Products,
		Swidtags:                aggregation.Swidtags,
		NumLicensesAcquired:     aggregation.NumLicensesAcquired,
		AvgUnitPrice:            avgUnitPrice,
		Scope:                   aggregation.Scope,
		NumLicencesMaintainance: aggregation.NumLicencesMaintainance,
		AvgMaintenanceUnitPrice: avgMaintenanceUnitPrice,
		Comment:                 aggregation.Comment.String,
	}
	if aggregation.StartOfMaintenance.Valid {
		resp.StartOfMaintenance, _ = ptypes.TimestampProto(aggregation.StartOfMaintenance.Time)
	}
	if aggregation.EndOfMaintenance.Valid {
		resp.EndOfMaintenance, _ = ptypes.TimestampProto(aggregation.EndOfMaintenance.Time)
	}
	return resp
}

func dbAggProductsToSrvAggProductsAll(aggprods []db.ListProductsForAggregationRow) []*v1.AggregatedRightsProducts {
	servAggProds := make([]*v1.AggregatedRightsProducts, 0, len(aggprods))
	for _, aggprod := range aggprods {
		servAggProds = append(servAggProds, dbAggProductsToSrvAggProducts(aggprod))
	}
	return servAggProds
}

func dbAggProductsToSrvAggProducts(aggprod db.ListProductsForAggregationRow) *v1.AggregatedRightsProducts {
	return &v1.AggregatedRightsProducts{
		Swidtag:     aggprod.Swidtag,
		ProductName: aggprod.ProductName,
		Editor:      aggprod.ProductEditor,
	}
}

func dbSelectedProductsToSrvSelectedProductsAll(selectedProds []db.ListSelectedProductsForAggregrationRow) []*v1.AggregatedRightsProducts {
	servSelectProds := make([]*v1.AggregatedRightsProducts, 0, len(selectedProds))
	for _, selectedProd := range selectedProds {
		servSelectProds = append(servSelectProds, dbSelectedProductsToSrvSelectedProducts(selectedProd))
	}
	return servSelectProds
}

func dbSelectedProductsToSrvSelectedProducts(selectedProd db.ListSelectedProductsForAggregrationRow) *v1.AggregatedRightsProducts {
	return &v1.AggregatedRightsProducts{
		Swidtag:     selectedProd.Swidtag,
		ProductName: selectedProd.ProductName,
		Editor:      selectedProd.ProductEditor,
	}
}
