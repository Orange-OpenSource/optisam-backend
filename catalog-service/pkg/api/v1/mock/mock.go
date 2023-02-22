// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/catalog-service/pkg/api/v1 (interfaces: ProductCatalogClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	reflect "reflect"
)

// MockProductCatalogClient is a mock of ProductCatalogClient interface
type MockProductCatalogClient struct {
	ctrl     *gomock.Controller
	recorder *MockProductCatalogClientMockRecorder
}

// MockProductCatalogClientMockRecorder is the mock recorder for MockProductCatalogClient
type MockProductCatalogClientMockRecorder struct {
	mock *MockProductCatalogClient
}

// NewMockProductCatalogClient creates a new mock instance
func NewMockProductCatalogClient(ctrl *gomock.Controller) *MockProductCatalogClient {
	mock := &MockProductCatalogClient{ctrl: ctrl}
	mock.recorder = &MockProductCatalogClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockProductCatalogClient) EXPECT() *MockProductCatalogClientMockRecorder {
	return m.recorder
}

// BulkFileUpload mocks base method
func (m *MockProductCatalogClient) BulkFileUpload(arg0 context.Context, arg1 *v1.UploadRecords, arg2 ...grpc.CallOption) (*v1.UploadResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BulkFileUpload", varargs...)
	ret0, _ := ret[0].(*v1.UploadResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BulkFileUpload indicates an expected call of BulkFileUpload
func (mr *MockProductCatalogClientMockRecorder) BulkFileUpload(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkFileUpload", reflect.TypeOf((*MockProductCatalogClient)(nil).BulkFileUpload), varargs...)
}

// BulkFileUploadLogs mocks base method
func (m *MockProductCatalogClient) BulkFileUploadLogs(arg0 context.Context, arg1 *v1.UploadCatalogDataLogsRequest, arg2 ...grpc.CallOption) (*v1.UploadCatalogDataLogsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BulkFileUploadLogs", varargs...)
	ret0, _ := ret[0].(*v1.UploadCatalogDataLogsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BulkFileUploadLogs indicates an expected call of BulkFileUploadLogs
func (mr *MockProductCatalogClientMockRecorder) BulkFileUploadLogs(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkFileUploadLogs", reflect.TypeOf((*MockProductCatalogClient)(nil).BulkFileUploadLogs), varargs...)
}

// CreateEditor mocks base method
func (m *MockProductCatalogClient) CreateEditor(arg0 context.Context, arg1 *v1.CreateEditorRequest, arg2 ...grpc.CallOption) (*v1.Editor, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateEditor", varargs...)
	ret0, _ := ret[0].(*v1.Editor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEditor indicates an expected call of CreateEditor
func (mr *MockProductCatalogClientMockRecorder) CreateEditor(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEditor", reflect.TypeOf((*MockProductCatalogClient)(nil).CreateEditor), varargs...)
}

// DeleteEditor mocks base method
func (m *MockProductCatalogClient) DeleteEditor(arg0 context.Context, arg1 *v1.GetEditorRequest, arg2 ...grpc.CallOption) (*v1.DeleteResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteEditor", varargs...)
	ret0, _ := ret[0].(*v1.DeleteResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteEditor indicates an expected call of DeleteEditor
func (mr *MockProductCatalogClientMockRecorder) DeleteEditor(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEditor", reflect.TypeOf((*MockProductCatalogClient)(nil).DeleteEditor), varargs...)
}

// DeleteProduct mocks base method
func (m *MockProductCatalogClient) DeleteProduct(arg0 context.Context, arg1 *v1.GetProductRequest, arg2 ...grpc.CallOption) (*v1.DeleteResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteProduct", varargs...)
	ret0, _ := ret[0].(*v1.DeleteResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteProduct indicates an expected call of DeleteProduct
func (mr *MockProductCatalogClientMockRecorder) DeleteProduct(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProduct", reflect.TypeOf((*MockProductCatalogClient)(nil).DeleteProduct), varargs...)
}

// GetEditor mocks base method
func (m *MockProductCatalogClient) GetEditor(arg0 context.Context, arg1 *v1.GetEditorRequest, arg2 ...grpc.CallOption) (*v1.Editor, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEditor", varargs...)
	ret0, _ := ret[0].(*v1.Editor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEditor indicates an expected call of GetEditor
func (mr *MockProductCatalogClientMockRecorder) GetEditor(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEditor", reflect.TypeOf((*MockProductCatalogClient)(nil).GetEditor), varargs...)
}

// GetProduct mocks base method
func (m *MockProductCatalogClient) GetProduct(arg0 context.Context, arg1 *v1.GetProductRequest, arg2 ...grpc.CallOption) (*v1.Product, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetProduct", varargs...)
	ret0, _ := ret[0].(*v1.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProduct indicates an expected call of GetProduct
func (mr *MockProductCatalogClientMockRecorder) GetProduct(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProduct", reflect.TypeOf((*MockProductCatalogClient)(nil).GetProduct), varargs...)
}

// InsertProduct mocks base method
func (m *MockProductCatalogClient) InsertProduct(arg0 context.Context, arg1 *v1.Product, arg2 ...grpc.CallOption) (*v1.Product, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "InsertProduct", varargs...)
	ret0, _ := ret[0].(*v1.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertProduct indicates an expected call of InsertProduct
func (mr *MockProductCatalogClientMockRecorder) InsertProduct(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertProduct", reflect.TypeOf((*MockProductCatalogClient)(nil).InsertProduct), varargs...)
}

// UpdateEditor mocks base method
func (m *MockProductCatalogClient) UpdateEditor(arg0 context.Context, arg1 *v1.Editor, arg2 ...grpc.CallOption) (*v1.Editor, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateEditor", varargs...)
	ret0, _ := ret[0].(*v1.Editor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateEditor indicates an expected call of UpdateEditor
func (mr *MockProductCatalogClientMockRecorder) UpdateEditor(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateEditor", reflect.TypeOf((*MockProductCatalogClient)(nil).UpdateEditor), varargs...)
}

// UpdateProduct mocks base method
func (m *MockProductCatalogClient) UpdateProduct(arg0 context.Context, arg1 *v1.Product, arg2 ...grpc.CallOption) (*v1.Product, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateProduct", varargs...)
	ret0, _ := ret[0].(*v1.Product)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateProduct indicates an expected call of UpdateProduct
func (mr *MockProductCatalogClientMockRecorder) UpdateProduct(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProduct", reflect.TypeOf((*MockProductCatalogClient)(nil).UpdateProduct), varargs...)
}
