package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductServiceServer) CreateAggregatedRights(ctx context.Context, req *v1.AggregatedRightsRequest) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateProductAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.ProductRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
			return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
		} else {
			_, err := s.ProductRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
				AcqrightSku: req.Sku,
				Scope:       req.Scope,
			})
			if err != nil {
				if err != sql.ErrNoRows {
					logger.Log.Error("service/v1 - CreateProductAggregation - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
					return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
				}
			} else {
				return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku already exists")
			}
		}
	} else {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "sku already exists")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRight(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	uperr := s.ProductRepo.UpsertAggregatedRights(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

// nolint: maligned, gocyclo, funlen
func (s *ProductServiceServer) validateAggregatedRight(ctx context.Context, req *v1.AggregatedRightsRequest) (db.UpsertAggregatedRightsParams, *dgworker.UpsertAggregatedRight, error) {
	_, err := s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
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
		// if err := s.metricCheckForProcessorAndNup(ctx, metrics.Metrices, idx, req.AggregationID, "", req.Scope); err != nil {
		// 	return db.UpsertAggregatedRightsParams{}, nil, err
		// }
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
	supNum := strings.Split(req.GetSupportNumber(), ",")
	for _, snum := range supNum {
		if len(snum) > 16 {
			logger.Log.Sugar().Errorf("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", "Support Number %v is greater than 16 characters"))
			return db.UpsertAggregatedRightsParams{}, nil, status.Error(codes.InvalidArgument, "Support Number is greater than 16 characters")
		}
	}
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
			SupportNumbers:            supNum,
			MaintenanceProvider:       req.MaintenanceProvider,
			FileName:                  req.FileName,
			FileData:                  req.FileData,
			Repartition:               req.Repartition,
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
			Repartition:               req.Repartition,
		}, nil
}

func (s *ProductServiceServer) UpdateAggregatedRights(ctx context.Context, req *v1.AggregatedRightsRequest) (*v1.AggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAggregation ", zap.String("reason", "ScopeError"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	aggRight, err := s.ProductRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
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
	resp, err := s.GetAvailableLicenses(ctx, &v1.GetAvailableLicensesRequest{Sku: req.Sku, Scope: req.Scope})
	if err != nil {
		logger.Log.Error("service/v1 - UpdateAggregatedRights - GetAvailableLicenses", zap.String("reason", err.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Internal, "DBError")
	}
	if req.NumLicensesAcquired < resp.TotalSharedLicenses {
		logger.Log.Error("service/v1 - UpdateAggregatedRights - GetAvailableLicenses", zap.String("reason", "AcquiredLicences less than sharedLicences"))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.InvalidArgument, "AcquiredLicences less than sharedLicences")
	}
	dbAggRight, upsertAggRight, err := s.validateAggregatedRight(ctx, req)
	if err != nil {
		return &v1.AggregatedRightsResponse{}, err
	}
	dbAggRight.CreatedBy = userClaims.UserID
	uperr := s.ProductRepo.UpsertAggregatedRights(ctx, dbAggRight)
	if uperr != nil {
		logger.Log.Error("service/v1 - UpdateAggregation - UpsertAggregation", zap.String("reason", uperr.Error()))
		return &v1.AggregatedRightsResponse{}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAggrightWorkerJob(ctx, *upsertAggRight)
	return &v1.AggregatedRightsResponse{Success: true}, nil
}

func (s *ProductServiceServer) UpdateAggrightsSharedLicenses(ctx context.Context, req *v1.UpdateAggrightsSharedLicensesRequest) (*v1.UpdateSharedLicensesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses ", zap.String("reason", "ScopeError"))
		return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	resp, err := s.GetAvailableLicenses(ctx, &v1.GetAvailableLicensesRequest{Sku: req.Sku, Scope: req.Scope})
	if err != nil {
		logger.Log.Error("service/v1 -UpdateAggrightsSharedLicenses - GetAvailableLicenses", zap.String("reason", err.Error()))
		return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "DBError")
	}
	senderSku, err := s.ProductRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
			return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "DBError")
		}
		return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.InvalidArgument, "aggregation does not exist")
	}
	totalAvailLic := 0
	for _, v := range req.LicenseData {
		if v.SharedLicenses == 0 {
			for _, i := range resp.SharedData {
				if v.RecieverScope == i.Scope {
					totalAvailLic += int(i.SharedLicenses)
				}
			}
		}
	}
	licenses := 0
	for _, v := range req.LicenseData {
		licenses = licenses + int(v.SharedLicenses)
	}
	totalAvailLic += int(resp.AvailableLicenses)
	licenses = licenses - int(resp.TotalSharedLicenses)
	if int32(licenses) > int32(totalAvailLic) {
		return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.InvalidArgument, "LicencesNotAvailable")
	}
	for _, v := range req.LicenseData {
		_, err := s.ProductRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
			Sku:   req.Sku,
			Scope: v.RecieverScope,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = s.CreateMetricIfNotExists(ctx, req.Scope, v.RecieverScope, senderSku.Metric)
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - CreateMetricIfNotExists", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{}, err
				}
				err = s.CreateAggregationIfNotExists(ctx, req.Scope, v.RecieverScope, req.AggregationName)
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - CreateAggregationIfNotExists", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{}, err
				}
				resp, err := s.ProductRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: req.AggregationName,
					Scope:           v.RecieverScope,
				})
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - GetAggregationByName", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "DBError")
				}
				unitPrice, _ := senderSku.AvgUnitPrice.Float64()
				maintenanceUnitPrice, _ := senderSku.AvgMaintenanceUnitPrice.Float64()
				startOfMaintenance := senderSku.StartOfMaintenance.Time.Format(time.RFC3339)
				endOfMaintenance := senderSku.EndOfMaintenance.Time.Format(time.RFC3339)
				if senderSku.NumLicencesMaintenance == 0 {
					startOfMaintenance = ""
					endOfMaintenance = ""
				}
				orderingDate := senderSku.OrderingDate.Time.Format(time.RFC3339)
				if orderingDate == "0001-01-01T00:00:00Z" {
					orderingDate = ""
				}
				_, err = s.CreateAggregatedRights(ctx, &v1.AggregatedRightsRequest{
					Sku:                       senderSku.Sku,
					AggregationID:             resp.ID,
					MetricName:                senderSku.Metric,
					NumLicensesAcquired:       0,
					AvgUnitPrice:              unitPrice,
					StartOfMaintenance:        startOfMaintenance,
					EndOfMaintenance:          endOfMaintenance,
					NumLicencesMaintenance:    senderSku.NumLicencesMaintenance,
					LastPurchasedOrder:        senderSku.LastPurchasedOrder,
					SupportNumber:             strings.Join(senderSku.SupportNumbers, ","),
					MaintenanceProvider:       senderSku.MaintenanceProvider,
					Comment:                   senderSku.Comment.String,
					OrderingDate:              orderingDate,
					CorporateSourcingContract: senderSku.CorporateSourcingContract,
					SoftwareProvider:          senderSku.SoftwareProvider,
					FileName:                  senderSku.FileName,
					AvgMaintenanceUnitPrice:   maintenanceUnitPrice,
					Scope:                     v.RecieverScope,
				})
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - CreateAggregatedRights", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "InternalError")
				}
			} else {
				logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - GetAggregatedRightBySKU", zap.String("reason", err.Error()))
				return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "DBError")
			}
		}
		if dbresp := s.ProductRepo.UpsertSharedLicenses(ctx, db.UpsertSharedLicensesParams{
			Sku:            req.Sku,
			Scope:          req.Scope,
			SharingScope:   v.RecieverScope,
			SharedLicences: int32(v.SharedLicenses),
		}); dbresp != nil {
			logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - UpsertSharedLicenses", zap.String("reason", dbresp.Error()))
			return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
		}
		if dbresp := s.ProductRepo.UpsertRecievedLicenses(ctx, db.UpsertRecievedLicensesParams{
			Sku:              req.Sku,
			Scope:            v.RecieverScope,
			SharingScope:     req.Scope,
			RecievedLicences: int32(v.SharedLicenses),
		}); dbresp != nil {
			logger.Log.Error("service/v1 - UpdateAggrightsSharedLicenses - UpsertRecievedLicenses", zap.String("reason", dbresp.Error()))
			return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
		}
	}
	return &v1.UpdateSharedLicensesResponse{Success: true}, nil
}

func (s *ProductServiceServer) CreateAggregationIfNotExists(ctx context.Context, senderScope string, recieverScope string, aggName string) error {
	_, err := s.ProductRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
		AggregationName: aggName,
		Scope:           recieverScope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			agg, err := s.ProductRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
				AggregationName: aggName,
				Scope:           senderScope,
			})
			if err != nil {
				logger.Log.Error("service/v1 - CreateAggregationIfNotExists - GetAggregationByName", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "DBError")
			}
			for i := range agg.Swidtags {
				err = s.CreateProduct(ctx, agg.Swidtags[i], senderScope, recieverScope)
				if err != nil {
					logger.Log.Error("service/v1 - CreateAggregationIfNotExists - CreateProduct", zap.String("reason", err.Error()))
					return status.Error(codes.Internal, "InternalError")
				}
			}
			resp, err := s.CreateAggregation(ctx, &v1.Aggregation{
				ID:              0,
				AggregationName: aggName,
				ProductEditor:   agg.ProductEditor,
				ProductNames:    agg.Products,
				Swidtags:        agg.Swidtags,
				Scope:           recieverScope,
			})
			if err != nil {
				logger.Log.Error("service/v1 - CreateAggregationIfNotExists - CreateAggregation", zap.String("reason", err.Error()))
				return err
			}
			if !resp.Success {
				logger.Log.Error("service/v1 - CreateAggregationIfNotExists - CreateAggregation", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "InternalError")
			}
		} else {
			logger.Log.Error("service/v1 - CreateAggregationIfNotExists - GetAggregationByName", zap.String("reason", err.Error()))
			return status.Error(codes.Internal, "DBError")
		}
	}
	return nil
}

func (s *ProductServiceServer) CreateProduct(ctx context.Context, swidtag string, senderScope string, recieverScope string) error {
	_, err := s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{
		Swidtag: swidtag,
		Scope:   recieverScope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			product, err := s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{
				Swidtag: swidtag,
				Scope:   senderScope,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					err = s.CreateAcqrights(ctx, swidtag, senderScope, recieverScope)
					if err != nil {
						logger.Log.Error("service/v1 - CreateProduct - CreateAcqrights", zap.String("reason", err.Error()))
						return status.Error(codes.Internal, "AcqRightNotCreated")
					}
					return nil
				} else {
					logger.Log.Error("service/v1 - CreateProduct - GetProductInformation", zap.String("reason", err.Error()))
					return status.Error(codes.NotFound, "ProductNotFound")
				}
			}
			resp, err := s.UpsertProduct(ctx, &v1.UpsertProductRequest{
				SwidTag: swidtag,
				Name:    product.ProductName,
				Editor:  product.ProductEditor,
				Version: product.ProductVersion,
				Scope:   recieverScope,
			})
			if err != nil {
				logger.Log.Error("service/v1 - CreateProduct - UpsertProduct", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "ProductNotCreated")
			}
			if !resp.Success {
				logger.Log.Error("service/v1 - CreateProduct - UpsertProduct", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "ProductNotCreated")
			}
		} else {
			logger.Log.Error("service/v1 - CreateProduct - GetProductInformation", zap.String("reason", err.Error()))
			return status.Error(codes.Internal, "DBError")
		}
	}
	return nil
}

func (s *ProductServiceServer) CreateAcqrights(ctx context.Context, swidtag string, senderScope string, recieverScope string) error {
	_, err := s.ProductRepo.GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
		Swidtag: swidtag,
		Scope:   recieverScope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			product, err := s.ProductRepo.GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
				Swidtag: swidtag,
				Scope:   senderScope,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return nil
				}
				logger.Log.Error("service/v1 - CreateAcqrights - GetAcqBySwidtag", zap.String("reason", err.Error()))
				return status.Error(codes.NotFound, "AcRightNotFound")
			}
			err = s.CreateMetricIfNotExists(ctx, senderScope, recieverScope, product.Metric)
			if err != nil {
				logger.Log.Error("service/v1 - CreateAcqrights - CreateMetricIfNotExists", zap.String("reason", err.Error()))
				return err
			}
			resp, err := s.UpsertAcqRights(ctx, &v1.UpsertAcqRightsRequest{
				Sku:                 product.Sku,
				Swidtag:             swidtag,
				ProductName:         product.ProductName,
				ProductEditor:       product.ProductEditor,
				Version:             product.Version,
				MetricType:          product.Metric,
				NumLicensesAcquired: 0,
				Scope:               recieverScope,
			})
			if err != nil {
				logger.Log.Error("service/v1 - CreateAcqrights - UpsertAcqRights", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "AcqrightNotCreated")
			}
			if !resp.Success {
				logger.Log.Error("service/v1 - CreateAcqrights - UpsertProduct", zap.String("reason", err.Error()))
				return status.Error(codes.Internal, "AcqrightNotCreated")
			}
		} else {
			logger.Log.Error("service/v1 - CreateAcqrights - GetAcqBySwidtag", zap.String("reason", err.Error()))
			return status.Error(codes.Internal, "DBError")
		}
	}
	return nil
}

func (s *ProductServiceServer) DeleteAggregatedRights(ctx context.Context, req *v1.DeleteAggregatedRightsRequest) (*v1.DeleteAggregatedRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteAggregatedRightsResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DeleteAggregatedRightsResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.ProductRepo.DeleteAggregatedRightBySKU(ctx, db.DeleteAggregatedRightBySKUParams{
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

func (s *ProductServiceServer) pushUpsertAggrightWorkerJob(ctx context.Context, req dgworker.UpsertAggregatedRight) {
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

func (s *ProductServiceServer) DownloadAggregatedRightsFile(ctx context.Context, req *v1.DownloadAggregatedRightsFileRequest) (*v1.DownloadAggregatedRightsFileResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Sugar().Errorf("service/v1 - DownloadAggregatedRightsFile - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return &v1.DownloadAggregatedRightsFileResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	acq, err := s.ProductRepo.GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
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
	acqFileData, err := s.ProductRepo.GetAggregatedRightsFileDataBySKU(ctx, db.GetAggregatedRightsFileDataBySKUParams{
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
