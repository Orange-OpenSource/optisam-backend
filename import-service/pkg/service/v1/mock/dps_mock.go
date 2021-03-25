// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: ../../../../dps-service/pkg/api/v1/dps.pb.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	reflect "reflect"
)

// MockDpsServiceClient is a mock of DpsServiceClient interface
type MockDpsServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockDpsServiceClientMockRecorder
}

// MockDpsServiceClientMockRecorder is the mock recorder for MockDpsServiceClient
type MockDpsServiceClientMockRecorder struct {
	mock *MockDpsServiceClient
}

// NewMockDpsServiceClient creates a new mock instance
func NewMockDpsServiceClient(ctrl *gomock.Controller) *MockDpsServiceClient {
	mock := &MockDpsServiceClient{ctrl: ctrl}
	mock.recorder = &MockDpsServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDpsServiceClient) EXPECT() *MockDpsServiceClientMockRecorder {
	return m.recorder
}

// NotifyUpload mocks base method
func (m *MockDpsServiceClient) NotifyUpload(ctx context.Context, in *v1.NotifyUploadRequest, opts ...grpc.CallOption) (*v1.NotifyUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NotifyUpload", varargs...)
	ret0, _ := ret[0].(*v1.NotifyUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NotifyUpload indicates an expected call of NotifyUpload
func (mr *MockDpsServiceClientMockRecorder) NotifyUpload(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyUpload", reflect.TypeOf((*MockDpsServiceClient)(nil).NotifyUpload), varargs...)
}

// DashboardQualityOverview mocks base method
func (m *MockDpsServiceClient) DashboardQualityOverview(ctx context.Context, in *v1.DashboardQualityOverviewRequest, opts ...grpc.CallOption) (*v1.DashboardQualityOverviewResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DashboardQualityOverview", varargs...)
	ret0, _ := ret[0].(*v1.DashboardQualityOverviewResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DashboardQualityOverview indicates an expected call of DashboardQualityOverview
func (mr *MockDpsServiceClientMockRecorder) DashboardQualityOverview(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DashboardQualityOverview", reflect.TypeOf((*MockDpsServiceClient)(nil).DashboardQualityOverview), varargs...)
}

// DashboardDataFailureRate mocks base method
func (m *MockDpsServiceClient) DashboardDataFailureRate(ctx context.Context, in *v1.DataFailureRateRequest, opts ...grpc.CallOption) (*v1.DataFailureRateResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DashboardDataFailureRate", varargs...)
	ret0, _ := ret[0].(*v1.DataFailureRateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DashboardDataFailureRate indicates an expected call of DashboardDataFailureRate
func (mr *MockDpsServiceClientMockRecorder) DashboardDataFailureRate(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DashboardDataFailureRate", reflect.TypeOf((*MockDpsServiceClient)(nil).DashboardDataFailureRate), varargs...)
}

// ListUploadData mocks base method
func (m *MockDpsServiceClient) ListUploadData(ctx context.Context, in *v1.ListUploadRequest, opts ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadData indicates an expected call of ListUploadData
func (mr *MockDpsServiceClientMockRecorder) ListUploadData(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadData), varargs...)
}

// ListUploadMetaData mocks base method
func (m *MockDpsServiceClient) ListUploadMetaData(ctx context.Context, in *v1.ListUploadRequest, opts ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadMetaData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadMetaData indicates an expected call of ListUploadMetaData
func (mr *MockDpsServiceClientMockRecorder) ListUploadMetaData(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadMetaData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadMetaData), varargs...)
}

// ListUploadGlobalData mocks base method
func (m *MockDpsServiceClient) ListUploadGlobalData(ctx context.Context, in *v1.ListUploadRequest, opts ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadGlobalData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadGlobalData indicates an expected call of ListUploadGlobalData
func (mr *MockDpsServiceClientMockRecorder) ListUploadGlobalData(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadGlobalData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadGlobalData), varargs...)
}

// ListFailedRecord mocks base method
func (m *MockDpsServiceClient) ListFailedRecord(ctx context.Context, in *v1.ListFailedRequest, opts ...grpc.CallOption) (*v1.ListFailedResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListFailedRecord", varargs...)
	ret0, _ := ret[0].(*v1.ListFailedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFailedRecord indicates an expected call of ListFailedRecord
func (mr *MockDpsServiceClientMockRecorder) ListFailedRecord(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFailedRecord", reflect.TypeOf((*MockDpsServiceClient)(nil).ListFailedRecord), varargs...)
}

// ListFailureReasonsRatio mocks base method
func (m *MockDpsServiceClient) ListFailureReasonsRatio(ctx context.Context, in *v1.ListFailureReasonRequest, opts ...grpc.CallOption) (*v1.ListFailureReasonResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListFailureReasonsRatio", varargs...)
	ret0, _ := ret[0].(*v1.ListFailureReasonResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFailureReasonsRatio indicates an expected call of ListFailureReasonsRatio
func (mr *MockDpsServiceClientMockRecorder) ListFailureReasonsRatio(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFailureReasonsRatio", reflect.TypeOf((*MockDpsServiceClient)(nil).ListFailureReasonsRatio), varargs...)
}

// DeleteInventory mocks base method
func (m *MockDpsServiceClient) DeleteInventory(ctx context.Context, in *v1.DeleteInventoryRequest, opts ...grpc.CallOption) (*v1.DeleteInventoryResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteInventory", varargs...)
	ret0, _ := ret[0].(*v1.DeleteInventoryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteInventory indicates an expected call of DeleteInventory
func (mr *MockDpsServiceClientMockRecorder) DeleteInventory(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInventory", reflect.TypeOf((*MockDpsServiceClient)(nil).DeleteInventory), varargs...)
}

// MockDpsServiceServer is a mock of DpsServiceServer interface
type MockDpsServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockDpsServiceServerMockRecorder
}

// MockDpsServiceServerMockRecorder is the mock recorder for MockDpsServiceServer
type MockDpsServiceServerMockRecorder struct {
	mock *MockDpsServiceServer
}

// NewMockDpsServiceServer creates a new mock instance
func NewMockDpsServiceServer(ctrl *gomock.Controller) *MockDpsServiceServer {
	mock := &MockDpsServiceServer{ctrl: ctrl}
	mock.recorder = &MockDpsServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDpsServiceServer) EXPECT() *MockDpsServiceServerMockRecorder {
	return m.recorder
}

// NotifyUpload mocks base method
func (m *MockDpsServiceServer) NotifyUpload(arg0 context.Context, arg1 *v1.NotifyUploadRequest) (*v1.NotifyUploadResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NotifyUpload", arg0, arg1)
	ret0, _ := ret[0].(*v1.NotifyUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NotifyUpload indicates an expected call of NotifyUpload
func (mr *MockDpsServiceServerMockRecorder) NotifyUpload(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyUpload", reflect.TypeOf((*MockDpsServiceServer)(nil).NotifyUpload), arg0, arg1)
}

// DashboardQualityOverview mocks base method
func (m *MockDpsServiceServer) DashboardQualityOverview(arg0 context.Context, arg1 *v1.DashboardQualityOverviewRequest) (*v1.DashboardQualityOverviewResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DashboardQualityOverview", arg0, arg1)
	ret0, _ := ret[0].(*v1.DashboardQualityOverviewResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DashboardQualityOverview indicates an expected call of DashboardQualityOverview
func (mr *MockDpsServiceServerMockRecorder) DashboardQualityOverview(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DashboardQualityOverview", reflect.TypeOf((*MockDpsServiceServer)(nil).DashboardQualityOverview), arg0, arg1)
}

// DashboardDataFailureRate mocks base method
func (m *MockDpsServiceServer) DashboardDataFailureRate(arg0 context.Context, arg1 *v1.DataFailureRateRequest) (*v1.DataFailureRateResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DashboardDataFailureRate", arg0, arg1)
	ret0, _ := ret[0].(*v1.DataFailureRateResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DashboardDataFailureRate indicates an expected call of DashboardDataFailureRate
func (mr *MockDpsServiceServerMockRecorder) DashboardDataFailureRate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DashboardDataFailureRate", reflect.TypeOf((*MockDpsServiceServer)(nil).DashboardDataFailureRate), arg0, arg1)
}

// ListUploadData mocks base method
func (m *MockDpsServiceServer) ListUploadData(arg0 context.Context, arg1 *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadData", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadData indicates an expected call of ListUploadData
func (mr *MockDpsServiceServerMockRecorder) ListUploadData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadData", reflect.TypeOf((*MockDpsServiceServer)(nil).ListUploadData), arg0, arg1)
}

// ListUploadMetaData mocks base method
func (m *MockDpsServiceServer) ListUploadMetaData(arg0 context.Context, arg1 *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadMetaData", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadMetaData indicates an expected call of ListUploadMetaData
func (mr *MockDpsServiceServerMockRecorder) ListUploadMetaData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadMetaData", reflect.TypeOf((*MockDpsServiceServer)(nil).ListUploadMetaData), arg0, arg1)
}

// ListUploadGlobalData mocks base method
func (m *MockDpsServiceServer) ListUploadGlobalData(arg0 context.Context, arg1 *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadGlobalData", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadGlobalData indicates an expected call of ListUploadGlobalData
func (mr *MockDpsServiceServerMockRecorder) ListUploadGlobalData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadGlobalData", reflect.TypeOf((*MockDpsServiceServer)(nil).ListUploadGlobalData), arg0, arg1)
}

// ListFailedRecord mocks base method
func (m *MockDpsServiceServer) ListFailedRecord(arg0 context.Context, arg1 *v1.ListFailedRequest) (*v1.ListFailedResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFailedRecord", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListFailedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFailedRecord indicates an expected call of ListFailedRecord
func (mr *MockDpsServiceServerMockRecorder) ListFailedRecord(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFailedRecord", reflect.TypeOf((*MockDpsServiceServer)(nil).ListFailedRecord), arg0, arg1)
}

// ListFailureReasonsRatio mocks base method
func (m *MockDpsServiceServer) ListFailureReasonsRatio(arg0 context.Context, arg1 *v1.ListFailureReasonRequest) (*v1.ListFailureReasonResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListFailureReasonsRatio", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListFailureReasonResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFailureReasonsRatio indicates an expected call of ListFailureReasonsRatio
func (mr *MockDpsServiceServerMockRecorder) ListFailureReasonsRatio(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFailureReasonsRatio", reflect.TypeOf((*MockDpsServiceServer)(nil).ListFailureReasonsRatio), arg0, arg1)
}

// DeleteInventory mocks base method
func (m *MockDpsServiceServer) DeleteInventory(arg0 context.Context, arg1 *v1.DeleteInventoryRequest) (*v1.DeleteInventoryResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteInventory", arg0, arg1)
	ret0, _ := ret[0].(*v1.DeleteInventoryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteInventory indicates an expected call of DeleteInventory
func (mr *MockDpsServiceServerMockRecorder) DeleteInventory(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInventory", reflect.TypeOf((*MockDpsServiceServer)(nil).DeleteInventory), arg0, arg1)
}
