package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/simulation-service/pkg/repository/v1"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimulationServiceRepo_CreateConfig(t *testing.T) {
	type args struct {
		ctx        context.Context
		masterData *v1.MasterData
		data       []*v1.ConfigData
		scope      string
	}
	tests := []struct {
		name    string
		r       *SimulationServiceRepo
		args    args
		setup   func(h *SimulationServiceRepo) func() error
		verify  func(h *SimulationServiceRepo) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewSimulationServiceRepository(sqldb),
			args: args{
				ctx: context.Background(),
				masterData: &v1.MasterData{
					Name:          "server_1",
					Status:        1,
					EquipmentType: "Server",
					CreatedBy:     "admin@superuser.com",
					CreatedOn:     time.Now().UTC(),
					UpdatedBy:     "admin@superuser.com",
					UpdatedOn:     time.Now().UTC(),
				},
				scope: "scope1",
				data: []*v1.ConfigData{
					{
						ConfigMetadata: &v1.Metadata{
							AttributeName:  "cpu",
							ConfigFileName: "1.csv",
						},
						ConfigValues: []*v1.ConfigValue{
							{
								Key:   "xenon",
								Value: []byte(`{"cpu":"xenon","cf":"1"}`),
							},
							{
								Key:   "phenon",
								Value: []byte(`{"cpu":"phenon","cf":"2"}`),
							},
						},
					},
					{
						ConfigMetadata: &v1.Metadata{
							AttributeName:  "cf",
							ConfigFileName: "2.csv",
						},
						ConfigValues: []*v1.ConfigValue{
							{
								Key:   "1",
								Value: []byte(`{"cf":"1"}`),
							},
							{
								Key:   "2",
								Value: []byte(`{"cf":"2"}`),
							},
						},
					},
				},
			},
			setup: func(h *SimulationServiceRepo) func() error {
				return func() error {
					return deleteConfig()
				}
			},
			verify: func(h *SimulationServiceRepo) error {
				// verify config_master table data
				masterData, err := h.ListConfig(context.Background(), db.ListConfigParams{
					Status:        1,
					IsEquipType:   false,
					EquipmentType: "",
					Scope:         "scope1",
				})
				if err != nil {
					return err
				}
				index := configByName(masterData, "server_1")
				if index == -1 {
					return fmt.Errorf("config master data does not exists")
				}
				var expectedMasterData db.ConfigMaster
				expectedMasterData = db.ConfigMaster{
					Name:          "server_1",
					Status:        1,
					EquipmentType: "Server",
					CreatedBy:     "admin@superuser.com",
					UpdatedBy:     "admin@superuser.com",
				}
				compareMasterData(t, "CreateConfig", expectedMasterData, masterData[index])

				// verify config_metadata table data
				actualMetadataAll, err := h.GetMetadatabyConfigID(context.Background(), masterData[index].ID)
				if err != nil {
					return err
				}
				var expectedMetadataAll []db.GetMetadatabyConfigIDRow
				expectedMetadataAll = []db.GetMetadatabyConfigIDRow{
					{
						EquipmentType:  "Server",
						AttributeName:  "cpu",
						ConfigFilename: "1.csv",
					},
					{
						EquipmentType:  "Server",
						AttributeName:  "cf",
						ConfigFilename: "2.csv",
					},
				}

				compareMetadataAll(t, "CreateConfig", expectedMetadataAll, actualMetadataAll)

				//verify config_data table data
				actualConfigValue1, err := h.GetDataByMetadataID(context.Background(), actualMetadataAll[0].ID)
				actualConfigValue2, err := h.GetDataByMetadataID(context.Background(), actualMetadataAll[1].ID)

				var expectedConfigValues []byte
				var act []byte
				expectedConfigValues = []byte(`[{"cpu": "xenon", "cf": "1"},{"cpu": "phenon", "cf": "2"}]`)

				if actualMetadataAll[0].AttributeName == "cpu" {
					var data []string
					for _, val := range actualConfigValue1 {
						data = append(data, string(val.JsonData))
					}
					act = []byte(`[` + strings.Join(data, ",") + `]`)
				} else {
					var data []string
					for _, val := range actualConfigValue2 {
						data = append(data, string(val.JsonData))
					}
					act = []byte(`[` + strings.Join(data, ",") + `]`)
				}

				assert.JSONEq(t, string(expectedConfigValues), string(act), "CreateConfig")

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(tt.r)
			defer func() {
				if !assert.Empty(t, cleanup(), "no error is expected from cleanup") {
					return
				}
			}()
			if err := tt.r.CreateConfig(tt.args.ctx, tt.args.masterData, tt.args.data, tt.args.scope); (err != nil) != tt.wantErr {
				t.Errorf("SimulationServiceRepo.CreateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func TestSimulationServiceRepo_UpdateConfig(t *testing.T) {
	//context
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
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
				{
					Key:   "xenon",
					Value: []byte(`{"cpu":"xenon","cf":"1"}`),
				},
				{
					Key:   "phenon",
					Value: []byte(`{"cpu":"phenon","cf":"2"}`),
				},
			},
		},
		{
			ConfigMetadata: &v1.Metadata{
				ID:             2,
				AttributeName:  "cf",
				ConfigFileName: "2.csv",
			},
			ConfigValues: []*v1.ConfigValue{
				{
					Key:   "1",
					Value: []byte(`{"cf":"1"}`),
				},
				{
					Key:   "2",
					Value: []byte(`{"cf":"2"}`),
				},
			},
		},
	}
	type args struct {
		ctx         context.Context
		configID    int32
		eqType      string
		updatedBy   string
		metadataIDs []int32
		data        []*v1.ConfigData
		scope       string
	}
	tests := []struct {
		name    string
		r       *SimulationServiceRepo
		args    args
		setup   func(h *SimulationServiceRepo) (func() error, error)
		verify  func(h *SimulationServiceRepo) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewSimulationServiceRepository(sqldb),
			args: args{
				ctx:         ctx,
				configID:    1,
				eqType:      "Server",
				metadataIDs: []int32{1, 2},
				scope:       "scope1",
				updatedBy:   "admin@superuser.com",
				data: []*v1.ConfigData{
					{
						ConfigMetadata: &v1.Metadata{
							AttributeName:  "cpu",
							ConfigFileName: "3.csv",
						},
						ConfigValues: []*v1.ConfigValue{
							{
								Key:   "xenon",
								Value: []byte(`{"cpu":"xenon","cf":"1"}`),
							},
							{
								Key:   "phenon",
								Value: []byte(`{"cpu":"phenon","cf":"2"}`),
							},
						},
					},
					{
						ConfigMetadata: &v1.Metadata{
							AttributeName:  "cf",
							ConfigFileName: "4.csv",
						},
						ConfigValues: []*v1.ConfigValue{
							{
								Key:   "1",
								Value: []byte(`{"cf":"1"}`),
							},
							{
								Key:   "2",
								Value: []byte(`{"cf":"2"}`),
							},
						},
					},
				},
			},
			setup: func(h *SimulationServiceRepo) (func() error, error) {
				err := createConfig(createMasterData, createData, "scope1")
				if err != nil {
					return nil, err
				}
				return func() error {
					return deleteConfig()
				}, nil
			},
			verify: func(h *SimulationServiceRepo) error {
				// verify config_master table data
				masterData, err := h.ListConfig(context.Background(), db.ListConfigParams{
					Status:        1,
					IsEquipType:   false,
					EquipmentType: "",
					Scope:         "scope1",
				})
				if err != nil {
					return err
				}
				index := configByName(masterData, "server_1")
				if index == -1 {
					return fmt.Errorf("config master data does not exists")
				}

				// verify config_metadata table data
				actualMetadataAll, err := h.GetMetadatabyConfigID(context.Background(), masterData[index].ID)
				if err != nil {
					return err
				}
				var expectedMetadataAll []db.GetMetadatabyConfigIDRow
				expectedMetadataAll = []db.GetMetadatabyConfigIDRow{
					{
						EquipmentType:  "Server",
						AttributeName:  "cpu",
						ConfigFilename: "3.csv",
					},
					{
						EquipmentType:  "Server",
						AttributeName:  "cf",
						ConfigFilename: "4.csv",
					},
				}

				compareMetadataAll(t, "CreateConfig", expectedMetadataAll, actualMetadataAll)

				//verify config_data table data
				actualConfigValue1, err := h.GetDataByMetadataID(context.Background(), actualMetadataAll[0].ID)
				actualConfigValue2, err := h.GetDataByMetadataID(context.Background(), actualMetadataAll[1].ID)

				var expectedConfigValues []byte
				var act []byte
				expectedConfigValues = []byte(`[{"cpu": "xenon", "cf": "1"},{"cpu": "phenon", "cf": "2"}]`)

				if actualMetadataAll[0].AttributeName == "cpu" {
					var data []string
					for _, val := range actualConfigValue1 {
						data = append(data, string(val.JsonData))
					}
					act = []byte(`[` + strings.Join(data, ",") + `]`)
				} else {
					var data []string
					for _, val := range actualConfigValue2 {
						data = append(data, string(val.JsonData))
					}
					act = []byte(`[` + strings.Join(data, ",") + `]`)
				}

				assert.JSONEq(t, string(expectedConfigValues), string(act), "CreateConfig")

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
			if err := tt.r.UpdateConfig(tt.args.ctx, tt.args.configID, tt.args.eqType, tt.args.updatedBy, tt.args.metadataIDs, tt.args.data, tt.args.scope); (err != nil) != tt.wantErr {
				t.Errorf("SimulationServiceRepo.UpdateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func configByName(configs []db.ListConfigRow, configName string) int {
	for i, config := range configs {
		if config.Name == configName {
			return i
		}
	}
	return -1
}

// func compareMasterDataAll(t *testing.T, name string, exp, act []db.ConfigMaster) {
// 	if exp == nil && act == nil {
// 		return
// 	}
// 	if exp == nil {
// 		assert.Nil(t, act, "Metadata is expected to be nil")
// 	}
// 	if !assert.Lenf(t, act, len(exp), "expected number of elements are: %d", len(exp)) {
// 		return
// 	}
// 	for i := range exp {
// 		if idx := masterDataIndex(exp[i], act); idx != -1 {
// 			compareMasterData(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
// 		}
// 	}

// }

func masterDataIndex(exp db.ConfigMaster, act []db.ConfigMaster) int {
	for i := range act {
		if exp.ID == act[i].ID {
			return i
		}
	}
	return -1
}

func compareMasterData(t *testing.T, name string, exp db.ConfigMaster, act db.ListConfigRow) {
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

func createConfig(masterData *v1.MasterData, data []*v1.ConfigData, scope string) error {
	insertMasterdata := `INSERT INTO config_master (id,name,equipment_type,status,created_by,created_on,updated_by,updated_on,scope) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`

	txn, err := sqldb.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	//Insert into master table
	_, err = txn.ExecContext(context.Background(), insertMasterdata, masterData.ID, masterData.Name, masterData.EquipmentType, masterData.Status, masterData.CreatedBy, masterData.CreatedOn, masterData.UpdatedBy, masterData.UpdatedOn, scope)
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
