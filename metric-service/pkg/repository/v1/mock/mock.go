// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/metric-service/pkg/repository/v1 (interfaces: Metric)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	reflect "reflect"
)

// MockMetric is a mock of Metric interface
type MockMetric struct {
	ctrl     *gomock.Controller
	recorder *MockMetricMockRecorder
}

// MockMetricMockRecorder is the mock recorder for MockMetric
type MockMetricMockRecorder struct {
	mock *MockMetric
}

// NewMockMetric creates a new mock instance
func NewMockMetric(ctrl *gomock.Controller) *MockMetric {
	mock := &MockMetric{ctrl: ctrl}
	mock.recorder = &MockMetricMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMetric) EXPECT() *MockMetricMockRecorder {
	return m.recorder
}

// CreateMetricACS mocks base method
func (m *MockMetric) CreateMetricACS(arg0 context.Context, arg1 *v1.MetricACS, arg2 *v1.Attribute, arg3 []string) (*v1.MetricACS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricACS", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1.MetricACS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricACS indicates an expected call of CreateMetricACS
func (mr *MockMetricMockRecorder) CreateMetricACS(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricACS", reflect.TypeOf((*MockMetric)(nil).CreateMetricACS), arg0, arg1, arg2, arg3)
}

// CreateMetricIPS mocks base method
func (m *MockMetric) CreateMetricIPS(arg0 context.Context, arg1 *v1.MetricIPS, arg2 []string) (*v1.MetricIPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricIPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricIPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricIPS indicates an expected call of CreateMetricIPS
func (mr *MockMetricMockRecorder) CreateMetricIPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricIPS", reflect.TypeOf((*MockMetric)(nil).CreateMetricIPS), arg0, arg1, arg2)
}

// CreateMetricInstanceNumberStandard mocks base method
func (m *MockMetric) CreateMetricInstanceNumberStandard(arg0 context.Context, arg1 *v1.MetricINM, arg2 []string) (*v1.MetricINM, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricInstanceNumberStandard", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricINM)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricInstanceNumberStandard indicates an expected call of CreateMetricInstanceNumberStandard
func (mr *MockMetricMockRecorder) CreateMetricInstanceNumberStandard(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricInstanceNumberStandard", reflect.TypeOf((*MockMetric)(nil).CreateMetricInstanceNumberStandard), arg0, arg1, arg2)
}

// CreateMetricOPS mocks base method
func (m *MockMetric) CreateMetricOPS(arg0 context.Context, arg1 *v1.MetricOPS, arg2 []string) (*v1.MetricOPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricOPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricOPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricOPS indicates an expected call of CreateMetricOPS
func (mr *MockMetricMockRecorder) CreateMetricOPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricOPS", reflect.TypeOf((*MockMetric)(nil).CreateMetricOPS), arg0, arg1, arg2)
}

// CreateMetricOracleNUPStandard mocks base method
func (m *MockMetric) CreateMetricOracleNUPStandard(arg0 context.Context, arg1 *v1.MetricNUPOracle, arg2 []string) (*v1.MetricNUPOracle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricOracleNUPStandard", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricNUPOracle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricOracleNUPStandard indicates an expected call of CreateMetricOracleNUPStandard
func (mr *MockMetricMockRecorder) CreateMetricOracleNUPStandard(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricOracleNUPStandard", reflect.TypeOf((*MockMetric)(nil).CreateMetricOracleNUPStandard), arg0, arg1, arg2)
}

// CreateMetricSPS mocks base method
func (m *MockMetric) CreateMetricSPS(arg0 context.Context, arg1 *v1.MetricSPS, arg2 []string) (*v1.MetricSPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMetricSPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricSPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMetricSPS indicates an expected call of CreateMetricSPS
func (mr *MockMetricMockRecorder) CreateMetricSPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMetricSPS", reflect.TypeOf((*MockMetric)(nil).CreateMetricSPS), arg0, arg1, arg2)
}

// EquipmentTypes mocks base method
func (m *MockMetric) EquipmentTypes(arg0 context.Context, arg1 []string) ([]*v1.EquipmentType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentTypes", arg0, arg1)
	ret0, _ := ret[0].([]*v1.EquipmentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentTypes indicates an expected call of EquipmentTypes
func (mr *MockMetricMockRecorder) EquipmentTypes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentTypes", reflect.TypeOf((*MockMetric)(nil).EquipmentTypes), arg0, arg1)
}

// GetMetricConfigACS mocks base method
func (m *MockMetric) GetMetricConfigACS(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricACS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigACS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricACS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigACS indicates an expected call of GetMetricConfigACS
func (mr *MockMetricMockRecorder) GetMetricConfigACS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigACS", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigACS), arg0, arg1, arg2)
}

// GetMetricConfigINM mocks base method
func (m *MockMetric) GetMetricConfigINM(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricINMConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigINM", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricINMConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigINM indicates an expected call of GetMetricConfigINM
func (mr *MockMetricMockRecorder) GetMetricConfigINM(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigINM", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigINM), arg0, arg1, arg2)
}

// GetMetricConfigIPS mocks base method
func (m *MockMetric) GetMetricConfigIPS(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricIPSConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigIPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricIPSConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigIPS indicates an expected call of GetMetricConfigIPS
func (mr *MockMetricMockRecorder) GetMetricConfigIPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigIPS", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigIPS), arg0, arg1, arg2)
}

// GetMetricConfigNUP mocks base method
func (m *MockMetric) GetMetricConfigNUP(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricNUPConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigNUP", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricNUPConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigNUP indicates an expected call of GetMetricConfigNUP
func (mr *MockMetricMockRecorder) GetMetricConfigNUP(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigNUP", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigNUP), arg0, arg1, arg2)
}

// GetMetricConfigOPS mocks base method
func (m *MockMetric) GetMetricConfigOPS(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricOPSConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigOPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricOPSConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigOPS indicates an expected call of GetMetricConfigOPS
func (mr *MockMetricMockRecorder) GetMetricConfigOPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigOPS", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigOPS), arg0, arg1, arg2)
}

// GetMetricConfigSPS mocks base method
func (m *MockMetric) GetMetricConfigSPS(arg0 context.Context, arg1 string, arg2 []string) (*v1.MetricSPSConfig, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricConfigSPS", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.MetricSPSConfig)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetricConfigSPS indicates an expected call of GetMetricConfigSPS
func (mr *MockMetricMockRecorder) GetMetricConfigSPS(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricConfigSPS", reflect.TypeOf((*MockMetric)(nil).GetMetricConfigSPS), arg0, arg1, arg2)
}

// ListMetricACS mocks base method
func (m *MockMetric) ListMetricACS(arg0 context.Context, arg1 []string) ([]*v1.MetricACS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricACS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricACS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricACS indicates an expected call of ListMetricACS
func (mr *MockMetricMockRecorder) ListMetricACS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricACS", reflect.TypeOf((*MockMetric)(nil).ListMetricACS), arg0, arg1)
}

// ListMetricIPS mocks base method
func (m *MockMetric) ListMetricIPS(arg0 context.Context, arg1 []string) ([]*v1.MetricIPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricIPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricIPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricIPS indicates an expected call of ListMetricIPS
func (mr *MockMetricMockRecorder) ListMetricIPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricIPS", reflect.TypeOf((*MockMetric)(nil).ListMetricIPS), arg0, arg1)
}

// ListMetricNUP mocks base method
func (m *MockMetric) ListMetricNUP(arg0 context.Context, arg1 []string) ([]*v1.MetricNUPOracle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricNUP", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricNUPOracle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricNUP indicates an expected call of ListMetricNUP
func (mr *MockMetricMockRecorder) ListMetricNUP(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricNUP", reflect.TypeOf((*MockMetric)(nil).ListMetricNUP), arg0, arg1)
}

// ListMetricOPS mocks base method
func (m *MockMetric) ListMetricOPS(arg0 context.Context, arg1 []string) ([]*v1.MetricOPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricOPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricOPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricOPS indicates an expected call of ListMetricOPS
func (mr *MockMetricMockRecorder) ListMetricOPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricOPS", reflect.TypeOf((*MockMetric)(nil).ListMetricOPS), arg0, arg1)
}

// ListMetricSPS mocks base method
func (m *MockMetric) ListMetricSPS(arg0 context.Context, arg1 []string) ([]*v1.MetricSPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricSPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricSPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricSPS indicates an expected call of ListMetricSPS
func (mr *MockMetricMockRecorder) ListMetricSPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricSPS", reflect.TypeOf((*MockMetric)(nil).ListMetricSPS), arg0, arg1)
}

// ListMetricTypeInfo mocks base method
func (m *MockMetric) ListMetricTypeInfo(arg0 context.Context, arg1 []string) ([]*v1.MetricTypeInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricTypeInfo", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricTypeInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricTypeInfo indicates an expected call of ListMetricTypeInfo
func (mr *MockMetricMockRecorder) ListMetricTypeInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricTypeInfo", reflect.TypeOf((*MockMetric)(nil).ListMetricTypeInfo), arg0, arg1)
}

// ListMetrices mocks base method
func (m *MockMetric) ListMetrices(arg0 context.Context, arg1 []string) ([]*v1.MetricInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetrices", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetrices indicates an expected call of ListMetrices
func (mr *MockMetricMockRecorder) ListMetrices(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetrices", reflect.TypeOf((*MockMetric)(nil).ListMetrices), arg0, arg1)
}
