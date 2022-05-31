// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/dps-service/pkg/api/v1 (interfaces: DpsServiceClient)

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

// DashboardQualityOverview mocks base method
func (m *MockDpsServiceClient) DashboardQualityOverview(arg0 context.Context, arg1 *v1.DashboardQualityOverviewRequest, arg2 ...grpc.CallOption) (*v1.DashboardQualityOverviewResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DashboardQualityOverview", varargs...)
	ret0, _ := ret[0].(*v1.DashboardQualityOverviewResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DashboardQualityOverview indicates an expected call of DashboardQualityOverview
func (mr *MockDpsServiceClientMockRecorder) DashboardQualityOverview(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DashboardQualityOverview", reflect.TypeOf((*MockDpsServiceClient)(nil).DashboardQualityOverview), varargs...)
}

// DataAnalysis mocks base method
func (m *MockDpsServiceClient) DataAnalysis(arg0 context.Context, arg1 *v1.DataAnalysisRequest, arg2 ...grpc.CallOption) (*v1.DataAnalysisResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DataAnalysis", varargs...)
	ret0, _ := ret[0].(*v1.DataAnalysisResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DataAnalysis indicates an expected call of DataAnalysis
func (mr *MockDpsServiceClientMockRecorder) DataAnalysis(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DataAnalysis", reflect.TypeOf((*MockDpsServiceClient)(nil).DataAnalysis), varargs...)
}

// DeleteInventory mocks base method
func (m *MockDpsServiceClient) DeleteInventory(arg0 context.Context, arg1 *v1.DeleteInventoryRequest, arg2 ...grpc.CallOption) (*v1.DeleteInventoryResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteInventory", varargs...)
	ret0, _ := ret[0].(*v1.DeleteInventoryResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteInventory indicates an expected call of DeleteInventory
func (mr *MockDpsServiceClientMockRecorder) DeleteInventory(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInventory", reflect.TypeOf((*MockDpsServiceClient)(nil).DeleteInventory), varargs...)
}

// DropUploadedFileData mocks base method
func (m *MockDpsServiceClient) DropUploadedFileData(arg0 context.Context, arg1 *v1.DropUploadedFileDataRequest, arg2 ...grpc.CallOption) (*v1.DropUploadedFileDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DropUploadedFileData", varargs...)
	ret0, _ := ret[0].(*v1.DropUploadedFileDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropUploadedFileData indicates an expected call of DropUploadedFileData
func (mr *MockDpsServiceClientMockRecorder) DropUploadedFileData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropUploadedFileData", reflect.TypeOf((*MockDpsServiceClient)(nil).DropUploadedFileData), varargs...)
}

// GetAnalysisFileInfo mocks base method
func (m *MockDpsServiceClient) GetAnalysisFileInfo(arg0 context.Context, arg1 *v1.GetAnalysisFileInfoRequest, arg2 ...grpc.CallOption) (*v1.GetAnalysisFileInfoResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetAnalysisFileInfo", varargs...)
	ret0, _ := ret[0].(*v1.GetAnalysisFileInfoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAnalysisFileInfo indicates an expected call of GetAnalysisFileInfo
func (mr *MockDpsServiceClientMockRecorder) GetAnalysisFileInfo(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAnalysisFileInfo", reflect.TypeOf((*MockDpsServiceClient)(nil).GetAnalysisFileInfo), varargs...)
}

// ListDeletionRecords mocks base method
func (m *MockDpsServiceClient) ListDeletionRecords(arg0 context.Context, arg1 *v1.ListDeletionRequest, arg2 ...grpc.CallOption) (*v1.ListDeletionResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListDeletionRecords", varargs...)
	ret0, _ := ret[0].(*v1.ListDeletionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListDeletionRecords indicates an expected call of ListDeletionRecords
func (mr *MockDpsServiceClientMockRecorder) ListDeletionRecords(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListDeletionRecords", reflect.TypeOf((*MockDpsServiceClient)(nil).ListDeletionRecords), varargs...)
}

// ListFailedRecord mocks base method
func (m *MockDpsServiceClient) ListFailedRecord(arg0 context.Context, arg1 *v1.ListFailedRequest, arg2 ...grpc.CallOption) (*v1.ListFailedResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListFailedRecord", varargs...)
	ret0, _ := ret[0].(*v1.ListFailedResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListFailedRecord indicates an expected call of ListFailedRecord
func (mr *MockDpsServiceClientMockRecorder) ListFailedRecord(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListFailedRecord", reflect.TypeOf((*MockDpsServiceClient)(nil).ListFailedRecord), varargs...)
}

// ListUploadData mocks base method
func (m *MockDpsServiceClient) ListUploadData(arg0 context.Context, arg1 *v1.ListUploadRequest, arg2 ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadData indicates an expected call of ListUploadData
func (mr *MockDpsServiceClientMockRecorder) ListUploadData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadData), varargs...)
}

// ListUploadGlobalData mocks base method
func (m *MockDpsServiceClient) ListUploadGlobalData(arg0 context.Context, arg1 *v1.ListUploadRequest, arg2 ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadGlobalData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadGlobalData indicates an expected call of ListUploadGlobalData
func (mr *MockDpsServiceClientMockRecorder) ListUploadGlobalData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadGlobalData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadGlobalData), varargs...)
}

// ListUploadMetaData mocks base method
func (m *MockDpsServiceClient) ListUploadMetaData(arg0 context.Context, arg1 *v1.ListUploadRequest, arg2 ...grpc.CallOption) (*v1.ListUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListUploadMetaData", varargs...)
	ret0, _ := ret[0].(*v1.ListUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadMetaData indicates an expected call of ListUploadMetaData
func (mr *MockDpsServiceClientMockRecorder) ListUploadMetaData(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadMetaData", reflect.TypeOf((*MockDpsServiceClient)(nil).ListUploadMetaData), varargs...)
}

// NotifyUpload mocks base method
func (m *MockDpsServiceClient) NotifyUpload(arg0 context.Context, arg1 *v1.NotifyUploadRequest, arg2 ...grpc.CallOption) (*v1.NotifyUploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "NotifyUpload", varargs...)
	ret0, _ := ret[0].(*v1.NotifyUploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NotifyUpload indicates an expected call of NotifyUpload
func (mr *MockDpsServiceClientMockRecorder) NotifyUpload(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NotifyUpload", reflect.TypeOf((*MockDpsServiceClient)(nil).NotifyUpload), varargs...)
}

// StoreCoreFactorReference mocks base method
func (m *MockDpsServiceClient) StoreCoreFactorReference(arg0 context.Context, arg1 *v1.StoreReferenceDataRequest, arg2 ...grpc.CallOption) (*v1.StoreReferenceDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "StoreCoreFactorReference", varargs...)
	ret0, _ := ret[0].(*v1.StoreReferenceDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StoreCoreFactorReference indicates an expected call of StoreCoreFactorReference
func (mr *MockDpsServiceClientMockRecorder) StoreCoreFactorReference(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StoreCoreFactorReference", reflect.TypeOf((*MockDpsServiceClient)(nil).StoreCoreFactorReference), varargs...)
}

// ViewCoreFactorLogs mocks base method
func (m *MockDpsServiceClient) ViewCoreFactorLogs(arg0 context.Context, arg1 *v1.ViewCoreFactorLogsRequest, arg2 ...grpc.CallOption) (*v1.ViewCoreFactorLogsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ViewCoreFactorLogs", varargs...)
	ret0, _ := ret[0].(*v1.ViewCoreFactorLogsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ViewCoreFactorLogs indicates an expected call of ViewCoreFactorLogs
func (mr *MockDpsServiceClientMockRecorder) ViewCoreFactorLogs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ViewCoreFactorLogs", reflect.TypeOf((*MockDpsServiceClient)(nil).ViewCoreFactorLogs), varargs...)
}

// ViewFactorReference mocks base method
func (m *MockDpsServiceClient) ViewFactorReference(arg0 context.Context, arg1 *v1.ViewReferenceDataRequest, arg2 ...grpc.CallOption) (*v1.ViewReferenceDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ViewFactorReference", varargs...)
	ret0, _ := ret[0].(*v1.ViewReferenceDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ViewFactorReference indicates an expected call of ViewFactorReference
func (mr *MockDpsServiceClientMockRecorder) ViewFactorReference(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ViewFactorReference", reflect.TypeOf((*MockDpsServiceClient)(nil).ViewFactorReference), varargs...)
}
