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
