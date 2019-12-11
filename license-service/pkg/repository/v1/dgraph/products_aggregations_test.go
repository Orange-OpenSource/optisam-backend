// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/protos/api"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_CreateProductAggregation(t *testing.T) {
	type args struct {
		ctx    context.Context
		pa     *v1.ProductAggregation
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		verify  func(repo *LicenseRepository, wantAgg *v1.ProductAggregation) error
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				pa: &v1.ProductAggregation{
					Name:     "Agg1",
					Editor:   "Oracle",
					Product:  "Database",
					Metric:   "0x999",
					Products: []string{"0x345", "0x346"},
				},
				scopes: []string{"Asia", "France"},
			},
			verify: func(repo *LicenseRepository, wantAgg *v1.ProductAggregation) error {
				gotAgg, err := repo.ListProductAggregations(context.Background(), []string{"Asia", "France"})
				if err != nil {
					return err
				}
				if len(gotAgg) < 1 {
					return errors.New("no data found")
				}
				compareProductAggregation(t, "ProductAggregation", wantAgg, gotAgg[0])
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetPa, err := tt.r.CreateProductAggregation(tt.args.ctx, tt.args.pa, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.CreateProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer func() {
				err := deleteNodes(gotRetPa.ID, "0x345", "0x346", "0x999")
				assert.Empty(t, err, "error is not expect in deleteNode")
			}()

			err = tt.verify(tt.r, gotRetPa)
			if !assert.Empty(t, err, "error is not expect in verify") {
				return
			}
		})
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
	assert.ElementsMatchf(t, exp.Products, act.Products, "%s.Products are not same", name)

}

func TestLicenseRepository_ProductAggregationsByName(t *testing.T) {
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
					Metric:   "0x999",
					Products: []string{"0x345", "0x346"},
				}
				repo := NewLicenseRepository(dgClient)
				prodAgg, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return nil, err
				}
				return func() error {
					if err := deleteNodes(prodAgg.ID, "0x999", "0x345", "0x346"); err != nil {
						return err
					}
					return nil
				}, nil
			},
			want: &v1.ProductAggregation{
				Name:     "Agg1",
				Editor:   "Oracle",
				Product:  "Database",
				Metric:   "0x999",
				Products: []string{"0x345", "0x346"},
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
					Metric:   "0x999",
					Products: []string{"0x345", "0x346"},
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
					Metric:   "0x998",
					Products: []string{"0x445", "0x446"},
				}

				prodAgg2, err := repo.CreateProductAggregation(context.Background(), pa, []string{"Asia"})
				if err != nil {
					return "", nil, err
				}

				return prodAgg1.ID, func() error {
					if err := deleteNodes("0x999", "0x345", "0x346", prodAgg2.ID, "0x998", "0x445", "0x446"); err != nil {
						return err
					}
					return nil
				}, nil
			},
			wantRetPa: []*v1.ProductAggregation{
				&v1.ProductAggregation{
					Name:     "Agg2",
					Editor:   "Oracle",
					Product:  "Server",
					Metric:   "0x998",
					Products: []string{"0x445", "0x446"},
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

func compareProductAggregationAll(t *testing.T, name string, exp []*v1.ProductAggregation, act []*v1.ProductAggregation) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareProductAggregation(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
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
						Predicate: "product_aggreagtion.products",
						ObjectId:  prodblankID,
					},
					&api.NQuad{
						Subject:   prodAggblankID,
						Predicate: "product_aggreagtion.metric",
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
