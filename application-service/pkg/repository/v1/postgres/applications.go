// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	"database/sql"
	"errors"
	v1 "optisam-backend/application-service/pkg/api/v1"
	gendb "optisam-backend/application-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
)

//ApplicationRepository for Dgraph
type ApplicationRepository struct {
	*gendb.Queries
	db *sql.DB
}

//NewApplicationRepository creates new Repository
func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}

//ApplicationRepository
type ApplicationRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

func NewApplicationRepositoryTx(db *sql.Tx) *ApplicationRepositoryTx {
	return &ApplicationRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

func (p *ApplicationRepository) UpsertInstanceTX(ctx context.Context, req *v1.UpsertInstanceRequest) error {
	//Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	at := NewApplicationRepositoryTx(tx)
	instance, err := at.GetApplicationInstance(ctx, req.GetInstanceId())

	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error("service/v1 - UpsertInstance - GetApplicationInstane", zap.Error(err))
		return errors.New("DBError")
	}

	if req.Products.GetOperation() == "add" {
		instance.Products = helper.AppendElementsIfNotExists(instance.Products, req.GetProducts().GetProductId())
	}

	if req.Products.GetOperation() == "delete" {
		instance.Products = helper.RemoveElements(instance.Products, req.GetProducts().GetProductId())
	}

	if req.Equipments.GetOperation() == "add" {
		instance.Equipments = helper.AppendElementsIfNotExists(instance.Equipments, req.GetEquipments().GetEquipmentId())
	}
	if req.Equipments.GetOperation() == "delete" {
		instance.Equipments = helper.RemoveElements(instance.Equipments, req.GetEquipments().GetEquipmentId())
	}
	if req.GetApplicationId() != "" {
		instance.ApplicationID = req.GetApplicationId()
	}
	if req.GetInstanceName() != "" {
		instance.InstanceEnvironment = req.GetInstanceName()
	}
	logger.Log.Sugar().Infof("products %v,equipment %v", req.GetProducts().GetProductId, req.GetEquipments().GetEquipmentId)
	err = at.UpsertApplicationInstance(ctx, gendb.UpsertApplicationInstanceParams{
		ApplicationID:       instance.ApplicationID,
		InstanceID:          req.GetInstanceId(),
		InstanceEnvironment: instance.InstanceEnvironment,
		Products:            instance.Products,
		Equipments:          instance.Equipments,
		Scope:               req.GetScope(),
	})
	if err != nil {
		_ = tx.Rollback()
	} else {
		_ = tx.Commit()
	}
	return err
}
