// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strconv"

	"go.uber.org/zap"
)

// ParentsHirerachyForEquipment ...
func (r *LicenseRepository) ParentsHirerachyForEquipment(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, scopes []string) (*v1.Equipment, error) {
	q := `{
		ParentsHirerachy(func: eq(equipment.id,` + equipID + `) , first: 1) @recurse(depth: ` + strconv.Itoa(int(hirearchyLevel)) + `, loop: false) ` + agregateFilters(scopeFilters(scopes)) + ` {
			ID: uid
		 	EquipID: equipment.id
			Type: equipment.type
			Parent:equipment.parent
		}
	}`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ParentsHirerachyForEquipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ParentsHirerachyForEquipment - cannot complete query transaction")
	}
	type eq struct {
		ID      string
		EquipID string
		Type    string
		Parent  []*eq
	}
	type data struct {
		ParentsHirerachy []*eq
	}

	d := &data{}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		logger.Log.Error("ParentsHirerachyForEquipment - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ParentsHirerachyForEquipment - cannot unmarshal Json object")
	}

	if len(d.ParentsHirerachy) == 0 {
		return nil, v1.ErrNodeNotFound
	}
	equip := d.ParentsHirerachy[0]
	equipment := &v1.Equipment{
		ID:      equip.ID,
		EquipID: equip.EquipID,
		Type:    equip.Type,
	}

	tmp := equipment

	for len(equip.Parent) != 0 {
		equip = equip.Parent[0]
		tmp.Parent = &v1.Equipment{
			ID:      equip.ID,
			EquipID: equip.EquipID,
			Type:    equip.Type,
		}
		tmp = tmp.Parent
	}
	return equipment, nil
}

// ProductsForEquipmentForMetricOracleProcessorStandard gives products for oracle processor.standard
func (r *LicenseRepository) ProductsForEquipmentForMetricOracleProcessorStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *v1.MetricOPSComputed, scopes []string) ([]*v1.ProductData, error) {
	return r.productsForEquipmentForMetric(ctx, equipID, equipType, hirearchyLevel, metric.Name, scopes)
}

// ProductsForEquipmentForMetricOracleNUPStandard gives products for oracle processor.standard
func (r *LicenseRepository) ProductsForEquipmentForMetricOracleNUPStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *v1.MetricNUPComputed, scopes []string) ([]*v1.ProductData, error) {
	return r.productsForEquipmentForMetric(ctx, equipID, equipType, hirearchyLevel, metric.Name, scopes)
}

// ProductsForEquipmentForMetricIPSStandard gives products for oracle processor.standard
func (r *LicenseRepository) ProductsForEquipmentForMetricIPSStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *v1.MetricIPSComputed, scopes []string) ([]*v1.ProductData, error) {
	return r.productsForEquipmentForMetric(ctx, equipID, equipType, hirearchyLevel, metric.Name, scopes)
}

// ProductsForEquipmentForMetricSAGStandard gives products for oracle processor.standard
func (r *LicenseRepository) ProductsForEquipmentForMetricSAGStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *v1.MetricSPSComputed, scopes []string) ([]*v1.ProductData, error) {
	return r.productsForEquipmentForMetric(ctx, equipID, equipType, hirearchyLevel, metric.Name, scopes)
}

func (r *LicenseRepository) productsForEquipmentForMetric(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metricName string, scopes []string) ([]*v1.ProductData, error) {
	q := `{
		var (func:eq(equipment.id,` + equipID + `))@recurse(depth:  ` + strconv.Itoa(int(hirearchyLevel)) + `, loop: false) ` + agregateFilters(scopeFilters(scopes)) + `{
			id as  ~product.equipment
			~equipment.parent
		}
		pid as var(func:uid(id))@cascade{
			product.acqRights @filter(eq(acqRights.metric,` + metricName + `))
		}
		Products (func:uid(pid)){
			Name :              product.name
			Version  :          product.version
			Category :          product.category
			Editor :            product.editor
			Swidtag :           product.swidtag
		}  
	  }`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ProductsForEquipmentForMetricOracleProcessorStandard - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductsForEquipmentForMetricOracleProcessorStandard - cannot complete query transaction")
	}
	type data struct {
		Products []*v1.ProductData
	}
	prodList := &data{}
	if err := json.Unmarshal(resp.GetJson(), &prodList); err != nil {
		logger.Log.Error("ProductsForEquipmentForMetricOracleProcessorStandard - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("ProductsForEquipmentForMetricOracleProcessorStandard - cannot unmarshal Json object")
	}
	if len(prodList.Products) == 0 {
		return nil, v1.ErrNoData
	}
	return prodList.Products, nil
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll implements license.ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll
func (r *LicenseRepository) ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx context.Context, equipID, equipType string, mat *v1.MetricOPSComputed, scopes []string) (int64, float64, error) {
	templ, ok := r.templates[opsEquipTemplate]
	if !ok {
		return 0, 0, errors.New("dgraph/ComputedLicensesForEquipmentForMetricOracleProcessorStandard - cannot find template for:  " + string(opsEquipTemplate))
	}
	q, err := queryBuilderEquipOPS(mat, templ, equipID, equipType)
	if err != nil {
		logger.Log.Error("dgraph/ComputedLicensesForEquipmentForMetricOracleProcessorStandard - queryBuilderEquipOPS", zap.Error(err))
		return 0, 0, errors.New("dgraph/ComputedLicensesForEquipmentForMetricOracleProcessorStandard - query cannot be built")
	}
	//fmt.Println(q)
	licenses, err := r.licensesForQueryAll(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/ComputedLicensesForEquipmentForMetricOracleProcessorStandard - query failed", zap.Error(err), zap.String("query", q))
		return 0, 0, errors.New("dgraph/ComputedLicensesForEquipmentForMetricOracleProcessorStandard - query failed")
	}

	return int64(licenses.Licenses), licenses.LicensesNoCeil, nil
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandard gives licenses for product
func (r *LicenseRepository) ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx context.Context, equipID, equipType string, mat *v1.MetricOPSComputed, scopes []string) (int64, error) {
	l, _, err := r.ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx, equipID, equipType, mat, scopes)
	if err != nil {
		return 0, err
	}
	return l, nil
}

//UsersForEquipmentForMetricOracleNUP implements License UsersForEquipmentForMetricOracleNUP function
func (r *LicenseRepository) UsersForEquipmentForMetricOracleNUP(ctx context.Context, equipID, equipType, productID string, hirearchyLevel uint8, metric *v1.MetricNUPComputed, scopes []string) ([]*v1.User, error) {
	q := `{
		var(func:eq(equipment.id,"` + equipID + `"))@recurse(depth: ` + strconv.Itoa(int(hirearchyLevel)) + `, loop: false)` + agregateFilters(scopeFilters(scopes)) + `{
		  userIDs as  equipment.users
		  ~equipment.parent
		}
		var(func:eq(product.swidtag,"` + productID + `")){
		  productUserIDs as product.users
		}
	  
		Users(func:uid(productUserIDs))@filter(uid(userIDs)){
		  ID : uid
		  UserID : users.id
		  UserCount : users.count
		}
	  }
	`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("UsersForEquipmentForMetricOracleNUP - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("UsersForEquipmentForMetricOracleNUP - cannot complete query transaction")
	}
	type data struct {
		Users []*v1.User
	}
	userInstances := &data{}
	if err := json.Unmarshal(resp.GetJson(), &userInstances); err != nil {
		logger.Log.Error("UsersForEquipmentForMetricOracleNUP - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, fmt.Errorf("UsersForEquipmentForMetricOracleNUP - cannot unmarshal Json object")
	}
	if len(userInstances.Users) == 0 {
		return nil, v1.ErrNoData
	}
	return userInstances.Users, nil
}
