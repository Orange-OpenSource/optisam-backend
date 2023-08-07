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

// ProductExistsForApplication implements ProductExistsForApplication function
func (l *LicenseRepository) ProductExistsForApplication(ctx context.Context, prodID, appID string, scopes ...string) (bool, error) {
	q := `{
		AppProduct(func: eq(application.id,"` + appID + `"))@filter(eq(scopes,[` + strings.Join(scopes, ",") + `])){
			count(Product:application.product@filter(eq(product.swidtag,"` + prodID + `")))
		}
	  }`

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductExistsForApplication - ", zap.String("reason", err.Error()), zap.String("query", q))
		return false, fmt.Errorf("productExistsForApplication - cannot complete query transaction")
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
		return false, fmt.Errorf("productExistsForApplication - cannot unmarshal Json object")
	}

	if len(d.AppProduct) == 0 {
		return false, nil
	}

	return true, nil

}

// ProductApplicationEquipments implements ProductApplicationEquipments function
func (l *LicenseRepository) ProductApplicationEquipments(ctx context.Context, prodID, appID string, scopes ...string) ([]*v1.Equipment, error) {
	q := `{
		var(func: eq(product.swidtag,"` + prodID + `"))@filter(eq(scopes,` + strings.Join(scopes, ",") + `) AND eq(type_name,"product")) {
		 prod_uid as  uid
		}
	  
		 var(func: eq(application.id,"` + appID + `"))@filter(eq(scopes,` + strings.Join(scopes, ",") + `) AND eq(type_name,"application")) {
			application.product@filter(uid(prod_uid)){
				Uid as uid
				}
		}
	  
		Equipments(func: uid(Uid))@filter(eq(scopes,` + strings.Join(scopes, ",") + `) AND eq(type_name,"product")) {
		  Equipment: product.equipment {
			uid
			EquipID: equipment.id
			Type: equipment.type
		  }
		}
		}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductApplicationEquipments - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("productApplicationEquipments - cannot complete query transaction")
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
		return nil, fmt.Errorf("productApplicationEquipments - cannot unmarshal Json object")
	}

	if len(d.Equipments) == 0 {
		return []*v1.Equipment{}, nil
	}

	return d.Equipments[0].Equipment, nil

}
