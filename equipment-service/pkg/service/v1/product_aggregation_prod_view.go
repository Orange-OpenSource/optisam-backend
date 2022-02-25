package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *equipmentServiceServer) ListEquipmentsForProductAggregation(ctx context.Context, req *v1.ListEquipmentsForProductAggregationRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProductAggregation - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.EqTypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProductAggregation - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	idx := attributeIndexByName(req.SortBy, eqType.Attributes)
	if idx < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot find sort by attribute: %s", req.SortBy)
	}

	if !eqType.Attributes[idx].IsDisplayed {
		return nil, status.Errorf(codes.InvalidArgument, "cannot sort by attribute: %s is not displayable", req.SortBy)
	}

	filter, err := parseEquipmentQueryParam(req.SearchParams, eqType.Attributes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProductAggregation - parsing equipment query params", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "cannot parse equipment query params")
	}

	queryParams := &repo.QueryEquipments{
		PageSize:  req.PageSize,
		Offset:    offset(req.PageSize, req.PageNum),
		SortBy:    req.SortBy,
		SortOrder: sortOrder(req.SortOrder),
		Filter:    filter,
	}

	numOfrecords, equipments, err := s.equipmentRepo.ListEquipmentsForProductAggregation(ctx, req.Name, eqType, queryParams, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProductAggregation - ", zap.String("reason", err.Error()), zap.Any("request params", queryParams))
		return nil, status.Error(codes.Internal, "cannot fetch equipments for product aggregation")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: numOfrecords,
		Equipments:   equipments,
	}, nil
}
