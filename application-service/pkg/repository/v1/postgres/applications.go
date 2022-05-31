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

// ApplicationRepository for Dgraph
type ApplicationRepository struct {
	*gendb.Queries
	db *sql.DB
}

// NewApplicationRepository creates new Repository
func NewApplicationRepository(db *sql.DB) *ApplicationRepository {
	return &ApplicationRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}

// ApplicationRepositoryTx ...
type ApplicationRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

// NewApplicationRepositoryTx ...
func NewApplicationRepositoryTx(db *sql.Tx) *ApplicationRepositoryTx {
	return &ApplicationRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}

// UpsertInstanceTX ...
func (p *ApplicationRepository) UpsertInstanceTX(ctx context.Context, req *v1.UpsertInstanceRequest) error {
	// Create Transaction
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	at := NewApplicationRepositoryTx(tx)
	instance, err := at.GetApplicationInstance(ctx, gendb.GetApplicationInstanceParams{
		InstanceID: req.InstanceId,
		Scope:      req.Scope,
	})

	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error("service/v1 - UpsertInstance - GetApplicationInstance", zap.Error(err))
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
		_ = tx.Rollback() // nolint: errcheck
	} else {
		_ = tx.Commit()
	}
	return err
}

// DropApplicationDataTX drops all the applications and linking data from a particular scope
func (p *ApplicationRepository) DropApplicationDataTX(ctx context.Context, scope string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	at := NewApplicationRepositoryTx(tx)
	if err := at.DeleteApplicationsByScope(ctx, scope); err != nil {
		logger.Log.Error("failed to delete application data", zap.Error(err))
		if err = tx.Rollback(); err != nil {
			logger.Log.Error("Rollback is failed for application data", zap.Error(err))
		}
		return err

	}
	if err := at.DeleteInstancesByScope(ctx, scope); err != nil {
		logger.Log.Error("failed to delete application,instance,equipment linking data", zap.Error(err))
		if err = tx.Rollback(); err != nil {
			logger.Log.Error("Failed to rollback application, instance, equipment linking", zap.Error(err))
		}
		return err

	}
	return tx.Commit()
}

// DropObscolenscenceDataTX drops all obscelence config from a particular scope
func (p *ApplicationRepository) DropObscolenscenceDataTX(ctx context.Context, scope string) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Log.Error("Failed to start Transaction", zap.Error(err))
		return err
	}
	at := NewApplicationRepositoryTx(tx)

	// delete domain criticity
	if err = at.DeleteDomainCriticityByScope(ctx, scope); err != nil {
		logger.Log.Error("failed to delete domain criticity", zap.Error(err))
		if err = tx.Rollback(); err != nil {
			logger.Log.Error("failed to rollback domain criticity", zap.Error(err))
		}
		return err
	}

	// delete maintenance criticity
	if err = at.DeleteMaintenanceCirticityByScope(ctx, scope); err != nil {
		logger.Log.Error("failed to delete maintenance ciritcity", zap.Error(err))
		if err = tx.Rollback(); err != nil {
			logger.Log.Error("failed to rollback maintenance ciritcity", zap.Error(err))
		}
		return err

	}

	// delete risk matrix, cascade deletion of risk matric config
	if err = at.DeleteRiskMatricbyScope(ctx, scope); err != nil {
		logger.Log.Error("failed to delete risk metric and its config", zap.Error(err))
		if err = tx.Rollback(); err != nil {
			logger.Log.Error("failed to rollback metric and config", zap.Error(err))
		}
		return err

	}
	if err = tx.Commit(); err != nil {
		logger.Log.Error("Failed to commit", zap.Error(err))
		return err
	}
	return nil
}
