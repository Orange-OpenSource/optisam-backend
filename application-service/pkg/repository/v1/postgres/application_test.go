package postgres

import (
	"context"
	"database/sql"
	"log"
	v1 "optisam-backend/application-service/pkg/api/v1"
	dbm "optisam-backend/application-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"testing"

	"go.uber.org/zap"
)

func upsertInstanceCleanup() {
	query := "delete from applications_instances where application_id ='a1' ;"
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Cleanup is failed for application id : a1")
	} else {
		log.Println("cleanup for Test_UpsertInstanceTX successful ")
	}

}

func Test_DropObscolenscenceDataTX(t *testing.T) {
	tests := []struct {
		name   string
		ob     *ApplicationRepository
		setup  func() error
		check  func() bool
		outErr bool
		ctx    context.Context
	}{
		{
			name:   "deletionSuccess",
			outErr: false,
			setup: func() error {
				q := "insert into applications (application_id,application_name,application_version,application_owner,application_domain,scope,obsolescence_risk)values('ttt','a','v','o','d','s1','Low');"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				q = "insert into applications_instances (application_id,instance_id,instance_environment,scope)values('ttt','iii','e','s1');"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				q = "insert into domain_criticity values(1000,'s1',1,'{d}','admin@test.com',NOW());"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				q = "insert into maintenance_time_criticity values(1000,'s1',1,48,72,'admin@test.com',NOW());"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				q = "insert into risk_matrix values(1000,'s1','admin@test.com',NOW());"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				q = "insert into risk_matrix_config values(1000,1,1,1);"
				if _, err := db.Exec(q); err != nil {
					return err
				}

				return nil
			},
			check: func() bool {
				var idata int
				q := "select critic_id from domain_criticity where critic_id = 1000 ;"
				row := db.QueryRow(q)
				if err := row.Scan(&idata); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("domain_criticity scan failed", zap.Error(err))
					return true
				} else if idata > 0 {
					return true
				}

				q = "select configuration_id from risk_matrix where configuration_id = 1000 ;"
				row = db.QueryRow(q)
				if err := row.Scan(&idata); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("risk_matrix scan failed", zap.Error(err))
					return true
				} else if idata > 0 {
					return true
				}

				q = "select maintenance_critic_id from maintenance_time_criticity where maintenance_critic_id = 1000 ;"
				row = db.QueryRow(q)
				if err := row.Scan(&idata); err != nil && err != sql.ErrNoRows {
					logger.Log.Error("maintenance_time_criticity scan failed", zap.Error(err))
					return true
				} else if idata > 0 {
					return true
				}
				return false
			},
			ctx: context.Background(),
			ob:  NewApplicationRepository(db),
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := test.ob.DropObscolenscenceDataTX(test.ctx, "s1")
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if test.check() {
				t.Errorf("application resource should be deleted")
				return
			}
		})
	}
	return
}

func Test_UpsertInstanceTX(t *testing.T) {
	type input struct {
		ctx context.Context
		arg *v1.UpsertInstanceRequest
	}
	defer upsertInstanceCleanup()
	tests := []struct {
		name   string
		ob     *ApplicationRepository
		input  input
		outErr bool
	}{
		{
			name: "Upserting_data_first_time",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertInstanceRequest{
					ApplicationId: "a1",
					InstanceId:    "i1",
					InstanceName:  "i1name",
					Products: &v1.UpsertInstanceRequestProduct{
						Operation: "add",
						ProductId: []string{"p1", "p2"},
					},
					Equipments: &v1.UpsertInstanceRequestEquipment{
						Operation:   "add",
						EquipmentId: []string{"e1", "e2"},
					},
					Scope: "s1",
				},
			},
			outErr: false,
		},
		{
			name: "Primary_key_conflict_validation_in_upsertApplicationInstance",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertInstanceRequest{
					ApplicationId: "a1",
					InstanceId:    "i1",
					InstanceName:  "i1name",
					Products: &v1.UpsertInstanceRequestProduct{
						Operation: "add",
						ProductId: []string{"p1", "p2"},
					},
					Equipments: &v1.UpsertInstanceRequestEquipment{
						Operation:   "add",
						EquipmentId: []string{"e1", "e2"},
					},
					Scope: "s1",
				},
			},
			outErr: false,
		},
		{
			name: "Upserting_old_instance_with_new_scope",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertInstanceRequest{
					ApplicationId: "a1",
					InstanceId:    "i1",
					InstanceName:  "iname",
					Products: &v1.UpsertInstanceRequestProduct{
						Operation: "add",
						ProductId: []string{"p1", "p2", "p3"},
					},
					Equipments: &v1.UpsertInstanceRequestEquipment{
						Operation:   "add",
						EquipmentId: []string{"e1", "e2"},
					},
					Scope: "s2",
				},
			},
			outErr: false,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := test.ob.UpsertInstanceTX(test.input.ctx, test.input.arg)
			log.Println("ERR", err)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
	return
}

func upsertApplicationCleanup() {
	query := "delete from applications where application_id ='a1' ;"
	_, err := db.Exec(query)
	if err != nil {
		log.Println("Cleanup is failed for application id : a1")
	} else {
		log.Println("cleanup for Test_UpsertApplication successful")
	}

}

func Test_UpsertApplication(t *testing.T) {
	type input struct {
		ctx context.Context
		arg *v1.UpsertApplicationRequest
	}
	defer upsertApplicationCleanup()
	tests := []struct {
		name   string
		ob     *ApplicationRepository
		input  input
		outErr bool
	}{
		{
			name: "Upserting_data_first_time",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertApplicationRequest{
					ApplicationId: "a1",
					Name:          "app1",
					Version:       "ver1",
					Owner:         "own1",
					Scope:         "s1",
				},
			},
			outErr: false,
		},
		{
			name: "Primary_key_conflict_validation_in_upsertApplication",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertApplicationRequest{
					ApplicationId: "a1",
					Name:          "app1",
					Version:       "ver1",
					Owner:         "own1",
					Scope:         "s1",
				},
			},
			outErr: false,
		},
		{
			name: "Old_application_upsertion_with_new_scope",
			ob:   NewApplicationRepository(db),
			input: input{
				ctx: context.Background(),
				arg: &v1.UpsertApplicationRequest{
					ApplicationId: "a1",
					Name:          "app1",
					Version:       "ver1",
					Owner:         "own1",
					Scope:         "s2",
				},
			},
			outErr: false,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := test.ob.Queries.UpsertApplication(test.input.ctx, dbm.UpsertApplicationParams{
				ApplicationID:      test.input.arg.ApplicationId,
				ApplicationName:    test.input.arg.Name,
				ApplicationOwner:   test.input.arg.Owner,
				ApplicationVersion: test.input.arg.Version,
				Scope:              test.input.arg.Scope,
			})
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
	return
}
