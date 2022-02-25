package v1

import (
	"context"
	"database/sql"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"time"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: gocyclo
func (s *productServiceServer) ListAggregatedAcqRights(ctx context.Context, req *v1.ListAggregatedAcqRightsRequest) (*v1.ListAggregatedAcqRightsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	dbresp, err := s.productRepo.ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
		Scope:    req.Scopes,
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
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
		temp.ID = dbresp[i].ID
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
		temp.NumLicencesMaintainance = dbresp[i].NumLicencesMaintainance
		temp.AvgUnitPrice, _ = dbresp[i].AvgUnitPrice.Float64()
		temp.AvgMaintenanceUnitPrice, _ = dbresp[i].AvgMaintenanceUnitPrice.Float64()
		temp.LicenceUnderMaintenance = no
		if dbresp[i].StartOfMaintenance.Valid {
			temp.StartOfMaintenance = dbresp[i].StartOfMaintenance.Time.Format(time.RFC3339)
		}
		if dbresp[i].EndOfMaintenance.Valid {
			temp.EndOfMaintenance = dbresp[i].EndOfMaintenance.Time.Format(time.RFC3339)
			if dbresp[i].EndOfMaintenance.Time.After(time.Now()) {
				temp.LicenceUnderMaintenance = yes
			}
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

func (s *productServiceServer) ListAcqRightsAggregationRecords(ctx context.Context, req *v1.ListAcqRightsAggregationRecordsRequest) (*v1.ListAcqRightsAggregationRecordsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}

	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	dbresp, err := s.productRepo.ListAcqRightsAggregationIndividual(ctx, db.ListAcqRightsAggregationIndividualParams{
		Scope:         req.Scopes,
		AggregationID: req.AggregationId,
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListAcqRightsAggregationRecords - ListAcqRightsAggregationIndividual", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, err.Error())
	}

	apiresp := v1.ListAcqRightsAggregationRecordsResponse{}
	apiresp.AcquiredRights = make([]*v1.AcqRights, len(dbresp))

	for i := range dbresp {
		apiresp.AcquiredRights[i] = &v1.AcqRights{}
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
		apiresp.AcquiredRights[i].Version = dbresp[i].Version
	}

	return &apiresp, nil
}

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
		Swidtag: resp.Swidtags,
		Scope:   req.Scope,
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

		temp.Comment = v.Comment.String
		temp.TotalPurchaseCost, _ = v.TotalPurchaseCost.Float64()
		temp.TotalMaintenanceCost, _ = v.TotalMaintenanceCost.Float64()
		apiresp.AcqRights = append(apiresp.AcqRights, temp)
	}
	return apiresp, nil
}
