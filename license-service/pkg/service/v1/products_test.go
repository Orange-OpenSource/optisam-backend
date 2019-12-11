// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type productQueryMatcher struct {
	q *repo.QueryProducts
	t *testing.T
}

func (p *productQueryMatcher) Matches(x interface{}) bool {
	expQ, ok := x.(*repo.QueryProducts)
	if !ok {
		return ok
	}

	// TODO write code to compare p.q and expQ here
	return compareQueryProducts(p, expQ)
}
func compareQueryProducts(p *productQueryMatcher, exp *repo.QueryProducts) bool {
	if exp == nil {
		return false
	}
	assert.Equalf(p.t, p.q.PageSize, exp.PageSize, "Pagesize are not same")
	assert.Equalf(p.t, p.q.Offset, exp.Offset, "Offset are not same")
	assert.Equalf(p.t, p.q.SortBy, exp.SortBy, "SortBy are not same")
	assert.Equalf(p.t, p.q.SortOrder, exp.SortOrder, "SortOrder are not same")
	assert.Equalf(p.t, p.q.Filter.Filters, exp.Filter.Filters, "Filter are not same")
	assert.Equalf(p.t, p.q.AcqFilter.Filters, exp.AcqFilter.Filters, "AcqFilter are not same")
	assert.Equalf(p.t, p.q.AggFilter.Filters, exp.AggFilter.Filters, "AggFilter are not same")
	return true
}
func (p *productQueryMatcher) String() string {
	return "productQueryMatcher"
}

func Test_licenseServiceServer_ListProducts(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockLicense := mock.NewMockLicense(mockCtrl)

	s := NewLicenseServiceServer(mockLicense).(*licenseServiceServer)

	type args struct {
		ctx context.Context
		req *v1.ListProductsRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		mock    func()
		want    *v1.ListProductsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListProductsRequest{
					PageNum:   1,
					PageSize:  10,
					SortBy:    "name",
					SortOrder: "asc",
					SearchParams: &v1.ProductSearchParams{
						Name: &v1.StringFilter{
							Filteringkey: "afra",
						},
						SwidTag: &v1.StringFilter{
							Filteringkey: "MICR001",
						},
						Editor: &v1.StringFilter{
							Filteringkey: "oracle",
						},
						AgFilter: &v1.AggregationFilter{
							NotForMetric: "oracle type 1",
						},
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetProducts(ctx, &productQueryMatcher{
					q: &repo.QueryProducts{
						PageSize:  10,
						Offset:    int32(0),
						SortBy:    "name",
						SortOrder: "orderasc",
						Filter: &repo.AggregateFilter{
							Filters: []repo.Queryable{
								&repo.Filter{
									FilterKey:   "swidtag",
									FilterValue: "MICR001",
								},
								&repo.Filter{
									FilterKey:   "name",
									FilterValue: "afra",
								},
								&repo.Filter{
									FilterKey:   "editor",
									FilterValue: "oracle",
								},
							},
						},
						AcqFilter: &repo.AggregateFilter{
							Filters: []repo.Queryable{
								&repo.Filter{
									FilterKey:   "metric",
									FilterValue: "oracle type 1",
								},
							},
						},
						AggFilter: &repo.AggregateFilter{
							Filters: []repo.Queryable{
								&repo.Filter{
									FilterKey:   "Name",
									FilterValue: "oracle type 1",
								},
							},
						},
					},
					t: t,
				}, []string{"A", "B"}).Return(&repo.ProductInfo{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 10,
						},
					},
					Products: []repo.ProductData{
						repo.ProductData{
							Name:              "ProductName",
							Version:           "1",
							Category:          "database",
							Editor:            "oracle",
							Swidtag:           "MICR001",
							NumOfEquipments:   2,
							NumOfApplications: 3,
						},
					},
				}, nil).Times(1)
			},
			want: &v1.ListProductsResponse{
				TotalRecords: 10,
				Products: []*v1.Product{
					&v1.Product{
						Name:              "ProductName",
						Version:           "1",
						Category:          "database",
						Editor:            "oracle",
						SwidTag:           "MICR001",
						NumofEquipments:   2,
						NumOfApplications: 3,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.ListProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductResponse(t, "ListProducts", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_ListApplicationsForProduct(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockLicense := mock.NewMockLicense(mockCtrl)

	s := NewLicenseServiceServer(mockLicense).(*licenseServiceServer)

	type args struct {
		ctx context.Context
		req *v1.ListApplicationsForProductRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		want    *v1.ListApplicationsForProductResponse
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS - sortby:name , sortorder:asc",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListApplicationsForProductRequest{
					SwidTag:   "MICR001",
					PageNum:   1,
					PageSize:  10,
					SortBy:    "name",
					SortOrder: "asc",
					SearchParams: &v1.ApplicationSearchParams{
						Name: &v1.StringFilter{
							Filteringkey: "afra",
						},
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplicationsForProduct(ctx, &repo.QueryApplicationsForProduct{
					SwidTag:   "MICR001",
					PageSize:  10,
					Offset:    0,
					SortBy:    "name",
					SortOrder: repo.SortASC,
					Filter: &repo.AggregateFilter{
						Filters: []repo.Queryable{
							&repo.Filter{
								FilterKey:   "name",
								FilterValue: "afra",
							},
						},
					},
				}, []string{"A", "B"}).Return(&repo.ApplicationsForProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 2,
						},
					},
					Applications: []repo.ApplicationsForProductData{
						repo.ApplicationsForProductData{
							ApplicationID:   "1",
							Name:            "Aversea",
							Owner:           "Yumber",
							NumOfEquipments: 4,
							NumOfInstances:  3,
						},
						repo.ApplicationsForProductData{
							ApplicationID:   "2",
							Name:            "Avellete",
							Owner:           "Pional",
							NumOfEquipments: 6,
							NumOfInstances:  5,
						},
					},
				}, nil).Times(1)
			},
			want: &v1.ListApplicationsForProductResponse{
				TotalRecords: 2,
				Applications: []*v1.ApplicationForProduct{
					&v1.ApplicationForProduct{
						ApplicationId:   "1",
						Name:            "Aversea",
						AppOwner:        "Yumber",
						NumOfInstances:  3,
						NumofEquipments: 4,
					},
					&v1.ApplicationForProduct{
						ApplicationId:   "2",
						Name:            "Avellete",
						AppOwner:        "Pional",
						NumOfInstances:  5,
						NumofEquipments: 6,
					},
				},
			},
			wantErr: false,
		},

		{name: "SUCCESS - sortby:name , sortorder:desc",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListApplicationsForProductRequest{
					SwidTag:   "MICR001",
					PageNum:   1,
					PageSize:  10,
					SortBy:    "name",
					SortOrder: "desc",
					SearchParams: &v1.ApplicationSearchParams{
						Name: &v1.StringFilter{
							Filteringkey: "afra",
						},
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplicationsForProduct(ctx, &repo.QueryApplicationsForProduct{
					SwidTag:   "MICR001",
					PageSize:  10,
					Offset:    0,
					SortBy:    "name",
					SortOrder: repo.SortDESC,
					Filter: &repo.AggregateFilter{
						Filters: []repo.Queryable{
							&repo.Filter{
								FilterKey:   "name",
								FilterValue: "afra",
							},
						},
					},
				}, []string{"A", "B"}).Return(&repo.ApplicationsForProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 2,
						},
					},
					Applications: []repo.ApplicationsForProductData{
						repo.ApplicationsForProductData{
							ApplicationID:   "1",
							Name:            "Aversea",
							Owner:           "Yumber",
							NumOfEquipments: 4,
							NumOfInstances:  3,
						},
						repo.ApplicationsForProductData{
							ApplicationID:   "2",
							Name:            "Avellete",
							Owner:           "Pional",
							NumOfEquipments: 6,
							NumOfInstances:  5,
						},
					},
				}, nil).Times(1)
			},
			want: &v1.ListApplicationsForProductResponse{
				TotalRecords: 2,
				Applications: []*v1.ApplicationForProduct{
					&v1.ApplicationForProduct{
						ApplicationId:   "1",
						Name:            "Aversea",
						AppOwner:        "Yumber",
						NumOfInstances:  3,
						NumofEquipments: 4,
					},
					&v1.ApplicationForProduct{
						ApplicationId:   "2",
						Name:            "Avellete",
						AppOwner:        "Pional",
						NumOfInstances:  5,
						NumofEquipments: 6,
					},
				},
			},
			wantErr: false,
		},

		{name: "FAILURE",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListApplicationsForProductRequest{
					SwidTag:   "MICR002",
					PageNum:   1,
					PageSize:  10,
					SortBy:    "name",
					SortOrder: "asc",
				},
			},
			want: &v1.ListApplicationsForProductResponse{
				TotalRecords: 2,
				Applications: []*v1.ApplicationForProduct{
					&v1.ApplicationForProduct{
						ApplicationId:   "1",
						Name:            "Aversea",
						AppOwner:        "Yumber",
						NumOfInstances:  3,
						NumofEquipments: 4,
					},
					&v1.ApplicationForProduct{
						ApplicationId:   "2",
						Name:            "Avellete",
						AppOwner:        "Pional",
						NumOfInstances:  5,
						NumofEquipments: 6,
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplicationsForProduct(ctx, &repo.QueryApplicationsForProduct{
					SwidTag:   "MICR002",
					PageSize:  10,
					Offset:    0,
					SortBy:    "name",
					SortOrder: repo.SortASC}, []string{"A", "B"}).Return(&repo.ApplicationsForProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 2,
						},
					},
					Applications: []repo.ApplicationsForProductData{
						repo.ApplicationsForProductData{
							ApplicationID:   "1",
							Name:            "Aversea",
							Owner:           "Yumber",
							NumOfEquipments: 4,
							NumOfInstances:  3,
						},
						repo.ApplicationsForProductData{
							ApplicationID:   "2",
							Name:            "Avellete",
							Owner:           "Pional",
							NumOfEquipments: 6,
							NumOfInstances:  5,
						},
					},
				}, fmt.Errorf("Test error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - can not retrieve claims",
			s: s,
			args: args{
				ctx: context.Background(),
				req: &v1.ListApplicationsForProductRequest{
					SwidTag:   "MICR002",
					PageNum:   1,
					PageSize:  10,
					SortBy:    "name",
					SortOrder: "asc",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.ListApplicationsForProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListApplicationsForProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareApplicationsForProductResponse(t, "ApplicationsForProduct", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_ListInstancesForApplicationsProduct(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockLicense := mock.NewMockLicense(mockCtrl)

	s := NewLicenseServiceServer(mockLicense).(*licenseServiceServer)

	type args struct {
		ctx context.Context
		req *v1.ListInstancesForApplicationProductRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		want    *v1.ListInstancesForApplicationProductResponse
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS - sortby: env, sortorder:asc",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_ASC,
				},
			},
			want: &v1.ListInstancesForApplicationProductResponse{
				TotalRecords: 1,
				Instances: []*v1.InstancesForApplicationProduct{
					&v1.InstancesForApplicationProduct{
						Id:              "2",
						Name:            "ORACLE",
						Environment:     "Production",
						NumofEquipments: 3,
						NumofProducts:   3,
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplication(ctx, "4", []string{"A", "B"}).Return(&repo.ApplicationDetails{
					Name: "Oracle",
				}, nil).Times(1)
				mockLicense.EXPECT().GetInstancesForApplicationsProduct(ctx, &repo.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC249",
					AppID:     "4",
					PageSize:  10,
					Offset:    0,
					SortBy:    1,
					SortOrder: repo.SortASC}, []string{"A", "B"}).Return(&repo.InstancesForApplicationProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 1,
						},
					},
					Instances: []repo.InstancesForApplicationProductData{
						repo.InstancesForApplicationProductData{
							ID:              "2",
							Name:            "ORACLE",
							Environment:     "Production",
							NumOfEquipments: 3,
							NumOfProducts:   3,
						},
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},
		{name: "SUCCESS - instance name is empty",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_ASC,
				},
			},
			want: &v1.ListInstancesForApplicationProductResponse{
				TotalRecords: 1,
				Instances: []*v1.InstancesForApplicationProduct{
					&v1.InstancesForApplicationProduct{
						Id:              "2",
						Name:            "Oracle",
						Environment:     "Production",
						NumofEquipments: 3,
						NumofProducts:   3,
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplication(ctx, "4", []string{"A", "B"}).Return(&repo.ApplicationDetails{
					Name: "Oracle",
				}, nil).Times(1)
				mockLicense.EXPECT().GetInstancesForApplicationsProduct(ctx, &repo.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC249",
					AppID:     "4",
					PageSize:  10,
					Offset:    0,
					SortBy:    1,
					SortOrder: repo.SortASC}, []string{"A", "B"}).Return(&repo.InstancesForApplicationProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 1,
						},
					},
					Instances: []repo.InstancesForApplicationProductData{
						repo.InstancesForApplicationProductData{
							ID:              "2",
							Name:            "",
							Environment:     "Production",
							NumOfEquipments: 3,
							NumOfProducts:   3,
						},
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},

		{name: "SUCCESS - sortby:env , sortorder:desc",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_DESC,
				},
			},
			want: &v1.ListInstancesForApplicationProductResponse{
				TotalRecords: 1,
				Instances: []*v1.InstancesForApplicationProduct{
					&v1.InstancesForApplicationProduct{
						Id:              "2",
						Name:            "ORACLE",
						Environment:     "Production",
						NumofEquipments: 3,
						NumofProducts:   3,
					},
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplication(ctx, "4", []string{"A", "B"}).Return(&repo.ApplicationDetails{
					Name: "Oracle",
				}, nil).Times(1)
				mockLicense.EXPECT().GetInstancesForApplicationsProduct(ctx, &repo.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC249",
					AppID:     "4",
					PageSize:  10,
					Offset:    0,
					SortBy:    1,
					SortOrder: repo.SortDESC}, []string{"A", "B"}).Return(&repo.InstancesForApplicationProduct{
					NumOfRecords: []repo.TotalRecords{
						repo.TotalRecords{
							TotalCnt: 1,
						},
					},
					Instances: []repo.InstancesForApplicationProductData{
						repo.InstancesForApplicationProductData{
							ID:              "2",
							Environment:     "Production",
							NumOfEquipments: 3,
							NumOfProducts:   3,
						},
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_ASC,
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot get application",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_ASC,
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplication(ctx, "4", []string{"A", "B"}).Return(nil, errors.New("cannot get application")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE",
			s: s,
			args: args{
				ctx: ctx,
				req: &v1.ListInstancesForApplicationProductRequest{
					SwidTag:       "ORAC249",
					ApplicationId: "4",
					PageNum:       1,
					PageSize:      10,
					SortBy:        v1.ListInstancesForApplicationProductRequest_ENV,
					SortOrder:     v1.SortOrder_ASC,
				},
			},
			mock: func() {
				mockLicense.EXPECT().GetApplication(ctx, "4", []string{"A", "B"}).Return(&repo.ApplicationDetails{
					Name: "Oracle",
				}, nil).Times(1)
				mockLicense.EXPECT().GetInstancesForApplicationsProduct(ctx, &repo.QueryInstancesForApplicationProduct{
					SwidTag:   "ORAC249",
					AppID:     "4",
					PageSize:  10,
					Offset:    0,
					SortBy:    1,
					SortOrder: repo.SortASC}, []string{"A", "B"}).Return(nil, fmt.Errorf("Test error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			got, err := tt.s.ListInstancesForApplicationsProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListInstancesForApplicationsProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareInstancesForApplicationsProductResponse(t, "InstancesForApplicationsProduct", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_ListAcqRightsForProduct(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.ListAcquiredRightsForProductRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		setup   func()
		want    *v1.ListAcquiredRightsForProductResponse
		wantErr bool
	}{
		{name: "SUCCESS - computelicenseOPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s3",
						Metric:       "ONS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(8), nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(6), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s2",
						SwidTag:        "P1",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s3",
						SwidTag:        "P1",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{name: "SUCCESS - no equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s3",
						Metric:       "ONS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 0,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s2",
						SwidTag:        "P1",
						Metric:         "WS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s3",
						SwidTag:        "P1",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch products acquired rights",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("", nil, errors.New(""))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch products",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrices  list",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test srror"))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New(""))

			},
			wantErr: true,
		},
		{name: "SUCCESS - computelicenseOPS failed- cannot fetch metric OPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e6",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- parent hierarchy not found",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e6",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- cannot find base level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e9",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- cannot find aggregate level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e9",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- cannot find end level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e9",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- levels are not in valid order",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e3",
						AggerateLevelEqTypeID: "e2",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- numOfcores attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- numOfcpu attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}

				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- corefactor attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseOPS failed- cannot compute metric licenses",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricOPS{
					&repo.MetricOPS{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					&repo.MetricOPS{
						Name: "IMB",
					},
				}, nil)
				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(0), errors.New(""))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS - licenseProd<=licenseNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s3",
						Metric:       "ONS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(8), uint64(8), nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(6), uint64(6), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s2",
						SwidTag:        "P1",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s3",
						SwidTag:        "P1",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS - licenseProd>licenseNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s3",
						Metric:       "ONS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(8), uint64(6), nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(6), uint64(4), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s2",
						SwidTag:        "P1",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s3",
						SwidTag:        "P1",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseSPS failed - cannot fetch metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e6",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseSPS failed - metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS failed- cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS failed- numofcores attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS failed- coreFactor attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseSPS failed- cannot compute licenses for metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricSPS{
					&repo.MetricSPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricSPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(0), uint64(0), errors.New(""))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseIPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s2",
						Metric:       "WS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
					&repo.ProductAcquiredRight{
						SKU:          "s3",
						Metric:       "ONS",
						AcqLicenses:  10,
						TotalCost:    50,
						AvgUnitPrice: 5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(8), nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(6), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s2",
						SwidTag:        "P1",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					&v1.ProductAcquiredRights{
						SKU:            "s3",
						SwidTag:        "P1",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseIPS failed - cannot fetch metric IPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e6",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computelicenseIPS failed - metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseIPS failed- cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseIPS failed- numofcores attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseIPS failed- coreFactor attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseIPS failed- cannot compute licenses for metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					&repo.Metric{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricIPS{
					&repo.MetricIPS{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					&repo.MetricIPS{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A", "B"}).Times(1).Return(uint64(0), errors.New(""))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{name: "SUCCESS - default",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []string{"A", "B"}).Times(1).Return("pp1", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "s1",
						Metric:       "OPS",
						AcqLicenses:  5,
						TotalCost:    20,
						AvgUnitPrice: 4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "OPS",
						Type: "",
					},
					&repo.Metric{
						Name: "WS",
						Type: "",
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpus",
					Type: repo.DataTypeInt,
				}
				corefactor := &repo.Attribute{
					ID:   "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					ParentID:   "e3",
					Attributes: []*repo.Attribute{cores, cpu, corefactor},
				}
				start := &repo.EquipmentType{
					ID:       "e1",
					ParentID: "e2",
				}
				agg := &repo.EquipmentType{
					ID:       "e3",
					ParentID: "e4",
				}
				end := &repo.EquipmentType{
					ID:       "e4",
					ParentID: "e5",
				}
				endP := &repo.EquipmentType{
					ID: "e5",
				}

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListAcqRightsForProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcqRightsForProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductAcquiredRights(t, "ListAcquiredRightsForProductResponse", tt.want, got)
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_licenseServiceServer_ListEquipmentsForProduct(t *testing.T) {

	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	//s := NewLicenseServiceServer(mockLicense).(*licenseServiceServer)

	eqTypes := []*repo.EquipmentType{
		&repo.EquipmentType{
			Type: "typ1",
			ID:   "1",
			Attributes: []*repo.Attribute{
				&repo.Attribute{
					ID:           "1",
					Name:         "attr1",
					Type:         repo.DataTypeString,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				&repo.Attribute{
					ID:           "2",
					Name:         "attr2",
					Type:         repo.DataTypeString,
					IsDisplayed:  true,
					IsSearchable: true,
				},
			},
		},
		&repo.EquipmentType{
			Type: "typ2",
			ID:   "2",
			Attributes: []*repo.Attribute{
				&repo.Attribute{
					ID:          "1",
					Name:        "attr1",
					Type:        repo.DataTypeString,
					IsDisplayed: true,
				},
				&repo.Attribute{
					ID:          "2",
					Name:        "attr2",
					Type:        repo.DataTypeString,
					IsDisplayed: true,
				},
			},
		},
	}
	// TODO
	// queryParams := &repo.QueryEquipments{
	// 	PageSize:  10,
	// 	Offset:    90,
	// 	SortBy:    "attr1",
	// 	SortOrder: repo.SortDESC,
	// 	Filter: &repo.AggregateFilter{
	// 		Filters: []repo.Queryable{
	// 			&repo.Filter{
	// 				FilterKey:   "attr1",
	// 				FilterValue: "a11",
	// 			},
	// 			&repo.Filter{
	// 				FilterKey:   "attr2",
	// 				FilterValue: "a22",
	// 			},
	// 		},
	// 	},
	// }
	type args struct {
		ctx context.Context
		req *v1.ListEquipmentsForProductRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		want    *v1.ListEquipmentsResponse
		wantErr bool
		setup   func()
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ProductEquipments(ctx, "P1", eqTypes[0], gomock.Any(), []string{"A", "B"}).Times(1).Return(int32(2), json.RawMessage(`[{ID:"1"}]`), nil)

			},
			want: &v1.ListEquipmentsResponse{
				TotalRecords: 2,
				Equipments:   json.RawMessage(`[{ID:"1"}]`),
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch equipment type with given Id",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot find sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr3",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return([]*repo.EquipmentType{
					&repo.EquipmentType{
						Type: "typ1",
						ID:   "1",
						Attributes: []*repo.Attribute{
							&repo.Attribute{
								ID:           "1",
								Name:         "attr1",
								Type:         repo.DataTypeString,
								IsDisplayed:  false,
								IsSearchable: true,
							},
							&repo.Attribute{
								ID:           "2",
								Name:         "attr2",
								Type:         repo.DataTypeString,
								IsDisplayed:  false,
								IsSearchable: true,
							},
						},
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot parse equipment query param",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr3=att3",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- cannot fetch product equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					SwidTag:      "P1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_DESC,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"A", "B"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ProductEquipments(ctx, "P1", eqTypes[0], gomock.Any(), []string{"A", "B"}).Times(1).Return(int32(2), nil, errors.New("test error"))

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListEquipmentsForProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListEquipmentsForProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("licenseServiceServer.ListEquipmentsForProduct() = %v, want %v", got, tt.want)
			}
			if tt.setup == nil {
				defer mockCtrl.Finish()
			}
		})
	}
}

func compareProductResponse(t *testing.T, name string, exp *v1.ListProductsResponse, act *v1.ListProductsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.TotalRecords, act.TotalRecords, "%s.Records are not same", name)
	compareProductsAll(t, name+".Products", exp.Products, act.Products)
}

func compareProductsAll(t *testing.T, name string, exp []*v1.Product, act []*v1.Product) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareProducts(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareProducts(t *testing.T, name string, exp *v1.Product, act *v1.Product) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Version, act.Version, "%s.Version are not same", name)
	assert.Equalf(t, exp.Category, act.Category, "%s.Category are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.NumOfApplications, act.NumOfApplications, "%s.NumOfApplications are not same", name)
	assert.Equalf(t, exp.NumofEquipments, act.NumofEquipments, "%s.NumofEquipments are not same", name)
}

func compareApplicationsForProductResponse(t *testing.T, name string, exp *v1.ListApplicationsForProductResponse, act *v1.ListApplicationsForProductResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.TotalRecords, act.TotalRecords, "%s.Records are not same", name)

	compareApplicationsForProductAll(t, name+".Applications", exp.Applications, act.Applications)
}

func compareApplicationsForProductAll(t *testing.T, name string, exp []*v1.ApplicationForProduct, act []*v1.ApplicationForProduct) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareApplicationForProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareApplicationForProduct(t *testing.T, name string, exp *v1.ApplicationForProduct, act *v1.ApplicationForProduct) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.ApplicationId, act.ApplicationId, "%s.ApplicationId are not same", name)
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.AppOwner, act.AppOwner, "%s.AppOwner are not same", name)
	assert.Equalf(t, exp.NumofEquipments, act.NumofEquipments, "%s.NumofEquipments are not same", name)
	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances are not same", name)
}

func compareInstancesForApplicationsProductResponse(t *testing.T, name string, exp *v1.ListInstancesForApplicationProductResponse, act *v1.ListInstancesForApplicationProductResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.TotalRecords, act.TotalRecords, "%s.Records are not same", name)
	compareInstancesForApplicationsProductAll(t, name+".Instances", exp.Instances, act.Instances)
}

func compareInstancesForApplicationsProductAll(t *testing.T, name string, exp []*v1.InstancesForApplicationProduct, act []*v1.InstancesForApplicationProduct) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareInstanceForApplicationsProduct(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareInstanceForApplicationsProduct(t *testing.T, name string, exp *v1.InstancesForApplicationProduct, act *v1.InstancesForApplicationProduct) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Id, act.Id, "%s.Id are not same", name)
	assert.Equalf(t, exp.Environment, act.Environment, "%s.Environment are not same", name)
	assert.Equalf(t, exp.NumofEquipments, act.NumofEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumofProducts, act.NumofProducts, "%s.NumOfProducts are not same", name)
}

func compareProductAcquiredRights(t *testing.T, name string, exp *v1.ListAcquiredRightsForProductResponse, act *v1.ListAcquiredRightsForProductResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	compareProductAcquiredRightsAll(t, name+".AcqRights", exp.AcqRights, act.AcqRights)
}

func compareProductAcquiredRightsAll(t *testing.T, name string, exp []*v1.ProductAcquiredRights, act []*v1.ProductAcquiredRights) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAcqRight(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}
func compareAcqRight(t *testing.T, name string, exp *v1.ProductAcquiredRights, act *v1.ProductAcquiredRights) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU are not same", name)
	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTags are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metrics are not same", name)
	assert.Equalf(t, exp.NumCptLicences, act.NumCptLicences, "%s.Computed Licenses are not same", name)
	assert.Equalf(t, exp.NumAcqLicences, act.NumAcqLicences, "%s.Acquired Licenses are not same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.Total Cost is not same", name)
	assert.Equalf(t, exp.DeltaNumber, act.DeltaNumber, "%s.Delta Numbers are not same", name)
	assert.Equalf(t, exp.DeltaCost, act.DeltaCost, "%s.Delta Cost is not same", name)

}
