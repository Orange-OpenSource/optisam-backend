// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package db_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	v1 "optisam-backend/simulation-service/pkg/repository/v1"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimulationServiceRepo_DeleteConfig(t *testing.T) {
	// Creating data using createConfig.
	createMasterData := &v1.MasterData{
		ID:            1,
		Name:          "server_1",
		Status:        1,
		EquipmentType: "Server",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData := []*v1.ConfigData{
		{
			ConfigMetadata: &v1.Metadata{
				ID:             1,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx context.Context
		arg db.DeleteConfigParams
	}
	tests := []struct {
		name    string
		r       *postgres.SimulationServiceRepo
		args    args
		setup   func(h *postgres.SimulationServiceRepo) (func() error, error)
		verify  func(h *postgres.SimulationServiceRepo) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx: context.Background(),
				arg: db.DeleteConfigParams{
					ID:     1,
					Status: 2,
				},
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData, createData)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			verify: func(h *postgres.SimulationServiceRepo) error {
				// verify config_master table data
				masterData, err := h.ListConfig(context.Background(), db.ListConfigParams{
					Status:        2,
					IsEquipType:   false,
					EquipmentType: "",
				})
				if err != nil {
					return err
				}
				index := configByName(masterData, "server_1")
				if index == -1 {
					return fmt.Errorf("config is not deleted")
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.DeleteConfig(tt.args.ctx, tt.args.arg); (err != nil) != tt.wantErr {
				t.Errorf("Queries.DeleteConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}

		})
	}
}

func TestSimulationServiceRepo_DeleteConfigData(t *testing.T) {
	// Creating data using createConfig.
	createMasterData := &v1.MasterData{
		ID:            1,
		Name:          "server_1",
		Status:        1,
		EquipmentType: "Server",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             1,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx      context.Context
		configID int32
	}
	tests := []struct {
		name    string
		r       *postgres.SimulationServiceRepo
		args    args
		setup   func(h *postgres.SimulationServiceRepo) (func() error, error)
		verify  func(h *postgres.SimulationServiceRepo) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx:      context.Background(),
				configID: 1,
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData, createData)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			verify: func(h *postgres.SimulationServiceRepo) error {

				// verify config_metadata table data
				metadata, err := h.GetMetadatabyConfigID(context.Background(), 1)
				if err != nil {
					return err
				}
				if len(metadata) == 0 {
					return nil
				}
				return fmt.Errorf("Configuration data is not deleted")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.DeleteConfigData(tt.args.ctx, tt.args.configID); (err != nil) != tt.wantErr {
				t.Errorf("Queries.DeleteConfigData() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func TestSimulationServiceRepo_ListConfig(t *testing.T) {
	// Creating data using createConfig.
	createMasterData := &v1.MasterData{
		ID:            1,
		Name:          "server_2",
		Status:        1,
		EquipmentType: "Cluster",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             3,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             4,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	createMasterData1 := &v1.MasterData{
		ID:            2,
		Name:          "server_1",
		Status:        1,
		EquipmentType: "Server",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData1 := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             1,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	createMasterData2 := &v1.MasterData{
		ID:            3,
		Name:          "server_3",
		Status:        2,
		EquipmentType: "Cluster",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData2 := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             5,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             6,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx context.Context
		arg db.ListConfigParams
	}
	tests := []struct {
		name    string
		r       *postgres.SimulationServiceRepo
		args    args
		setup   func(h *postgres.SimulationServiceRepo) (func() error, error)
		want    []db.ConfigMaster
		wantErr bool
	}{
		{name: "SUCCESS - With Equipment Type",
			r: postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx: context.Background(),
				arg: db.ListConfigParams{
					IsEquipType:   true,
					EquipmentType: "Server",
					Status:        1,
				},
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData, createData)
				if err != nil {
					return nil, err
				}
				err = createConfig(createMasterData1, createData1)
				if err != nil {
					return nil, err
				}
				err = createConfig(createMasterData2, createData2)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			want: []db.ConfigMaster{
				db.ConfigMaster{
					ID:            2,
					Name:          "server_1",
					EquipmentType: "Server",
					Status:        1,
					CreatedBy:     "admin@superuser.com",
					UpdatedBy:     "admin@superuser.com",
				},
			},
		},
		{name: "SUCCESS - Without Equipment Type",
			r: postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx: context.Background(),
				arg: db.ListConfigParams{
					IsEquipType:   false,
					EquipmentType: "",
					Status:        1,
				},
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData, createData)
				if err != nil {
					return nil, err
				}
				err = createConfig(createMasterData1, createData1)
				if err != nil {
					return nil, err
				}
				err = createConfig(createMasterData2, createData2)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			want: []db.ConfigMaster{
				db.ConfigMaster{
					ID:            1,
					Name:          "server_2",
					Status:        1,
					EquipmentType: "Cluster",
					CreatedBy:     "admin@superuser.com",
					UpdatedBy:     "admin@superuser.com",
				},
				db.ConfigMaster{
					ID:            2,
					Name:          "server_1",
					EquipmentType: "Server",
					Status:        1,
					CreatedBy:     "admin@superuser.com",
					UpdatedBy:     "admin@superuser.com",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.ListConfig(tt.args.ctx, tt.args.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Queries.ListConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMasterDataAll(t, "ListConfig", tt.want, got)
			}

		})
	}
}

func TestSimulationServiceRepo_GetMetadatabyConfigID(t *testing.T) {
	createMasterData1 := &v1.MasterData{
		ID:            2,
		Name:          "server_1",
		Status:        1,
		EquipmentType: "Server",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData1 := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             1,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx      context.Context
		configID int32
	}
	tests := []struct {
		name    string
		r       *postgres.SimulationServiceRepo
		args    args
		setup   func(h *postgres.SimulationServiceRepo) (func() error, error)
		want    []db.GetMetadatabyConfigIDRow
		wantErr bool
	}{
		{
			name: "SUCCESS",
			r:    postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx:      context.Background(),
				configID: int32(2),
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData1, createData1)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			want: []db.GetMetadatabyConfigIDRow{
				db.GetMetadatabyConfigIDRow{
					ID:             1,
					AttributeName:  "cpu",
					EquipmentType:  "Server",
					ConfigFilename: "1.csv",
				},
				db.GetMetadatabyConfigIDRow{
					ID:             2,
					AttributeName:  "cf",
					EquipmentType:  "Server",
					ConfigFilename: "2.csv",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.GetMetadatabyConfigID(tt.args.ctx, tt.args.configID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Queries.GetMetadatabyConfigID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetadataAll(t, "GetMetadatabyConfigID", tt.want, got)
			}
		})
	}
}

func TestSimulationServiceRepo_GetDatabyMetadataID(t *testing.T) {
	createMasterData1 := &v1.MasterData{
		ID:            2,
		Name:          "server_1",
		Status:        1,
		EquipmentType: "Server",
		CreatedBy:     "admin@superuser.com",
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     "admin@superuser.com",
		UpdatedOn:     time.Now().UTC(),
	}
	createData1 := []*v1.ConfigData{
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             1,
				AttributeName:  "cpu",
				ConfigFileName: "1.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		&v1.ConfigData{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				&v1.ConfigValue{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				&v1.ConfigValue{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx        context.Context
		MetadataID int32
	}
	tests := []struct {
		name    string
		r       *postgres.SimulationServiceRepo
		args    args
		setup   func(h *postgres.SimulationServiceRepo) (func() error, error)
		want    []db.GetDataByMetadataIDRow
		wantErr bool
	}{
		{
			name: "SUCCESS",
			r:    postgres.NewSimulationServiceRepository(sqldb),
			args: args{
				ctx:        context.Background(),
				MetadataID: int32(1),
			},
			setup: func(h *postgres.SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData1, createData1)
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			want: []db.GetDataByMetadataIDRow{
				db.GetDataByMetadataIDRow{
					AttributeValue: "xenon",
					JsonData:       []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				db.GetDataByMetadataIDRow{
					AttributeValue: "phenon",
					JsonData:       []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.r)
			if !assert.Empty(t, err, "no error is expected in setup") {
				return
			}
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			got, err := tt.r.GetDataByMetadataID(tt.args.ctx, tt.args.MetadataID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Queries.GetDatabyMetadataID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var expectedConfigValues json.RawMessage
			expectedConfigValues = []byte(`[{"cpu": "xenon", "cf": "1"},{"cpu": "phenon", "cf": "2"}]`)

			var data []string
			var act []byte
			for _, val := range got {
				data = append(data, string(val.JsonData))
			}
			act = []byte(`[` + strings.Join(data, ",") + `]`)

			if !tt.wantErr {
				assert.JSONEq(t, string(expectedConfigValues), string(act), "GetDatabyMetadataID")
			}
		})
	}
}

func configByName(configs []db.ConfigMaster, configName string) int {
	for i, config := range configs {
		if config.Name == configName {
			return i
		}
	}
	return -1
}

func compareMasterDataAll(t *testing.T, name string, exp, act []db.ConfigMaster) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "Metadata is expected to be nil")
	}
	if !assert.Lenf(t, act, len(exp), "expected number of elements are: %d", len(exp)) {
		return
	}
	for i := range exp {
		if idx := masterDataIndex(exp[i], act); idx != -1 {
			compareMasterData(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
		}
	}

}

func masterDataIndex(exp db.ConfigMaster, act []db.ConfigMaster) int {
	for i := range act {
		if exp.ID == act[i].ID {
			return i
		}
	}
	return -1
}

func compareMasterData(t *testing.T, name string, exp, act db.ConfigMaster) {
	assert.Equalf(t, exp.EquipmentType, act.EquipmentType, "%s %s.EquipmentType are not same", name, exp.EquipmentType)
	assert.Equalf(t, exp.Name, act.Name, "%s %s.ConfigName are not same", name, exp.Name)
	assert.Equalf(t, exp.Status, act.Status, "%s %s.Status are not same", name, exp.Status)
	assert.Equalf(t, exp.CreatedBy, act.CreatedBy, "%s %s.CreatedBy are not same", name, exp.CreatedBy)
	assert.Equalf(t, exp.UpdatedBy, act.UpdatedBy, "%s %s.UpdatedBy are not same", name, exp.UpdatedBy)
}

func compareMetadataAll(t *testing.T, name string, expected, actual []db.GetMetadatabyConfigIDRow) {
	if expected == nil && actual == nil {
		return
	}
	if expected == nil {
		assert.Nil(t, actual, "Metadata is expected to be nil")
	}
	if !assert.Lenf(t, actual, len(expected), "expected number of elements are: %d", len(expected)) {
		return
	}
	for i := range expected {
		if idx := metadataIndex(expected[i], actual); idx != -1 {
			compareMetadata(t, fmt.Sprintf("%s[%d]", name, i), expected[i], actual[idx])
		}
	}

}

func metadataIndex(exp db.GetMetadatabyConfigIDRow, act []db.GetMetadatabyConfigIDRow) int {
	for i := range act {
		if exp.ID == act[i].ID {
			return i
		}
	}
	return -1
}

func compareMetadata(t *testing.T, name string, exp, act db.GetMetadatabyConfigIDRow) {
	assert.Equalf(t, exp.ID, act.ID, "%s %s.ID are not same", name, exp.ID)
	assert.Equalf(t, exp.EquipmentType, act.EquipmentType, "%s %s.EquipmentType are not same", name, exp.EquipmentType)
	assert.Equalf(t, exp.AttributeName, act.AttributeName, "%s %s.AttributeName are not same", name, exp.AttributeName)
	assert.Equalf(t, exp.ConfigFilename, act.ConfigFilename, "%s %s.ConfigFileName are not same", name, exp.ConfigFilename)

}

func deleteConfig() error {
	deleteMaster := "DELETE FROM config_master"

	txn, err := sqldb.BeginTx(context.Background(), &sql.TxOptions{})

	if err != nil {
		return err
	}

	_, err = txn.ExecContext(context.Background(), deleteMaster)
	if err != nil {
		return err
	}
	if err = txn.Commit(); err != nil {
		return err
	}

	return nil
}

func createConfig(masterData *v1.MasterData, data []*v1.ConfigData) error {
	insertMasterdata := `INSERT INTO config_master (id,name,equipment_type,status,created_by,created_on,updated_by,updated_on) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`

	txn, err := sqldb.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	//Insert into master table
	_, err = txn.ExecContext(context.Background(), insertMasterdata, masterData.ID, masterData.Name, masterData.EquipmentType, masterData.Status, masterData.CreatedBy, masterData.CreatedOn, masterData.UpdatedBy, masterData.UpdatedOn)
	if err != nil {
		return err
	}
	// Insert into metadata and data table
	//Insert into metadata and data table
	for _, d := range data {
		//insert data into config_metadata and config_data table
		err = insertConfigData1(context.Background(), txn, masterData.ID, masterData.EquipmentType, d.ConfigMetadata, d.ConfigValues)
		if err != nil {
			return err
		}
	}
	if err = txn.Commit(); err != nil {
		return err
	}
	return nil

}

func insertConfigData1(ctx context.Context, txn *sql.Tx, configID int32, eqType string, metadata *v1.Metadata, values []*v1.ConfigValue) error {
	insertMetadata := `INSERT INTO config_metadata (id,config_id,equipment_type,attribute_name, config_filename) VALUES($1,$2,$3,$4,$5) RETURNING id`
	_, err := txn.ExecContext(ctx, insertMetadata, metadata.ID, configID, eqType, metadata.AttributeName, metadata.ConfigFileName)
	if err != nil {
		return err
	}
	// insert into data table
	dataQuery, args := getInsertConfigQuery1(metadata.ID, values)
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

func getInsertConfigQuery1(metadataID int32, values []*v1.ConfigValue) (string, []interface{}) {
	insertData := `INSERT INTO config_data (metadata_id,attribute_value,json_data) VALUES`
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
