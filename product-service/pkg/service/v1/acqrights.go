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

// nolint: funlen, gocyclo
func (s *productServiceServer) UpsertAcqRights(ctx context.Context, req *v1.UpsertAcqRightsRequest) (*v1.UpsertAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var startOfMaintenance, endOfMaintenance, orderingDate sql.NullTime
	var startTime, endTime, orderingTime time.Time
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
			logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse ordering time", zap.String("reason", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "unable to parse ordering time")
		}
		orderingDate = sql.NullTime{Time: orderingTime, Valid: true}
	}
	if err := s.productRepo.UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
		Sku:                       req.GetSku(),
		Swidtag:                   req.GetSwidtag(),
		ProductName:               req.GetProductName(),
		CorporateSourcingContract: req.GetCorporateSourcingContract(),
		OrderingDate:              orderingDate,
		ProductEditor:             req.GetProductEditor(),
		Metric:                    req.GetMetricType(),
		SoftwareProvider:          req.GetSoftwareProvider(),
		MaintenanceProvider:       req.GetMaintenanceProvider(),
		NumLicensesAcquired:       req.GetNumLicensesAcquired(),
		NumLicencesMaintainance:   req.GetNumLicencesMaintainance(),
		AvgUnitPrice:              decimal.NewFromFloat(req.GetAvgUnitPrice()),
		AvgMaintenanceUnitPrice:   decimal.NewFromFloat(req.GetAvgMaintenanceUnitPrice()),
		TotalPurchaseCost:         decimal.NewFromFloat(req.GetTotalPurchaseCost()),
		TotalMaintenanceCost:      decimal.NewFromFloat(req.GetTotalMaintenanceCost()),
		TotalCost:                 decimal.NewFromFloat(req.GetTotalCost()),
		Scope:                     req.GetScope(),
		StartOfMaintenance:        startOfMaintenance,
		EndOfMaintenance:          endOfMaintenance,
		Version:                   req.GetVersion(),
		CreatedBy:                 userClaims.UserID,
		LastPurchasedOrder:        req.GetLastPurchasedOrder(),
		SupportNumber:             req.GetSupportNumber(),
		Repartition:               req.GetRepartition(),
	}); err != nil {
		logger.Log.Error("service/v1 - UpsertAcqRights - UpsertAcquiredRights", zap.String("reason", err.Error()))
		return &v1.UpsertAcqRightsResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}
	// For Worker Queue
	s.pushUpsertAcqrightsWorkerJob(ctx, dgworker.UpsertAcqRightsRequest{
		Sku:                       req.Sku,
		Swidtag:                   req.Swidtag,
		ProductName:               req.ProductName,
		ProductEditor:             req.ProductEditor,
		MetricType:                req.MetricType,
		NumLicensesAcquired:       req.NumLicensesAcquired,
		AvgUnitPrice:              req.AvgUnitPrice,
		AvgMaintenanceUnitPrice:   req.AvgMaintenanceUnitPrice,
		TotalPurchaseCost:         req.TotalPurchaseCost,
		TotalMaintenanceCost:      req.TotalMaintenanceCost,
		TotalCost:                 req.TotalCost,
		Scope:                     req.Scope,
		StartOfMaintenance:        req.StartOfMaintenance,
		EndOfMaintenance:          req.EndOfMaintenance,
		NumLicencesMaintenance:    req.NumLicencesMaintainance,
		Version:                   req.Version,
		CorporateSourcingContract: req.CorporateSourcingContract,
		SoftwareProvider:          req.SoftwareProvider,
		OrderingDate:              req.OrderingDate,
		MaintenanceProvider:       req.MaintenanceProvider,
		LastPurchasedOrder:        req.LastPurchasedOrder,
		SupportNumber:             req.SupportNumber,
		Repartition:               req.Repartition,
	})
	return &v1.UpsertAcqRightsResponse{Success: true}, nil
}

// nolint: gocyclo, funlen
func (s *productServiceServer) ListAcqRights(ctx context.Context, req *v1.ListAcqRightsRequest) (*v1.ListAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		logger.Log.Sugar().Infof("acrights-service - ListAcqRights - user don't have access to the scopes: %v, requested scopes: %v", userClaims.Socpes, req.Scopes)
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var orderingTime time.Time
	var orderingdate sql.NullTime
	if req.GetSearchParams().GetOrderingDate().GetFilteringkey() != "" {
		var err error
		if strings.Contains(req.GetSearchParams().GetOrderingDate().GetFilteringkey(), "/") && len(req.GetSearchParams().GetOrderingDate().GetFilteringkey()) <= 8 {
			orderingTime, err = time.Parse("1/2/06", req.GetSearchParams().GetOrderingDate().GetFilteringkey())
		} else if len(req.GetSearchParams().GetOrderingDate().GetFilteringkey()) == 10 {
			orderingTime, err = time.Parse("02-01-2006", req.GetSearchParams().GetOrderingDate().GetFilteringkey())
		} else {
			orderingTime, err = time.Parse(time.RFC3339Nano, req.GetSearchParams().GetOrderingDate().GetFilteringkey())
		}
		if err != nil {
			logger.Log.Sugar().Errorf("service/v1 - ListAcqRights - unable to parse ordering date search params", zap.String("reason", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "not correct timestamp for ordering date search params")
		}
		orderingdate = sql.NullTime{Time: orderingTime, Valid: true}
	}
	dbresp, err := s.productRepo.ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
		Scope:                       req.Scopes,
		Sku:                         req.GetSearchParams().GetSKU().GetFilteringkey(),
		IsSku:                       req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		LkSku:                       !req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		SoftwareProvider:            req.GetSearchParams().GetSoftwareProvider().GetFilteringkey(),
		IsSoftwareProvider:          req.GetSearchParams().GetSoftwareProvider().GetFilterType() && req.GetSearchParams().GetSoftwareProvider().GetFilteringkey() != "",
		LkSoftwareProvider:          !req.GetSearchParams().GetSoftwareProvider().GetFilterType() && req.GetSearchParams().GetSoftwareProvider().GetFilteringkey() != "",
		OrderingDate:                orderingdate,
		IsOrderingDate:              !req.GetSearchParams().GetOrderingDate().GetFilterType() && req.GetSearchParams().GetOrderingDate().GetFilteringkey() != "",
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
		licenses, _ := s.GetAvailableLicenses(ctx, &v1.GetAvailableLicensesRequest{Sku: dbresp[i].Sku, Scope: req.Scopes[0]})
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
		apiresp.AcquiredRights[i].SoftwareProvider = dbresp[i].SoftwareProvider
		apiresp.AcquiredRights[i].SupportNumber = dbresp[i].SupportNumber
		apiresp.AcquiredRights[i].LastPurchasedOrder = dbresp[i].LastPurchasedOrder
		apiresp.AcquiredRights[i].MaintenanceProvider = dbresp[i].MaintenanceProvider
		apiresp.AcquiredRights[i].CorporateSourcingContract = dbresp[i].CorporateSourcingContract
		apiresp.AcquiredRights[i].FileName = dbresp[i].FileName
		apiresp.AcquiredRights[i].ProductSwidTag = dbresp[i].ProductSwidTag.String
		apiresp.AcquiredRights[i].VersionSwidTag = dbresp[i].VersionSwidTag.String

		apiresp.AcquiredRights[i].EditorId = dbresp[i].EditorID.String
		apiresp.AcquiredRights[i].ProductId = dbresp[i].ProductID.String

		apiresp.AcquiredRights[i].Repartition = dbresp[i].Repartition
		apiresp.AcquiredRights[i].SharedLicenses = licenses.TotalSharedLicenses
		apiresp.AcquiredRights[i].RecievedLicenses = licenses.TotalRecievedLicenses
		apiresp.AcquiredRights[i].AvailableLicenses = licenses.AvailableLicenses
		apiresp.AcquiredRights[i].SharedData = licenses.SharedData
		if dbresp[i].OrderingDate.Valid {
			apiresp.AcquiredRights[i].OrderingDate, _ = ptypes.TimestampProto(dbresp[i].OrderingDate.Time)
		}
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

	dbAcqRight, upsertAcqRight, err := s.validateAcqRight(ctx, req)
	if err != nil {
		return &v1.AcqRightResponse{}, err
	}
	dbAcqRight.CreatedBy = userClaims.UserID
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
	resp, err := s.GetAvailableLicenses(ctx, &v1.GetAvailableLicensesRequest{Sku: req.Sku, Scope: req.Scope})
	if err != nil {
		logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAvailableLicenses", zap.String("reason", err.Error()))
		return &v1.AcqRightResponse{}, status.Error(codes.Internal, "DBError")
	}
	if req.NumLicensesAcquired < resp.TotalSharedLicenses {
		logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAvailableLicenses", zap.String("reason", "AcquiredLicences less than sharedLicences"))
		return &v1.AcqRightResponse{}, status.Error(codes.InvalidArgument, "AcquiredLicences less than sharedLicences")
	}
	dbAcqRight, upsertAcqRight, err := s.validateAcqRight(ctx, req)
	if err != nil {
		return &v1.AcqRightResponse{}, err
	}
	dbAcqRight.CreatedBy = userClaims.UserID
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
	if err := s.productRepo.DeleteSharedLicences(ctx, db.DeleteSharedLicencesParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	}); err != nil {
		return &v1.DeleteAcqRightResponse{Success: false}, status.Error(codes.Internal, "DBError")
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

func (s *productServiceServer) GetAvailableLicenses(ctx context.Context, req *v1.GetAvailableLicensesRequest) (*v1.GetAvailableLicensesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.GetAvailableLicensesResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Sugar().Errorf("service/v1 - GetAvailableLicenses - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return &v1.GetAvailableLicensesResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}

	totalSharedLicences := 0
	totalRecievedLicences := 0
	acqLicenses, aggLicenses := 0, 0

	aggregatedLicenses, err := s.productRepo.GetAvailableAggLicenses(ctx, db.GetAvailableAggLicensesParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			aggLicenses = 0
		} else {
			logger.Log.Error("service/v1 - GetAvailableLicenses - GetAvailableAggLicenses", zap.String("reason", err.Error()))
			return &v1.GetAvailableLicensesResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		aggLicenses = int(aggregatedLicenses)
	}

	acquiredLicenses, err := s.productRepo.GetAvailableAcqLicenses(ctx, db.GetAvailableAcqLicensesParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			acqLicenses = 0
		} else {
			logger.Log.Error("service/v1 - GetAvailableLicenses - GetAvailableAcqLicenses", zap.String("reason", err.Error()))
			return &v1.GetAvailableLicensesResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		acqLicenses = int(acquiredLicenses)
	}

	totalSharedData, err := s.productRepo.GetTotalSharedLicenses(ctx, db.GetTotalSharedLicensesParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			totalSharedLicences = 0
			totalRecievedLicences = 0
		} else {
			logger.Log.Error("service/v1 - GetAvailableLicenses - GetTotalSharedLicenses", zap.String("reason", err.Error()))
			return &v1.GetAvailableLicensesResponse{}, status.Error(codes.Internal, "DBError")
		}
	} else {
		totalSharedLicences = int(totalSharedData.TotalSharedLicences)
		totalRecievedLicences = int(totalSharedData.TotalRecievedLicences)
	}

	sharedData, err := s.productRepo.GetSharedLicenses(ctx, db.GetSharedLicensesParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			totalSharedLicences = 0
			totalRecievedLicences = 0
		} else {
			logger.Log.Error("service/v1 - GetAvailableLicenses - GetSharedLicenses", zap.String("reason", err.Error()))
			return &v1.GetAvailableLicensesResponse{}, status.Error(codes.Internal, "DBError")
		}
	}
	availableLicenses := acqLicenses + aggLicenses - totalSharedLicences + totalRecievedLicences
	apiresp := &v1.GetAvailableLicensesResponse{}
	apiresp.SharedData = make([]*v1.SharedData, len(sharedData))
	apiresp.AvailableLicenses = int32(availableLicenses)
	apiresp.TotalSharedLicenses = int32(totalSharedLicences)
	apiresp.TotalRecievedLicenses = int32(totalRecievedLicences)
	for i := range sharedData {
		apiresp.SharedData[i] = &v1.SharedData{}
		apiresp.SharedData[i].Scope = sharedData[i].SharingScope
		apiresp.SharedData[i].SharedLicenses = sharedData[i].SharedLicences
		apiresp.SharedData[i].RecievedLicenses = sharedData[i].RecievedLicences
	}
	return apiresp, nil
}

func (s *productServiceServer) DeleteSharedLicenses(ctx context.Context, req *v1.DeleteSharedLicensesRequest) (*v1.DeleteSharedLicensesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteSharedLicensesResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - DeleteSharedLicenses ", zap.String("reason", "ScopeError"))
		return &v1.DeleteSharedLicensesResponse{Success: false}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	if dbresp := s.productRepo.UpsertSharedLicenses(ctx, db.UpsertSharedLicensesParams{
		Sku:            req.Sku,
		Scope:          req.Scope,
		SharingScope:   req.RecieverScope,
		SharedLicences: 0,
	}); dbresp != nil {
		logger.Log.Error("service/v1 - DeleteSharedLicenses - UpsertSharedLicenses", zap.String("reason", dbresp.Error()))
		return &v1.DeleteSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}
	if dbresp := s.productRepo.UpsertRecievedLicenses(ctx, db.UpsertRecievedLicensesParams{
		Sku:              req.Sku,
		Scope:            req.RecieverScope,
		SharingScope:     req.Scope,
		RecievedLicences: 0,
	}); dbresp != nil {
		logger.Log.Error("service/v1 - DeleteSharedLicenses - UpsertRecievedLicenses", zap.String("reason", dbresp.Error()))
		return &v1.DeleteSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}
	return &v1.DeleteSharedLicensesResponse{Success: true}, nil
}

func (s *productServiceServer) UpdateAcqrightsSharedLicenses(ctx context.Context, req *v1.UpdateSharedLicensesRequest) (*v1.UpdateSharedLicensesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses ", zap.String("reason", "ScopeError"))
		return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	resp, err := s.GetAvailableLicenses(ctx, &v1.GetAvailableLicensesRequest{Sku: req.Sku, Scope: req.Scope})
	if err != nil {
		logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAvailableLicenses", zap.String("reason", err.Error()))
		return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	senderSku, err := s.productRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
		AcqrightSku: req.Sku,
		Scope:       req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &v1.UpdateSharedLicensesResponse{}, status.Error(codes.InvalidArgument, "SKU does not exist")
		}
		logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAcqRightBySKU", zap.String("reason", err.Error()))
		return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Internal, "DBError")
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
		_, err := s.productRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
			AcqrightSku: req.Sku,
			Scope:       v.RecieverScope,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = s.CreateMetricIfNotExists(ctx, req.Scope, v.RecieverScope, senderSku.Metric)
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - CreateMetricIfNotExists", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{Success: false}, err
				}
				unitPrice, _ := senderSku.AvgUnitPrice.Float64()
				maintenanceUnitPrice, _ := senderSku.AvgMaintenanceUnitPrice.Float64()
				startOfMaintenance := senderSku.StartOfMaintenance.Time.Format(time.RFC3339)
				endOfMaintenance := senderSku.EndOfMaintenance.Time.Format(time.RFC3339)
				if senderSku.NumLicencesMaintainance == 0 {
					startOfMaintenance = ""
					endOfMaintenance = ""
				}
				orderingDate := senderSku.OrderingDate.Time.Format(time.RFC3339)
				if orderingDate == "0001-01-01T00:00:00Z" {
					orderingDate = ""
				}
				_, err = s.CreateAcqRight(ctx, &v1.AcqRightRequest{
					Sku:                       senderSku.Sku,
					ProductName:               senderSku.ProductName,
					ProductEditor:             senderSku.ProductEditor,
					Version:                   senderSku.Version,
					MetricName:                senderSku.Metric,
					NumLicensesAcquired:       0,
					AvgUnitPrice:              unitPrice,
					StartOfMaintenance:        startOfMaintenance,
					EndOfMaintenance:          endOfMaintenance,
					NumLicencesMaintainance:   senderSku.NumLicencesMaintainance,
					LastPurchasedOrder:        senderSku.LastPurchasedOrder,
					SupportNumber:             senderSku.SupportNumber,
					MaintenanceProvider:       senderSku.MaintenanceProvider,
					Comment:                   senderSku.Comment.String,
					OrderingDate:              orderingDate,
					CorporateSourcingContract: senderSku.CorporateSourcingContract,
					SoftwareProvider:          senderSku.SoftwareProvider,
					FileName:                  senderSku.FileName,
					Repartition:               senderSku.Repartition,
					FileData:                  senderSku.FileData,
					AvgMaintenanceUnitPrice:   maintenanceUnitPrice,
					Scope:                     v.RecieverScope,
				})
				if err != nil {
					logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - CreateAcqRight", zap.String("reason", err.Error()))
					return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Internal, "InternalError")
				}
			} else {
				logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - GetAcqRightBySKU", zap.String("reason", err.Error()))
				return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Internal, "DBError")
			}
		}
		if dbresp := s.productRepo.UpsertSharedLicenses(ctx, db.UpsertSharedLicensesParams{
			Sku:            req.Sku,
			Scope:          req.Scope,
			SharingScope:   v.RecieverScope,
			SharedLicences: v.SharedLicenses,
		}); dbresp != nil {
			logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - UpsertSharedLicenses", zap.String("reason", dbresp.Error()))
			return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
		}
		if dbresp := s.productRepo.UpsertRecievedLicenses(ctx, db.UpsertRecievedLicensesParams{
			Sku:              req.Sku,
			Scope:            v.RecieverScope,
			SharingScope:     req.Scope,
			RecievedLicences: v.SharedLicenses,
		}); dbresp != nil {
			logger.Log.Error("service/v1 - UpdateAcqrightsSharedLicenses - UpsertRecievedLicenses", zap.String("reason", dbresp.Error()))
			return &v1.UpdateSharedLicensesResponse{Success: false}, status.Error(codes.Unknown, "DBError")
		}
	}
	return &v1.UpdateSharedLicensesResponse{Success: true}, nil
}

func (s *productServiceServer) CreateMetricIfNotExists(ctx context.Context, senderScope string, recieverScope string, metric string) error {
	metrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{senderScope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricIfNotExists - ListMetrices", zap.String("reason", err.Error()))
		return status.Error(codes.Internal, "ServiceError")
	}
	recieverMetrics, err := s.metric.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{recieverScope},
	})
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricIfNotExists - ListMetrices", zap.String("reason", err.Error()))
		return status.Error(codes.Internal, "ServiceError")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		return status.Error(codes.InvalidArgument, "MetricNotExists")
	}
	for _, met := range strings.Split(metric, ",") {
		idx := metricExists(metrics.Metrices, met)
		if idx == -1 {
			logger.Log.Error("service/v1 - CreateMetricIfNotExists - metric does not exist", zap.String("metric:", metric))
			return status.Error(codes.InvalidArgument, "MetricNotExists")
		}
		if metrics == nil || len(metrics.Metrices) == 0 {
			_, err = s.metric.CreateMetric(ctx, &metv1.CreateMetricRequest{
				Metric:        metrics.Metrices[idx],
				SenderScope:   senderScope,
				RecieverScope: recieverScope,
			})
			if err != nil {
				return err
			}
		} else {
			index := metricExists(recieverMetrics.Metrices, met)
			if index == -1 {
				_, err = s.metric.CreateMetric(ctx, &metv1.CreateMetricRequest{
					Metric:        metrics.Metrices[idx],
					SenderScope:   senderScope,
					RecieverScope: recieverScope,
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *productServiceServer) GetMetric(ctx context.Context, req *v1.GetMetricRequest) (*v1.GetMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.GetMetricResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Sugar().Errorf("service/v1 - GetMetric - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return &v1.GetMetricResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	resp, err := s.productRepo.GetMetricsBySku(ctx, db.GetMetricsBySkuParams{
		Sku:   req.Sku,
		Scope: req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetMetric - GetMetricsBySku", zap.String("reason", err.Error()))
		return &v1.GetMetricResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetMetricResponse{Metric: resp.Metric}, nil
}

func (s *productServiceServer) DownloadAcqRightFile(ctx context.Context, req *v1.DownloadAcqRightFileRequest) (*v1.DownloadAcqRightFileResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		logger.Log.Sugar().Errorf("service/v1 - DownloadAcqRightFile - req scope: %s, available scopes: %v", req.Scope, userClaims.Socpes)
		return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.InvalidArgument, "ScopeValidationError")
	}
	acq, err := s.productRepo.GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
		AcqrightSku: req.Sku,
		Scope:       req.Scope,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.InvalidArgument, "SKU does not exist")
		}
		logger.Log.Error("service/v1 - DownloadAcqRightFile - GetAcqRightBySKU", zap.String("reason", err.Error()))
		return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.Internal, "DBError")
	}
	if acq.FileName == "" {
		return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.InvalidArgument, "Acquired Right does not contain file")
	}
	acqFileData, err := s.productRepo.GetAcqRightFileDataBySKU(ctx, db.GetAcqRightFileDataBySKUParams{
		AcqrightSku: req.Sku,
		Scope:       req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - DownloadAcqRightFile - GetAcqRightFileDataBySKU", zap.String("reason", err.Error()))
		return &v1.DownloadAcqRightFileResponse{}, status.Error(codes.Internal, "DBError")
	}
	return &v1.DownloadAcqRightFileResponse{
		FileData: acqFileData,
	}, nil
}

// nolint: gocyclo,funlen
func (s *productServiceServer) validateAcqRight(ctx context.Context, req *v1.AcqRightRequest) (db.UpsertAcqRightsParams, *dgworker.UpsertAcqRightsRequest, error) {

	var swidtag string
	if req.Version != "" {
		swidtag = strings.ReplaceAll(strings.Join([]string{req.ProductName, req.ProductEditor, req.Version}, "_"), " ", "_")
	} else {
		swidtag = strings.ReplaceAll(strings.Join([]string{req.ProductName, req.ProductEditor}, "_"), " ", "_")
	}
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
		idx := metricExists(metrics.Metrices, met)
		if idx == -1 {
			logger.Log.Error("service/v1 - validateAcqRight - metric does not exist", zap.String("metric:", met))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "MetricNotExists")
		}
		// if err := s.metricCheckForProcessorAndNup(ctx, metrics.Metrices, idx, 0, swidtag, req.Scope); err != nil {
		// 	return db.UpsertAcqRightsParams{}, nil, err
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
			logger.Log.Error("service/v1 - validateAcqRight - unable to parse ordering time", zap.String("reason", err.Error()))
			return db.UpsertAcqRightsParams{}, nil, status.Error(codes.InvalidArgument, "unable to parse ordering time")
		}
		orderingDate = sql.NullTime{Time: orderingTime, Valid: true}
	}
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
	return db.UpsertAcqRightsParams{
			Sku:                       req.Sku,
			Swidtag:                   swidtag,
			ProductName:               req.ProductName,
			ProductEditor:             req.ProductEditor,
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
			NumLicencesMaintainance:   req.NumLicencesMaintainance,
			Version:                   req.Version,
			Comment:                   sql.NullString{String: req.Comment, Valid: true},
			OrderingDate:              orderingDate,
			CorporateSourcingContract: req.CorporateSourcingContract,
			SoftwareProvider:          req.SoftwareProvider,
			LastPurchasedOrder:        req.LastPurchasedOrder,
			SupportNumber:             req.SupportNumber,
			MaintenanceProvider:       req.MaintenanceProvider,
			FileName:                  req.FileName,
			FileData:                  req.FileData,
			Repartition:               req.Repartition,
		}, &dgworker.UpsertAcqRightsRequest{
			Sku:                       req.Sku,
			Swidtag:                   swidtag,
			ProductName:               req.ProductName,
			ProductEditor:             req.ProductEditor,
			MetricType:                req.MetricName,
			NumLicensesAcquired:       req.NumLicensesAcquired,
			AvgUnitPrice:              req.AvgUnitPrice,
			AvgMaintenanceUnitPrice:   req.AvgMaintenanceUnitPrice,
			TotalPurchaseCost:         totalPurchaseCost,
			TotalMaintenanceCost:      totalMaintenanceCost,
			TotalCost:                 (totalPurchaseCost + totalMaintenanceCost),
			Scope:                     req.Scope,
			StartOfMaintenance:        req.StartOfMaintenance,
			EndOfMaintenance:          req.EndOfMaintenance,
			NumLicencesMaintenance:    req.NumLicencesMaintainance,
			Version:                   req.Version,
			OrderingDate:              req.OrderingDate,
			CorporateSourcingContract: req.CorporateSourcingContract,
			SoftwareProvider:          req.SoftwareProvider,
			LastPurchasedOrder:        req.LastPurchasedOrder,
			SupportNumber:             req.SupportNumber,
			MaintenanceProvider:       req.MaintenanceProvider,
			Repartition:               req.Repartition,
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

// func (s *productServiceServer) metricCheckForProcessorAndNup(ctx context.Context, metrics []*metv1.Metric, idx int, aggid int32, swidtag, scope string) error {
// 	switch metrics[idx].Type {
// 	case metModels.MetricOPSOracleProcessorStandard.String(), metModels.MetricOracleNUPStandard.String():
// 		if aggid != 0 {
// 			aggRightsMetrics, err := s.productRepo.GetAggRightMetricsByAggregationId(ctx, db.GetAggRightMetricsByAggregationIdParams{
// 				AggID: aggid,
// 				Scope: scope,
// 			})
// 			if err != nil {
// 				logger.Log.Sugar().Errorf("service/v1 - validateAcqRight - metricCheckForProcessorAndNup - repo/GetAggRightMetricsByAggregationId - unable to get acqrights metrics by aggregation id:%v", zap.Error(err))
// 				return status.Error(codes.Internal, "DBError")
// 			}
// 			for _, acqMetric := range aggRightsMetrics {
// 				if err := metricCheckForProcessorAndNupInRights(metrics, acqMetric.Metric, idx); err != nil {
// 					return err
// 				}
// 			}
// 		} else {
// 			acqMetrics, err := s.productRepo.GetAcqRightMetricsBySwidtag(ctx, db.GetAcqRightMetricsBySwidtagParams{
// 				Swidtag: swidtag,
// 				Scope:   scope,
// 			})
// 			if err != nil {
// 				logger.Log.Sugar().Errorf("service/v1 - validateAcqRight - metricCheckForProcessorAndNup - repo/GetAcqRightMetricsBySwidtag - unable to get acqrights metrics by swidtag:%v", zap.Error(err))
// 				return status.Error(codes.Internal, "DBError")
// 			}
// 			for _, acqMetric := range acqMetrics {
// 				if err := metricCheckForProcessorAndNupInRights(metrics, acqMetric.Metric, idx); err != nil {
// 					return err
// 				}
// 			}
// 		}

// 		return nil
// 	default:
// 		return nil
// 	}
// }

// func metricCheckForProcessorAndNupInRights(metrics []*metv1.Metric, acqmet string, idx int) error {
// 	ind := metricExists(metrics, acqmet)
// 	if ind == -1 {
// 		logger.Log.Error("service/v1 - validateAcqRight - acquired right metric does not exist", zap.String("metric:", acqmet))
// 		return status.Error(codes.Internal, "Internal Error")
// 	}
// 	if metrics[ind].Type == metModels.MetricOPSOracleProcessorStandard.String() && metrics[idx].Type == metModels.MetricOPSOracleProcessorStandard.String() {
// 		if metrics[idx].Name != metrics[ind].Name {
// 			return status.Error(codes.InvalidArgument, "You can not use 2 different metrics of type oracle.processor.standard for the same product/aggregation.")
// 		}
// 	}
// 	if metrics[ind].Type == metModels.MetricOracleNUPStandard.String() && metrics[idx].Type == metModels.MetricOracleNUPStandard.String() {
// 		if metrics[idx].Name != metrics[ind].Name {
// 			return status.Error(codes.InvalidArgument, "You can not use 2 different metrics of type oracle.nup.standard for the same product/aggregation.")
// 		}
// 	}
// 	return nil
// }
