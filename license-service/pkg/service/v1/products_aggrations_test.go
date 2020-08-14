// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_CreateProductAggregation(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"P1", "P2", "P3"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.ProductAggregation
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		want    *v1.ProductAggregation
		mock    func()
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			want: &v1.ProductAggregation{
				ID:           "ProID1",
				Name:         "ProName",
				Metric:       "m1",
				Editor:       "e1",
				ProductNames: []string{"pro1"},
				Products:     []string{"P1", "P2"},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"P1", "P2", "P3"}).Return([]*repo.Metric{
					&repo.Metric{
						ID:   "m1ID",
						Name: "m1",
					},
					&repo.Metric{
						ID:   "m2ID",
						Name: "m2",
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"P1", "P2", "P3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"P1", "P2", "P3"}).Return("P2ID", nil).Times(1),
				)
				mockLicense.EXPECT().CreateProductAggregation(ctx, &repo.ProductAggregation{
					Name:     "ProName",
					Metric:   "m1ID",
					Editor:   "e1",
					Product:  "pro1",
					Products: []string{"P1ID", "P2ID"},
				}, []string{"P1", "P2", "P3"}).Return(&repo.ProductAggregation{
					ID:       "ProID1",
					Name:     "ProName",
					Metric:   "m1ID",
					Editor:   "e1",
					Product:  "pro1",
					Products: []string{"P1ID", "P2ID"},
				}, nil).Times(1)
			},
			wantErr: false,
		},
		{name: "FAILURE-cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot get product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - product aggregation node already exists",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(&repo.ProductAggregation{
					Name: "ProName",
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch metrics",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"P1", "P2", "P3"}).Return(nil, errors.New("Internal error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - metric not found",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"P1", "P2", "P3"}).Return([]*repo.Metric{
					&repo.Metric{
						ID:   "m2ID",
						Name: "m2",
					},
					&repo.Metric{
						ID:   "m3ID",
						Name: "m3",
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot get product id for swid tag",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"P1", "P2", "P3"}).Return([]*repo.Metric{
					&repo.Metric{
						ID:   "m1ID",
						Name: "m1",
					},
					&repo.Metric{
						ID:   "m2ID",
						Name: "m2",
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"P1", "P2", "P3"}).Return("", errors.New("Internal")).Times(1),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot create product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProName", []string{"P1", "P2", "P3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
				mockLicense.EXPECT().ListMetrices(ctx, []string{"P1", "P2", "P3"}).Return([]*repo.Metric{
					&repo.Metric{
						ID:   "m1ID",
						Name: "m1",
					},
					&repo.Metric{
						ID:   "m2ID",
						Name: "m2",
					},
				}, nil).Times(1)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"P1", "P2", "P3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"P1", "P2", "P3"}).Return("P2ID", nil).Times(1),
				)
				mockLicense.EXPECT().CreateProductAggregation(ctx, &repo.ProductAggregation{
					Name:     "ProName",
					Metric:   "m1ID",
					Editor:   "e1",
					Product:  "pro1",
					Products: []string{"P1ID", "P2ID"},
				}, []string{"P1", "P2", "P3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - user doesnot have access to create product aggregation",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"P1", "P2", "P3"},
				}),
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - unknown role",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"P1", "P2", "P3"},
				}),
				req: &v1.ProductAggregation{
					Name:         "ProName",
					Metric:       "m1",
					Editor:       "e1",
					ProductNames: []string{"pro1"},
					Products:     []string{"P1", "P2"},
				},
			},
			mock:    func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := NewLicenseServiceServer(rep)
			got, err := s.CreateProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.CreateProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductAggregation(t, "CreateProductAggregation", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_DeleteProductAggregation(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"P1", "P2", "P3"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.DeleteProductAggregationRequest
	}
	tests := []struct {
		name    string
		s       *licenseServiceServer
		args    args
		mock    func()
		want    *v1.ListProductAggregationResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteProductAggregationRequest{
					ID: "ProID1",
				},
			},
			want: &v1.ListProductAggregationResponse{
				Aggregations: []*v1.ProductAggregation{
					&v1.ProductAggregation{
						ID:           "ProID2",
						Name:         "ProName2",
						Metric:       "m2ID",
						Editor:       "e2",
						ProductNames: []string{"pro1"},
						Products:     []string{"P1ID", "P3ID"},
					},
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().DeleteProductAggregation(ctx, "ProID1", []string{"P1", "P2", "P3"}).Return([]*repo.ProductAggregation{
					&repo.ProductAggregation{
						ID:       "ProID2",
						Name:     "ProName2",
						Metric:   "m2ID",
						Editor:   "e2",
						Product:  "pro2",
						Products: []string{"P1ID", "P3ID"},
					},
				}, nil).Times(1)
			},
			wantErr: false,
		},
		{name: "FAILURE-cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteProductAggregationRequest{
					ID: "ProID1",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "FAILURE-user doesnot have access to delete product aggregation",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.DeleteProductAggregationRequest{
					ID: "ProID1",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{
			name: "FAILURE-unknown role",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.DeleteProductAggregationRequest{
					ID: "ProID1",
				},
			},
			mock:    func() {},
			wantErr: true,
		},
		{name: "FAILURE - cannot delete product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.DeleteProductAggregationRequest{
					ID: "ProID1",
				},
			},
			mock: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().DeleteProductAggregation(ctx, "ProID1", []string{"P1", "P2", "P3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()
			s := NewLicenseServiceServer(rep)
			got, err := s.DeleteProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.DeleteProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductAggregationAll(t, "DeleteProductAggregation", got, tt.want)
			}
		})
	}
}

func Test_licenseServiceServer_UpdateProductAggregation(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License

	type args struct {
		ctx context.Context
		req *v1.UpdateProductAggregationRequest
	}
	tests := []struct {
		name    string
		setup   func()
		args    args
		want    *v1.ProductAggregation
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProIDC1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1", "pro2"},
					},
				},
			},
			want: &v1.ProductAggregation{
				ID:           "PA1",
				Name:         "ProIDC1",
				Metric:       "m1",
				Editor:       "e1",
				ProductNames: []string{"pro1", "pro2"},
				Products:     []string{"P1", "P2", "P5"},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProIDC1", []string{"Scope1", "Scope2", "Scope3"}).Return(
						nil, repo.ErrNodeNotFound).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1", "pro2"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1", "pro2"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P3ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P4ID", nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					Name:            "ProIDC1",
					AddedProducts:   []string{"P1ID", "P2ID"},
					RemovedProducts: []string{"P3ID", "P4ID"},
					Product:         "pro1,pro2",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(nil).Times(1)
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProIDC1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
					ID:         "PA1",
					Name:       "ProIDC1",
					MetricName: "m1",
					Editor:     "e1",
					Product:    "pro1,pro2",
					Products:   []string{"P1", "P2", "P5"},
				}, nil).Times(1)
			},
		},
		{name: "SUCCESS - With same product Aggregation Name",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			want: &v1.ProductAggregation{
				ID:           "PA1",
				Name:         "ProID1",
				Metric:       "m1",
				Editor:       "e1",
				ProductNames: []string{"pro1"},
				Products:     []string{"P1", "P2", "P5"},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P3ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P4ID", nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					Name:            "",
					AddedProducts:   []string{"P1ID", "P2ID"},
					RemovedProducts: []string{"P3ID", "P4ID"},
					Product:         "pro1",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(nil).Times(1)
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
					ID:         "PA1",
					Name:       "ProID1",
					MetricName: "m1",
					Editor:     "e1",
					Product:    "pro1",
					Products:   []string{"P1", "P2", "P5"},
				}, nil).Times(1)
			},
		},
		{name: "SUCCESS - With same product Aggregation Name, empty removed and added products and no change in product names",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{},
						RemovedProducts: []string{},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			want: &v1.ProductAggregation{
				ID:           "PA1",
				Name:         "ProID1",
				Metric:       "m1",
				Editor:       "e1",
				ProductNames: []string{"pro1"},
				Products:     []string{"P3", "P4", "P5"},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					Name:            "",
					AddedProducts:   []string{},
					RemovedProducts: []string{},
					Product:         "pro1",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(nil).Times(1)
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
					ID:         "PA1",
					Name:       "ProID1",
					MetricName: "m1",
					Editor:     "e1",
					Product:    "pro1",
					Products:   []string{"P3", "P4", "P5"},
				}, nil).Times(1)
			},
		},
		{
			name: "FAILURE-user doesnot have access to update product aggregation",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE- can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE-cannot get product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE-product aggregation node does not exist",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(nil, repo.ErrNodeNotFound).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE-cannot get product id for swid tag for added products",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("", errors.New("Internal")).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("", errors.New("Internal")).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE-cannot get product id for swid tag for removed products",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("", errors.New("Internal")).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("", errors.New("Internal")).Times(1),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE- cannot update product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P3ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P4ID", nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					AddedProducts:   []string{"P1ID", "P2ID"},
					RemovedProducts: []string{"P3ID", "P4ID"},
					Product:         "pro1",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{
			name: "FAILURE- cannot get product aggregation",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				gomock.InOrder(
					mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
						ID:         "PA1",
						Name:       "ProID1",
						MetricName: "m1",
						Editor:     "e1",
						Product:    "pro1",
						Products:   []string{"P3", "P4", "P5"},
					}, nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P3ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P4ID", nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					AddedProducts:   []string{"P1ID", "P2ID"},
					RemovedProducts: []string{"P3ID", "P4ID"},
					Product:         "pro1",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(nil).Times(1)
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(nil, errors.New("Internal")).Times(1)

			},
			wantErr: true,
		},
		{
			name: "FAILURE- product aggregation node does not exist",
			args: args{
				ctx: ctx,
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockLicense(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(&repo.ProductAggregation{
					ID:         "PA1",
					Name:       "ProID1",
					MetricName: "m1",
					Editor:     "e1",
					Product:    "pro1",
					Products:   []string{"P3", "P4", "P5"},
				}, nil).Times(1)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P1", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P1ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P2", &productQueryMatcher{
						q: &repo.QueryProducts{
							Filter: &repo.AggregateFilter{
								Filters: []repo.Queryable{
									&repo.Filter{
										FilterMatchingType:  repo.EqFilter,
										FilterKey:           "name",
										FilterValueMultiple: stringToInterface([]string{"pro1"}),
									},
									&repo.Filter{
										FilterMatchingType: repo.EqFilter,
										FilterKey:          "editor",
										FilterValue:        "e1",
									},
								},
							},
							AcqFilter: productAcqRightFilter("m1"),
							AggFilter: productAggregateFilter("m1"),
						},
						t: t,
					}, []string{"Scope1", "Scope2", "Scope3"}).Return("P2ID", nil).Times(1),
				)
				gomock.InOrder(
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P3", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P3ID", nil).Times(1),
					mockLicense.EXPECT().ProductIDForSwidtag(ctx, "P4", nil, []string{"Scope1", "Scope2", "Scope3"}).Return("P4ID", nil).Times(1),
				)
				mockLicense.EXPECT().UpdateProductAggregation(ctx, "PA1", &repo.UpdateProductAggregationRequest{
					AddedProducts:   []string{"P1ID", "P2ID"},
					RemovedProducts: []string{"P3ID", "P4ID"},
					Product:         "pro1",
				}, []string{"Scope1", "Scope2", "Scope3"}).Return(nil).Times(1)

				mockLicense.EXPECT().ProductAggregationsByName(ctx, "ProID1", []string{"Scope1", "Scope2", "Scope3"}).Return(nil, repo.ErrNodeNotFound).Times(1)

			},
			wantErr: true,
		},
		{
			name: "FAILURE-unknown role",
			args: args{
				ctx: ctxmanage.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "abc",
					Socpes: []string{"Scope1", "Scope2", "Scope3"},
				}),
				req: &v1.UpdateProductAggregationRequest{
					Name: "ProID1",
					Aggregation: &v1.UpdateAggregation{
						Name:            "ProID1",
						AddedProducts:   []string{"P1", "P2"},
						RemovedProducts: []string{"P3", "P4"},
						ProductNames:    []string{"pro1"},
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.UpdateProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.UpdateProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			compareProductAggregation(t, "ProductAggregation", tt.want, got)
		})
	}
}

func compareProductAggregationAll(t *testing.T, name string, exp *v1.ListProductAggregationResponse, act *v1.ListProductAggregationResponse) {
	for i := 0; i < len(exp.Aggregations)-1; i++ {
		compareProductAggregation(t, name, exp.Aggregations[i], act.Aggregations[i])
	}
}

func compareProductAggregation(t *testing.T, name string, exp *v1.ProductAggregation, act *v1.ProductAggregation) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Editor are not same", name)
	assert.Equalf(t, exp.ProductNames, act.ProductNames, "%s.Products are not same", name)
	for i := 0; i < len(exp.Products)-1; i++ {
		assert.Equalf(t, exp.Products[i], act.Products[i], "%s.Products are not same", name)
	}
}
