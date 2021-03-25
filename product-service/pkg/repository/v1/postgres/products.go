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

	"go.uber.org/zap"
)

//ProductRepository ...
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

//ProductRepositoryTx ...
type ProductRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

//NewProductRepositoryTx ...
func NewProductRepositoryTx(db *sql.Tx) *ProductRepositoryTx {
	return &ProductRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

//UpsertProductTx upserts products/ linking data
func (p *ProductRepository) UpsertProductTx(ctx context.Context, req *v1.UpsertProductRequest, user string) error {
	var addApplications, deleteApplications, deleteEquipment []string
	var addEquipments []*v1.UpsertProductRequestEquipmentEquipmentuser
	var upsertPartialFlag bool

	//Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductRepositoryTx(tx)

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
		deleteEquipment = make([]string, len(deleteEquipmentUsers))
		for _, d := range deleteEquipmentUsers {
			deleteEquipment = append(deleteEquipment, d.GetEquipmentId())
		}
	}

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

	for _, app := range addApplications {
		err := pt.UpsertProductApplications(ctx, gendb.UpsertProductApplicationsParams{
			Swidtag:       req.GetSwidTag(),
			ApplicationID: app,
			Scope:         req.GetScope()})
		if err != nil {
			tx.Rollback()
			logger.Log.Error("Failed to execute UpsertProductApplications", zap.Error(err))
			return err
		}
	}

	for _, equip := range addEquipments {
		err := pt.UpsertProductEquipments(ctx, gendb.UpsertProductEquipmentsParams{
			Swidtag:     req.GetSwidTag(),
			EquipmentID: equip.EquipmentId,
			NumOfUsers:  sql.NullInt32{Int32: equip.NumUser, Valid: true},
			Scope:       req.GetScope()})
		if err != nil {
			tx.Rollback()
			logger.Log.Error("Failed to execute UpsertProductEquipments", zap.Error(err))
			return err
		}
	}

	if len(deleteApplications) > 0 {
		err = pt.DeleteProductApplications(ctx, gendb.DeleteProductApplicationsParams{
			ProductID:     req.GetSwidTag(),
			ApplicationID: deleteApplications,
			Scope:         req.GetScope(),
		})
		if err != nil {
			tx.Rollback()
			logger.Log.Error("failed to delete product-applicaiton", zap.Error(err))
			return err
		}
	}

	if len(deleteEquipment) > 0 {
		err = pt.DeleteProductEquipments(ctx, gendb.DeleteProductEquipmentsParams{
			ProductID:   req.GetSwidTag(),
			EquipmentID: deleteEquipment,
			//SCOPE BASED CHANGE
			Scope: req.GetScope(),
		})
		if err != nil {
			tx.Rollback()
			logger.Log.Error("failed to delete product-equipments", zap.Error(err))
			return err
		}
	}

	tx.Commit()
	return nil
}

//DropProductDataTx drops all the products data/ and linking in a particular scope
func (p *ProductRepository) DropProductDataTx(ctx context.Context, scope string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	pt := NewProductRepositoryTx(tx)
	if err := pt.DeleteProductsByScope(ctx, scope); err != nil {
		tx.Rollback()
		logger.Log.Error("failed to delete products data", zap.Error(err))
		return err
	}
	if err := pt.DeleteAcqrightsByScope(ctx, scope); err != nil {
		tx.Rollback()
		logger.Log.Error("failed to delete acqrights data", zap.Error(err))
		return err
	}
	if err := pt.DeleteProductAggregationByScope(ctx, scope); err != nil {
		tx.Rollback()
		logger.Log.Error("failed to delete product aggregations data", zap.Error(err))
		return err
	}
	tx.Commit()
	return nil
}
