// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
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
	return compareQueryProducts(p, expQ)
}
func compareQueryProducts(p *productQueryMatcher, exp *repo.QueryProducts) bool {
	if exp == nil {
		return false
	}
	if !assert.Equalf(p.t, p.q.PageSize, exp.PageSize, "Pagesize are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.Offset, exp.Offset, "Offset are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.SortBy, exp.SortBy, "SortBy are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.SortOrder, exp.SortOrder, "SortOrder are not same") {
		return false
	}
	if !compareQueryFilters(p.t, "productQueryMatcher", p.q.Filter.Filters, exp.Filter.Filters) {
		return false
	}
	if !compareQueryFilters(p.t, "productQueryMatcher", p.q.AcqFilter.Filters, exp.AcqFilter.Filters) {
		return false
	}
	if !compareQueryFilters(p.t, "productQueryMatcher", p.q.AggFilter.Filters, exp.AggFilter.Filters) {
		return false
	}
	return true
}

func compareQueryFilters(t *testing.T, name string, expFilter []repo.Queryable, actFilter []repo.Queryable) bool {
	for i := range expFilter {
		if !compareQueryFilter(t, fmt.Sprintf("%s[%d]", name, i), expFilter[i], actFilter[i]) {
			return false
		}
	}
	return true
}

func compareQueryFilter(t *testing.T, name string, expFilter repo.Queryable, actFilter repo.Queryable) bool {
	if !assert.Equalf(t, expFilter.Key(), actFilter.Key(), "%s.Filter key is not same", name) {
		return false
	}
	if !assert.Equalf(t, expFilter.Value(), actFilter.Value(), "%s.Filter value is not same", name) {
		return false
	}
	if !compareQueryFilterValues(t, name, expFilter.Values(), actFilter.Values()) {
		return false
	}
	// if !assert.Equalf(t, expFilter.Values(), actFilter.Values(), "%s.Filter values is not same", name) {
	//     return false
	// }
	if !assert.Equalf(t, expFilter.Priority(), actFilter.Priority(), "%s.Filter priority is not same", name) {
		return false
	}
	if !assert.Equalf(t, expFilter.Type(), actFilter.Type(), "%s.Filter type is not same", name) {
		return false
	}
	return true
}
func compareQueryFilterValues(t *testing.T, name string, exp []interface{}, act []interface{}) bool {
	if exp == nil && act == nil {
		return true
	}
	for i := range exp {
		if !assert.Equalf(t, exp[i], act[i], "%s.Filter values is not same", name) {
			return false
		}
	}
	return true
}

func (p *productQueryMatcher) String() string {
	return "productQueryMatcher"
}

type queryMatcherApplicationsForProduct struct {
	q *repo.QueryApplicationsForProduct
	t *testing.T
}

func (p *queryMatcherApplicationsForProduct) Matches(x interface{}) bool {
	expQ, ok := x.(*repo.QueryApplicationsForProduct)
	if !ok {
		return ok
	}
	return compareQueryApplicationForProduct(p, expQ)
}
func compareQueryApplicationForProduct(p *queryMatcherApplicationsForProduct, exp *repo.QueryApplicationsForProduct) bool {
	if exp == nil {
		return false
	}
	if !assert.Equalf(p.t, p.q.PageSize, exp.PageSize, "Pagesize are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.Offset, exp.Offset, "Offset are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.SortBy, exp.SortBy, "SortBy are not same") {
		return false
	}
	if !assert.Equalf(p.t, p.q.SortOrder, exp.SortOrder, "SortOrder are not same") {
		return false
	}
	if !compareQueryFilters(p.t, "queryMatcherApplicationsForProduct", p.q.Filter.Filters, exp.Filter.Filters) {
		return false
	}
	return true
}
func (p *queryMatcherApplicationsForProduct) String() string {
	return "queryMatcherApplicationsForProduct"
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
		{name: "SUCCESS - computeLicenseACS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "attribute.counter.standard",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					&repo.MetricACS{
						Name:          "ACS1",
						EqType:        "server",
						AttributeName: "cpu",
						Value:         "2",
					},
				}, nil)

				mat := &repo.MetricACSComputed{
					Name:      "attribute.counter.standard",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "uidORAC001", mat, []string{"A", "B"}).Times(1).Return(uint64(10), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumCptLicences: 10,
						NumAcqLicences: 20,
						TotalCost:      9270,
						DeltaNumber:    10,
						DeltaCost:      200,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseACS failed - cannot fetch acs metrics",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseACS failed - cannot find metric name acs",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "acs",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseACS failed - cannot find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "attribute.counter.standard",
						EqType:        "cluster",
						AttributeName: "corefactor",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseACS failed - attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "attribute.counter.standard",
						EqType:        "server",
						AttributeName: "servermodel",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{name: "SUCCESS - computeLicenseACS failed - cannot compute licenses for metric OPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					&repo.ProductAcquiredRight{
						SKU:          "ORAC001ACS",
						Metric:       "attribute.counter.standard",
						AcqLicenses:  20,
						TotalCost:    9270,
						AvgUnitPrice: 20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A", "B"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						repo.ProductAdditionalData{
							NumofEquipments: 56,
						},
					},
				}, nil)

				mockRepo.EXPECT().ListMetrices(ctx, []string{"A", "B"}).Times(1).Return([]*repo.Metric{
					&repo.Metric{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					&repo.Metric{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					&repo.Metric{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					&repo.Metric{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					&repo.Metric{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
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
					Name: "corefactor",
					Type: repo.DataTypeInt,
				}

				base := &repo.EquipmentType{
					ID:         "e2",
					Type:       "server",
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
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A", "B"}).Times(1).Return([]*repo.MetricACS{
					&repo.MetricACS{
						Name:          "attribute.counter.standard",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
				}, nil)
				mat := &repo.MetricACSComputed{
					Name:      "attribute.counter.standard",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "uidORAC001", mat, []string{"A", "B"}).Times(1).Return(uint64(0), errors.New("Internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					&v1.ProductAcquiredRights{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
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
