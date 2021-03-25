// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"optisam-backend/equipment-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type productQueryMatcherProductView struct {
	q *repo.QueryProductAggregations
	t *testing.T
}

func (p *productQueryMatcherProductView) Matches(x interface{}) bool {
	expQ, ok := x.(*repo.QueryProductAggregations)
	if !ok {
		return ok
	}
	return compareQueryProductView(p, expQ)
}
func compareQueryProductView(p *productQueryMatcherProductView, exp *repo.QueryProductAggregations) bool {
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
	if !compareQueryFilters(p.t, "productQueryMatcherProductView", p.q.Filter.Filters, exp.Filter.Filters) {
		return false
	}
	if !compareQueryFilters(p.t, "productQueryMatcherProductView", p.q.ProductFilter.Filters, exp.ProductFilter.Filters) {
		return false
	}
	return true
}
func (p *productQueryMatcherProductView) String() string {
	return "productQueryMatcherProductView"
}

type productQueryMatcherApplications struct {
	q *repo.QueryApplicationsForProductAggregation
	t *testing.T
}

func (p *productQueryMatcherApplications) Matches(x interface{}) bool {
	expQ, ok := x.(*repo.QueryApplicationsForProductAggregation)
	if !ok {
		return ok
	}
	return compareQueryApplication(p, expQ)
}
func compareQueryApplication(p *productQueryMatcherApplications, exp *repo.QueryApplicationsForProductAggregation) bool {
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
	if !compareQueryFilters(p.t, "productQueryMatcherApplications", p.q.Filter.Filters, exp.Filter.Filters) {
		return false
	}
	return true
}
func (p *productQueryMatcherApplications) String() string {
	return "productQueryMatcherApplications"
}

type productQueryMatcherEquipments struct {
	q *repo.QueryEquipments
	t *testing.T
}

func (p *productQueryMatcherEquipments) Matches(x interface{}) bool {
	expQ, ok := x.(*repo.QueryEquipments)
	if !ok {
		return ok
	}
	return compareQueryEquipment(p, expQ)
}
func compareQueryEquipment(p *productQueryMatcherEquipments, exp *repo.QueryEquipments) bool {
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
	if !compareQueryFiltersWithoutOrder(p.t, "productQueryMatcherEquipments", p.q.Filter.Filters, exp.Filter.Filters) {
		return false
	}
	return true
}
func (p *productQueryMatcherEquipments) String() string {
	return "productQueryMatcherEquipments"
}
func Test_equipmentServiceServer_ListEquipmentsForProductAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.Equipment

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
	}
	queryParams := &repo.QueryEquipments{
		PageSize:  10,
		Offset:    90,
		SortBy:    "attr1",
		SortOrder: repo.SortASC,
		Filter: &repo.AggregateFilter{
			Filters: []repo.Queryable{
				&repo.Filter{
					FilterKey:   "attr1",
					FilterValue: "a11",
				},
				&repo.Filter{
					FilterKey:   "attr2",
					FilterValue: "a22",
				},
			},
		},
	}

	type args struct {
		ctx context.Context
		req *v1.ListEquipmentsForProductAggregationRequest
	}
	tests := []struct {
		name    string
		s       *equipmentServiceServer
		args    args
		setup   func()
		want    *v1.ListEquipmentsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					Name:         "agg1",
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ListEquipmentsForProductAggregation(ctx, "agg1", eqTypes[0], &productQueryMatcherEquipments{
					q: queryParams,
					t: t,
				}, []string{"Scope1"}).Times(1).Return(int32(2), json.RawMessage(`[{ID:"1"}]`), nil)

			},
			want: &v1.ListEquipmentsResponse{
				TotalRecords: 2,
				Equipments:   json.RawMessage(`[{ID:"1"}]`),
			},
			wantErr: false,
		},
		{name: "FAILURE - ListEquipmentsForProductAggregation - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ListEquipmentsForProductAggregation - some claims are not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope3"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ListEquipmentsForProductAggregation - cannot fetch equipment types",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - ListEquipmentsForProductAggregation - cannot fetch equipment type with given Id",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "3",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE- ListEquipmentsForProductAggregation - cannot find sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr3",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE-ListEquipmentsForProductAggregation- cannot sort by attribute",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return([]*repo.EquipmentType{
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
		{name: "FAILURE-ListEquipmentsForProductAggregation- cannot parse equipment query param",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr3=att3",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypes, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - ListEquipmentsForProductAggregation - cannot fetch product aggregation equipments",
			args: args{
				ctx: ctx,
				req: &v1.ListEquipmentsForProductAggregationRequest{
					EqTypeId:     "1",
					SortBy:       "attr1",
					SearchParams: "attr1=a11,attr2=a22",
					Name:         "agg1",
					PageNum:      10,
					PageSize:     10,
					SortOrder:    v1.SortOrder_ASC,
					Scopes:       []string{"Scope1"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicense := mock.NewMockEquipment(mockCtrl)
				rep = mockLicense
				mockLicense.EXPECT().EquipmentTypes(ctx, []string{"Scope1"}).Times(1).Return(eqTypes, nil)
				mockLicense.EXPECT().ListEquipmentsForProductAggregation(ctx, "agg1", eqTypes[0], &productQueryMatcherEquipments{
					q: queryParams,
					t: t,
				}, []string{"Scope1"}).Times(1).Return(int32(2), nil, errors.New("test error"))

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewEquipmentServiceServer(rep)
			got, err := s.ListEquipmentsForProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("equipmentServiceServer.ListEquipmentsForProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListEquipmentResponse(t, "ListEquipmentsForProductAggregation", got, tt.want)
			}
		})
	}
}

func compareListEquipmentResponse(t *testing.T, name string, exp *v1.ListEquipmentsResponse, act *v1.ListEquipmentsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	assert.Equalf(t, exp.TotalRecords, act.TotalRecords, "%s.TotalRecords are not same", name)
	assert.Equalf(t, exp.Equipments, act.Equipments, "%s.Equipments are not same", name)
}

func compareQueryFiltersWithoutOrder(t *testing.T, name string, expFilter []repo.Queryable, actFilter []repo.Queryable) bool {
	for i := range expFilter {
		idx := queryFilterindex(expFilter[i].Key(), actFilter)
		if idx == -1 {
			return false
		}
		if !compareQueryFilter(t, fmt.Sprintf("%s[%d]", name, i), expFilter[i], actFilter[idx]) {
			return false
		}
	}
	return true
}

func queryFilterindex(key string, filter []repo.Queryable) int {
	for i := range filter {
		if key == filter[i].Key() {
			return i
		}
	}
	return -1
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
	//  return false
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
