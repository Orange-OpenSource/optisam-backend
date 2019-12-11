// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestLicenseRepository_GetProducts(t *testing.T) {

// 	r := &LicenseRepository{dg: dgClient}

// 	ctx := context.Background()

// 	type args struct {
// 		ctx    context.Context
// 		params *v1.QueryProducts
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *LicenseRepository
// 		args    args
// 		want    *v1.ProductInfo
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "name",
// 					SortOrder: "asc",
// 				},
// 			},
// 			want: &v1.ProductInfo{
// 				NumOfRecords: []v1.TotalRecords{
// 					v1.TotalRecords{
// 						TotalCnt: 10,
// 					},
// 				},
// 				Products: []v1.ProductData{
// 					v1.ProductData{
// 						Name:              "ORACLE PARTITIONNING",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC008",
// 						NumOfEquipments:   2,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "ORACLE SGBD Enterprise",
// 						Version:           "9.2.0.8.0",
// 						Category:          "Database",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC003",
// 						NumOfEquipments:   7,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC006",
// 						NumOfEquipments:   4,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.8.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC001",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 4,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Internet Directory Client",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC007",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 2,
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "swidtag",
// 					SortOrder: "asc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   "name",
// 								FilterValue: "Oracle Instant",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: &v1.ProductInfo{
// 				NumOfRecords: []v1.TotalRecords{
// 					v1.TotalRecords{
// 						TotalCnt: 2,
// 					},
// 				},
// 				Products: []v1.ProductData{
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.8.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC001",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 4,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC006",
// 						NumOfEquipments:   4,
// 						NumOfApplications: 2,
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "name",
// 					SortOrder: "desc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   "name",
// 								FilterValue: "Oracle In",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: &v1.ProductInfo{
// 				NumOfRecords: []v1.TotalRecords{
// 					v1.TotalRecords{
// 						TotalCnt: 3,
// 					},
// 				},
// 				Products: []v1.ProductData{
// 					v1.ProductData{
// 						Name:              "Oracle Internet Directory Client",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC007",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.8.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC001",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 4,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC006",
// 						NumOfEquipments:   4,
// 						NumOfApplications: 2,
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx: ctx,
// 				params: &v1.QueryProducts{
// 					PageSize:  5,
// 					Offset:    0,
// 					SortBy:    "numofEquipments",
// 					SortOrder: "asc",
// 					Filter: &v1.AggregateFilter{
// 						Filters: []v1.Queryable{
// 							&v1.Filter{
// 								FilterKey:   "editor",
// 								FilterValue: "Ora",
// 							},
// 						},
// 					},
// 				},
// 			},
// 			want: &v1.ProductInfo{
// 				NumOfRecords: []v1.TotalRecords{
// 					v1.TotalRecords{
// 						TotalCnt: 10,
// 					},
// 				},
// 				Products: []v1.ProductData{
// 					v1.ProductData{
// 						Name:              "Oracle Net Listener",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC004",
// 						NumOfEquipments:   2,
// 						NumOfApplications: 1,
// 					},
// 					v1.ProductData{
// 						Name:              "ORACLE PARTITIONNING",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC008",
// 						NumOfEquipments:   2,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.8.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC001",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 4,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle Net",
// 						Version:           "9.2.0.7.0",
// 						Category:          "Other",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC005",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 2,
// 					},
// 					v1.ProductData{
// 						Name:              "Oracle SGBD Noyau",
// 						Version:           "9.2.0.9.0",
// 						Category:          "Database",
// 						Editor:            "oracle",
// 						Swidtag:           "ORAC010",
// 						NumOfEquipments:   3,
// 						NumOfApplications: 3,
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			got, err := tt.r.GetProducts(tt.args.ctx, tt.args.params)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LicenseRepository.GetProducts() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LicenseRepository.GetProducts() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestLicenseRepository_GetProductInformation(t *testing.T) {

// 	r := &LicenseRepository{dg: dgClient}

// 	ctx := context.Background()

// 	type args struct {
// 		ctx     context.Context
// 		swidtag string
// 	}
// 	tests := []struct {
// 		name    string
// 		r       *LicenseRepository
// 		args    args
// 		want    *v1.ProductAdditionalInfo
// 		wantErr bool
// 	}{
// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx:     ctx,
// 				swidtag: "ORAC002",
// 			},

// 			want: &v1.ProductAdditionalInfo{
// 				Products: []v1.ProductAdditionalData{
// 					v1.ProductAdditionalData{
// 						Swidtag:           "ORAC002",
// 						Name:              "Oracle SGBD Noyau",
// 						Version:           "9.2.0.8.0",
// 						Editor:            "oracle",
// 						NumofEquipments:   6,
// 						NumOfApplications: 3,
// 						NumofOptions:      2,
// 						Child: []v1.ProductChildData{
// 							v1.ProductChildData{
// 								SwidTag: "ORAC007",
// 								Name:    "Oracle Internet Directory Client",
// 								Editor:  "oracle",
// 								Version: "9.2.0.7.0",
// 							},
// 							v1.ProductChildData{
// 								SwidTag: "ORAC010",
// 								Name:    "Oracle SGBD Noyau",
// 								Editor:  "oracle",
// 								Version: "9.2.0.9.0",
// 							},
// 						},
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx:     ctx,
// 				swidtag: "ORAC005",
// 			},

// 			want: &v1.ProductAdditionalInfo{
// 				Products: []v1.ProductAdditionalData{
// 					v1.ProductAdditionalData{
// 						Swidtag:           "ORAC005",
// 						Name:              "Oracle Net",
// 						Version:           "9.2.0.7.0",
// 						Editor:            "oracle",
// 						NumofEquipments:   3,
// 						NumOfApplications: 2,
// 						NumofOptions:      3,
// 						Child: []v1.ProductChildData{
// 							{
// 								SwidTag: "ORAC001",
// 								Name:    "Oracle Instant Client",
// 								Editor:  "oracle",
// 								Version: "9.2.0.8.0",
// 							},
// 							{
// 								SwidTag: "ORAC008",
// 								Name:    "ORACLE PARTITIONNING",
// 								Editor:  "oracle",
// 								Version: "9.2.0.7.0",
// 							},
// 							{
// 								SwidTag: "ORAC004",
// 								Name:    "Oracle Net Listener",
// 								Editor:  "oracle",
// 								Version: "9.2.0.7.0",
// 							},
// 						},
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},

// 		{
// 			name: "SUCCESS",
// 			r:    r,
// 			args: args{
// 				ctx:     ctx,
// 				swidtag: "ORAC001",
// 			},

// 			want: &v1.ProductAdditionalInfo{
// 				Products: []v1.ProductAdditionalData{
// 					v1.ProductAdditionalData{
// 						Swidtag:           "ORAC001",
// 						Name:              "Oracle Instant Client",
// 						Version:           "9.2.0.8.0",
// 						Editor:            "oracle",
// 						NumofEquipments:   3,
// 						NumOfApplications: 4,
// 						NumofOptions:      0,
// 					},
// 				},
// 			},

// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			got, err := tt.r.GetProductInformation(tt.args.ctx, tt.args.swidtag)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LicenseRepository.GetProductInformation() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LicenseRepository.GetProductInformation() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestLicenseRepository_GetApplicationsForProduct(t *testing.T) {

	r := &LicenseRepository{dg: dgClient}

	ctx := context.Background()

	type args struct {
		ctx    context.Context
		params *v1.QueryApplicationsForProduct
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    *v1.ApplicationsForProduct
		wantErr bool
	}{
		{name: "SUCCESS - sortby:name, sortorder:desc",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "ORAC003",
					PageSize:  5,
					Offset:    0,
					SortBy:    "name",
					SortOrder: v1.SortDESC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					v1.ApplicationsForProductData{
						Name:            "Avelletta",
						Owner:           "Yumber",
						NumOfEquipments: 0,
						NumOfInstances:  1,
					},
					v1.ApplicationsForProductData{
						Name:            "Afragusa",
						Owner:           "Pional",
						NumOfEquipments: 9,
						NumOfInstances:  2,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - sortby:numofEquipments, sortorder:asc, filter:afr",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "ORAC009",
					PageSize:  5,
					Offset:    0,
					SortBy:    "numofEquipments",
					SortOrder: v1.SortASC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "name",
								FilterValue: "afr",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					v1.ApplicationsForProductData{
						Name:            "Afragusa",
						Owner:           "Pional",
						NumOfEquipments: 2,
						NumOfInstances:  1,
					},
					v1.ApplicationsForProductData{
						Name:            "Afragusa",
						Owner:           "Pional",
						NumOfEquipments: 9,
						NumOfInstances:  2,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - sortby:name, sortorder:asc, filter:yumb",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "ORAC003",
					PageSize:  5,
					Offset:    0,
					SortBy:    "name",
					SortOrder: v1.SortASC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "application_owner",
								FilterValue: "yumb",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 1,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					v1.ApplicationsForProductData{
						Name:            "Avelletta",
						Owner:           "Yumber",
						NumOfEquipments: 0,
						NumOfInstances:  1,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - sortby:name, sortorder:desc, filter:are",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "ORAC005",
					PageSize:  5,
					Offset:    0,
					SortBy:    "name",
					SortOrder: v1.SortDESC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "name",
								FilterValue: "are",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 1,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					{
						Name:            "Aresena",
						Owner:           "Tractive",
						NumOfEquipments: 0,
						NumOfInstances:  1,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - sortby:numOfInstances, sortorder:desc",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "ORAC001",
					PageSize:  5,
					Offset:    0,
					SortBy:    "numOfInstances",
					SortOrder: v1.SortDESC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 4,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					v1.ApplicationsForProductData{
						Name:            "Acireales",
						Owner:           "Biogercorp",
						NumOfEquipments: 15,
						NumOfInstances:  3,
					},
					v1.ApplicationsForProductData{
						Name:            "Avelletta",
						Owner:           "Yumber",
						NumOfEquipments: 0,
						NumOfInstances:  2,
					},
					v1.ApplicationsForProductData{
						Name:            "Afragusa",
						Owner:           "Pional",
						NumOfEquipments: 2,
						NumOfInstances:  1,
					},
					v1.ApplicationsForProductData{
						Name:            "Avelletta",
						Owner:           "Yumber",
						NumOfEquipments: 0,
						NumOfInstances:  1,
					},
				},
			},
			wantErr: false,
		},
		{name: "SUCCESS - sortby:name, sortorder:desc, scope3",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryApplicationsForProduct{
					SwidTag:   "WIN2",
					PageSize:  5,
					Offset:    0,
					SortBy:    "name",
					SortOrder: v1.SortDESC,
				},
				scopes: []string{"scope3"},
			},
			want: &v1.ApplicationsForProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Applications: []v1.ApplicationsForProductData{
					v1.ApplicationsForProductData{
						Name:            "Afragusa",
						Owner:           "Pional",
						NumOfEquipments: 12,
						NumOfInstances:  2,
					},
					v1.ApplicationsForProductData{
						Name:            "Acireales",
						Owner:           "Biogercorp",
						NumOfEquipments: 3,
						NumOfInstances:  1,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetApplicationsForProduct(tt.args.ctx, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.GetApplicationsForProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareApplicationsForProductResponse(t, "ApplicationsForProduct", got, tt.want)
			}
		})
	}
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

func TestLicenseRepository_GetInstancesForApplicationsProduct(t *testing.T) {

	r := &LicenseRepository{dg: dgClient}

	ctx := context.Background()

	type args struct {
		ctx    context.Context
		params *v1.QueryInstancesForApplicationProduct
		scopes []string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    *v1.InstancesForApplicationProduct
		wantErr bool
	}{
		{name: "SUCCESS",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC001",
					AppID:     "8",
					PageSize:  5,
					Offset:    0,
					SortBy:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.InstancesForApplicationProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 1,
					},
				},
				Instances: []v1.InstancesForApplicationProductData{
					v1.InstancesForApplicationProductData{
						ID:              "16",
						Environment:     "Production",
						NumOfEquipments: 0,
						NumOfProducts:   4,
					},
				},
			},
		},

		{name: "SUCCESS - sortby:env, sortorder:asc",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC005",
					AppID:     "2",
					PageSize:  5,
					Offset:    0,
					SortBy:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.InstancesForApplicationProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Instances: []v1.InstancesForApplicationProductData{

					v1.InstancesForApplicationProductData{
						ID:              "1",
						Environment:     "Development",
						NumOfEquipments: 3,
						NumOfProducts:   3,
					},
					v1.InstancesForApplicationProductData{
						ID:              "2",
						Environment:     "Development",
						NumOfEquipments: 6,
						NumOfProducts:   2,
					},
				},
			},
		},

		{name: "SUCCESS - sortby:numofEquipments, sortorder:asc",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC007",
					AppID:     "1",
					PageSize:  5,
					Offset:    0,
					SortBy:    3,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.InstancesForApplicationProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 3,
					},
				},
				Instances: []v1.InstancesForApplicationProductData{
					v1.InstancesForApplicationProductData{
						ID:              "3",
						Environment:     "Production",
						NumOfEquipments: 2,
						NumOfProducts:   2,
					},
					v1.InstancesForApplicationProductData{
						ID:              "4",
						Environment:     "Test",
						NumOfEquipments: 3,
						NumOfProducts:   2,
					},
					v1.InstancesForApplicationProductData{
						ID:              "5",
						Environment:     "Development",
						NumOfEquipments: 10,
						NumOfProducts:   1,
					},
				},
			},
		},

		{name: "SUCCESS - sortby:numOfProducts, sortorder:desc, scope3",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC005",
					AppID:     "2",
					PageSize:  5,
					Offset:    0,
					SortBy:    2,
					SortOrder: v1.SortDESC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: &v1.InstancesForApplicationProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Instances: []v1.InstancesForApplicationProductData{
					v1.InstancesForApplicationProductData{
						ID:              "1",
						Environment:     "Development",
						NumOfEquipments: 3,
						NumOfProducts:   3,
					},
					v1.InstancesForApplicationProductData{
						ID:              "2",
						Environment:     "Development",
						NumOfEquipments: 6,
						NumOfProducts:   2,
					},
				},
			},
		},

		{name: "SUCCESS - scope3",
			r: r,
			args: args{
				ctx: ctx,
				params: &v1.QueryInstancesForApplicationProduct{
					SwidTag:   "WIN2",
					AppID:     "A2",
					PageSize:  5,
					Offset:    0,
					SortBy:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope3"},
			},
			want: &v1.InstancesForApplicationProduct{
				NumOfRecords: []v1.TotalRecords{
					v1.TotalRecords{
						TotalCnt: 2,
					},
				},
				Instances: []v1.InstancesForApplicationProductData{
					v1.InstancesForApplicationProductData{
						ID:              "I2",
						Environment:     "Development",
						NumOfEquipments: 5,
						NumOfProducts:   3,
					},
					v1.InstancesForApplicationProductData{
						ID:              "I3",
						Environment:     "Production",
						NumOfEquipments: 7,
						NumOfProducts:   2,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.GetInstancesForApplicationsProduct(tt.args.ctx, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.GetInstancesForApplicationsProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareInstancesForApplicationsProductResponse(t, "InstancesForApplicationsProduct", got, tt.want)
			}
		})
	}
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

func TestLicenseRepository_ProductAcquiredRights(t *testing.T) {
	type args struct {
		ctx     context.Context
		swidTag string
		scopes  []string
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
				scopes:  []string{"scope1", "scope2"},
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
				scopes:  []string{"scope3"},
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

func TestLicenseRepository_ProductEquipments(t *testing.T) {

	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}

	defer func() {
		assert.Empty(t, cleanup(), "error  not expected from clean up")
	}()

	//	return
	// eqType := eqTypes[0]

	// equipments, err := equipmentsJSONFromCSV("testdata/equip_3.csv", eqType, true)
	// if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
	// 	return
	// }

	eqType := eqTypes[1]

	equipments, err := equipmentsJSONFromCSV("testdata/scope1/v1/equip_4.csv", eqType, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}
	equipmentsNew, err := equipmentsJSONFromCSV("testdata/scope3/v1/equip_4.csv", eqType, true)
	if !assert.Empty(t, err, "error not expected from equipmentsJSONFromCSV") {
		return
	}

	type args struct {
		ctx     context.Context
		swidTag string
		eqType  *v1.EquipmentType
		params  *v1.QueryEquipments
		scopes  []string
	}

	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    int32
		want1   json.RawMessage
		wantErr bool
	}{
		{name: "success : some sorting",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC001",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : no sort by choose default,page size 2 offset 1",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC001",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  2,
					Offset:    1,
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sort by non displayable attribute",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC001",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr4",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sort by unknown attribute",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC001",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr4.111",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipments[0], equipments[1], equipments[2]}, ",") + "]"),
		},
		{name: "success : sorting, searching by multiple params",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "ORAC001",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   "attr1",
								FilterValue: "equip4",
							},
							&v1.Filter{
								FilterKey:   "attr4.1",
								FilterValue: "mmmmmm44_1",
							},
							&v1.Filter{
								FilterKey:   "attr2",
								FilterValue: 333333422,
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want:  1,
			want1: []byte("[" + strings.Join([]string{equipments[1]}, ",") + "]"),
		},
		{name: "success : some sorting - scope3",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				swidTag: "WIN3",
				eqType:  eqType,
				params: &v1.QueryEquipments{
					PageSize:  3,
					Offset:    0,
					SortBy:    "attr1",
					SortOrder: v1.SortASC,
				},
				scopes: []string{"scope3"},
			},
			want:  3,
			want1: []byte("[" + strings.Join([]string{equipmentsNew[0], equipmentsNew[1], equipmentsNew[2]}, ",") + "]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.r.ProductEquipments(tt.args.ctx, tt.args.swidTag, tt.args.eqType, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductEquipments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.ProductEquipments() got = %v, want %v", got, tt.want)
			}
			fields := strings.Split(string(got1), ",")

			idIndexes := []int{}
			for idx, field := range fields {
				if strings.Contains(field, `[{"ID"`) {
					if idx < len(fields)-1 {
						fields[idx+1] = "[{" + fields[idx+1]
					}
					idIndexes = append(idIndexes, idx)
					continue
				}
				if strings.Contains(field, `{"ID"`) {
					if idx < len(fields)-1 {
						fields[idx+1] = "{" + fields[idx+1]
					}
					idIndexes = append(idIndexes, idx)
				}
			}

			// remove indexes from fields
			idLessfields := make([]string, 0, len(fields)-len(idIndexes))
			count := 0
			for idx := range fields {
				if count < len(idIndexes) && idx == idIndexes[count] {
					count++
					continue
				}
				idLessfields = append(idLessfields, fields[idx])
			}

			assert.Equal(t, strings.Join(strings.Split(string(tt.want1), ","), ","), strings.Join(idLessfields, ","))
		})
	}
}
