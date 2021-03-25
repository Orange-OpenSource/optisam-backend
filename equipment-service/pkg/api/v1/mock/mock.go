// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/equipment-service/pkg/api/v1 (interfaces: EquipmentServiceClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	reflect "reflect"
)

// MockEquipmentServiceClient is a mock of EquipmentServiceClient interface
type MockEquipmentServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockEquipmentServiceClientMockRecorder
}

// MockEquipmentServiceClientMockRecorder is the mock recorder for MockEquipmentServiceClient
type MockEquipmentServiceClientMockRecorder struct {
	mock *MockEquipmentServiceClient
}

// NewMockEquipmentServiceClient creates a new mock instance
func NewMockEquipmentServiceClient(ctrl *gomock.Controller) *MockEquipmentServiceClient {
	mock := &MockEquipmentServiceClient{ctrl: ctrl}
	mock.recorder = &MockEquipmentServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEquipmentServiceClient) EXPECT() *MockEquipmentServiceClientMockRecorder {
	return m.recorder
}

// CreateEquipmentType mocks base method
func (m *MockEquipmentServiceClient) CreateEquipmentType(arg0 context.Context, arg1 *v1.EquipmentType, arg2 ...grpc.CallOption) (*v1.EquipmentType, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateEquipmentType", varargs...)
	ret0, _ := ret[0].(*v1.EquipmentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEquipmentType indicates an expected call of CreateEquipmentType
func (mr *MockEquipmentServiceClientMockRecorder) CreateEquipmentType(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEquipmentType", reflect.TypeOf((*MockEquipmentServiceClient)(nil).CreateEquipmentType), varargs...)
}

// DeleteEquipmentType mocks base method
func (m *MockEquipmentServiceClient) DeleteEquipmentType(arg0 context.Context, arg1 *v1.DeleteEquipmentTypeRequest, arg2 ...grpc.CallOption) (*v1.DeleteEquipmentTypeResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteEquipmentType", varargs...)
	ret0, _ := ret[0].(*v1.DeleteEquipmentTypeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteEquipmentType indicates an expected call of DeleteEquipmentType
func (mr *MockEquipmentServiceClientMockRecorder) DeleteEquipmentType(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEquipmentType", reflect.TypeOf((*MockEquipmentServiceClient)(nil).DeleteEquipmentType), varargs...)
}

// DropEquipmentData mocks base method
func (m *MockEquipmentServiceClient) DropEquipmentData(arg0 context.Context, arg1 *v1.DropEquipmentDataRequest, arg2 ...grpc.CallOption) (*v1.DropEquipmentDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DropEquipmentData", varargs...)
	ret0, _ := ret[0].(*v1.DropEquipmentDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropEquipmentData indicates an expected call of DropEquipmentData
func (mr *MockEquipmentServiceClientMockRecorder) DropEquipmentData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropEquipmentData", reflect.TypeOf((*MockEquipmentServiceClient)(nil).DropEquipmentData), varargs...)
}

// EquipmentsPerEquipmentType mocks base method
func (m *MockEquipmentServiceClient) EquipmentsPerEquipmentType(arg0 context.Context, arg1 *v1.EquipmentsPerEquipmentTypeRequest, arg2 ...grpc.CallOption) (*v1.EquipmentsPerEquipmentTypeResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EquipmentsPerEquipmentType", varargs...)
	ret0, _ := ret[0].(*v1.EquipmentsPerEquipmentTypeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentsPerEquipmentType indicates an expected call of EquipmentsPerEquipmentType
func (mr *MockEquipmentServiceClientMockRecorder) EquipmentsPerEquipmentType(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentsPerEquipmentType", reflect.TypeOf((*MockEquipmentServiceClient)(nil).EquipmentsPerEquipmentType), varargs...)
}

// EquipmentsTypes mocks base method
func (m *MockEquipmentServiceClient) EquipmentsTypes(arg0 context.Context, arg1 *v1.EquipmentTypesRequest, arg2 ...grpc.CallOption) (*v1.EquipmentTypesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EquipmentsTypes", varargs...)
	ret0, _ := ret[0].(*v1.EquipmentTypesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentsTypes indicates an expected call of EquipmentsTypes
func (mr *MockEquipmentServiceClientMockRecorder) EquipmentsTypes(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentsTypes", reflect.TypeOf((*MockEquipmentServiceClient)(nil).EquipmentsTypes), varargs...)
}

// GetEquipment mocks base method
func (m *MockEquipmentServiceClient) GetEquipment(arg0 context.Context, arg1 *v1.GetEquipmentRequest, arg2 ...grpc.CallOption) (*v1.GetEquipmentResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEquipment", varargs...)
	ret0, _ := ret[0].(*v1.GetEquipmentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipment indicates an expected call of GetEquipment
func (mr *MockEquipmentServiceClientMockRecorder) GetEquipment(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipment", reflect.TypeOf((*MockEquipmentServiceClient)(nil).GetEquipment), varargs...)
}

// GetEquipmentMetadata mocks base method
func (m *MockEquipmentServiceClient) GetEquipmentMetadata(arg0 context.Context, arg1 *v1.EquipmentMetadataRequest, arg2 ...grpc.CallOption) (*v1.EquipmentMetadata, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEquipmentMetadata", varargs...)
	ret0, _ := ret[0].(*v1.EquipmentMetadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipmentMetadata indicates an expected call of GetEquipmentMetadata
func (mr *MockEquipmentServiceClientMockRecorder) GetEquipmentMetadata(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipmentMetadata", reflect.TypeOf((*MockEquipmentServiceClient)(nil).GetEquipmentMetadata), varargs...)
}

// ListEquipmentChildren mocks base method
func (m *MockEquipmentServiceClient) ListEquipmentChildren(arg0 context.Context, arg1 *v1.ListEquipmentChildrenRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipmentChildren", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipmentChildren indicates an expected call of ListEquipmentChildren
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipmentChildren(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipmentChildren", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipmentChildren), varargs...)
}

// ListEquipmentParents mocks base method
func (m *MockEquipmentServiceClient) ListEquipmentParents(arg0 context.Context, arg1 *v1.ListEquipmentParentsRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipmentParents", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipmentParents indicates an expected call of ListEquipmentParents
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipmentParents(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipmentParents", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipmentParents), varargs...)
}

// ListEquipments mocks base method
func (m *MockEquipmentServiceClient) ListEquipments(arg0 context.Context, arg1 *v1.ListEquipmentsRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipments", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipments indicates an expected call of ListEquipments
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipments(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipments", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipments), varargs...)
}

// ListEquipmentsForProduct mocks base method
func (m *MockEquipmentServiceClient) ListEquipmentsForProduct(arg0 context.Context, arg1 *v1.ListEquipmentsForProductRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipmentsForProduct", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipmentsForProduct indicates an expected call of ListEquipmentsForProduct
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipmentsForProduct(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipmentsForProduct", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipmentsForProduct), varargs...)
}

// ListEquipmentsForProductAggregation mocks base method
func (m *MockEquipmentServiceClient) ListEquipmentsForProductAggregation(arg0 context.Context, arg1 *v1.ListEquipmentsForProductAggregationRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipmentsForProductAggregation", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipmentsForProductAggregation indicates an expected call of ListEquipmentsForProductAggregation
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipmentsForProductAggregation(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipmentsForProductAggregation", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipmentsForProductAggregation), varargs...)
}

// ListEquipmentsMetadata mocks base method
func (m *MockEquipmentServiceClient) ListEquipmentsMetadata(arg0 context.Context, arg1 *v1.ListEquipmentMetadataRequest, arg2 ...grpc.CallOption) (*v1.ListEquipmentMetadataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListEquipmentsMetadata", varargs...)
	ret0, _ := ret[0].(*v1.ListEquipmentMetadataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipmentsMetadata indicates an expected call of ListEquipmentsMetadata
func (mr *MockEquipmentServiceClientMockRecorder) ListEquipmentsMetadata(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipmentsMetadata", reflect.TypeOf((*MockEquipmentServiceClient)(nil).ListEquipmentsMetadata), varargs...)
}

// UpdateEquipmentType mocks base method
func (m *MockEquipmentServiceClient) UpdateEquipmentType(arg0 context.Context, arg1 *v1.UpdateEquipmentTypeRequest, arg2 ...grpc.CallOption) (*v1.EquipmentType, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateEquipmentType", varargs...)
	ret0, _ := ret[0].(*v1.EquipmentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEquipmentType indicates an expected call of UpdateEquipmentType
func (mr *MockEquipmentServiceClientMockRecorder) UpdateEquipmentType(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEquipmentType", reflect.TypeOf((*MockEquipmentServiceClient)(nil).UpdateEquipmentType), varargs...)
}

// UpsertEquipment mocks base method
func (m *MockEquipmentServiceClient) UpsertEquipment(arg0 context.Context, arg1 *v1.UpsertEquipmentRequest, arg2 ...grpc.CallOption) (*v1.UpsertEquipmentResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertEquipment", varargs...)
	ret0, _ := ret[0].(*v1.UpsertEquipmentResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertEquipment indicates an expected call of UpsertEquipment
func (mr *MockEquipmentServiceClientMockRecorder) UpsertEquipment(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertEquipment", reflect.TypeOf((*MockEquipmentServiceClient)(nil).UpsertEquipment), varargs...)
}

// UpsertMetadata mocks base method
func (m *MockEquipmentServiceClient) UpsertMetadata(arg0 context.Context, arg1 *v1.UpsertMetadataRequest, arg2 ...grpc.CallOption) (*v1.UpsertMetadataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertMetadata", varargs...)
	ret0, _ := ret[0].(*v1.UpsertMetadataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertMetadata indicates an expected call of UpsertMetadata
func (mr *MockEquipmentServiceClientMockRecorder) UpsertMetadata(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertMetadata", reflect.TypeOf((*MockEquipmentServiceClient)(nil).UpsertMetadata), varargs...)
}