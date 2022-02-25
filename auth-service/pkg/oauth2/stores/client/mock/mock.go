// Code generated by MockGen. DO NOT EDIT.
// Source: gopkg.in/oauth2.v3 (interfaces: ClientStore)

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	oauth2 "gopkg.in/oauth2.v3"
	reflect "reflect"
)

// MockClientStore is a mock of ClientStore interface
type MockClientStore struct {
	ctrl     *gomock.Controller
	recorder *MockClientStoreMockRecorder
}

// MockClientStoreMockRecorder is the mock recorder for MockClientStore
type MockClientStoreMockRecorder struct {
	mock *MockClientStore
}

// NewMockClientStore creates a new mock instance
func NewMockClientStore(ctrl *gomock.Controller) *MockClientStore {
	mock := &MockClientStore{ctrl: ctrl}
	mock.recorder = &MockClientStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClientStore) EXPECT() *MockClientStoreMockRecorder {
	return m.recorder
}

// GetByID mocks base method
func (m *MockClientStore) GetByID(arg0 string) (oauth2.ClientInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", arg0)
	ret0, _ := ret[0].(oauth2.ClientInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID
func (mr *MockClientStoreMockRecorder) GetByID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockClientStore)(nil).GetByID), arg0)
}
