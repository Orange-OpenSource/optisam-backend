// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ProductAcquiredRights(t *testing.T) {
	type args struct {
		ctx     context.Context
		swidTag string
		scopes  string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    string
		want1   []*v1.ProductAcquiredRight
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC003",
				scopes:  "scope1",
			},
			want1: []*v1.ProductAcquiredRight{
				&v1.ProductAcquiredRight{
					SKU:          "ORAC003PROC",
					Metric:       "oracle.processor.standard",
					AcqLicenses:  967,
					TotalCost:    23312248,
					AvgUnitPrice: 1426,
				},
			},
		},
		{name: "SUCCESS - scope3",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "WIN3",
				scopes:  "scope3",
			},
			want1: []*v1.ProductAcquiredRight{
				&v1.ProductAcquiredRight{
					SKU:          "WIN3PROC",
					Metric:       "Windows.processor.standard",
					AcqLicenses:  967,
					TotalCost:    23312248,
					AvgUnitPrice: 1426,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.r.ProductAcquiredRights(tt.args.ctx, tt.args.swidTag, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductAcquiredRights() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == "" {
					t.Errorf("LicenseRepository.ProductAcquiredRights() - ID is empty")
				}
				compareProductAcquiredRightsAll(t, "ProductAcquiredRights", tt.want1, got1)
			}
		})
	}
}

func compareProductDataAllProducts(t *testing.T, name string, exp, act []*v1.ProductData) {
	for i := range exp {
		if idx := productDataIndex(exp[i], act); idx != -1 {
			compareProductData(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
		}
	}
}
func productDataIndex(expProduct *v1.ProductData, actProducts []*v1.ProductData) int {
	for i := range actProducts {
		if expProduct.Swidtag == actProducts[i].Swidtag {
			return i
		}
	}
	return -1
}
func compareProductAcquiredRightsAll(t *testing.T, name string, exp, act []*v1.ProductAcquiredRight) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareProductAcquiredRights(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareProductAcquiredRights(t *testing.T, name string, exp, act *v1.ProductAcquiredRight) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "ProductRight is expected to be nil")
	}

	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU should be same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric should be same", name)
	assert.Equalf(t, exp.AcqLicenses, act.AcqLicenses, "%s.AcquiredLicenses should be same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost should be same", name)
	assert.Equalf(t, exp.AvgUnitPrice, act.AvgUnitPrice, "%s.EndEqTypeID should be same", name)

}

func compareInstancesForApplicationsProductResponse(t *testing.T, name string, exp *v1.InstancesForApplicationProduct, act *v1.InstancesForApplicationProduct) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.NumOfRecords[0].TotalCnt, act.NumOfRecords[0].TotalCnt, "%s.Records are not same", name)
	compareInstancesForApplicationsProductAll(t, name+".Instances", exp.Instances, act.Instances)
}

func compareInstancesForApplicationsProductAll(t *testing.T, name string, exp []v1.InstancesForApplicationProductData, act []v1.InstancesForApplicationProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareInstanceForApplicationsProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareInstanceForApplicationsProduct(t *testing.T, name string, exp v1.InstancesForApplicationProductData, act v1.InstancesForApplicationProductData) {
	// if exp == nil && act == nil {
	// 	return
	// }
	// if exp == nil {
	// 	assert.Nil(t, act, "attribute is expected to be nil")
	// }

	assert.Equalf(t, exp.ID, act.ID, "%s.Id are not same", name)
	assert.Equalf(t, exp.Environment, act.Environment, "%s.Environment are not same", name)
	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumOfProducts, act.NumOfProducts, "%s.NumOfProducts are not same", name)
}

func compareApplicationsForProductResponse(t *testing.T, name string, exp *v1.ApplicationsForProduct, act *v1.ApplicationsForProduct) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.NumOfRecords[0].TotalCnt, act.NumOfRecords[0].TotalCnt, "%s.Records are not same", name)

	compareApplicationsForProductAll(t, name+".Applications", exp.Applications, act.Applications)
}

func compareApplicationsForProductAll(t *testing.T, name string, exp []v1.ApplicationsForProductData, act []v1.ApplicationsForProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareApplicationForProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareApplicationForProduct(t *testing.T, name string, exp v1.ApplicationsForProductData, act v1.ApplicationsForProductData) {
	// if exp == nil && act == nil {
	// 	return
	// }
	// if exp == nil {
	// 	assert.Nil(t, act, "attribute is expected to be nil")
	// }

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Owner, act.Owner, "%s.Owner are not same", name)
	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances are not same", name)
}

func stringToInterface(vals []string) []interface{} {
	interfaceSlice := make([]interface{}, len(vals))
	for i := range vals {
		interfaceSlice[i] = vals[i]
	}
	return interfaceSlice
}
