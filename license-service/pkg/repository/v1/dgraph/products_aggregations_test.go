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

func TestLicenseRepository_ProductAggregationsByName(t *testing.T) {
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
		t.Fatal(err)
	}

	metID, ok := assigned.Uids["metric"]
	if !ok {
		t.Fatalf("cannot metric id for metric xid: %v", "metric")
	}
	defer func() {
		err := deleteNodes(metID)
		assert.Empty(t, err, "error is not expect in deleteNode")
	}()

	prod1, err := getUIDForProductXID("WIN4")
	if err != nil {
		t.Fatal(err)
	}

	prod2, err := getUIDForProductXID("WIN5")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		ctx    context.Context
		name   string
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		setup   func() (func() error, error)
		want    *v1.ProductAggregation
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				name:   "Agg1",
				scopes: []string{"Asia"},
			},
			setup: func() (func() error, error) {
				pa := &v1.ProductAggregation{
					Name:     "Agg1",
					Editor:   "Oracle",
					Product:  "Database",
					Metric:   metID,
					Products: []string{prod1, prod2},
					ProductsFull: []*v1.ProductData{
						&v1.ProductData{
							Swidtag: "ORAC001",
							Name:    "Database",
						},
					},
				}
				repo := NewLicenseRepository(dgClient)
				prodAgg, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return nil, err
				}
				return func() error {
					if err := deleteNodes(prodAgg.ID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			want: &v1.ProductAggregation{
				Name:       "Agg1",
				Editor:     "Oracle",
				Product:    "Database",
				Metric:     metID,
				MetricName: "Windows.processor.standard",
				Products:   []string{"WIN4", "WIN5"},
				ProductsFull: []*v1.ProductData{
					&v1.ProductData{
						Swidtag: "ORAC001",
						Name:    "Database",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			got, err := tt.r.ProductAggregationsByName(tt.args.ctx, tt.args.name, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductAggregationsByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductAggregation(t, "ProductAggregation", tt.want, got)
			}
		})
	}
}

func TestLicenseRepository_DeleteProductAggregation(t *testing.T) {
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
		t.Fatal(err)
	}

	metID, ok := assigned.Uids["metric"]
	if !ok {
		t.Fatalf("cannot metric id for metric xid: %v", "metric")
	}
	defer func() {
		err := deleteNodes(metID)
		assert.Empty(t, err, "error is not expect in deleteNode")
	}()

	prod1, err := getUIDForProductXID("WIN4")
	if err != nil {
		t.Fatal(err)
	}

	prod2, err := getUIDForProductXID("WIN5")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		ctx    context.Context
		id     string
		scopes []string
	}
	tests := []struct {
		name      string
		r         *LicenseRepository
		args      args
		setup     func() (string, func() error, error)
		wantRetPa []*v1.ProductAggregation
		wantErr   bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: []string{"Asia"},
			},
			setup: func() (string, func() error, error) {
				pa := &v1.ProductAggregation{
					Name:     "Agg1",
					Editor:   "Oracle",
					Product:  "Database",
					Metric:   metID,
					Products: []string{prod1, prod2},
					ProductsFull: []*v1.ProductData{
						&v1.ProductData{
							Swidtag: "ORAC001",
							Name:    "Database",
						},
					},
				}
				repo := NewLicenseRepository(dgClient)
				prodAgg1, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return "", nil, err
				}
				pa = &v1.ProductAggregation{
					Name:     "Agg2",
					Editor:   "Oracle",
					Product:  "Server",
					Metric:   metID,
					Products: []string{prod1, prod2},
					ProductsFull: []*v1.ProductData{
						&v1.ProductData{
							Swidtag: "ORAC001",
							Name:    "Server",
						},
					},
				}

				prodAgg2, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return "", nil, err
				}

				return prodAgg1.ID, func() error {
					if err := deleteNodes(prodAgg2.ID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			wantRetPa: []*v1.ProductAggregation{
				&v1.ProductAggregation{
					Name:       "Agg2",
					Editor:     "Oracle",
					Product:    "Server",
					Metric:     metID,
					MetricName: "Windows.processor.standard",
					Products:   []string{"WIN4", "WIN5"},
					ProductsFull: []*v1.ProductData{
						&v1.ProductData{
							Swidtag: "ORAC001",
							Name:    "Server",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ID, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			gotRetPa, err := tt.r.DeleteProductAggregation(tt.args.ctx, ID, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.DeleteProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductAggregationAll(t, "ProductAggregations", tt.wantRetPa, gotRetPa)
			}
		})
	}
}

func TestLicenseRepository_ProductIDForSwidtag(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		params *v1.QueryProducts
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		setup   func() (func() error, error)
		want    string
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				id:     "ORAC001",
				params: &v1.QueryProducts{},
				scopes: []string{"scope1", "scope2", "scope3"},
			},
			setup: func() (func() error, error) {
				return func() error {
					return nil
				}, nil
			},
			want: "Not Null",
		},
		{name: "SUCCESS - acqRights filter - node found",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  "P001",
				params: &v1.QueryProducts{
					AcqFilter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "xyz",
							},
						},
					},
				},
				scopes: []string{"scope1"},
			},
			setup: func() (func() error, error) {
				prod := "P1"
				acq := "A1"
				prodblankID := blankID(prod)
				acqblankID := blankID(acq)
				nquads := []*api.NQuad{
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "product.swidtag",
						ObjectValue: stringObjectValue("P001"),
					},
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "scopes",
						ObjectValue: stringObjectValue("scope1"),
					},
					&api.NQuad{
						Subject:   prodblankID,
						Predicate: "product.acqRights",
						ObjectId:  acqblankID,
					},
					&api.NQuad{
						Subject:     acqblankID,
						Predicate:   "acqRights.metric",
						ObjectValue: stringObjectValue("xyz"),
					},
				}
				r := NewLicenseRepository(dgClient)
				mu := &api.Mutation{
					Set:       nquads,
					CommitNow: true,
				}
				txn := r.dg.NewTxn()
				assigned, err := txn.Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil
				}
				prodID, ok := assigned.Uids[prod]
				if !ok {
					return nil, nil
				}
				acqID, ok := assigned.Uids[acq]
				if !ok {
					return nil, nil
				}
				return func() error {
					if err := deleteNodes(prodID, acqID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			want: "Not null",
		},
		{name: "SUCCESS - acqRights filter - node not found",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  "P001",
				params: &v1.QueryProducts{
					AcqFilter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "abc",
							},
						},
					},
				},
				scopes: []string{"scope1"},
			},
			setup: func() (func() error, error) {
				prod := "P1"
				acq := "A1"
				prodblankID := blankID(prod)
				acqblankID := blankID(acq)
				nquads := []*api.NQuad{
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "product.swidtag",
						ObjectValue: stringObjectValue("P001"),
					},
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "scopes",
						ObjectValue: stringObjectValue("scope1"),
					},
					&api.NQuad{
						Subject:   prodblankID,
						Predicate: "product.acqRights",
						ObjectId:  acqblankID,
					},
					&api.NQuad{
						Subject:     acqblankID,
						Predicate:   "acqRights.metric",
						ObjectValue: stringObjectValue("xyz"),
					},
				}
				r := NewLicenseRepository(dgClient)
				mu := &api.Mutation{
					Set:       nquads,
					CommitNow: true,
				}
				txn := r.dg.NewTxn()
				assigned, err := txn.Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil
				}
				prodID, ok := assigned.Uids[prod]
				if !ok {
					return nil, nil
				}
				acqID, ok := assigned.Uids[acq]
				if !ok {
					return nil, nil
				}
				return func() error {
					if err := deleteNodes(prodID, acqID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			wantErr: true,
		},
		{name: "SUCCESS - acqRights filter - agg filter - node found",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  "P001",
				params: &v1.QueryProducts{
					AcqFilter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "xyz",
							},
						},
					},
					AggFilter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "abc",
							},
						},
					},
				},
				scopes: []string{"scope1"},
			},
			setup: func() (func() error, error) {
				prod := "P1"
				acq := "A1"
				met := "M1"
				prodAgg := "PAG1"
				prodblankID := blankID(prod)
				acqblankID := blankID(acq)
				metblankID := blankID(met)
				prodAggblankID := blankID(prodAgg)
				nquads := []*api.NQuad{
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "product.swidtag",
						ObjectValue: stringObjectValue("P001"),
					},
					&api.NQuad{
						Subject:     prodblankID,
						Predicate:   "scopes",
						ObjectValue: stringObjectValue("scope1"),
					},
					&api.NQuad{
						Subject:   prodblankID,
						Predicate: "product.acqRights",
						ObjectId:  acqblankID,
					},
					&api.NQuad{
						Subject:     acqblankID,
						Predicate:   "acqRights.metric",
						ObjectValue: stringObjectValue("xyz"),
					},
					&api.NQuad{
						Subject:   prodAggblankID,
						Predicate: "product_aggregation.products",
						ObjectId:  prodblankID,
					},
					&api.NQuad{
						Subject:   prodAggblankID,
						Predicate: "product_aggregation.metric",
						ObjectId:  metblankID,
					},
					&api.NQuad{
						Subject:     metblankID,
						Predicate:   "metric.name",
						ObjectValue: stringObjectValue("xyz"),
					},
				}
				r := NewLicenseRepository(dgClient)
				mu := &api.Mutation{
					Set:       nquads,
					CommitNow: true,
				}
				txn := r.dg.NewTxn()
				assigned, err := txn.Mutate(context.Background(), mu)
				if err != nil {
					return nil, nil
				}
				prodID, ok := assigned.Uids[prod]
				if !ok {
					return nil, nil
				}
				acqID, ok := assigned.Uids[acq]
				if !ok {
					return nil, nil
				}
				metID, ok := assigned.Uids[met]
				if !ok {
					return nil, nil
				}
				aggID, ok := assigned.Uids[prodAgg]
				if !ok {
					return nil, nil
				}
				return func() error {
					if err := deleteNodes(prodID, acqID, metID, aggID); err != nil {
						return err
					}
					return nil
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			got, err := tt.r.ProductIDForSwidtag(tt.args.ctx, tt.args.id, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductIDForSwidtag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != "" {
				if got == "" {
					t.Errorf("LicenseRepository.ProductIDForSwidtag() = %v ", got)
				}
			}
		})
	}
}

func TestLicenseRepository_UpdateProductAggregation(t *testing.T) {
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
		t.Fatal(err)
	}

	metID, ok := assigned.Uids["metric"]
	if !ok {
		t.Fatalf("cannot metric id for metric xid: %v", "metric")
	}
	defer func() {
		err := deleteNodes(metID)
		assert.Empty(t, err, "error is not expect in deleteNode")
	}()

	prod1, err := getUIDForProductXID("WIN1")
	if err != nil {
		t.Fatal(err)
	}

	prod2, err := getUIDForProductXID("WIN2")
	if err != nil {
		t.Fatal(err)
	}

	prod3, err := getUIDForProductXID("WIN3")
	if err != nil {
		t.Fatal(err)
	}

	prod4, err := getUIDForProductXID("WIN4")
	if err != nil {
		t.Fatal(err)
	}

	prod5, err := getUIDForProductXID("WIN5")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		ctx    context.Context
		ID     string
		upa    *v1.UpdateProductAggregationRequest
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		setup   func() (string, func() error, error)
		args    args
		verify  func(r *LicenseRepository) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: []string{"Asia"},
				upa: &v1.UpdateProductAggregationRequest{
					Name:            "ProIDC1",
					AddedProducts:   []string{prod1, prod2},
					RemovedProducts: []string{prod3, prod4},
					Product:         "pro1",
				},
			},
			setup: func() (string, func() error, error) {
				pa := &v1.ProductAggregation{
					Name:       "ProID1",
					Editor:     "Oracle",
					Product:    "Database",
					Metric:     metID,
					MetricName: "Windows.processor.standard",
					Products:   []string{prod3, prod4, prod5},
				}
				repo := NewLicenseRepository(dgClient)
				prodAgg1, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return "", nil, err
				}

				return prodAgg1.ID, func() error {
					if err := deleteNodes(prodAgg1.ID); err != nil {
						return err
					}
					return nil
				}, nil
			},
			verify: func(r *LicenseRepository) error {
				proAgg, err := r.ProductAggregationsByName(context.Background(), "ProIDC1", []string{"Asia"})
				if err != nil {
					return err
				}

				expectedProAgg := &v1.ProductAggregation{
					Name:       "ProIDC1",
					Editor:     "Oracle",
					Product:    "pro1",
					Metric:     metID,
					MetricName: "Windows.processor.standard",
					Products:   []string{"WIN1", "WIN2", "WIN5"},
				}

				compareProductAggregation(t, "ProIDC1", expectedProAgg, proAgg)
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ID, cleanup, err := tt.setup()
			if !assert.Empty(t, err, "error is not expected in setup") {
				return
			}
			defer func() {
				err := cleanup()
				assert.Empty(t, err, "error is not expected in cleanup")
			}()
			if err := tt.r.UpdateProductAggregation(tt.args.ctx, ID, tt.args.upa, tt.args.scopes); (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.UpdateProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				assert.Empty(t, tt.verify(tt.r))
			}
		})
	}
}

func compareProductAggregationAll(t *testing.T, name string, exp []*v1.ProductAggregation, act []*v1.ProductAggregation) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareProductAggregation(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareProductAggregation(t *testing.T, name string, exp *v1.ProductAggregation, act *v1.ProductAggregation) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "Product Agg is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Product, act.Product, "%s.Product are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.MetricName are not same", name)
	assert.ElementsMatchf(t, exp.Products, act.Products, "%s.Products are not same", name)
	assert.ElementsMatchf(t, exp.AcqRights, act.AcqRights, "%s.AcqRights are not same", name)
	compareProductDataAllProducts(t, fmt.Sprintf("%s.ProductsFull are not same", name), exp.ProductsFull, act.ProductsFull)
	compareAcquiredRightsAllNoOrder(t, fmt.Sprintf("%s.AcqRightsFull are not same", name), exp.AcqRightsFull, act.AcqRightsFull)
}

func compareProductDataAll(t *testing.T, name string, exp, act []*v1.ProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareProductData(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareProductDataAllNoOrder(t *testing.T, name string, exp, act []*v1.ProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		idx := productDataIdxWithSwidTag(exp[i].Swidtag, act)
		if !assert.NotEqualf(t, -1, idx, "product with sidtag if not found", exp[i].Swidtag) {
			continue
		}
		compareProductData(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
	}
}

func productDataIdxWithSwidTag(swidTag string, act []*v1.ProductData) int {
	for i := range act {
		if act[i].Swidtag == swidTag {
			return i
		}
	}
	return -1
}

func compareProductData(t *testing.T, name string, exp, act *v1.ProductData) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "Product Agg is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.Swidtag, act.Swidtag, "%%s.Product are not same", name)
	assert.Equalf(t, exp.Category, act.Category, "%s.Category are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)
	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumOfApplications, act.NumOfApplications, "%s.NumOfApplications are not same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost are not same", name)
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
