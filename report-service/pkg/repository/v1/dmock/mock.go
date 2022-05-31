// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/report-service/pkg/repository/v1 (interfaces: DgraphReport)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	json "encoding/json"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/report-service/pkg/repository/v1"
	reflect "reflect"
)

// MockDgraphReport is a mock of DgraphReport interface
type MockDgraphReport struct {
	ctrl     *gomock.Controller
	recorder *MockDgraphReportMockRecorder
}

// MockDgraphReportMockRecorder is the mock recorder for MockDgraphReport
type MockDgraphReportMockRecorder struct {
	mock *MockDgraphReport
}

// NewMockDgraphReport creates a new mock instance
func NewMockDgraphReport(ctrl *gomock.Controller) *MockDgraphReport {
	mock := &MockDgraphReport{ctrl: ctrl}
	mock.recorder = &MockDgraphReportMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDgraphReport) EXPECT() *MockDgraphReportMockRecorder {
	return m.recorder
}

// EquipmentAttributes mocks base method
func (m *MockDgraphReport) EquipmentAttributes(arg0 context.Context, arg1, arg2 string, arg3 []*v1.EquipmentAttributes, arg4 string) (json.RawMessage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentAttributes", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(json.RawMessage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentAttributes indicates an expected call of EquipmentAttributes
func (mr *MockDgraphReportMockRecorder) EquipmentAttributes(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentAttributes", reflect.TypeOf((*MockDgraphReport)(nil).EquipmentAttributes), arg0, arg1, arg2, arg3, arg4)
}

// EquipmentParents mocks base method
func (m *MockDgraphReport) EquipmentParents(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1.Equipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentParents", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1.Equipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentParents indicates an expected call of EquipmentParents
func (mr *MockDgraphReportMockRecorder) EquipmentParents(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentParents", reflect.TypeOf((*MockDgraphReport)(nil).EquipmentParents), arg0, arg1, arg2, arg3)
}

// EquipmentTypeAttrs mocks base method
func (m *MockDgraphReport) EquipmentTypeAttrs(arg0 context.Context, arg1, arg2 string) ([]*v1.EquipmentAttributes, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentTypeAttrs", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*v1.EquipmentAttributes)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentTypeAttrs indicates an expected call of EquipmentTypeAttrs
func (mr *MockDgraphReportMockRecorder) EquipmentTypeAttrs(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentTypeAttrs", reflect.TypeOf((*MockDgraphReport)(nil).EquipmentTypeAttrs), arg0, arg1, arg2)
}

// EquipmentTypeParents mocks base method
func (m *MockDgraphReport) EquipmentTypeParents(arg0 context.Context, arg1, arg2 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentTypeParents", arg0, arg1, arg2)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentTypeParents indicates an expected call of EquipmentTypeParents
func (mr *MockDgraphReportMockRecorder) EquipmentTypeParents(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentTypeParents", reflect.TypeOf((*MockDgraphReport)(nil).EquipmentTypeParents), arg0, arg1, arg2)
}

// ProductEquipments mocks base method
func (m *MockDgraphReport) ProductEquipments(arg0 context.Context, arg1, arg2, arg3 string) ([]*v1.ProductEquipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductEquipments", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]*v1.ProductEquipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductEquipments indicates an expected call of ProductEquipments
func (mr *MockDgraphReportMockRecorder) ProductEquipments(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductEquipments", reflect.TypeOf((*MockDgraphReport)(nil).ProductEquipments), arg0, arg1, arg2, arg3)
}
