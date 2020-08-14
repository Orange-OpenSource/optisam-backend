// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	"database/sql"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/product-service/pkg/api/v1"
	gendb "optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
)

//ProductRepository
type ProductRepository struct {
	*gendb.Queries
	db *sql.DB
}

//NewProductRepository creates new Repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}

//ProductRepository
type ProductRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

func NewProductRepositoryTx(db *sql.Tx) *ProductRepositoryTx {
	return &ProductRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

func (p *ProductRepository) UpsertProductTx(ctx context.Context, req *v1.UpsertProductRequest, user string) error {
	var addApplications, deleteApplications, deleteEquipment []string
	var addEquipments []*v1.UpsertProductRequestEquipmentEquipmentuser

	var upsertPartialFlag bool

	if req.Applications.GetOperation() == "add" {
		upsertPartialFlag = true
		addApplications = req.GetApplications().GetApplicationId()
	} else {
		deleteApplications = req.GetApplications().GetApplicationId()
	}
	if req.Equipments.GetOperation() == "add" {
		upsertPartialFlag = true
		addEquipments = req.Equipments.GetEquipmentusers()
	} else {
		deleteEquipmentUsers := req.GetEquipments().GetEquipmentusers()
		deleteEquipment := make([]string, len(deleteEquipmentUsers))
		for _, d := range deleteEquipmentUsers {
			deleteEquipment = append(deleteEquipment, d.GetEquipmentId())
		}
	}

	//Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductRepositoryTx(tx)

	//Upsert Product Master
	if upsertPartialFlag {
		err = pt.UpsertProductPartial(ctx, gendb.UpsertProductPartialParams{Scope: req.GetScope(), Swidtag: req.GetSwidTag(), CreatedBy: user})
	} else {
		err = pt.UpsertProduct(ctx, gendb.UpsertProductParams{
			Swidtag:         req.GetSwidTag(),
			ProductName:     req.GetName(),
			ProductCategory: req.GetCategory(),
			ProductEdition:  req.GetEdition(),
			ProductEditor:   req.GetEditor(),
			ProductVersion:  req.GetVersion(),
			Scope:           req.GetScope(),
			OptionOf:        req.GetOptionOf(),
			CreatedBy:       user,
			CreatedOn:       time.Now(),
			UpdatedBy:       sql.NullString{String: user},
			UpdatedOn:       time.Now(),
		})
	}
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to upsert product", zap.Error(err))
		return err
	}

	//Bulk Insert Applications
	stmt, err := pt.db.PrepareContext(ctx, pq.CopyIn("products_applications", "swidtag", "application_id"))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Failed to prepare statement", zap.Error(err))
		return err
	}
	for _, app := range addApplications {
		_, err = stmt.Exec(req.GetSwidTag(), app)
		if err != nil {
			tx.Rollback()
			logger.Log.Error("Failed to execute statement", zap.Error(err))
			return err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Failed to flush bulk copy", zap.Error(err))
		return err
	}

	//Bulk Insert Equipments
	stmt, err = pt.db.PrepareContext(ctx, pq.CopyIn("products_equipments", "swidtag", "equipment_id", "num_of_users"))
	defer stmt.Close()
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Failed to prepare statement", zap.Error(err))
		return err
	}
	for _, equip := range addEquipments {
		_, err = stmt.Exec(req.GetSwidTag(), equip.GetEquipmentId(), equip.GetNumUser())
		if err != nil {
			tx.Rollback()
			logger.Log.Error("Failed to execute statement", zap.Error(err))
			return err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		tx.Rollback()
		logger.Log.Error("Failed to flush bulk copy", zap.Error(err))
		return err
	}

	// Delete Product Applications
	err = pt.DeleteProductApplications(ctx, gendb.DeleteProductApplicationsParams{
		ProductID:     req.GetSwidTag(),
		ApplicationID: deleteApplications,
	})
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to delete product-applicaiton", zap.Error(err))
		return err
	}

	//Delete Product Equipments
	err = pt.DeleteProductEquipments(ctx, gendb.DeleteProductEquipmentsParams{
		ProductID:   req.GetSwidTag(),
		EquipmentID: deleteEquipment,
	})
	if err != nil {
		tx.Rollback()
		logger.Log.Error("failed to delete product-equipments", zap.Error(err))
		return err
	}

	tx.Commit()
	return nil
}
