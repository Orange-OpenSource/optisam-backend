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

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ProductAggregationDetails(t *testing.T) {
	acquiredRights := []*v1.AcquiredRights{
		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "WIN1PROC",
			SwidTag:                        "WIN1",
			ProductName:                    "Windows Client",
			Editor:                         "Windows",
			Metric:                         "Windows.processor.standard",
			AcquiredLicensesNumber:         1016,
			LicensesUnderMaintenanceNumber: 1008,
			AvgLicenesUnitPrice:            2042,
			AvgMaintenanceUnitPrice:        14294,
			TotalPurchaseCost:              2074672,
			TotalMaintenanceCost:           14408352,
			TotalCost:                      35155072,
		},
		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "WIN2PROC",
			SwidTag:                        "WIN2",
			ProductName:                    "Windows XML Development Kit",
			Editor:                         "Windows",
			Metric:                         "Windows.processor.standard",
			AcquiredLicensesNumber:         181,
			LicensesUnderMaintenanceNumber: 181,
			AvgLicenesUnitPrice:            1759,
			AvgMaintenanceUnitPrice:        12313,
			TotalPurchaseCost:              318379,
			TotalMaintenanceCost:           2228653,
			TotalCost:                      5412443,
		},
	}

	products := []*v1.ProductData{
		&v1.ProductData{
			Name:              "Windows Instant Client",
			Version:           "9.2.0.8.0",
			Category:          "Other",
			Editor:            "Windows",
			Swidtag:           "WIN1",
			NumOfEquipments:   3,
			NumOfApplications: 1,
			TotalCost:         70310144,
		},
		&v1.ProductData{
			Name:              "Windows SGBD Noyau",
			Version:           "9.2.0.8.0",
			Category:          "Database",
			Editor:            "Windows",
			Swidtag:           "WIN2",
			NumOfEquipments:   5,
			NumOfApplications: 2,
			TotalCost:         5412443,
		},
	}

	aggregations := []*v1.ProductAggregation{
		&v1.ProductAggregation{
			ID:                "",
			Name:              "agg1",
			Editor:            "Windows",
			Product:           "Windows Instant Client",
			Metric:            "Windows.processor.standard",
			NumOfApplications: 3,
			NumOfEquipments:   8,
			TotalCost:         40567515,
			Products:          []string{"WIN1", "WIN2"},
			ProductsFull:      []*v1.ProductData{products[0], products[1]},
			AcqRights:         []string{"WIN1PROC", "WIN2PROC"},
			AcqRightsFull:     []*v1.AcquiredRights{acquiredRights[0], acquiredRights[1]},
		},
	}

	cleanup, err := func() (func() error, error) {
		mu := &api.Mutation{
			CommitNow: true,
			Set: []*api.NQuad{
				&api.NQuad{
					Subject:     blankID("metric"),
					Predicate:   "type_name",
					ObjectValue: stringObjectValue("metric"),
				},
				&api.NQuad{
					Subject:     blankID("metric"),
					Predicate:   "metric.name",
					ObjectValue: stringObjectValue("Windows.processor.standard"),
				},
			},
		}

		assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
		if err != nil {
			return nil, err
		}

		metID, ok := assigned.Uids["metric"]
		if !ok {
			return nil, fmt.Errorf("cannot metric id for metric xid: %v", "metric")
		}
		prod1, err := getUIDForProductXID("WIN1")
		if err != nil {
			return nil, err
		}

		prod2, err := getUIDForProductXID("WIN2")
		if err != nil {
			return nil, err
		}

		pa := &v1.ProductAggregation{
			Name:     "agg1",
			Editor:   "Windows",
			Product:  "Windows Instant Client",
			Metric:   metID,
			Products: []string{prod1, prod2},
		}
		repo := NewLicenseRepository(dgClient)
		prodAgg, err := repo.CreateProductAggregation(context.Background(), pa, []string{"scope3"})
		if err != nil {
			return nil, err
		}
		return func() error {
			if err := deleteNodes(prodAgg.ID, metID); err != nil {
				return err
			}
			return nil
		}, nil
	}()

	if !assert.Empty(t, err, "error is not expect in setup") {
		return
	}

	defer func() {
		err := cleanup()
		assert.Empty(t, err, "error is not expect in cleanup")
	}()
	type args struct {
		ctx    context.Context
		name   string
		params *v1.QueryProductAggregations
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		want    *v1.ProductAggregation
		wantErr bool
	}{
		{name: "success",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				name:   "agg1",
				scopes: []string{"scope3"},
			},
			want: aggregations[0],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.lr.ProductAggregationDetails(tt.args.ctx, tt.args.name, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductAggregationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			compareProductAggregation(t, "productAggregation", tt.want, got)
		})
	}
}

func compareApplicationForProductAggregationAll(t *testing.T, name string, exp []*v1.ApplicationsForProductData, act []*v1.ApplicationsForProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareApplicationForProductAggregation(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareApplicationForProductAggregation(t *testing.T, name string, exp *v1.ApplicationsForProductData, act *v1.ApplicationsForProductData) {
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Owner, act.Owner, "%s.Owner are not same", name)
	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances are not same", name)
}
