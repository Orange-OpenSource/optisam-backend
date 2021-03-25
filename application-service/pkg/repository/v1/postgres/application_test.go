// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
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
