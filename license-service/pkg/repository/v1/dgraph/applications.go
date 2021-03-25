// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
)

//ProductExistsForApplication implements ProductExistsForApplication function
func (lr *LicenseRepository) ProductExistsForApplication(ctx context.Context, prodID, appID string, scopes ...string) (bool, error) {
	q := `{
		AppProduct(func: eq(application.id,` + appID + `))@filter(eq(scopes,[` + strings.Join(scopes, ",") + `])){
			count(Product:application.product@filter(eq(product.swidtag,` + prodID + `)))
		}
	  }`

	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductExistsForApplication - ", zap.String("reason", err.Error()), zap.String("query", q))
		return false, fmt.Errorf("ProductExistsForApplication - cannot complete query transaction")
	}

	type Object struct {
		Product int
	}

	type data struct {
		AppProduct []*Object
	}
	d := &data{}
	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("ProductExistsForApplication - ", zap.String("reason", err.Error()), zap.String("query", q))
		return false, fmt.Errorf("ProductExistsForApplication - cannot unmarshal Json object")
	}

	if len(d.AppProduct) == 0 {
		return false, nil
	}

	return true, nil

}

//ProductApplicationEquipments implements ProductApplicationEquipments function
func (lr *LicenseRepository) ProductApplicationEquipments(ctx context.Context, prodID, appID string, scopes ...string) ([]*v1.Equipment, error) {
	q := `{
		var(func: eq(application.id,` + appID + `))@filter(eq(scopes,` + strings.Join(scopes, ",") + `)) {
		 app_inst as  application.instance
		}
	  
		 var(func: eq(product.swidtag,` + prodID + `))@filter(eq(scopes,` + strings.Join(scopes, ",") + `)) {
		  prod_inst as ~instance.product@filter(uid(app_inst))
		}
	  
		var(func: uid(prod_inst)) {
		  ins_equip as instance.equipment
		}
	  
		Equipments(func: eq(product.swidtag,` + prodID + `))@filter(eq(scopes,` + strings.Join(scopes, ",") + `)) {
		  Equipment: product.equipment@filter(uid(ins_equip)) {
			uid
			EquipID: equipment.id
			Type: equipment.type
		  }
		}
		}`
	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductApplicationEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductApplicationEquipments - cannot complete query transaction")
	}
	type object struct {
		Equipment []*v1.Equipment
	}
	type data struct {
		Equipments []*object
	}
	d := &data{}
	if err := json.Unmarshal(resp.GetJson(), d); err != nil {
		logger.Log.Error("ProductApplicationEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductApplicationEquipments - cannot unmarshal Json object")
	}

	if len(d.Equipments) == 0 {
		return []*v1.Equipment{}, nil
	}

	return d.Equipments[0].Equipment, nil

}
