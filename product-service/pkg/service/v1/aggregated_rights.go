package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServiceServer) CreateAggregatedRights(ctx context.Context, req *v1.AggregatedRightsRequest) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.productRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku already exists")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRight(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	uperr := s.productRepo.UpsertAggregatedRights(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

// nolint: maligned, gocyclo, funlen
func (s *productServiceServer) validateAggregatedRight(ctx context.Context, req *v1.AggregatedRightsRequest) (db.UpsertAggregatedRightsParams, *dgworker.UpsertAggregatedRight, error) {
	_, err := s.productRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
		ID:    req.AggregationID,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregationByName", zap.String("reason", err.Error()))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.Internal, "DBError")
		}
		return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "aggregation does not exists")
	}
	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - validateAggregatedRights - ListMetrices", zap.String("reason", err.Error()))
		return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.Internal, "ServiceError")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
	}
	for _, met := range strings.Split(req.MetricName, ",") {
		idx := metricExists(metrics.Metrices, met)
		if idx == -1 {
			logger.Log.Error("service/v1 - validateAggregatedRights - metric does not exist", zap.String("metric:", met))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
		}
		if err := s.metricCheckForProcessorAndNup(ctx, metrics.Metrices, idx, req.AggregationID, "", req.Scope); err != nil {
			return db.UpsertAggregatedRightsParams{}, nil, err
		}
	}
	var totalPurchaseCost, totalMaintenanceCost float64
	totalPurchaseCost = req.AvgUnitPrice * float64(req.NumLicensesAcquired)
	var startOfMaintenance, endOfMaintenance, orderingDate sql.NullTime
	var startTime, endTime, orderingTime time.Time
	var err1, err2 error

	if req.OrderingDate != "" {
		var err error
		if strings.Contains(req.OrderingDate, "/") && len(req.OrderingDate) <= 8 {
			orderingTime, err = time.Parse("1/2/06", req.OrderingDate)
		} else if len(req.OrderingDate) == 10 {
			orderingTime, err = time.Parse("02-01-2006", req.OrderingDate)
		} else {
			orderingTime, err = time.Parse(time.RFC3339Nano, req.OrderingDate)
		}
		if err != nil {
			logger.Log.Error("service/v1 - validateAggregation - unable to parse ordering time", zap.String("reason", err.Error()))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse ordering time")
		}
		orderingDate = sql.NullTime{Time: orderingTime, Valid: true}
	}
	if req.NumLicencesMaintenance == 0 && req.StartOfMaintenance == "" && req.EndOfMaintenance == "" {
		// do nothing
	} else if req.NumLicencesMaintenance != 0 && req.StartOfMaintenance != "" && req.EndOfMaintenance != "" {
		if strings.Contains(req.StartOfMaintenance, "/") && len(req.StartOfMaintenance) <= 8 {
			startTime, err1 = time.Parse("1/2/06", req.StartOfMaintenance)
		} else if len(req.StartOfMaintenance) == 10 {
			startTime, err1 = time.Parse("02-01-2006", req.StartOfMaintenance)
		} else {
			startTime, err1 = time.Parse(time.RFC3339Nano, req.StartOfMaintenance)
		}
		startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}
		if err1 != nil {
			logger.Log.Error("service/v1 - validateAggregation - unable to parse start time", zap.String("reason", err1.Error()))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse start time")
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
			logger.Log.Error("service/v1 - validateAggregation - unable to parse end time", zap.String("reason", err2.Error()))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse end time")
		}
		if !endTime.After(startTime) {
			logger.Log.Error("service/v1 - validateAggregation", zap.String("reason", "maintenance end time must be greater than maintenance start time"))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "end time is less than start time")
		}
	} else {
		return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "all or no fields should be present( maintenance licenses, start date, end date)")
	}
	totalMaintenanceCost = req.AvgMaintenanceUnitPrice * float64(req.NumLicencesMaintenance)
	return db.UpsertAggregatedRightsParams{
			Sku:                       req.Sku,
			AggregationID:             req.AggregationID,
			Scope:                     req.Scope,
			Metric:                    req.MetricName,
			NumLicensesAcquired:       req.NumLicensesAcquired,
			AvgUnitPrice:              decimal.NewFromFloat(req.AvgUnitPrice),
			AvgMaintenanceUnitPrice:   decimal.NewFromFloat(req.AvgMaintenanceUnitPrice),
			TotalPurchaseCost:         decimal.NewFromFloat(totalPurchaseCost),
			TotalMaintenanceCost:      decimal.NewFromFloat(totalMaintenanceCost),
			TotalCost:                 decimal.NewFromFloat(totalPurchaseCost + totalMaintenanceCost),
			StartOfMaintenance:        startOfMaintenance,
			EndOfMaintenance:          endOfMaintenance,
			NumLicencesMaintenance:    req.NumLicencesMaintenance,
			Comment:                   sql.NullString{String: req.Comment, Valid: true},
			OrderingDate:              orderingDate,
			CorporateSourcingContract: req.CorporateSourcingContract,
			SoftwareProvider:          req.SoftwareProvider,
			LastPurchasedOrder:        req.LastPurchasedOrder,
			SupportNumber:             req.SupportNumber,
			MaintenanceProvider:       req.MaintenanceProvider,
			FileName:                  req.FileName,
			FileData:                  req.FileData,
		}, &dgworker.UpsertAggregatedRight{
			Sku:                       req.Sku,
			AggregationID:             req.AggregationID,
			Metric:                    req.MetricName,
			NumLicensesAcquired:       req.NumLicensesAcquired,
			AvgUnitPrice:              req.AvgUnitPrice,
			AvgMaintenanceUnitPrice:   req.AvgMaintenanceUnitPrice,
			TotalPurchaseCost:         totalPurchaseCost,
			TotalMaintenanceCost:      totalMaintenanceCost,
			TotalCost:                 (totalPurchaseCost + totalMaintenanceCost),
			Scope:                     req.Scope,
			StartOfMaintenance:        req.StartOfMaintenance,
			EndOfMaintenance:          req.EndOfMaintenance,
			NumLicencesMaintenance:    req.NumLicencesMaintenance,
			OrderingDate:              req.OrderingDate,
			CorporateSourcingContract: req.CorporateSourcingContract,
			SoftwareProvider:          req.SoftwareProvider,
			LastPurchasedOrder:        req.LastPurchasedOrder,
			SupportNumber:             req.SupportNumber,
			MaintenanceProvider:       req.MaintenanceProvider,
		}, nil
}

func (s *productServiceServer) UpdateAggregatedRights(ctx context.Context, req *v1.AggregatedRightsRequest) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	aggRight, err := s.productRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpdateAggregatedRights - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		}
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "aggregation does not exist")
	}
	if aggRight.Sku != req.Sku {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku cannot be updated")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRight(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	uperr := s.productRepo.UpsertAggregatedRights(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - UpdateAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

func (s *productServiceServer) DeleteAggregatedRights(ctx context.Context, req *v1.DeleteAggregatedRightsRequest) (*v1.DeleteAggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteAggregatedRightsResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DeleteAggregatedRightsResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.productRepo.DeleteAggregatedRightBySKU(ctx, db.DeleteAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	}); err != nil {
		return &v1.DeleteAggregatedRightsResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(dgworker.DeleteAggregatedRightRequest{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DeleteAggregatedRights, JSON: jsonData}

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
	return &v1.DeleteAggregatedRightsResponse{
		Success: true,
	}, nil
}

func (s *productServiceServer) pushUpsertAggrightWorkerJob(ctx context.Context, req dgworker.UpsertAggregatedRight) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertAggregatedRights, JSON: jsonData}

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

func (s *productServiceServer) DownloadAggregatedRightsFile(ctx context.Context, req *v1.DownloadAggregatedRightsFileRequest) (*v1.DownloadAggregatedRightsFileResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Sugar().Errorf("service/v1 - DownloadAggregatedRightsFile - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	acq, err := s.productRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.InvalidArgument, "SKU does not exist")
		}
		logger.Log.Error("service/v1 - DownloadAggregatedRightsFile - GetAcqRightBySKU", zap.String("reason", err.Error()))
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.Internal, "DBError")
	}
	if acq.FileName == "" {
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.InvalidArgument, "Aggregated Right does not contain file")
	}
	acqFileData, err := s.productRepo.GetAggregatedRightsFileDataBySKU(ctx, db.GetAggregatedRightsFileDataBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - DownloadAcqRightFile - GetAggregatedRightsFileDataBySKU", zap.String("reason", err.Error()))
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.DownloadAggregatedRightsFileResponse{
		FileData: acqFileData,
	}, nil
}
