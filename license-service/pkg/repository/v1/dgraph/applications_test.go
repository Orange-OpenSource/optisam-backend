// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

// import (
// 	"context"
// 	v1 "optisam-backend/license-service/pkg/repository/v1"
// 	"reflect"
// 	"testing"
// )

// func TestLicenseRepository_GetApplications(t *testing.T) {

// 	r := &LicenseRepository{dg: dgClient}

// 	ctx := context.Background()

// 	type args struct {
// 		ctx    context.Context
// 		params *v1.QueryApplications
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *LicenseRepository
// 		args    args
// 		want    *v1.ApplicationInfo
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:5,
// 					Offset:0,
// 					SortBy:"name",
// 					SortOrder: "asc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []Queryable{
// 							v1.Filter{
// 								FilterKey: "name",
// 								FilterValue: "Alerm" ,
// 							} ,
// 						}

// 					}
// 				} ,
// 			},
// 			want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:5,
// 					Offset:0,
// 					SortBy:"name",
// 					SortOrder: "desc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []Queryable{
// 							v1.Filter{
// 								FilterKey: "application_owner",
// 								FilterValue: "Pio"
// 							} ,
// 						}

// 					}
// 				} ,
// 			},
// 			want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:5,
// 					Offset:0,
// 					SortBy:"application_owner",
// 					SortOrder: "asc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []Queryable{
// 							v1.Filter{
// 								FilterKey: "application_owner",
// 								FilterValue: "Piona"
// 							} ,
// 						}

// 					}
// 				} ,
// 			},
// 			want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:5,
// 					Offset:0,
// 					SortBy:"applicationId",
// 					SortOrder: "asc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []Queryable{
// 							v1.Filter{
// 								FilterKey: "name",
// 								FilterValue: "Afr"
// 							} ,
// 						}

// 					}
// 				} ,
// 			},
// 			want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},

// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.r.GetApplications(tt.args.ctx, tt.args.params)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LicenseRepository.GetApplications() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LicenseRepository.GetApplications() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestLicenseRepository_GetProductsForApplication(t *testing.T) {

// 	r := &LicenseRepository{dg: dgClient}

// 	ctx := context.Background()

// 	type args struct {
// 		ctx context.Context
// 		id  string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *LicenseRepository
// 		args    args
// 		want    *v1.ProductsForApplication
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				id: "92"
// 			},

// 		    want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				id: "26"
// 			},

// 		    want: &v1.ListProductsResponse{Products: []*v1.Product{{Name: "Oracle Client", Version: "10.2", Category: "Other", Editor: "oracle", SwidTag: "ORAC001"}}},
// 			setup: func() (func() error, error) {

// 				if err := insertSchema(r); err != nil {
// 					return nil, err
// 				}

// 				return func() error {
// 					return deleteAllData(r)
// 				}, nil
// 			},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := tt.r.GetProductsForApplication(tt.args.ctx, tt.args.id)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LicenseRepository.GetProductsForApplication() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LicenseRepository.GetProductsForApplication() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestLicenseRepository_GetApplication(t *testing.T) {
	type args struct {
		ctx    context.Context
		appID  string
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    *v1.ApplicationDetails
		wantErr bool
	}{
		{name: "success",
			r: &LicenseRepository{
				dg: dgClient,
			},
			args: args{
				ctx:    context.Background(),
				appID:  "1",
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationDetails{
				Name:             "Acireales",
				ApplicationID:    "1",
				ApplicationOwner: "Biogercorp",
				NumOfInstances:   3,
				NumOfProducts:    2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetApplication(tt.args.ctx, tt.args.appID, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.GetApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareApplication(t, "ApplicationsForProduct", got, tt.want)
			}
		})
	}
}

func compareApplication(t *testing.T, name string, exp *v1.ApplicationDetails, act *v1.ApplicationDetails) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.ApplicationID, act.ApplicationID, "%s.ApplicationID are not same", name)
	assert.Equalf(t, exp.ApplicationOwner, act.ApplicationOwner, "%s.Owner are not same", name)
	assert.Equalf(t, exp.NumOfProducts, act.NumOfProducts, "%s.NumOfProducts are not same", name)
	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances are not same", name)
}
