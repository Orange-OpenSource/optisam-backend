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

	"github.com/golang/protobuf/ptypes"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	yes string = "yes"
	no  string = "no"
)

func (s *productServiceServer) UpsertAcqRights(ctx context.Context, req *v1.UpsertAcqRightsRequest) (*v1.UpsertAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	startOfMaintenance := sql.NullTime{Valid: false}
	endOfMaintenance := sql.NullTime{Valid: false}

	var startTime, endTime time.Time
	var err1, err2 error

	if req.NumLicencesMaintainance <= 0 {
		req.StartOfMaintenance = ""
		req.EndOfMaintenance = ""
	} else {
		if req.StartOfMaintenance == "" && req.EndOfMaintenance == "" {
			logger.Log.Error("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", "start date and end date can not be empty if maintenance licenses are present"))
			return &v1.UpsertAcqRightsResponse{Success: false}, status.Error(codes.InvalidArgument, "start of maintenance/ end of maintenance is empty but maintenance licenses are present")
		}
		maintenanceStartTime := req.StartOfMaintenance
		maintenanceEndTime := req.EndOfMaintenance

		if len(maintenanceStartTime) <= 10 {
			if strings.Contains(maintenanceStartTime, "/") && len(maintenanceStartTime) <= 8 {
				startTime, err1 = time.Parse("1/2/06", maintenanceStartTime)
			} else if strings.Contains(maintenanceStartTime, "/") {
				startTime, err1 = time.Parse("02/01/2006", maintenanceStartTime)
			} else {
				startTime, err1 = time.Parse("02-01-2006", maintenanceStartTime)
			}
			if err1 != nil {
				logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse start time", zap.String("reason", err1.Error()))
				return nil, status.Error(codes.InvalidArgument, "unable to parse start time ")
			}
		} else {
			startTime, err1 = time.Parse(time.RFC3339Nano, maintenanceStartTime)
			if err1 != nil {
				logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse start time", zap.String("reason", err1.Error()))
				return nil, status.Error(codes.InvalidArgument, "unable to parse start time")
			}
		}
		startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}

		if len(maintenanceEndTime) <= 10 {
			if strings.Contains(maintenanceEndTime, "/") && len(maintenanceEndTime) <= 8 {
				endTime, err2 = time.Parse("1/2/06", maintenanceEndTime)
			} else if strings.Contains(maintenanceEndTime, "/") {
				endTime, err2 = time.Parse("02/01/2006", maintenanceEndTime)
			} else {
				endTime, err2 = time.Parse("02-01-2006", maintenanceEndTime)
			}
			if err2 != nil {
				logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse end time", zap.String("reason", err2.Error()))
				return nil, status.Error(codes.InvalidArgument, "unable to parse end time ")
			}
		} else {
			endTime, err2 = time.Parse(time.RFC3339Nano, maintenanceEndTime)
			if err2 != nil {
				logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse end time", zap.String("reason", err2.Error()))
				return nil, status.Error(codes.InvalidArgument, "unable to parse end time")
			}
		}
		endOfMaintenance = sql.NullTime{Time: endTime, Valid: true}
		if !endTime.After(startTime) {
			logger.Log.Error("service/v1 - UpsertAcqRights", zap.String("reason", "maintenance end time must be greater than maintenance start time"))
			return nil, status.Error(codes.InvalidArgument, "end time is less than start time")
		}
	}
	if err := s.productRepo.UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
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
		Scope:                   req.GetScope(),
		StartOfMaintenance:      startOfMaintenance,
		EndOfMaintenance:        endOfMaintenance,
		Version:                 req.GetVersion(),
		CreatedBy:               userClaims.UserID,
	}); err != nil {
		logger.Log.Error("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", err.Error()))
		return &v1.UpsertAcqRightsResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	s.pushUpsertAcqrightsWorkerJob(ctx, dgworker.UpsertAcqRightsRequest{
		Sku:                     req.Sku,
		Swidtag:                 req.Swidtag,
		ProductName:             req.ProductName,
		ProductEditor:           req.ProductEditor,
		MetricType:              req.MetricType,
		NumLicensesAcquired:     req.NumLicensesAcquired,
		AvgUnitPrice:            req.AvgUnitPrice,
		AvgMaintenanceUnitPrice: req.AvgMaintenanceUnitPrice,
		TotalPurchaseCost:       req.TotalPurchaseCost,
		TotalMaintenanceCost:    req.TotalMaintenanceCost,
		TotalCost:               req.TotalCost,
		Scope:                   req.Scope,
		StartOfMaintenance:      req.StartOfMaintenance,
		EndOfMaintenance:        req.EndOfMaintenance,
		NumLicencesMaintenance:  req.NumLicencesMaintainance,
		Version:                 req.Version,
	})

	return &v1.UpsertAcqRightsResponse{Success: true}, nil
}

// nolint: gocyclo
func (s *productServiceServer) ListAcqRights(ctx context.Context, req *v1.ListAcqRightsRequest) (*v1.ListAcqRightsResponse, error) {

	// ctx, span := trace.StartSpan(ctx, "Service Layer")
	// defer span.End()
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	// log.Println("SCOPES ", userClaims.Socpes)

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		logger.Log.Sugar().Infof("acrights-service - ListAcqRights - user don't have access to the scopes: %v, requested scopes: %v", userClaims.Socpes, req.Scopes)
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := s.productRepo.ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
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
		// API expect pagenum from 1 but the offset in DB starts
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
		apiresp.AcquiredRights[i].SKU = dbresp[i].Sku
		apiresp.AcquiredRights[i].AcquiredLicensesNumber = dbresp[i].NumLicensesAcquired
		apiresp.AcquiredRights[i].LicensesUnderMaintenanceNumber = dbresp[i].NumLicencesMaintainance
		apiresp.AcquiredRights[i].AvgLicenesUnitPrice, _ = dbresp[i].AvgUnitPrice.Float64()
		apiresp.AcquiredRights[i].AvgMaintenanceUnitPrice, _ = dbresp[i].AvgMaintenanceUnitPrice.Float64()
		apiresp.AcquiredRights[i].TotalPurchaseCost, _ = dbresp[i].TotalPurchaseCost.Float64()
		apiresp.AcquiredRights[i].TotalMaintenanceCost, _ = dbresp[i].TotalMaintenanceCost.Float64()
		apiresp.AcquiredRights[i].TotalCost, _ = dbresp[i].TotalCost.Float64()
		apiresp.AcquiredRights[i].Comment = dbresp[i].Comment.String
		if dbresp[i].StartOfMaintenance.Valid {
			apiresp.AcquiredRights[i].StartOfMaintenance, _ = ptypes.TimestampProto(dbresp[i].StartOfMaintenance.Time)
		}
		if dbresp[i].EndOfMaintenance.Valid {
			apiresp.AcquiredRights[i].EndOfMaintenance, _ = ptypes.TimestampProto(dbresp[i].EndOfMaintenance.Time)
			if dbresp[i].EndOfMaintenance.Time.After(time.Now()) {
				apiresp.AcquiredRights[i].LicensesUnderMaintenance = yes
			} else {
				apiresp.AcquiredRights[i].LicensesUnderMaintenance = no
			}
		} else {
			apiresp.AcquiredRights[i].LicensesUnderMaintenance = no
		}
	}

	return &apiresp, nil
}

func (s *productServiceServer) CreateAcqRight(ctx context.Context, req *v1.AcqRightRequest) (*v1.AcqRightResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AcqRightResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateAcqRight ", zap.String("reason", "ScopeError"))
		return &v1.AcqRightResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	_, err := s.productRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
		AcqrightSku: req.Sku,
		Scope:       req.Scope,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - CreateAcqRight - GetAcqRightBySKU", zap.String("reason", err.Error()))
			return &v1.AcqRightResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		return &v1.AcqRightResponse{}, status.Error(codes.InvalidArgument, "SKU already exists")
	}

	dbAcqRight, upsertAcqRight, err := s.validateAcqRight(ctx, userClaims.UserID, req)
	if err != nil {
		return &v1.AcqRightResponse{}, err
	}
	if inserr := s.productRepo.UpsertAcqRights(ctx, dbAcqRight); inserr != nil {
		logger.Log.Error("service/v1 - CreateAcqRight - UpsertAcqRights", zap.String("reason", inserr.Error()))
		return &v1.AcqRightResponse{}, status.Error(codes.Unknown, "DBError")
	}

	// For Worker Queue
	s.pushUpsertAcqrightsWorkerJob(ctx, *upsertAcqRight)
	return &v1.AcqRightResponse{
		Success: true,
	}, nil
}

func (s *productServiceServer) UpdateAcqRight(ctx context.Context, req *v1.AcqRightRequest) (*v1.AcqRightResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.AcqRightResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - CreateAcqRight ", zap.String("reason", "ScopeError"))
		return &v1.AcqRightResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
		AcqrightSku: req.Sku,
		Scope:       req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &v1.AcqRightResponse{}, status.Error(codes.InvalidArgument, "SKU does not exist")
		}
		logger.Log.Error("service/v1 - CreateAcqRight - GetAcqRightBySKU", zap.String("reason", err.Error()))
		return &v1.AcqRightResponse{}, status.Error(codes.Internal, "DBError")
	}
	dbAcqRight, upsertAcqRight, err := s.validateAcqRight(ctx, userClaims.UserID, req)
	if err != nil {
		return &v1.AcqRightResponse{}, err
	}
	if uperr := s.productRepo.UpsertAcqRights(ctx, dbAcqRight); uperr != nil {
		logger.Log.Error("service/v1 - UpdateAcqright - UpsertAcqRights", zap.String("reason", uperr.Error()))
		return &v1.AcqRightResponse{}, status.Error(codes.Unknown, "DBError")
	}
	if dbresp.Swidtag != upsertAcqRight.Swidtag {
		upsertAcqRight.IsSwidtagModified = true
	}
	if dbresp.Metric != upsertAcqRight.MetricType {
		upsertAcqRight.IsMetricModifed = true
	}
	// For Worker Queue
	s.pushUpsertAcqrightsWorkerJob(ctx, *upsertAcqRight)
	return &v1.AcqRightResponse{
		Success: true,
	}, nil
}

func (s *productServiceServer) DeleteAcqRight(ctx context.Context, req *v1.DeleteAcqRightRequest) (*v1.DeleteAcqRightResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteAcqRightResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DeleteAcqRightResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.productRepo.DeleteAcqrightBySKU(ctx, db.DeleteAcqrightBySKUParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	}); err != nil {
		return &v1.DeleteAcqRightResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(dgworker.DeleteAcqRightRequest{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DeleteAcqright, JSON: jsonData}

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
	return &v1.DeleteAcqRightResponse{
		Success: true,
	}, nil
}

// nolint: gocyclo
func (s *productServiceServer) validateAcqRight(ctx context.Context, userID string, req *v1.AcqRightRequest) (db.UpsertAcqRightsParams, *dgworker.UpsertAcqRightsRequest, error) {
	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - validateAcqRight - ListMetrices", zap.String("reason", err.Error()))
		return db.UpsertAcqRightsParams{}, nil, status.Error(codes.Internal, "ServiceError")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
	}
	for _, met := range strings.Split(req.MetricName, ",") {
		if idx := metricExists(metrics.Metrices, met); idx == -1 {
			logger.Log.Error("service/v1 - validateAcqRight - metric does not exist", zap.String("metric:", met))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
		}
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
		maintenanceStartTime := req.StartOfMaintenance
		maintenanceEndTime := req.EndOfMaintenance
		if strings.Contains(maintenanceStartTime, "/") && len(maintenanceStartTime) <= 8 {
			startTime, err1 = time.Parse("1/2/06", maintenanceStartTime)
		} else if len(maintenanceStartTime) == 10 {
			startTime, err1 = time.Parse("02-01-2006", maintenanceStartTime)
		} else {
			startTime, err1 = time.Parse(time.RFC3339Nano, maintenanceStartTime)
		}
		startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}
		if err1 != nil {
			logger.Log.Error("service/v1 - validateAcqRight - unable to parse start time", zap.String("reason", err1.Error()))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse start time")
		}
		if strings.Contains(maintenanceEndTime, "/") && len(maintenanceEndTime) <= 8 {
			endTime, err2 = time.Parse("1/2/06", maintenanceEndTime)
		} else if len(maintenanceEndTime) == 10 {
			endTime, err2 = time.Parse("02-01-2006", maintenanceEndTime)
		} else {
			endTime, err2 = time.Parse(time.RFC3339Nano, maintenanceEndTime)
		}
		endOfMaintenance = sql.NullTime{Time: endTime, Valid: true}
		if err2 != nil {
			logger.Log.Error("service/v1 - validateAcqRight - unable to parse end time", zap.String("reason", err2.Error()))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse end time")
		}
		if !endTime.After(startTime) {
			logger.Log.Error("service/v1 - validateAcqRight", zap.String("reason", "maintenance end time must be greater than maintenance start time"))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "end time is less than start time")
		}
	} else {
		return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "all or no fields should be present( maintenance licenses, start date, end date)")
	}
	totalMaintenanceCost = req.AvgMaintenanceUnitPrice * float64(req.NumLicencesMaintainance)
	swidtag := strings.ReplaceAll(strings.Join([]string{req.ProductName, req.ProductEditor, req.Version}, "_"), " ", "_")
	return db.UpsertAcqRightsParams{
			Sku:                     req.Sku,
			Swidtag:                 swidtag,
			ProductName:             req.ProductName,
			ProductEditor:           req.ProductEditor,
			Scope:                   req.Scope,
			Metric:                  req.MetricName,
			NumLicensesAcquired:     req.NumLicensesAcquired,
			AvgUnitPrice:            decimal.NewFromFloat(req.AvgUnitPrice),
			AvgMaintenanceUnitPrice: decimal.NewFromFloat(req.AvgMaintenanceUnitPrice),
			TotalPurchaseCost:       decimal.NewFromFloat(totalPurchaseCost),
			TotalMaintenanceCost:    decimal.NewFromFloat(totalMaintenanceCost),
			TotalCost:               decimal.NewFromFloat(totalPurchaseCost + totalMaintenanceCost),
			CreatedBy:               userID,
			StartOfMaintenance:      startOfMaintenance,
			EndOfMaintenance:        endOfMaintenance,
			NumLicencesMaintainance: req.NumLicencesMaintainance,
			Version:                 req.Version,
			Comment:                 sql.NullString{String: req.Comment, Valid: true},
		}, &dgworker.UpsertAcqRightsRequest{
			Sku:                     req.Sku,
			Swidtag:                 swidtag,
			ProductName:             req.ProductName,
			ProductEditor:           req.ProductEditor,
			MetricType:              req.MetricName,
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
			Version:                 req.Version,
		}, nil

}

func metricExists(metrics []*metv1.Metric, name string) int {
	for idx, met := range metrics {
		if met.Name == name {
			return idx
		}
	}
	return -1
}

func (s *productServiceServer) pushUpsertAcqrightsWorkerJob(ctx context.Context, req dgworker.UpsertAcqRightsRequest) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

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
