package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *ProductCatalogRepository) UpdateEditorTx(ctx context.Context, req *v1.Editor) (err error) {

	editorname := strings.Trim(req.Name, " ")
	if editorname == "" {
		return status.Error(codes.Internal, "editor name should not be empty")
	}

	partnermanagers, err := json.Marshal(req.PartnerManagers)
	if err != nil {
		logger.Log.Error("v1/service - Update Editor - Marshal Error PartnerManagersJson")
		return status.Error(codes.Internal, err.Error())
	}
	audits, err := json.Marshal(req.Audits)
	if err != nil {
		logger.Log.Error("v1/service - Update Editor - Marshal Error audits")
		return status.Error(codes.Internal, err.Error())
	}
	vendors, err := json.Marshal(req.Vendors)
	if err != nil {
		logger.Log.Error("v1/service - Update Editor - Marshal Error Vendors")
		return status.Error(codes.Internal, err.Error())

	}

	tx, err := p.db.BeginTx(ctx, nil)

	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}

	pt := NewProductCatalogRepositoryTx(tx)
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = pt.Queries.UpdateEditorCatalog(ctx, db.UpdateEditorCatalogParams{
		GeneralInformation: sql.NullString{String: req.GetGenearlInformation(), Valid: true},
		PartnerManagers:    partnermanagers,
		Audits:             audits,
		Vendors:            vendors,
		UpdatedOn:          time.Now(),
		ID:                 req.Id,
		Name:               editorname,
	})
	if err != nil {
		logger.Log.Error("service/v1 | Update | Update Editor", zap.Any("Error while saving records", err))
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return status.Error(codes.Internal, "Error while saving record, Duplicate Editor Name")
		}
		return status.Error(codes.Internal, "Error while saving record")
	}

	err = pt.Queries.UpdateEditorNameForProductCatalog(ctx, db.UpdateEditorNameForProductCatalogParams{
		EditorName: editorname,
		Editorid:   req.Id,
	})
	if err != nil {
		logger.Log.Error("service/v1 | UpdateEditor | Update Editor", zap.Any("Error retriving saved product record", err))
		return status.Error(codes.Internal, "Error while retriving saved record")
	}

	err = pt.Queries.UpdateVersionsSysSwidatagsForEditor(ctx, req.Id)
	if err != nil {
		logger.Log.Error("service/v1 | UpdateEditor | Update Editor", zap.Any("Error retriving saved product record", err))
		return status.Error(codes.Internal, "Error while retriving saved record")
	}

	return err
}
