// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/product-service/pkg/repository/v1 (interfaces: Product)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/product-service/pkg/api/v1"
	db "optisam-backend/product-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockProduct is a mock of Product interface
type MockProduct struct {
	ctrl     *gomock.Controller
	recorder *MockProductMockRecorder
}

// MockProductMockRecorder is the mock recorder for MockProduct
type MockProductMockRecorder struct {
	mock *MockProduct
}

// NewMockProduct creates a new mock instance
func NewMockProduct(ctrl *gomock.Controller) *MockProduct {
	mock := &MockProduct{ctrl: ctrl}
	mock.recorder = &MockProductMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProduct) EXPECT() *MockProductMockRecorder {
	return m.recorder
}

// AddComputedLicenses mocks base method
func (m *MockProduct) AddComputedLicenses(arg0 context.Context, arg1 db.AddComputedLicensesParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddComputedLicenses", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddComputedLicenses indicates an expected call of AddComputedLicenses
func (mr *MockProductMockRecorder) AddComputedLicenses(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddComputedLicenses", reflect.TypeOf((*MockProduct)(nil).AddComputedLicenses), arg0, arg1)
}

// CounterFeitedProductsCosts mocks base method
func (m *MockProduct) CounterFeitedProductsCosts(arg0 context.Context, arg1 db.CounterFeitedProductsCostsParams) ([]db.CounterFeitedProductsCostsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CounterFeitedProductsCosts", arg0, arg1)
	ret0, _ := ret[0].([]db.CounterFeitedProductsCostsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CounterFeitedProductsCosts indicates an expected call of CounterFeitedProductsCosts
func (mr *MockProductMockRecorder) CounterFeitedProductsCosts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CounterFeitedProductsCosts", reflect.TypeOf((*MockProduct)(nil).CounterFeitedProductsCosts), arg0, arg1)
}

// CounterFeitedProductsLicences mocks base method
func (m *MockProduct) CounterFeitedProductsLicences(arg0 context.Context, arg1 db.CounterFeitedProductsLicencesParams) ([]db.CounterFeitedProductsLicencesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CounterFeitedProductsLicences", arg0, arg1)
	ret0, _ := ret[0].([]db.CounterFeitedProductsLicencesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CounterFeitedProductsLicences indicates an expected call of CounterFeitedProductsLicences
func (mr *MockProductMockRecorder) CounterFeitedProductsLicences(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CounterFeitedProductsLicences", reflect.TypeOf((*MockProduct)(nil).CounterFeitedProductsLicences), arg0, arg1)
}

// CounterfeitPercent mocks base method
func (m *MockProduct) CounterfeitPercent(arg0 context.Context, arg1 string) (db.CounterfeitPercentRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CounterfeitPercent", arg0, arg1)
	ret0, _ := ret[0].(db.CounterfeitPercentRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CounterfeitPercent indicates an expected call of CounterfeitPercent
func (mr *MockProductMockRecorder) CounterfeitPercent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CounterfeitPercent", reflect.TypeOf((*MockProduct)(nil).CounterfeitPercent), arg0, arg1)
}

// DeleteAcqrightsByScope mocks base method
func (m *MockProduct) DeleteAcqrightsByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAcqrightsByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAcqrightsByScope indicates an expected call of DeleteAcqrightsByScope
func (mr *MockProductMockRecorder) DeleteAcqrightsByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAcqrightsByScope", reflect.TypeOf((*MockProduct)(nil).DeleteAcqrightsByScope), arg0, arg1)
}

// DeleteAggregation mocks base method
func (m *MockProduct) DeleteAggregation(arg0 context.Context, arg1 db.DeleteAggregationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAggregation", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAggregation indicates an expected call of DeleteAggregation
func (mr *MockProductMockRecorder) DeleteAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAggregation", reflect.TypeOf((*MockProduct)(nil).DeleteAggregation), arg0, arg1)
}

// DeleteProductAggregation mocks base method
func (m *MockProduct) DeleteProductAggregation(arg0 context.Context, arg1 db.DeleteProductAggregationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductAggregation", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProductAggregation indicates an expected call of DeleteProductAggregation
func (mr *MockProductMockRecorder) DeleteProductAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductAggregation", reflect.TypeOf((*MockProduct)(nil).DeleteProductAggregation), arg0, arg1)
}

// DeleteProductAggregationByScope mocks base method
func (m *MockProduct) DeleteProductAggregationByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductAggregationByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProductAggregationByScope indicates an expected call of DeleteProductAggregationByScope
func (mr *MockProductMockRecorder) DeleteProductAggregationByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductAggregationByScope", reflect.TypeOf((*MockProduct)(nil).DeleteProductAggregationByScope), arg0, arg1)
}

// DeleteProductApplications mocks base method
func (m *MockProduct) DeleteProductApplications(arg0 context.Context, arg1 db.DeleteProductApplicationsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductApplications", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProductApplications indicates an expected call of DeleteProductApplications
func (mr *MockProductMockRecorder) DeleteProductApplications(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductApplications", reflect.TypeOf((*MockProduct)(nil).DeleteProductApplications), arg0, arg1)
}

// DeleteProductEquipments mocks base method
func (m *MockProduct) DeleteProductEquipments(arg0 context.Context, arg1 db.DeleteProductEquipmentsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductEquipments", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProductEquipments indicates an expected call of DeleteProductEquipments
func (mr *MockProductMockRecorder) DeleteProductEquipments(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductEquipments", reflect.TypeOf((*MockProduct)(nil).DeleteProductEquipments), arg0, arg1)
}

// DeleteProductsByScope mocks base method
func (m *MockProduct) DeleteProductsByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductsByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteProductsByScope indicates an expected call of DeleteProductsByScope
func (mr *MockProductMockRecorder) DeleteProductsByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductsByScope", reflect.TypeOf((*MockProduct)(nil).DeleteProductsByScope), arg0, arg1)
}

// DropProductDataTx mocks base method
func (m *MockProduct) DropProductDataTx(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropProductDataTx", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DropProductDataTx indicates an expected call of DropProductDataTx
func (mr *MockProductMockRecorder) DropProductDataTx(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropProductDataTx", reflect.TypeOf((*MockProduct)(nil).DropProductDataTx), arg0, arg1)
}

// EquipmentProducts mocks base method
func (m *MockProduct) EquipmentProducts(arg0 context.Context, arg1 string) ([]db.ProductsEquipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentProducts", arg0, arg1)
	ret0, _ := ret[0].([]db.ProductsEquipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentProducts indicates an expected call of EquipmentProducts
func (mr *MockProductMockRecorder) EquipmentProducts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentProducts", reflect.TypeOf((*MockProduct)(nil).EquipmentProducts), arg0, arg1)
}

// GetAcqRightsCost mocks base method
func (m *MockProduct) GetAcqRightsCost(arg0 context.Context, arg1 []string) (db.GetAcqRightsCostRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAcqRightsCost", arg0, arg1)
	ret0, _ := ret[0].(db.GetAcqRightsCostRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAcqRightsCost indicates an expected call of GetAcqRightsCost
func (mr *MockProductMockRecorder) GetAcqRightsCost(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAcqRightsCost", reflect.TypeOf((*MockProduct)(nil).GetAcqRightsCost), arg0, arg1)
}

// GetProductAggregation mocks base method
func (m *MockProduct) GetProductAggregation(arg0 context.Context, arg1 db.GetProductAggregationParams) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductAggregation", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductAggregation indicates an expected call of GetProductAggregation
func (mr *MockProductMockRecorder) GetProductAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductAggregation", reflect.TypeOf((*MockProduct)(nil).GetProductAggregation), arg0, arg1)
}

// GetProductInformation mocks base method
func (m *MockProduct) GetProductInformation(arg0 context.Context, arg1 db.GetProductInformationParams) (db.GetProductInformationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductInformation", arg0, arg1)
	ret0, _ := ret[0].(db.GetProductInformationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductInformation indicates an expected call of GetProductInformation
func (mr *MockProductMockRecorder) GetProductInformation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductInformation", reflect.TypeOf((*MockProduct)(nil).GetProductInformation), arg0, arg1)
}

// GetProductOptions mocks base method
func (m *MockProduct) GetProductOptions(arg0 context.Context, arg1 db.GetProductOptionsParams) ([]db.GetProductOptionsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductOptions", arg0, arg1)
	ret0, _ := ret[0].([]db.GetProductOptionsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductOptions indicates an expected call of GetProductOptions
func (mr *MockProductMockRecorder) GetProductOptions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductOptions", reflect.TypeOf((*MockProduct)(nil).GetProductOptions), arg0, arg1)
}

// GetProductQualityOverview mocks base method
func (m *MockProduct) GetProductQualityOverview(arg0 context.Context, arg1 string) (db.GetProductQualityOverviewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductQualityOverview", arg0, arg1)
	ret0, _ := ret[0].(db.GetProductQualityOverviewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductQualityOverview indicates an expected call of GetProductQualityOverview
func (mr *MockProductMockRecorder) GetProductQualityOverview(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductQualityOverview", reflect.TypeOf((*MockProduct)(nil).GetProductQualityOverview), arg0, arg1)
}

// GetProductsByEditor mocks base method
func (m *MockProduct) GetProductsByEditor(arg0 context.Context, arg1 db.GetProductsByEditorParams) ([]db.GetProductsByEditorRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductsByEditor", arg0, arg1)
	ret0, _ := ret[0].([]db.GetProductsByEditorRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductsByEditor indicates an expected call of GetProductsByEditor
func (mr *MockProductMockRecorder) GetProductsByEditor(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductsByEditor", reflect.TypeOf((*MockProduct)(nil).GetProductsByEditor), arg0, arg1)
}

// InsertAggregation mocks base method
func (m *MockProduct) InsertAggregation(arg0 context.Context, arg1 db.InsertAggregationParams) (db.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertAggregation", arg0, arg1)
	ret0, _ := ret[0].(db.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertAggregation indicates an expected call of InsertAggregation
func (mr *MockProductMockRecorder) InsertAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertAggregation", reflect.TypeOf((*MockProduct)(nil).InsertAggregation), arg0, arg1)
}

// ListAcqRightsAggregation mocks base method
func (m *MockProduct) ListAcqRightsAggregation(arg0 context.Context, arg1 db.ListAcqRightsAggregationParams) ([]db.ListAcqRightsAggregationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsAggregation", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsAggregationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsAggregation indicates an expected call of ListAcqRightsAggregation
func (mr *MockProductMockRecorder) ListAcqRightsAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsAggregation", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsAggregation), arg0, arg1)
}

// ListAcqRightsAggregationIndividual mocks base method
func (m *MockProduct) ListAcqRightsAggregationIndividual(arg0 context.Context, arg1 db.ListAcqRightsAggregationIndividualParams) ([]db.ListAcqRightsAggregationIndividualRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsAggregationIndividual", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsAggregationIndividualRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsAggregationIndividual indicates an expected call of ListAcqRightsAggregationIndividual
func (mr *MockProductMockRecorder) ListAcqRightsAggregationIndividual(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsAggregationIndividual", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsAggregationIndividual), arg0, arg1)
}

// ListAcqRightsEditors mocks base method
func (m *MockProduct) ListAcqRightsEditors(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsEditors", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsEditors indicates an expected call of ListAcqRightsEditors
func (mr *MockProductMockRecorder) ListAcqRightsEditors(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsEditors", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsEditors), arg0, arg1)
}

// ListAcqRightsIndividual mocks base method
func (m *MockProduct) ListAcqRightsIndividual(arg0 context.Context, arg1 db.ListAcqRightsIndividualParams) ([]db.ListAcqRightsIndividualRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsIndividual", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsIndividualRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsIndividual indicates an expected call of ListAcqRightsIndividual
func (mr *MockProductMockRecorder) ListAcqRightsIndividual(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsIndividual", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsIndividual), arg0, arg1)
}

// ListAcqRightsMetrics mocks base method
func (m *MockProduct) ListAcqRightsMetrics(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsMetrics", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsMetrics indicates an expected call of ListAcqRightsMetrics
func (mr *MockProductMockRecorder) ListAcqRightsMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsMetrics", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsMetrics), arg0, arg1)
}

// ListAcqRightsProducts mocks base method
func (m *MockProduct) ListAcqRightsProducts(arg0 context.Context, arg1 db.ListAcqRightsProductsParams) ([]db.ListAcqRightsProductsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsProducts", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsProductsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsProducts indicates an expected call of ListAcqRightsProducts
func (mr *MockProductMockRecorder) ListAcqRightsProducts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsProducts", reflect.TypeOf((*MockProduct)(nil).ListAcqRightsProducts), arg0, arg1)
}

// ListAcqrightsProducts mocks base method
func (m *MockProduct) ListAcqrightsProducts(arg0 context.Context) ([]db.ListAcqrightsProductsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqrightsProducts", arg0)
	ret0, _ := ret[0].([]db.ListAcqrightsProductsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqrightsProducts indicates an expected call of ListAcqrightsProducts
func (mr *MockProductMockRecorder) ListAcqrightsProducts(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqrightsProducts", reflect.TypeOf((*MockProduct)(nil).ListAcqrightsProducts), arg0)
}

// ListAggregation mocks base method
func (m *MockProduct) ListAggregation(arg0 context.Context, arg1 []string) ([]db.ListAggregationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAggregation", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAggregationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAggregation indicates an expected call of ListAggregation
func (mr *MockProductMockRecorder) ListAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAggregation", reflect.TypeOf((*MockProduct)(nil).ListAggregation), arg0, arg1)
}

// ListAggregationProductsView mocks base method
func (m *MockProduct) ListAggregationProductsView(arg0 context.Context, arg1 db.ListAggregationProductsViewParams) ([]db.ListAggregationProductsViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAggregationProductsView", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAggregationProductsViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAggregationProductsView indicates an expected call of ListAggregationProductsView
func (mr *MockProductMockRecorder) ListAggregationProductsView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAggregationProductsView", reflect.TypeOf((*MockProduct)(nil).ListAggregationProductsView), arg0, arg1)
}

// ListAggregationsView mocks base method
func (m *MockProduct) ListAggregationsView(arg0 context.Context, arg1 db.ListAggregationsViewParams) ([]db.ListAggregationsViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAggregationsView", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAggregationsViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAggregationsView indicates an expected call of ListAggregationsView
func (mr *MockProductMockRecorder) ListAggregationsView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAggregationsView", reflect.TypeOf((*MockProduct)(nil).ListAggregationsView), arg0, arg1)
}

// ListEditors mocks base method
func (m *MockProduct) ListEditors(arg0 context.Context, arg1 []string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEditors", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEditors indicates an expected call of ListEditors
func (mr *MockProductMockRecorder) ListEditors(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEditors", reflect.TypeOf((*MockProduct)(nil).ListEditors), arg0, arg1)
}

// ListProductsView mocks base method
func (m *MockProduct) ListProductsView(arg0 context.Context, arg1 db.ListProductsViewParams) ([]db.ListProductsViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProductsView", arg0, arg1)
	ret0, _ := ret[0].([]db.ListProductsViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProductsView indicates an expected call of ListProductsView
func (mr *MockProductMockRecorder) ListProductsView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProductsView", reflect.TypeOf((*MockProduct)(nil).ListProductsView), arg0, arg1)
}

// ListProductsViewRedirectedApplication mocks base method
func (m *MockProduct) ListProductsViewRedirectedApplication(arg0 context.Context, arg1 db.ListProductsViewRedirectedApplicationParams) ([]db.ListProductsViewRedirectedApplicationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProductsViewRedirectedApplication", arg0, arg1)
	ret0, _ := ret[0].([]db.ListProductsViewRedirectedApplicationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProductsViewRedirectedApplication indicates an expected call of ListProductsViewRedirectedApplication
func (mr *MockProductMockRecorder) ListProductsViewRedirectedApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProductsViewRedirectedApplication", reflect.TypeOf((*MockProduct)(nil).ListProductsViewRedirectedApplication), arg0, arg1)
}

// ListProductsViewRedirectedEquipment mocks base method
func (m *MockProduct) ListProductsViewRedirectedEquipment(arg0 context.Context, arg1 db.ListProductsViewRedirectedEquipmentParams) ([]db.ListProductsViewRedirectedEquipmentRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListProductsViewRedirectedEquipment", arg0, arg1)
	ret0, _ := ret[0].([]db.ListProductsViewRedirectedEquipmentRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListProductsViewRedirectedEquipment indicates an expected call of ListProductsViewRedirectedEquipment
func (mr *MockProductMockRecorder) ListProductsViewRedirectedEquipment(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListProductsViewRedirectedEquipment", reflect.TypeOf((*MockProduct)(nil).ListProductsViewRedirectedEquipment), arg0, arg1)
}

// OverDeployedProductsCosts mocks base method
func (m *MockProduct) OverDeployedProductsCosts(arg0 context.Context, arg1 db.OverDeployedProductsCostsParams) ([]db.OverDeployedProductsCostsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OverDeployedProductsCosts", arg0, arg1)
	ret0, _ := ret[0].([]db.OverDeployedProductsCostsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OverDeployedProductsCosts indicates an expected call of OverDeployedProductsCosts
func (mr *MockProductMockRecorder) OverDeployedProductsCosts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OverDeployedProductsCosts", reflect.TypeOf((*MockProduct)(nil).OverDeployedProductsCosts), arg0, arg1)
}

// OverDeployedProductsLicences mocks base method
func (m *MockProduct) OverDeployedProductsLicences(arg0 context.Context, arg1 db.OverDeployedProductsLicencesParams) ([]db.OverDeployedProductsLicencesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OverDeployedProductsLicences", arg0, arg1)
	ret0, _ := ret[0].([]db.OverDeployedProductsLicencesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OverDeployedProductsLicences indicates an expected call of OverDeployedProductsLicences
func (mr *MockProductMockRecorder) OverDeployedProductsLicences(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OverDeployedProductsLicences", reflect.TypeOf((*MockProduct)(nil).OverDeployedProductsLicences), arg0, arg1)
}

// OverdeployPercent mocks base method
func (m *MockProduct) OverdeployPercent(arg0 context.Context, arg1 string) (db.OverdeployPercentRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OverdeployPercent", arg0, arg1)
	ret0, _ := ret[0].(db.OverdeployPercentRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OverdeployPercent indicates an expected call of OverdeployPercent
func (mr *MockProductMockRecorder) OverdeployPercent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OverdeployPercent", reflect.TypeOf((*MockProduct)(nil).OverdeployPercent), arg0, arg1)
}

// ProductAggregationChildOptions mocks base method
func (m *MockProduct) ProductAggregationChildOptions(arg0 context.Context, arg1 db.ProductAggregationChildOptionsParams) ([]db.ProductAggregationChildOptionsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductAggregationChildOptions", arg0, arg1)
	ret0, _ := ret[0].([]db.ProductAggregationChildOptionsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductAggregationChildOptions indicates an expected call of ProductAggregationChildOptions
func (mr *MockProductMockRecorder) ProductAggregationChildOptions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductAggregationChildOptions", reflect.TypeOf((*MockProduct)(nil).ProductAggregationChildOptions), arg0, arg1)
}

// ProductAggregationDetails mocks base method
func (m *MockProduct) ProductAggregationDetails(arg0 context.Context, arg1 db.ProductAggregationDetailsParams) (db.ProductAggregationDetailsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductAggregationDetails", arg0, arg1)
	ret0, _ := ret[0].(db.ProductAggregationDetailsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductAggregationDetails indicates an expected call of ProductAggregationDetails
func (mr *MockProductMockRecorder) ProductAggregationDetails(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductAggregationDetails", reflect.TypeOf((*MockProduct)(nil).ProductAggregationDetails), arg0, arg1)
}

// ProductsNotAcquired mocks base method
func (m *MockProduct) ProductsNotAcquired(arg0 context.Context, arg1 string) ([]db.ProductsNotAcquiredRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsNotAcquired", arg0, arg1)
	ret0, _ := ret[0].([]db.ProductsNotAcquiredRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsNotAcquired indicates an expected call of ProductsNotAcquired
func (mr *MockProductMockRecorder) ProductsNotAcquired(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsNotAcquired", reflect.TypeOf((*MockProduct)(nil).ProductsNotAcquired), arg0, arg1)
}

// ProductsNotDeployed mocks base method
func (m *MockProduct) ProductsNotDeployed(arg0 context.Context, arg1 string) ([]db.ProductsNotDeployedRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsNotDeployed", arg0, arg1)
	ret0, _ := ret[0].([]db.ProductsNotDeployedRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsNotDeployed indicates an expected call of ProductsNotDeployed
func (mr *MockProductMockRecorder) ProductsNotDeployed(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsNotDeployed", reflect.TypeOf((*MockProduct)(nil).ProductsNotDeployed), arg0, arg1)
}

// ProductsPerMetric mocks base method
func (m *MockProduct) ProductsPerMetric(arg0 context.Context, arg1 []string) ([]db.ProductsPerMetricRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsPerMetric", arg0, arg1)
	ret0, _ := ret[0].([]db.ProductsPerMetricRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsPerMetric indicates an expected call of ProductsPerMetric
func (mr *MockProductMockRecorder) ProductsPerMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsPerMetric", reflect.TypeOf((*MockProduct)(nil).ProductsPerMetric), arg0, arg1)
}

// UpdateAggregation mocks base method
func (m *MockProduct) UpdateAggregation(arg0 context.Context, arg1 db.UpdateAggregationParams) (db.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAggregation", arg0, arg1)
	ret0, _ := ret[0].(db.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAggregation indicates an expected call of UpdateAggregation
func (mr *MockProductMockRecorder) UpdateAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAggregation", reflect.TypeOf((*MockProduct)(nil).UpdateAggregation), arg0, arg1)
}

// UpsertAcqRights mocks base method
func (m *MockProduct) UpsertAcqRights(arg0 context.Context, arg1 db.UpsertAcqRightsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertAcqRights", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertAcqRights indicates an expected call of UpsertAcqRights
func (mr *MockProductMockRecorder) UpsertAcqRights(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertAcqRights", reflect.TypeOf((*MockProduct)(nil).UpsertAcqRights), arg0, arg1)
}

// UpsertProduct mocks base method
func (m *MockProduct) UpsertProduct(arg0 context.Context, arg1 db.UpsertProductParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProduct", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProduct indicates an expected call of UpsertProduct
func (mr *MockProductMockRecorder) UpsertProduct(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProduct", reflect.TypeOf((*MockProduct)(nil).UpsertProduct), arg0, arg1)
}

// UpsertProductAggregation mocks base method
func (m *MockProduct) UpsertProductAggregation(arg0 context.Context, arg1 db.UpsertProductAggregationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProductAggregation", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProductAggregation indicates an expected call of UpsertProductAggregation
func (mr *MockProductMockRecorder) UpsertProductAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProductAggregation", reflect.TypeOf((*MockProduct)(nil).UpsertProductAggregation), arg0, arg1)
}

// UpsertProductApplications mocks base method
func (m *MockProduct) UpsertProductApplications(arg0 context.Context, arg1 db.UpsertProductApplicationsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProductApplications", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProductApplications indicates an expected call of UpsertProductApplications
func (mr *MockProductMockRecorder) UpsertProductApplications(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProductApplications", reflect.TypeOf((*MockProduct)(nil).UpsertProductApplications), arg0, arg1)
}

// UpsertProductEquipments mocks base method
func (m *MockProduct) UpsertProductEquipments(arg0 context.Context, arg1 db.UpsertProductEquipmentsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProductEquipments", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProductEquipments indicates an expected call of UpsertProductEquipments
func (mr *MockProductMockRecorder) UpsertProductEquipments(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProductEquipments", reflect.TypeOf((*MockProduct)(nil).UpsertProductEquipments), arg0, arg1)
}

// UpsertProductPartial mocks base method
func (m *MockProduct) UpsertProductPartial(arg0 context.Context, arg1 db.UpsertProductPartialParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProductPartial", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProductPartial indicates an expected call of UpsertProductPartial
func (mr *MockProductMockRecorder) UpsertProductPartial(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProductPartial", reflect.TypeOf((*MockProduct)(nil).UpsertProductPartial), arg0, arg1)
}

// UpsertProductTx mocks base method
func (m *MockProduct) UpsertProductTx(arg0 context.Context, arg1 *v1.UpsertProductRequest, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertProductTx", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertProductTx indicates an expected call of UpsertProductTx
func (mr *MockProductMockRecorder) UpsertProductTx(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertProductTx", reflect.TypeOf((*MockProduct)(nil).UpsertProductTx), arg0, arg1, arg2)
}
