package dgraph

import (
	"fmt"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

// func TestLicenseRepository_ProductIDForSwidtag(t *testing.T) {
// 	type args struct {
// 		ctx    context.Context
// 		id     string
// 		params *v1.QueryProducts
// 		scopes string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *LicenseRepository
// 		args    args
// 		setup   func() (func() error, error)
// 		want    string
// 		wantErr bool
// 	}{
// 		{name: "SUCCESS",
// 			r: NewLicenseRepository(dgClient),
// 			args: args{
// 				ctx:    context.Background(),
// 				id:     "ORAC001",
// 				params: &v1.QueryProducts{},
// 				scopes: "scope1",
// 			},
// 			setup: func() (func() error, error) {
// 				return func() error {
// 					return nil
// 				}, nil
// 			},
// 			want: "Not Null",
// 		},
// 		{name: "SUCCESS - acqRights filter - node found",
// 			r: NewLicenseRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 				id:  "P001",
// 				params: &v1.QueryProducts{
// 					AcqFilter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
// 								FilterValue: "xyz",
// 							},
// 						},
// 					},
// 				},
// 				scopes: "scope1",
// 			},
// 			setup: func() (func() error, error) {
// 				prod := "P1"
// 				acq := "A1"
// 				prodblankID := blankID(prod)
// 				acqblankID := blankID(acq)
// 				nquads := []*api.NQuad{
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "product.swidtag",
// 						ObjectValue: stringObjectValue("P001"),
// 					},
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "scopes",
// 						ObjectValue: stringObjectValue("scope1"),
// 					},
// 					{
// 						Subject:   prodblankID,
// 						Predicate: "product.acqRights",
// 						ObjectId:  acqblankID,
// 					},
// 					{
// 						Subject:     acqblankID,
// 						Predicate:   "acqRights.metric",
// 						ObjectValue: stringObjectValue("xyz"),
// 					},
// 				}
// 				r := NewLicenseRepository(dgClient)
// 				mu := &api.Mutation{
// 					Set:       nquads,
// 					CommitNow: true,
// 				}
// 				txn := r.dg.NewTxn()
// 				assigned, err := txn.Mutate(context.Background(), mu)
// 				if err != nil {
// 					return nil, nil
// 				}
// 				prodID, ok := assigned.Uids[prod]
// 				if !ok {
// 					return nil, nil
// 				}
// 				acqID, ok := assigned.Uids[acq]
// 				if !ok {
// 					return nil, nil
// 				}
// 				return func() error {
// 					if err := deleteNodes(prodID, acqID); err != nil {
// 						return err
// 					}
// 					return nil
// 				}, nil
// 			},
// 			want: "Not null",
// 		},
// 		{name: "SUCCESS - acqRights filter - node not found",
// 			r: NewLicenseRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 				id:  "P001",
// 				params: &v1.QueryProducts{
// 					AcqFilter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
// 								FilterValue: "abc",
// 							},
// 						},
// 					},
// 				},
// 				scopes: "scope1",
// 			},
// 			setup: func() (func() error, error) {
// 				prod := "P1"
// 				acq := "A1"
// 				prodblankID := blankID(prod)
// 				acqblankID := blankID(acq)
// 				nquads := []*api.NQuad{
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "product.swidtag",
// 						ObjectValue: stringObjectValue("P001"),
// 					},
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "scopes",
// 						ObjectValue: stringObjectValue("scope1"),
// 					},
// 					{
// 						Subject:   prodblankID,
// 						Predicate: "product.acqRights",
// 						ObjectId:  acqblankID,
// 					},
// 					{
// 						Subject:     acqblankID,
// 						Predicate:   "acqRights.metric",
// 						ObjectValue: stringObjectValue("xyz"),
// 					},
// 				}
// 				r := NewLicenseRepository(dgClient)
// 				mu := &api.Mutation{
// 					Set:       nquads,
// 					CommitNow: true,
// 				}
// 				txn := r.dg.NewTxn()
// 				assigned, err := txn.Mutate(context.Background(), mu)
// 				if err != nil {
// 					return nil, nil
// 				}
// 				prodID, ok := assigned.Uids[prod]
// 				if !ok {
// 					return nil, nil
// 				}
// 				acqID, ok := assigned.Uids[acq]
// 				if !ok {
// 					return nil, nil
// 				}
// 				return func() error {
// 					if err := deleteNodes(prodID, acqID); err != nil {
// 						return err
// 					}
// 					return nil
// 				}, nil
// 			},
// 			wantErr: true,
// 		},
// 		{name: "SUCCESS - acqRights filter - agg filter - node found",
// 			r: NewLicenseRepository(dgClient),
// 			args: args{
// 				ctx: context.Background(),
// 				id:  "P001",
// 				params: &v1.QueryProducts{
// 					AcqFilter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
// 								FilterValue: "xyz",
// 							},
// 						},
// 					},
// 					AggFilter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
// 								FilterValue: "abc",
// 							},
// 						},
// 					},
// 				},
// 				scopes: "scope1",
// 			},
// 			setup: func() (func() error, error) {
// 				prod := "P1"
// 				acq := "A1"
// 				met := "M1"
// 				prodAgg := "PAG1"
// 				prodblankID := blankID(prod)
// 				acqblankID := blankID(acq)
// 				metblankID := blankID(met)
// 				prodAggblankID := blankID(prodAgg)
// 				nquads := []*api.NQuad{
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "product.swidtag",
// 						ObjectValue: stringObjectValue("P001"),
// 					},
// 					{
// 						Subject:     prodblankID,
// 						Predicate:   "scopes",
// 						ObjectValue: stringObjectValue("scope1"),
// 					},
// 					{
// 						Subject:   prodblankID,
// 						Predicate: "product.acqRights",
// 						ObjectId:  acqblankID,
// 					},
// 					{
// 						Subject:     acqblankID,
// 						Predicate:   "acqRights.metric",
// 						ObjectValue: stringObjectValue("xyz"),
// 					},
// 					{
// 						Subject:   prodAggblankID,
// 						Predicate: "product_aggregation.products",
// 						ObjectId:  prodblankID,
// 					},
// 					{
// 						Subject:   prodAggblankID,
// 						Predicate: "product_aggregation.metric",
// 						ObjectId:  metblankID,
// 					},
// 					{
// 						Subject:     metblankID,
// 						Predicate:   "metric.name",
// 						ObjectValue: stringObjectValue("xyz"),
// 					},
// 				}
// 				r := NewLicenseRepository(dgClient)
// 				mu := &api.Mutation{
// 					Set:       nquads,
// 					CommitNow: true,
// 				}
// 				txn := r.dg.NewTxn()
// 				assigned, err := txn.Mutate(context.Background(), mu)
// 				if err != nil {
// 					return nil, nil
// 				}
// 				prodID, ok := assigned.Uids[prod]
// 				if !ok {
// 					return nil, nil
// 				}
// 				acqID, ok := assigned.Uids[acq]
// 				if !ok {
// 					return nil, nil
// 				}
// 				metID, ok := assigned.Uids[met]
// 				if !ok {
// 					return nil, nil
// 				}
// 				aggID, ok := assigned.Uids[prodAgg]
// 				if !ok {
// 					return nil, nil
// 				}
// 				return func() error {
// 					if err := deleteNodes(prodID, acqID, metID, aggID); err != nil {
// 						return err
// 					}
// 					return nil
// 				}, nil
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cleanup, err := tt.setup()
// 			if !assert.Empty(t, err, "error is not expected in setup") {
// 				return
// 			}
// 			defer func() {
// 				err := cleanup()
// 				assert.Empty(t, err, "error is not expected in cleanup")
// 			}()
// 			got, err := tt.r.ProductIDForSwidtag(tt.args.ctx, tt.args.id, tt.args.params, tt.args.scopes)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LicenseRepository.ProductIDForSwidtag() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if tt.want != "" {
// 				if got == "" {
// 					t.Errorf("LicenseRepository.ProductIDForSwidtag() = %v ", got)
// 				}
// 			}
// 		})
// 	}
// }

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
