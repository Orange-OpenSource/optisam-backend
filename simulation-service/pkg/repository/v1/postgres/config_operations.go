// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/simulation-service/pkg/repository/v1"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	insertMetadata   = `INSERT INTO config_metadata (config_id,equipment_type,attribute_name, config_filename) VALUES($1,$2,$3,$4) RETURNING id`
	insertData       = `INSERT INTO config_data (metadata_id,attribute_value,json_data) VALUES`
	deleteMetadata   = `DELETE FROM config_metadata WHERE config_id=$1 AND id IN (`
	insertMasterdata = `INSERT INTO config_master (name,equipment_type,status,created_by,created_on,updated_by,updated_on) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`
	updateMasterData = `UPDATE config_master SET updated_by = $1, updated_on = $2`
)

// CreateConfig implements SimulationService CreateConfig function
func (r *SimulationServiceRepo) CreateConfig(ctx context.Context, masterData *v1.MasterData, data []*v1.ConfigData) (retErr error) {
	// initiating  a database transaction
	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if err := txn.Rollback(); err != nil {
				logger.Log.Error(" CreateConfig - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf(" CreateConfig - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(); err != nil {
			logger.Log.Error(" CreateConfig - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf(" CreateConfig - cannot commit txn")
		}
	}()

	var configID int32
	//Insert into master table
	err = txn.QueryRowContext(ctx, insertMasterdata, masterData.Name, masterData.EquipmentType, masterData.Status, masterData.CreatedBy, masterData.CreatedOn, masterData.UpdatedBy, masterData.UpdatedOn).Scan(&configID)

	if err != nil {
		return err
	}

	//Insert into metadata and data table
	for _, d := range data {
		//insert data into config_metadata and config_data table
		err = insertConfigData(ctx, txn, configID, masterData.EquipmentType, d.ConfigMetadata, d.ConfigValues)
		if err != nil {
			return err
		}
	}

	return nil
}

//UpdateConfig implements SimulationService UpdateConfig function
func (r *SimulationServiceRepo) UpdateConfig(ctx context.Context, configID int32, eqType string, metadataIDs []int32, data []*v1.ConfigData) (retErr error) {
	// initiating  a database transaction
	txn, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			if err := txn.Rollback(); err != nil {
				logger.Log.Error(" UpdateConfig - failed to discard txn", zap.String("reason", err.Error()))
				retErr = fmt.Errorf(" UpdateConfig - cannot discard txn")
			}
			return
		}
		if err := txn.Commit(); err != nil {
			logger.Log.Error(" UpdateConfig - failed to commit txn", zap.String("reason", err.Error()))
			retErr = fmt.Errorf(" UpdateConfig - cannot commit txn")
		}
	}()

	if len(metadataIDs) != 0 {
		deleteMetadataQuery, args := getDeleteMetadataQuery(metadataIDs, configID)

		//Delete data from metadata table
		_, err = txn.ExecContext(ctx, deleteMetadataQuery, args...)
		if err != nil {
			return err
		}
	}

	if len(data) != 0 {
		for _, d := range data {
			//insert data into config_metadata and config_data table
			err = insertConfigData(ctx, txn, configID, eqType, d.ConfigMetadata, d.ConfigValues)
			if err != nil {
				return err
			}
		}
	}

	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return fmt.Errorf("cannot find claims in context")
	}

	//Update master data
	_, err = txn.ExecContext(ctx, updateMasterData, userClaims.UserID, time.Now().UTC())
	if err != nil {
		return err
	}

	return nil
}

func getInsertConfigQuery(metadataID int32, values []*v1.ConfigValue) (string, []interface{}) {
	query := insertData
	args := make([]interface{}, 2*len(values)+1)
	queryValues := make([]string, len(values))
	args[0] = metadataID
	for i := range values {
		queryValues[i] = fmt.Sprintf("($1,$%d,$%d)", 2*i+2, 2*i+3)
		args[2*i+1], args[2*i+2] = values[i].Key, string(values[i].Value)
	}
	return query + strings.Join(queryValues, ","), args
}

func getDeleteMetadataQuery(metadataIDs []int32, configID int32) (string, []interface{}) {
	query := deleteMetadata
	args := []interface{}{
		configID,
	}
	for i := range metadataIDs {
		query += fmt.Sprintf("$1,$%v", i+2)
		args = append(args, metadataIDs[i])
		if i != len(metadataIDs)-1 {
			query += ","
		}
	}
	query += ")"
	return query, args

}

func insertConfigData(ctx context.Context, txn *sql.Tx, configID int32, eqType string, metadata *v1.Metadata, values []*v1.ConfigValue) error {
	// Insert intometadata table
	var metadataID int32
	err := txn.QueryRowContext(ctx, insertMetadata, configID, eqType, metadata.AttributeName, metadata.ConfigFileName).Scan(&metadataID)
	if err != nil {
		return err
	}
	// insert into data table
	dataQuery, args := getInsertConfigQuery(int32(metadataID), values)
	dataResult, err := txn.ExecContext(ctx, dataQuery, args...)
	if err != nil {
		return err
	}
	n, err := dataResult.RowsAffected()
	if err != nil {
		return err
	}
	if n != int64(len(values)) {
		return fmt.Errorf("repo/postgres - UpdateConfig - expected%v row to be affected, actual affected: %v", len(values), n)
	}

	return nil
}
