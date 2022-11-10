package v1

import (
	"context"
	"errors"
	"fmt"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
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
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A"},
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
		{
			name: "SUCCESS - computelicenseOPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
				}, nil)

				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
					Name:           "OPS",
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(8), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS - 0 acquired licenses",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1,s2",
						Metric:            "OPS",
						AcqLicenses:       0,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
				}, nil)

				mat := &repo.MetricOPSComputed{
					EqTypeTree:     []*repo.EquipmentType{start, base, agg, end},
					BaseType:       base,
					AggregateLevel: agg,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
					Name:           "OPS",
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(8), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1,s2",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 0,
						TotalCost:      20,
						DeltaNumber:    -8,
						DeltaCost:      -140,
					},
				},
			},
		},
		{
			name: "SUCCESS - no equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
					{
						SKU:               "s3",
						Metric:            "ONS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 0,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    5,
						DeltaCost:      20,
						NotDeployed:    true,
					},
					{
						SKU:            "s2",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "WS",
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    10,
						DeltaCost:      50,
						NotDeployed:    true,
					},
					{
						SKU:            "s3",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    10,
						DeltaCost:      50,
						NotDeployed:    true,
					},
				},
			},
		},
		{
			name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - cannot fetch products acquired rights",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("", "", nil, errors.New(""))

			},
			wantErr: true,
		},
		{
			name: "SUCCESS - Acquired Rights for Product not found",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("", "", nil, repo.ErrNodeNotFound)

			},
			wantErr: false,
			want:    &v1.ListAcquiredRightsForProductResponse{},
		},
		{
			name: "FAILURE - cannot fetch products",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{
			name: "FAILURE - cannot fetch metrices  list",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test srror"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return(nil, errors.New(""))

			},
			wantErr: true,
		},
		{
			name: "SUCCESS - computelicenseOPS failed- cannot fetch metric OPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "NUP",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "NUP",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- parent hierarchy not found",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- cannot find base level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e9",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- cannot find aggregate level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e9",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- cannot find end level equipment",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e9",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- levels are not in valid order",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e3",
						AggerateLevelEqTypeID: "e2",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- numOfcores attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- numOfcpu attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- corefactor attribute not valid/exists",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseOPS failed- cannot compute metric licenses",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricOPSOracleProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricOPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricOPS{
					{
						Name:                  "OPS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
						Name:                  "WS",
						NumCoreAttrID:         "cores",
						NumCPUAttrID:          "cpus",
						CoreFactorAttrID:      "corefactor",
						BaseEqTypeID:          "e2",
						AggerateLevelEqTypeID: "e3",
						StartEqTypeID:         "e1",
						EndEqTypeID:           "e4",
					},
					{
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
					Name:           "OPS",
				}
				mockRepo.EXPECT().MetricOPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(0), errors.New(""))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS - licenseProd<=licenseNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
					{
						SKU:               "s3",
						Metric:            "ONS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(8), uint64(8), nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(6), uint64(6), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					{
						SKU:            "s2",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					{
						SKU:            "s3",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS - licenseProd>licenseNonProd",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
					{
						SKU:               "s3",
						Metric:            "ONS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(8), uint64(6), nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(6), uint64(4), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					{
						SKU:            "s2",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					{
						SKU:            "s3",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseSPS failed - cannot fetch metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseSPS failed - metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS failed- cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS failed- numofcores attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS failed- coreFactor attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseSPS failed- cannot compute licenses for metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricSPSSagProcessorStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricSPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricSPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricSPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricSPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(0), uint64(0), errors.New(""))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseIPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
					{
						SKU:               "s2",
						Metric:            "WS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
					{
						SKU:               "s3",
						Metric:            "ONS",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 50,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					NumCPUAttr:     cpu,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(8), nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						NumCPUAttrID:     "cpus",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(6), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "OPS",
						NumCptLicences: 8,
						NumAcqLicences: 5,
						TotalCost:      20,
						DeltaNumber:    -3,
						DeltaCost:      -12,
					},
					{
						SKU:            "s2",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "WS",
						NumCptLicences: 6,
						NumAcqLicences: 10,
						TotalCost:      50,
						DeltaNumber:    4,
						DeltaCost:      20,
					},
					{
						SKU:            "s3",
						SwidTag:        "P1",
						ProductName:    "pname",
						Metric:         "ONS",
						NumAcqLicences: 10,
						TotalCost:      50,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseIPS failed - cannot fetch metric IPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return(nil, errors.New("test error"))

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computelicenseIPS failed - metric name doesnot exist",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name: "IMB",
					},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseIPS failed- cannot find base level equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e6",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseIPS failed- numofcores attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseIPS failed- coreFactor attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseIPS failed- cannot compute licenses for metric SPS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "P1",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "P1", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "P1", []*repo.Metric{
					{
						Name: "OPS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
					{
						Name: "WS",
						Type: repo.MetricIPSIbmPvuStandard,
					},
				}, false, []string{"A"}).Times(1).Return("pp1", "pname", []*repo.ProductAcquiredRight{
					{
						SKU:               "s1",
						Metric:            "OPS",
						AcqLicenses:       5,
						TotalCost:         20,
						TotalPurchaseCost: 20,
						AvgUnitPrice:      4,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "P1", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)

				mockRepo.EXPECT().ListMetricIPS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricIPS{
					{
						Name:             "OPS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name:             "WS",
						NumCoreAttrID:    "cores",
						CoreFactorAttrID: "corefactor",
						BaseEqTypeID:     "e2",
					},
					{
						Name: "IMB",
					},
				}, nil)

				mat := &repo.MetricIPSComputed{
					BaseType:       base,
					NumCoresAttr:   cores,
					CoreFactorAttr: corefactor,
				}
				mockRepo.EXPECT().MetricIPSComputedLicenses(ctx, "pp1", mat, []string{"A"}).Times(1).Return(uint64(0), errors.New(""))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "s1",
						SwidTag:        "P1",
						Metric:         "OPS",
						NumAcqLicences: 5,
						TotalCost:      20,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         2100,
						TotalPurchaseCost: 2000,
						AvgUnitPrice:      100,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricACS{
					{
						Name:          "attribute.counter.standard",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					{
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
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "uidORAC001", mat, []string{"A"}).Times(1).Return(uint64(10), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumCptLicences: 10,
						NumAcqLicences: 20,
						TotalCost:      2100,
						DeltaNumber:    10,
						DeltaCost:      1100,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS failed - cannot fetch acs metrics",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         9270,
						TotalPurchaseCost: 9270,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS failed - cannot find metric name acs",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         9270,
						TotalPurchaseCost: 9270,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricACS{
					{
						Name:          "acs",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS failed - cannot find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         9270,
						TotalPurchaseCost: 9270,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricACS{
					{
						Name:          "attribute.counter.standard",
						EqType:        "cluster",
						AttributeName: "corefactor",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS failed - attribute doesnt exits",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         9270,
						TotalPurchaseCost: 9270,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricACS{
					{
						Name:          "attribute.counter.standard",
						EqType:        "server",
						AttributeName: "servermodel",
						Value:         "2",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseACS failed - cannot compute licenses for metric ACS",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS",
						Metric:            "attribute.counter.standard",
						AcqLicenses:       20,
						TotalCost:         9270,
						TotalPurchaseCost: 9270,
						AvgUnitPrice:      20,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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

				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{start, base, agg, end, endP}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"A"}).Times(1).Return([]*repo.MetricACS{
					{
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
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "uidORAC001", mat, []string{"A"}).Times(1).Return(uint64(0), errors.New("Internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.counter.standard",
						NumAcqLicences: 20,
						TotalCost:      9270,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "attribute.sum.standard",
						EqType:         "server",
						AttributeName:  "corefactor",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)

				mat := &repo.MetricAttrSumStandComputed{
					Name:           "attribute.sum.standard",
					BaseType:       base,
					Attribute:      corefactor,
					ReferenceValue: 10,
				}
				mockRepo.EXPECT().MetricAttrSumComputedLicenses(ctx, "uidORAC001", mat, []string{"A"}).Times(1).Return(uint64(166), uint64(1660), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:             "ORAC001ACS,ORAC002ACS",
						SwidTag:         "ORAC001",
						Metric:          "attribute.sum.standard",
						NumCptLicences:  166,
						NumAcqLicences:  200,
						TotalCost:       1000,
						DeltaNumber:     34,
						DeltaCost:       170,
						ComputedDetails: "Sum of values:1660",
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum failed - can not fetch attr sum metric",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return(nil, errors.New("internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum failed - can not find metric name attribute.sum.standard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpu",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum failed - can not find equipment type",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)

				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "attribute.sum.standard",
						EqType:         "cluster",
						AttributeName:  "corefactor",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum failed - can not find attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "attribute.sum.standard",
						EqType:         "server",
						AttributeName:  "corefactors",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseAttrSum failed - can not compute license",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, nil)

				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "attribute.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "attribute.sum.standard",
						EqType:         "server",
						AttributeName:  "corefactor",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)

				mat := &repo.MetricAttrSumStandComputed{
					Name:           "attribute.sum.standard",
					BaseType:       base,
					Attribute:      corefactor,
					ReferenceValue: 10,
				}
				mockRepo.EXPECT().MetricAttrSumComputedLicenses(ctx, "uidORAC001", mat, []string{"A"}).Times(1).Return(uint64(0), uint64(0), errors.New("internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "attribute.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseUserSum",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, []string{"A"}).Times(1).Return("uidORAC001", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "user.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricUserSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricUserSumStand{
					{
						ID:   "uid1",
						Name: "user.sum.standard",
					},
					{
						ID:   "uid2",
						Name: "USS1",
					},
				}, nil)
				mockRepo.EXPECT().MetricUserSumComputedLicenses(ctx, "uidORAC001", []string{"A"}).Times(1).Return(uint64(166), uint64(1660), nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:             "ORAC001ACS,ORAC002ACS",
						SwidTag:         "ORAC001",
						Metric:          "user.sum.standard",
						NumCptLicences:  166,
						NumAcqLicences:  200,
						TotalCost:       1000,
						DeltaNumber:     34,
						DeltaCost:       170,
						ComputedDetails: "Sum of users:1660",
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseUserSum failed - can not fetch user sum metric",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "user.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricUserSum(ctx, []string{"A"}).Times(1).Return(nil, errors.New("internal"))
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "user.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - computeLicenseUserSum failed - can not find metric name user.sum.standard",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("", nil).Times(1)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "ORAC001", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}, false, []string{"A"}).Times(1).Return("uidORAC001", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "ORAC001ACS,ORAC002ACS",
						Metric:            "user.sum.standard",
						AcqLicenses:       200,
						TotalCost:         1000,
						TotalPurchaseCost: 1000,
						AvgUnitPrice:      5,
					},
				}, nil)

				mockRepo.EXPECT().GetProductInformation(ctx, "ORAC001", []string{"A"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
					},
				}, nil)

				cores := &repo.Attribute{
					ID:   "cores",
					Type: repo.DataTypeInt,
				}
				cpu := &repo.Attribute{
					ID:   "cpu",
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"A"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricUserSum(ctx, []string{"A"}).Times(1).Return([]*repo.MetricUserSumStand{
					{
						Name: "uss1",
					},
				}, nil)
			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AcqRights: []*v1.ProductAcquiredRights{
					{
						SKU:            "ORAC001ACS,ORAC002ACS",
						SwidTag:        "ORAC001",
						Metric:         "user.sum.standard",
						NumAcqLicences: 200,
						TotalCost:      1000,
					},
				},
			},
		},
		{
			name: "SUCCESS - swidtag is part of aggregates",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				metrics := []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("agg1", nil).Times(1)
				mockRepo.EXPECT().AggregationDetails(ctx, "agg1", metrics, "A").Times(1).Return(&repo.AggregationInfo{
					Name: "agg1",
				}, []*repo.ProductAcquiredRight{
					{SKU: "indsku"},
				}, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{
				AggregationName: "agg1"},
		},
		{
			name: "SUCCESS - swidtag is part of aggregates - no license bought on aggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsForProductRequest{
					SwidTag: "ORAC001",
					Scope:   "A",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				metrics := []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "attribute.counter.standard",
						Type: "attribute.counter.standard",
					},
					{
						Name: "attribute.sum.standard",
						Type: "attribute.sum.standard",
					},
					{
						Name: "user.sum.standard",
						Type: "user.sum.standard",
					},
				}
				mockRepo.EXPECT().ListMetrices(ctx, []string{"A"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().IsProductPurchasedInAggregation(ctx, "ORAC001", "A").Return("agg1", nil).Times(1)
				mockRepo.EXPECT().AggregationDetails(ctx, "agg1", metrics, "A").Times(1).Return(&repo.AggregationInfo{
					Name: "agg1",
				}, nil, nil)

			},
			want: &v1.ListAcquiredRightsForProductResponse{},
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
			} else {
				fmt.Println("Test case passed : [ ", tt.name, "]")
			}
		})
	}
}

func Test_licenseServiceServer_ListComputationDetails(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.Metric{
		{
			Name: "oracle.processor.standard",
			Type: "oracle.processor.standard",
		},
		{
			Name: "oracle.nup.standard",
			Type: "oracle.nup.standard",
		},
		{
			Name: "sag.processor.standard",
			Type: "sag.processor.standard",
		},
		{
			Name: "ibm.pvu.standard",
			Type: "ibm.pvu.standard",
		},
		{
			Name: "met1",
			Type: "attribute.counter.standard",
		},
		{
			Name: "met2",
			Type: "attribute.sum.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.ListComputationDetailsRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.ListComputationDetailsResponse
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"scope1"}).Times(1).Return([]*repo.MetricACS{
					{
						Name:          "met1",
						EqType:        "server",
						AttributeName: "corefactor",
						Value:         "2",
					},
					{
						Name:          "ACS1",
						EqType:        "server",
						AttributeName: "cpu",
						Value:         "2",
					},
				}, nil)

				acsmat := &repo.MetricACSComputed{
					Name:      "met1",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				mockRepo.EXPECT().MetricACSComputedLicenses(ctx, "uidprodswid", acsmat, []string{"scope1"}).Times(1).Return(uint64(10), nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"scope1"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "met2",
						EqType:         "server",
						AttributeName:  "corefactor",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)

				assmat := &repo.MetricAttrSumStandComputed{
					Name:           "met2",
					BaseType:       base,
					Attribute:      corefactor,
					ReferenceValue: 10,
				}
				mockRepo.EXPECT().MetricAttrSumComputedLicenses(ctx, "uidprodswid", assmat, []string{"scope1"}).Times(1).Return(uint64(5), uint64(50), nil)
			},
			want: &v1.ListComputationDetailsResponse{
				ComputedDetails: []*v1.ComputedDetails{
					{
						MetricName:      "met1",
						NumAcqLicences:  10,
						NumCptLicences:  10,
						DeltaNumber:     0,
						DeltaCost:       0,
						ComputedDetails: "",
					},
					{
						MetricName:      "met2",
						NumAcqLicences:  10,
						NumCptLicences:  5,
						DeltaNumber:     5,
						DeltaCost:       25,
						ComputedDetails: "Sum of values: 50",
					},
				},
			},
		},
		{name: "SUCCESS - Aggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					AggName: "agg1",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().AggregationDetails(ctx, "agg1", metrics, "scope1").Times(1).Return(&repo.AggregationInfo{
					ID:                1,
					Name:              "agg1",
					ProductNames:      []string{"prod1", "prod2"},
					Swidtags:          []string{"swid1", "swid2"},
					ProductIDs:        []string{"prodid1", "prodid2"},
					Editor:            "prodeditor",
					NumOfApplications: 10,
					NumOfEquipments:   56,
				}, []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				acsmat := &repo.MetricACSComputed{
					Name:      "met1",
					BaseType:  base,
					Attribute: corefactor,
					Value:     "2",
				}
				assmat := &repo.MetricAttrSumStandComputed{
					Name:           "met2",
					BaseType:       base,
					Attribute:      corefactor,
					ReferenceValue: 10,
				}
				gomock.InOrder(
					mockRepo.EXPECT().AggregationIndividualRights(ctx, []string{"prodid1", "prodid2"}, []string{"met1"}, "scope1").Times(1).Return([]*repo.AcqRightsInfo{
						{
							SKU:                  "indsku1",
							Swidtag:              "prodswid1",
							ProductName:          "prod1",
							ProductEditor:        "prodeditor",
							ProductVersion:       "prodver1",
							Metric:               []string{"met1"},
							Licenses:             10,
							MaintenanceLicenses:  0,
							UnitPrice:            10,
							MaintenanceUnitPrice: 0,
							PurchaseCost:         100,
							MaintenanceCost:      0,
							TotalCost:            100,
							StartOfMaintenance:   "",
							EndOfMaintenance:     "",
						},
					}, nil),
					mockRepo.EXPECT().ListMetricACS(ctx, []string{"scope1"}).Times(1).Return([]*repo.MetricACS{
						{
							Name:          "met1",
							EqType:        "server",
							AttributeName: "corefactor",
							Value:         "2",
						},
						{
							Name:          "ACS1",
							EqType:        "server",
							AttributeName: "cpu",
							Value:         "2",
						},
					}, nil),
					mockRepo.EXPECT().MetricACSComputedLicensesAgg(ctx, "agg1", "met1", acsmat, []string{"scope1"}).Times(1).Return(uint64(10), nil),
					mockRepo.EXPECT().AggregationIndividualRights(ctx, []string{"prodid1", "prodid2"}, []string{"met2"}, "scope1").Times(1).Return([]*repo.AcqRightsInfo{}, nil),
					mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"scope1"}).Times(1).Return([]*repo.MetricAttrSumStand{
						{
							Name:           "met2",
							EqType:         "server",
							AttributeName:  "corefactor",
							ReferenceValue: 10,
						},
						{
							Name:           "ASS1",
							EqType:         "server",
							AttributeName:  "cpu",
							ReferenceValue: 2,
						},
					}, nil),
					mockRepo.EXPECT().MetricAttrSumComputedLicensesAgg(ctx, "agg1", "met2", assmat, []string{"scope1"}).Times(1).Return(uint64(5), uint64(50), nil),
				)
			},
			want: &v1.ListComputationDetailsResponse{
				ComputedDetails: []*v1.ComputedDetails{
					{
						MetricName:      "met1",
						NumAcqLicences:  20,
						NumCptLicences:  10,
						DeltaNumber:     10,
						DeltaCost:       100,
						ComputedDetails: "",
					},
					{
						MetricName:      "met2",
						NumAcqLicences:  10,
						NumCptLicences:  5,
						DeltaNumber:     5,
						DeltaCost:       25,
						ComputedDetails: "Sum of values: 50",
					},
				},
			},
		},
		{name: "SUCCESS - no equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 0,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
			},
			want: &v1.ListComputationDetailsResponse{
				ComputedDetails: []*v1.ComputedDetails{
					{
						MetricName:      "met1",
						NumAcqLicences:  10,
						NumCptLicences:  0,
						DeltaNumber:     10,
						DeltaCost:       50,
						ComputedDetails: "",
					},
					{
						MetricName:      "met2",
						NumAcqLicences:  10,
						NumCptLicences:  0,
						DeltaNumber:     10,
						DeltaCost:       50,
						ComputedDetails: "",
					},
				},
			},
		},
		{name: "SUCCESS - one metric does not exist",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "met2",
						Type: "attribute.sum.standard",
					},
				}, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "met2",
						Type: "attribute.sum.standard",
					},
				}, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricAttrSum(ctx, []string{"scope1"}).Times(1).Return([]*repo.MetricAttrSumStand{
					{
						Name:           "met2",
						EqType:         "server",
						AttributeName:  "corefactor",
						ReferenceValue: 10,
					},
					{
						Name:           "ASS1",
						EqType:         "server",
						AttributeName:  "cpu",
						ReferenceValue: 2,
					},
				}, nil)

				assmat := &repo.MetricAttrSumStandComputed{
					Name:           "met2",
					BaseType:       base,
					Attribute:      corefactor,
					ReferenceValue: 10,
				}
				mockRepo.EXPECT().MetricAttrSumComputedLicenses(ctx, "uidprodswid", assmat, []string{"scope1"}).Times(1).Return(uint64(5), uint64(50), nil)
			},
			want: &v1.ListComputationDetailsResponse{
				ComputedDetails: []*v1.ComputedDetails{
					{
						MetricName:      "met2",
						NumAcqLicences:  10,
						NumCptLicences:  5,
						DeltaNumber:     5,
						DeltaCost:       25,
						ComputedDetails: "Sum of values: 50",
					},
				},
			},
		},
		{name: "FAILURE - can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope3",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - db/ProductAcquiredRights - does not exist",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("", "", nil, repo.ErrNodeNotFound)
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/ProductAcquiredRights",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("", "", nil, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - acqruired right does not exist",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2,sku3",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
					},
				}, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/GetProductInformation",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(nil, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/ListMetrics",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(nil, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - repo/ListEquipments",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 30,
						},
					},
				}, nil)
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{}, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - metric type not supported",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return([]*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "met2",
						Type: "unknown",
					},
				}, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", []*repo.Metric{
					{
						Name: "oracle.processor.standard",
						Type: "oracle.processor.standard",
					},
					{
						Name: "oracle.nup.standard",
						Type: "oracle.nup.standard",
					},
					{
						Name: "sag.processor.standard",
						Type: "sag.processor.standard",
					},
					{
						Name: "ibm.pvu.standard",
						Type: "ibm.pvu.standard",
					},
					{
						Name: "met2",
						Type: "unknown",
					},
				}, []string{"scope1"}).Times(1).Return("uidprodswid", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1,met2",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - MetricCalculation",
			args: args{
				ctx: ctx,
				req: &v1.ListComputationDetailsRequest{
					SwidTag: "prodswid",
					Sku:     "sku1,sku2",
					Scope:   "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListMetrices(ctx, []string{"scope1"}).Times(1).Return(metrics, nil)
				mockRepo.EXPECT().ProductAcquiredRights(ctx, "prodswid", metrics, false, []string{"scope1"}).Times(1).Return("uidprodswid", "P1", []*repo.ProductAcquiredRight{
					{
						SKU:               "sku1,sku2",
						Metric:            "met1",
						AcqLicenses:       10,
						TotalCost:         50,
						TotalPurchaseCost: 60,
						AvgUnitPrice:      5,
					},
				}, nil)
				mockRepo.EXPECT().GetProductInformation(ctx, "prodswid", []string{"scope1"}).Times(1).Return(&repo.ProductAdditionalInfo{
					Products: []repo.ProductAdditionalData{
						{
							NumofEquipments: 56,
						},
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
				mockRepo.EXPECT().EquipmentTypes(ctx, []string{"scope1"}).Times(1).Return([]*repo.EquipmentType{base}, nil)
				mockRepo.EXPECT().ListMetricACS(ctx, []string{"scope1"}).Times(1).Return(nil, errors.New("internal"))
			},
			want: &v1.ListComputationDetailsResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewLicenseServiceServer(rep)
			got, err := s.ListComputationDetails(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
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

func compareListComputationDetailsResponse(t *testing.T, name string, exp *v1.ListComputationDetailsResponse, act *v1.ListComputationDetailsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "computed details are expected to be nil")
	}
	compareComputationDetailsAll(t, name+".ComputedDetails", exp.ComputedDetails, act.ComputedDetails)
}

func compareComputationDetailsAll(t *testing.T, name string, exp []*v1.ComputedDetails, act []*v1.ComputedDetails) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareComputationDetails(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareComputationDetails(t *testing.T, name string, exp *v1.ComputedDetails, act *v1.ComputedDetails) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "computed detail is expected to be nil")
	}
	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.MetricName are not same", name)
	assert.Equalf(t, exp.ComputedDetails, act.ComputedDetails, "%s.ComputedDetails are not same", name)
	assert.Equalf(t, exp.NumCptLicences, act.NumCptLicences, "%s.Computed Licenses are not same", name)
	assert.Equalf(t, exp.NumAcqLicences, act.NumAcqLicences, "%s.Acquired Licenses are not same", name)
	assert.Equalf(t, exp.DeltaNumber, act.DeltaNumber, "%s.Delta Numbers are not same", name)
	assert.Equalf(t, exp.DeltaCost, act.DeltaCost, "%s.Delta Cost is not same", name)
}
