package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *equipmentServiceServer) EquipmentsPerEquipmentType(ctx context.Context, req *v1.EquipmentsPerEquipmentTypeRequest) (*v1.EquipmentsPerEquipmentTypeResponse, error) {
	// Finding Claims of User
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	// Checking if user has the permission to see this scope
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "User do not have access to the scope")
	}

	// Convert single scope to slice of string
	var scopes []string
	scopes = append(scopes, req.Scope)

	// Find all equipment types in the scope
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - EquipmentsPerEquipmentType - db/EquipmentTypes", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Server Error")
	}

	typeEquipments := make([]*v1.TypeEquipments, 0)

	// Find Equipments by Equipment Type
	for _, eqType := range eqTypes {
		numEquipments, _, err := s.equipmentRepo.Equipments(ctx, eqType, &repo.QueryEquipments{}, scopes)
		if err != nil {
			if !errors.Is(err, repo.ErrNoData) {
				logger.Log.Error("service/v1 - EquipmentsPerEquipmentType - db/Equipments", zap.Error(err))
				return nil, status.Error(codes.Internal, "Internal Server Error")
			}
		}
		typeEquipments = append(typeEquipments, &v1.TypeEquipments{
			EquipType:     eqType.Type,
			NumEquipments: numEquipments,
		})
	}

	return &v1.EquipmentsPerEquipmentTypeResponse{
		TypesEquipments: typeEquipments,
	}, nil
}
