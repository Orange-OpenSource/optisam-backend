package v1

import (
	"context"
	"database/sql"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: gocyclo, funlen
func (s *productServiceServer) ListAggregatedAcqRights(ctx context.Context, req *v1.ListAggregatedAcqRightsRequest) (*v1.ListAggregatedAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
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
			logger.Log.Sugar().Errorf("service/v1 - ListAggregatedAcqRights - unable to parse ordering date search params", zap.String("reason", err.Error()))
			return nil, status.Error(codes.InvalidArgument, "not correct timestamp for ordering date search params")
		}
		orderingdate = sql.NullTime{Time: orderingTime, Valid: true}
	}
	dbresp, err := s.productRepo.ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
		AggregationName:             req.GetSearchParams().GetName().GetFilteringkey(),
		IsAggName:                   req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkAggName:                   !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ProductEditor:               req.GetSearchParams().GetEditor().GetFilteringkey(),
		IsEditor:                    req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		LkEditor:                    !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
		Sku:                         req.GetSearchParams().GetSKU().GetFilteringkey(),
		IsSku:                       req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		LkSku:                       !req.GetSearchParams().GetSKU().GetFilterType() && req.GetSearchParams().GetSKU().GetFilteringkey() != "",
		SoftwareProvider:            req.GetSearchParams().GetSoftwareProvider().GetFilteringkey(),
		IsSoftwareProvider:          req.GetSearchParams().GetSoftwareProvider().GetFilterType() && req.GetSearchParams().GetSoftwareProvider().GetFilteringkey() != "",
		LkSoftwareProvider:          !req.GetSearchParams().GetSoftwareProvider().GetFilterType() && req.GetSearchParams().GetSoftwareProvider().GetFilteringkey() != "",
		OrderingDate:                orderingdate,
		IsOrderingDate:              !req.GetSearchParams().GetOrderingDate().GetFilterType() && req.GetSearchParams().GetOrderingDate().GetFilteringkey() != "",
		IsMetric:                    req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		Metric:                      req.GetSearchParams().GetMetric().GetFilteringkey(),
		LkMetric:                    !req.GetSearchParams().GetMetric().GetFilterType() && req.GetSearchParams().GetMetric().GetFilteringkey() != "",
		SwidtagAsc:                  strings.Contains(req.GetSortBy().String(), "NUM_OF_SWIDTAGS") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SwidtagDesc:                 strings.Contains(req.GetSortBy().String(), "NUM_OF_SWIDTAGS") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AggNameAsc:                  strings.Contains(req.GetSortBy().String(), "AGG_NAME") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggNameDesc:                 strings.Contains(req.GetSortBy().String(), "AGG_NAME") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AvgUnitPriceAsc:             strings.Contains(req.GetSortBy().String(), "UNIT_PRICE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AvgUnitPriceDesc:            strings.Contains(req.GetSortBy().String(), "UNIT_PRICE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AvgMaintenanceUnitPriceAsc:  strings.Contains(req.GetSortBy().String(), "MAINTENANCE_PRICE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AvgMaintenanceUnitPriceDesc: strings.Contains(req.GetSortBy().String(), "MAINTENANCE_PRICE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:            strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:           strings.Contains(req.GetSortBy().String(), "EDITOR") && strings.Contains(req.GetSortOrder().String(), "desc"),
		SkuAsc:                      strings.Contains(req.GetSortBy().String(), "SKU") && strings.Contains(req.GetSortOrder().String(), "asc"),
		SkuDesc:                     strings.Contains(req.GetSortBy().String(), "SKU") && strings.Contains(req.GetSortOrder().String(), "desc"),
		MetricAsc:                   strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "asc"),
		MetricDesc:                  strings.Contains(req.GetSortBy().String(), "METRIC") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumLicensesAcquiredAsc:      strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicensesAcquiredDesc:     strings.Contains(req.GetSortBy().String(), "ACQUIRED_LICENSES") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumLicencesMaintenanceAsc:   strings.Contains(req.GetSortBy().String(), "MAINTENANCE_LICENCES") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumLicencesMaintenanceDesc:  strings.Contains(req.GetSortBy().String(), "MAINTENANCE_LICENCES") && strings.Contains(req.GetSortOrder().String(), "desc"),
		StartOfMaintenanceAsc:       strings.Contains(req.GetSortBy().String(), "MAINTENANCE_START") && strings.Contains(req.GetSortOrder().String(), "asc"),
		StartOfMaintenanceDesc:      strings.Contains(req.GetSortBy().String(), "MAINTENANCE_START") && strings.Contains(req.GetSortOrder().String(), "desc"),
		EndOfMaintenanceAsc:         strings.Contains(req.GetSortBy().String(), "MAINTENANCE_END") && strings.Contains(req.GetSortOrder().String(), "asc"),
		EndOfMaintenanceDesc:        strings.Contains(req.GetSortBy().String(), "MAINTENANCE_END") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalMaintenanceCostAsc:     strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalMaintenanceCostDesc:    strings.Contains(req.GetSortBy().String(), "TOTAL_MAINTENANCE_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalCostAsc:                strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalCostDesc:               strings.Contains(req.GetSortBy().String(), "TOTAL_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TotalPurchaseCostAsc:        strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASED_COST") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TotalPurchaseCostDesc:       strings.Contains(req.GetSortBy().String(), "TOTAL_PURCHASED_COST") && strings.Contains(req.GetSortOrder().String(), "desc"),
		LicenseUnderMaintenanceAsc:  strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "asc"),
		LicenseUnderMaintenanceDesc: strings.Contains(req.GetSortBy().String(), "LICENSES_UNDER_MAINTENANCE") && strings.Contains(req.GetSortOrder().String(), "desc"),
		Scope:                       req.Scope,
		PageNum:                     req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:                    req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregationAcqRightsView - ListAcqRightsAggregation", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ListAggregatedAcqRightsResponse{}
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}
	for i := range dbresp {
		temp := &v1.AggregatedRightsView{}
		temp.ID = dbresp[i].AggregationID
		temp.AggregationName = dbresp[i].AggregationName
		temp.Sku = dbresp[i].Sku
		temp.Swidtags = dbresp[i].Swidtags
		resp, err := s.productRepo.GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
			Swidtag: dbresp[i].Swidtags,
			Scope:   dbresp[i].Scope,
		})
		if err != nil {
			logger.Log.Error("Failed to get individual acqrights", zap.Any("swidtags", dbresp[i].Swidtags), zap.Error(err))
		} else if len(resp) > 0 {
			temp.IsIndividualRightExists = true
		}
		temp.ProductEditor = dbresp[i].ProductEditor
		temp.ProductNames = dbresp[i].Products
		temp.MetricName = dbresp[i].Metric
		temp.NumLicensesAcquired = dbresp[i].NumLicensesAcquired
		temp.NumLicencesMaintenance = dbresp[i].NumLicencesMaintenance
		temp.AvgUnitPrice, _ = dbresp[i].AvgUnitPrice.Float64()
		temp.AvgMaintenanceUnitPrice, _ = dbresp[i].AvgMaintenanceUnitPrice.Float64()
		temp.LicenceUnderMaintenance = no
		temp.CorporateSourcingContract = dbresp[i].CorporateSourcingContract
		temp.SoftwareProvider = dbresp[i].SoftwareProvider
		temp.LastPurchasedOrder = dbresp[i].LastPurchasedOrder
		temp.SupportNumber = dbresp[i].SupportNumber
		temp.MaintenanceProvider = dbresp[i].MaintenanceProvider
		temp.FileName = dbresp[i].FileName
		if dbresp[i].StartOfMaintenance.Valid {
			temp.StartOfMaintenance = dbresp[i].StartOfMaintenance.Time.Format(time.RFC3339)
		}
		if dbresp[i].EndOfMaintenance.Valid {
			temp.EndOfMaintenance = dbresp[i].EndOfMaintenance.Time.Format(time.RFC3339)
			if dbresp[i].EndOfMaintenance.Time.After(time.Now()) {
				temp.LicenceUnderMaintenance = yes
			}

		}
		if dbresp[i].OrderingDate.Valid {
			temp.OrderingDate = dbresp[i].OrderingDate.Time.Format(time.RFC3339)
		}
		temp.Comment = dbresp[i].Comment.String
		temp.Scope = dbresp[i].Scope
		temp.TotalCost, _ = dbresp[i].TotalCost.Float64()
		temp.TotalPurchaseCost, _ = dbresp[i].TotalPurchaseCost.Float64()
		temp.TotalMaintenanceCost, _ = dbresp[i].TotalMaintenanceCost.Float64()
		apiresp.Aggregations = append(apiresp.Aggregations, temp)
	}
	return &apiresp, nil
}

// func (s *productServiceServer) ListAcqRightsAggregationRecords(ctx context.Context, req *v1.ListAcqRightsAggregationRecordsRequest) (*v1.ListAcqRightsAggregationRecordsResponse, error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
// 	}

// 	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
// 		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
// 	}

// 	dbresp, err := s.productRepo.ListAcqRightsAggregationIndividual(ctx, db.ListAcqRightsAggregationIndividualParams{
// 		Scope:         req.Scopes,
// 		AggregationID: req.AggregationId,
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - ListAcqRightsAggregationRecords - ListAcqRightsAggregationIndividual", zap.String("reason", err.Error()))
// 		return nil, status.Error(codes.Unknown, err.Error())
// 	}

// 	apiresp := v1.ListAcqRightsAggregationRecordsResponse{}
// 	apiresp.AcquiredRights = make([]*v1.AcqRights, len(dbresp))

// 	for i := range dbresp {
// 		apiresp.AcquiredRights[i] = &v1.AcqRights{}
// 		apiresp.AcquiredRights[i].SwidTag = dbresp[i].Swidtag
// 		apiresp.AcquiredRights[i].ProductName = dbresp[i].ProductName
// 		apiresp.AcquiredRights[i].Metric = dbresp[i].Metric
// 		apiresp.AcquiredRights[i].Editor = dbresp[i].ProductEditor
// 		apiresp.AcquiredRights[i].SKU = dbresp[i].Sku
// 		apiresp.AcquiredRights[i].AcquiredLicensesNumber = dbresp[i].NumLicensesAcquired
// 		apiresp.AcquiredRights[i].LicensesUnderMaintenanceNumber = dbresp[i].NumLicencesMaintainance
// 		apiresp.AcquiredRights[i].AvgLicenesUnitPrice, _ = dbresp[i].AvgUnitPrice.Float64()
// 		apiresp.AcquiredRights[i].AvgMaintenanceUnitPrice, _ = dbresp[i].AvgMaintenanceUnitPrice.Float64()
// 		apiresp.AcquiredRights[i].TotalPurchaseCost, _ = dbresp[i].TotalPurchaseCost.Float64()
// 		apiresp.AcquiredRights[i].TotalMaintenanceCost, _ = dbresp[i].TotalMaintenanceCost.Float64()
// 		apiresp.AcquiredRights[i].TotalCost, _ = dbresp[i].TotalCost.Float64()
// 		apiresp.AcquiredRights[i].Version = dbresp[i].Version
// 	}

// 	return &apiresp, nil
// }

func (s *productServiceServer) GetAggregationAcqrightsExpandedView(ctx context.Context, req *v1.GetAggregationAcqrightsExpandedViewRequest) (*v1.GetAggregationAcqrightsExpandedViewResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	resp, err := s.productRepo.GetAggregationByName(ctx, db.GetAggregationByNameParams{
		AggregationName: req.AggregationName,
		Scope:           req.Scope,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "NoAggregatedAcqrightsFound")
		}
		logger.Log.Error("Couldn't fetch aggregated Acqrights", zap.Error(err), zap.String("aggName", req.AggregationName))
		return nil, status.Error(codes.Internal, "DBError")
	}

	expandedAcq, err := s.productRepo.GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
		Swidtag:  resp.Swidtags,
		Scope:    req.Scope,
		IsMetric: true,
		Metric:   req.Metric,
	})
	if err != nil {
		logger.Log.Error("Failed to get acqrights on expanding aggregated acqrights", zap.Error(err), zap.String("aggName", req.AggregationName))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := &v1.GetAggregationAcqrightsExpandedViewResponse{
		TotalRecords: int32(len(expandedAcq)),
	}
	for _, v := range expandedAcq {
		temp := &v1.AcqRights{}
		temp.SKU = v.Sku
		temp.SwidTag = v.Swidtag
		temp.ProductName = v.ProductName
		temp.Editor = v.ProductEditor
		temp.Metric = v.Metric
		temp.CorporateSourcingContract = v.CorporateSourcingContract
		temp.SoftwareProvider = v.SoftwareProvider
		temp.LastPurchasedOrder = v.LastPurchasedOrder
		temp.SupportNumber = v.SupportNumber
		temp.MaintenanceProvider = v.MaintenanceProvider
		temp.AcquiredLicensesNumber = v.NumLicensesAcquired
		temp.AvgLicenesUnitPrice, _ = v.AvgUnitPrice.Float64()
		temp.LicensesUnderMaintenanceNumber = v.NumLicencesMaintainance
		temp.AvgMaintenanceUnitPrice, _ = v.AvgMaintenanceUnitPrice.Float64()
		temp.TotalCost, _ = v.TotalCost.Float64()
		temp.Version = v.Version
		temp.LicensesUnderMaintenance = no
		if v.StartOfMaintenance.Valid {
			temp.StartOfMaintenance, _ = ptypes.TimestampProto(v.StartOfMaintenance.Time)
		}
		if v.EndOfMaintenance.Valid {
			temp.EndOfMaintenance, _ = ptypes.TimestampProto(v.EndOfMaintenance.Time)
			if temp.EndOfMaintenance.AsTime().After(time.Now()) {
				temp.LicensesUnderMaintenance = yes
			}
		}
		if v.OrderingDate.Valid {
			temp.OrderingDate, _ = ptypes.TimestampProto(v.OrderingDate.Time)
		}
		temp.Comment = v.Comment.String
		temp.TotalPurchaseCost, _ = v.TotalPurchaseCost.Float64()
		temp.TotalMaintenanceCost, _ = v.TotalMaintenanceCost.Float64()
		apiresp.AcqRights = append(apiresp.AcqRights, temp)
	}
	return apiresp, nil
}
