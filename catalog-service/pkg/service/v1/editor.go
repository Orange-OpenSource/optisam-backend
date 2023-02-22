package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//Create Editor
func (p *productCatalogServer) CreateEditor(ctx context.Context, req *v1.CreateEditorRequest) (e *v1.Editor, err error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	var partnermanagers, audits, vendors []byte
	partnermanagers, err = json.Marshal(req.PartnerManagers)
	if err != nil {
		logger.Log.Error("v1/service - Create Editor - Marshal Error PartnerManagersJson")
		return nil, status.Error(codes.Internal, err.Error())
	}
	audits, err = json.Marshal(req.Audits)
	if err != nil {
		logger.Log.Error("v1/service - Create Editor - Marshal Error audits")
		return nil, status.Error(codes.Internal, err.Error())
	}
	vendors, err = json.Marshal(req.Vendors)
	if err != nil {
		logger.Log.Error("v1/service - create Editor - Marshal Error Vendors")
		return nil, status.Error(codes.Internal, err.Error())
	}

	editorname := strings.Trim(req.Name, " ")
	if editorname == "" {
		return nil, status.Error(codes.Internal, "editor name should not be empty")
	}
	uid := uuid.New().String()
	editor := v1.Editor{
		Id:                 uid,
		Name:               editorname,
		GenearlInformation: string(req.GenearlInformation),
		PartnerManagers:    req.PartnerManagers,
		Audits:             req.Audits,
		Vendors:            req.Vendors,
		CreatedOn:          timestamppb.New(time.Now()),
		UpdatedOn:          timestamppb.New(time.Now()),
	}
	err = p.productRepo.InsertEditorCatalog(ctx, db.InsertEditorCatalogParams{
		ID:                 uid,
		Name:               editorname,
		GeneralInformation: sql.NullString{String: req.GenearlInformation, Valid: true},
		PartnerManagers:    partnermanagers,
		Audits:             audits,
		Vendors:            vendors,
		CreatedOn:          time.Now(),
		UpdatedOn:          time.Now(),
	})
	if err != nil {
		logger.Log.Error("service/v1 | CreateEditor | Cretae Editor", zap.Any("Error while saving records", err))
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, status.Error(codes.Internal, "Error while saving record, Duplicate Editor Name")
		}
		return nil, status.Error(codes.Internal, "Error while saving record")
	}
	return &editor, err
}

func (p *productCatalogServer) GetEditor(ctx context.Context, req *v1.GetEditorRequest) (*v1.Editor, error) {
	// logger.Log.Info("req being processed to Editor.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	var editor v1.Editor

	if !ok {
		return &editor, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return &editor, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	editor.Audits = make([]*v1.Audits, 0)
	editor.Vendors = make([]*v1.Vendors, 0)
	editor.PartnerManagers = make([]*v1.PartnerManagers, 0)

	editorResponse, err := p.productRepo.GetEditorCatalog(ctx, req.EditorId)
	if err != nil {
		logger.Log.Error("service/v1 - geteditor - geteditor", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "DBError")
	}
	editor.Id = editorResponse.ID
	editor.Name = editorResponse.Name
	editor.GenearlInformation = editorResponse.GeneralInformation.String
	audits, err := json.Marshal(editorResponse.Audits)
	if err != nil {
		logger.Log.Error("service/v1 - geteditor - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while ListMetric")
	}
	json.Unmarshal(audits, &editor.Audits)

	vendors, err := json.Marshal(editorResponse.Vendors)
	if err != nil {
		logger.Log.Error("service/v1 - geteditor - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while ListMetric")
	}
	json.Unmarshal(vendors, &editor.Vendors)

	manager, err := json.Marshal(editorResponse.PartnerManagers)
	if err != nil {
		logger.Log.Error("service/v1 - geteditor - Marshal", zap.String("Reason: ", err.Error()))
		return nil, status.Error(codes.Internal, "Error while ListMetric")
	}
	json.Unmarshal(manager, &editor.PartnerManagers)

	createdOnObject, _ := ptypes.TimestampProto(editorResponse.CreatedOn)
	editor.CreatedOn = createdOnObject

	updatedOnObject, _ := ptypes.TimestampProto(editorResponse.UpdatedOn)
	editor.UpdatedOn = updatedOnObject

	return &editor, nil
}

func (p *productCatalogServer) UpdateEditor(ctx context.Context, req *v1.Editor) (e *v1.Editor, err error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	err = p.productRepo.UpdateEditorTx(ctx, req)
	if err != nil {
		logger.Log.Error("service/v1 | UpdateEditor | Update Editor", zap.Any("Error retriving saved record", err))
		return nil, status.Error(codes.Internal, "Error while retriving saved record")
	}
	editor := req
	return editor, err
}

func (s *productCatalogServer) DeleteEditor(ctx context.Context, request *v1.GetEditorRequest) (*v1.DeleteResponse, error) {
	logger.Log.Info("req being processed to DeleteEditor.")
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteResponse{}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DeleteResponse{}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	delErr := s.productRepo.DeleteEditorCatalog(ctx, request.EditorId)
	if delErr != nil {
		logger.Log.Error("DeleteEditor- DeleteEditorByID : ", zap.String("DeleteEditorByID: ", delErr.Error()))
		return nil, status.Error(codes.Internal, "DeleteEditorByID Error.")
	}
	return &v1.DeleteResponse{Success: true}, nil
}
