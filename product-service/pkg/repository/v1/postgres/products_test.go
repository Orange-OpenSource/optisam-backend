// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"context"
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"testing"

	"go.uber.org/zap"
)

func cleanupUpsertProductTx() {
	q := "delete from products_applications where swidtag = 'p1' ; "
	_, err := db.Exec(q)
	if err != nil {
		log.Println("Failed to delete data from products_applications for id p1")
	}

	q = "delete from products_equipments where swidtag = 'p1' ; "
	_, err = db.Exec(q)
	if err != nil {
		log.Println("Failed to delete data from products_equipments for id p1")
	}

	q = "delete from products where swidtag = 'p1' ; "
	_, err = db.Exec(q)
	if err != nil {
		log.Println("Failed to delete data from products for id p1")
	}
}

func Test_UpsertProductTx(t *testing.T) {
	defer cleanupUpsertProductTx()
	type input struct {
		ctx    context.Context
		userID string
		arg    *v1.UpsertProductRequest
	}
	tests := []struct {
		name    string
		ob      *ProductRepository
		input   input
		cleanup func(id string) error
		outErr  bool
	}{
		{
			name: "Upserting_data_first_time",
			ob:   NewProductRepository(db),
			input: input{
				ctx:    context.Background(),
				userID: "IAM",
				arg: &v1.UpsertProductRequest{
					SwidTag:  "p1",
					Name:     "p",
					Category: "pc",
					Edition:  "ped",
					Editor:   "pe",
					Version:  "pv",
					OptionOf: "po",
					Scope:    "s",
					Applications: &v1.UpsertProductRequestApplication{
						Operation:     "add",
						ApplicationId: []string{"a1", "a2"},
					},
					Equipments: &v1.UpsertProductRequestEquipment{
						Operation: "add",
						Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
							&v1.UpsertProductRequestEquipmentEquipmentuser{
								EquipmentId: "e1",
								NumUser:     int32(1),
							},
							&v1.UpsertProductRequestEquipmentEquipmentuser{
								EquipmentId: "e2",
								NumUser:     int32(2),
							},
						},
					},
				},
			},
			outErr: false,
		},
		{
			name: "Upserting_new_data_with_same_primary_key",
			ob:   NewProductRepository(db),
			input: input{
				ctx:    context.Background(),
				userID: "IAM",
				arg: &v1.UpsertProductRequest{
					SwidTag:  "p1",
					Name:     "p2",
					Category: "p2c",
					Edition:  "ped",
					Editor:   "pe",
					Version:  "pv",
					OptionOf: "po",
					Scope:    "s",
				},
			},
			outErr: true,
		},
		{
			name: "Upserting_same_swidtag_with_another_scope",
			ob:   NewProductRepository(db),
			input: input{
				ctx:    context.Background(),
				userID: "IAM",
				arg: &v1.UpsertProductRequest{
					SwidTag:  "p1",
					Name:     "p",
					Category: "pc",
					Edition:  "ped",
					Editor:   "pe",
					Version:  "pv",
					OptionOf: "po",
					Scope:    "s",
					Applications: &v1.UpsertProductRequestApplication{
						Operation:     "add",
						ApplicationId: []string{"a5", "a6"},
					},
					Equipments: &v1.UpsertProductRequestEquipment{
						Operation: "add",
						Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
							&v1.UpsertProductRequestEquipmentEquipmentuser{
								EquipmentId: "e6",
								NumUser:     int32(1),
							},
							&v1.UpsertProductRequestEquipmentEquipmentuser{
								EquipmentId: "e7",
								NumUser:     int32(2),
							},
						},
					},
				},
			},
			outErr: false,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			err := test.ob.UpsertProductTx(test.input.ctx, test.input.arg, test.input.userID)
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