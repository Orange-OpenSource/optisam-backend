// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/simulation-service/pkg/repository/v1 (interfaces: Repository)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/simulation-service/pkg/repository/v1"
	db "optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockRepository is a mock of Repository interface
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// CreateConfig mocks base method
func (m *MockRepository) CreateConfig(arg0 context.Context, arg1 *v1.MasterData, arg2 []*v1.ConfigData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateConfig", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateConfig indicates an expected call of CreateConfig
func (mr *MockRepositoryMockRecorder) CreateConfig(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateConfig", reflect.TypeOf((*MockRepository)(nil).CreateConfig), arg0, arg1, arg2)
}

// DeleteConfig mocks base method
func (m *MockRepository) DeleteConfig(arg0 context.Context, arg1 db.DeleteConfigParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteConfig", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteConfig indicates an expected call of DeleteConfig
func (mr *MockRepositoryMockRecorder) DeleteConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteConfig", reflect.TypeOf((*MockRepository)(nil).DeleteConfig), arg0, arg1)
}

// DeleteConfigData mocks base method
func (m *MockRepository) DeleteConfigData(arg0 context.Context, arg1 int32) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteConfigData", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteConfigData indicates an expected call of DeleteConfigData
func (mr *MockRepositoryMockRecorder) DeleteConfigData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteConfigData", reflect.TypeOf((*MockRepository)(nil).DeleteConfigData), arg0, arg1)
}

// GetConfig mocks base method
func (m *MockRepository) GetConfig(arg0 context.Context, arg1 db.GetConfigParams) (db.ConfigMaster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig", arg0, arg1)
	ret0, _ := ret[0].(db.ConfigMaster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetConfig indicates an expected call of GetConfig
func (mr *MockRepositoryMockRecorder) GetConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockRepository)(nil).GetConfig), arg0, arg1)
}

// GetDataByMetadataID mocks base method
func (m *MockRepository) GetDataByMetadataID(arg0 context.Context, arg1 int32) ([]db.GetDataByMetadataIDRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataByMetadataID", arg0, arg1)
	ret0, _ := ret[0].([]db.GetDataByMetadataIDRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataByMetadataID indicates an expected call of GetDataByMetadataID
func (mr *MockRepositoryMockRecorder) GetDataByMetadataID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataByMetadataID", reflect.TypeOf((*MockRepository)(nil).GetDataByMetadataID), arg0, arg1)
}

// GetMetadatabyConfigID mocks base method
func (m *MockRepository) GetMetadatabyConfigID(arg0 context.Context, arg1 int32) ([]db.GetMetadatabyConfigIDRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetadatabyConfigID", arg0, arg1)
	ret0, _ := ret[0].([]db.GetMetadatabyConfigIDRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetadatabyConfigID indicates an expected call of GetMetadatabyConfigID
func (mr *MockRepositoryMockRecorder) GetMetadatabyConfigID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetadatabyConfigID", reflect.TypeOf((*MockRepository)(nil).GetMetadatabyConfigID), arg0, arg1)
}

// ListConfig mocks base method
func (m *MockRepository) ListConfig(arg0 context.Context, arg1 db.ListConfigParams) ([]db.ConfigMaster, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListConfig", arg0, arg1)
	ret0, _ := ret[0].([]db.ConfigMaster)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListConfig indicates an expected call of ListConfig
func (mr *MockRepositoryMockRecorder) ListConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListConfig", reflect.TypeOf((*MockRepository)(nil).ListConfig), arg0, arg1)
}

// UpdateConfig mocks base method
func (m *MockRepository) UpdateConfig(arg0 context.Context, arg1 int32, arg2 string, arg3 []int32, arg4 []*v1.ConfigData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateConfig", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateConfig indicates an expected call of UpdateConfig
func (mr *MockRepositoryMockRecorder) UpdateConfig(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateConfig", reflect.TypeOf((*MockRepository)(nil).UpdateConfig), arg0, arg1, arg2, arg3, arg4)
}
