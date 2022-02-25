// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/application-service/pkg/repository/v1 (interfaces: Application)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/application-service/pkg/api/v1"
	db "optisam-backend/application-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockApplication is a mock of Application interface
type MockApplication struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationMockRecorder
}

// MockApplicationMockRecorder is the mock recorder for MockApplication
type MockApplicationMockRecorder struct {
	mock *MockApplication
}

// NewMockApplication creates a new mock instance
func NewMockApplication(ctrl *gomock.Controller) *MockApplication {
	mock := &MockApplication{ctrl: ctrl}
	mock.recorder = &MockApplicationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockApplication) EXPECT() *MockApplicationMockRecorder {
	return m.recorder
}

// GetApplicationInstance mocks base method
func (m *MockApplication) GetApplicationInstance(arg0 context.Context, arg1 string) (db.ApplicationsInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationInstance", arg0, arg1)
	ret0, _ := ret[0].(db.ApplicationsInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationInstance indicates an expected call of GetApplicationInstance
func (mr *MockApplicationMockRecorder) GetApplicationInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationInstance", reflect.TypeOf((*MockApplication)(nil).GetApplicationInstance), arg0, arg1)
}

// GetApplicationsView mocks base method
func (m *MockApplication) GetApplicationsView(arg0 context.Context, arg1 db.GetApplicationsViewParams) ([]db.GetApplicationsViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationsView", arg0, arg1)
	ret0, _ := ret[0].([]db.GetApplicationsViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationsView indicates an expected call of GetApplicationsView
func (mr *MockApplicationMockRecorder) GetApplicationsView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationsView", reflect.TypeOf((*MockApplication)(nil).GetApplicationsView), arg0, arg1)
}

// GetInstancesView mocks base method
func (m *MockApplication) GetInstancesView(arg0 context.Context, arg1 db.GetInstancesViewParams) ([]db.GetInstancesViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInstancesView", arg0, arg1)
	ret0, _ := ret[0].([]db.GetInstancesViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInstancesView indicates an expected call of GetInstancesView
func (mr *MockApplicationMockRecorder) GetInstancesView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInstancesView", reflect.TypeOf((*MockApplication)(nil).GetInstancesView), arg0, arg1)
}

// UpsertApplication mocks base method
func (m *MockApplication) UpsertApplication(arg0 context.Context, arg1 db.UpsertApplicationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplication", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplication indicates an expected call of UpsertApplication
func (mr *MockApplicationMockRecorder) UpsertApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplication", reflect.TypeOf((*MockApplication)(nil).UpsertApplication), arg0, arg1)
}

// UpsertApplicationInstance mocks base method
func (m *MockApplication) UpsertApplicationInstance(arg0 context.Context, arg1 db.UpsertApplicationInstanceParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplicationInstance", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplicationInstance indicates an expected call of UpsertApplicationInstance
func (mr *MockApplicationMockRecorder) UpsertApplicationInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationInstance", reflect.TypeOf((*MockApplication)(nil).UpsertApplicationInstance), arg0, arg1)
}

// UpsertInstanceTX mocks base method
func (m *MockApplication) UpsertInstanceTX(arg0 context.Context, arg1 *v1.UpsertInstanceRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertInstanceTX", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertInstanceTX indicates an expected call of UpsertInstanceTX
func (mr *MockApplicationMockRecorder) UpsertInstanceTX(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertInstanceTX", reflect.TypeOf((*MockApplication)(nil).UpsertInstanceTX), arg0, arg1)
}
