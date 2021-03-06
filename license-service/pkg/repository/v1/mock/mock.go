// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/license-service/pkg/repository/v1 (interfaces: License)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	reflect "reflect"
)

// MockLicense is a mock of License interface
type MockLicense struct {
	ctrl     *gomock.Controller
	recorder *MockLicenseMockRecorder
}

// MockLicenseMockRecorder is the mock recorder for MockLicense
type MockLicenseMockRecorder struct {
	mock *MockLicense
}

// NewMockLicense creates a new mock instance
func NewMockLicense(ctrl *gomock.Controller) *MockLicense {
	mock := &MockLicense{ctrl: ctrl}
	mock.recorder = &MockLicenseMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLicense) EXPECT() *MockLicenseMockRecorder {
	return m.recorder
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandard mocks base method
func (m *MockLicense) ComputedLicensesForEquipmentForMetricOracleProcessorStandard(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricOPSComputed, arg4 []string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputedLicensesForEquipmentForMetricOracleProcessorStandard", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandard indicates an expected call of ComputedLicensesForEquipmentForMetricOracleProcessorStandard
func (mr *MockLicenseMockRecorder) ComputedLicensesForEquipmentForMetricOracleProcessorStandard(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputedLicensesForEquipmentForMetricOracleProcessorStandard", reflect.TypeOf((*MockLicense)(nil).ComputedLicensesForEquipmentForMetricOracleProcessorStandard), arg0, arg1, arg2, arg3, arg4)
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll mocks base method
func (m *MockLicense) ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricOPSComputed, arg4 []string) (int64, float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll indicates an expected call of ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll
func (mr *MockLicenseMockRecorder) ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll", reflect.TypeOf((*MockLicense)(nil).ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll), arg0, arg1, arg2, arg3, arg4)
}

// CreateEquipmentType mocks base method
func (m *MockLicense) CreateEquipmentType(arg0 context.Context, arg1 *v1.EquipmentType, arg2 []string) (*v1.EquipmentType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEquipmentType", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.EquipmentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEquipmentType indicates an expected call of CreateEquipmentType
func (mr *MockLicenseMockRecorder) CreateEquipmentType(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEquipmentType", reflect.TypeOf((*MockLicense)(nil).CreateEquipmentType), arg0, arg1, arg2)
}

// CreateProductAggregation mocks base method
func (m *MockLicense) CreateProductAggregation(arg0 context.Context, arg1 *v1.ProductAggregation, arg2 []string) (*v1.ProductAggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateProductAggregation", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.ProductAggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateProductAggregation indicates an expected call of CreateProductAggregation
func (mr *MockLicenseMockRecorder) CreateProductAggregation(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateProductAggregation", reflect.TypeOf((*MockLicense)(nil).CreateProductAggregation), arg0, arg1, arg2)
}

// DeleteProductAggregation mocks base method
func (m *MockLicense) DeleteProductAggregation(arg0 context.Context, arg1 string, arg2 []string) ([]*v1.ProductAggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteProductAggregation", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*v1.ProductAggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteProductAggregation indicates an expected call of DeleteProductAggregation
func (mr *MockLicenseMockRecorder) DeleteProductAggregation(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteProductAggregation", reflect.TypeOf((*MockLicense)(nil).DeleteProductAggregation), arg0, arg1, arg2)
}

// EquipmentTypes mocks base method
func (m *MockLicense) EquipmentTypes(arg0 context.Context, arg1 []string) ([]*v1.EquipmentType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EquipmentTypes", arg0, arg1)
	ret0, _ := ret[0].([]*v1.EquipmentType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EquipmentTypes indicates an expected call of EquipmentTypes
func (mr *MockLicenseMockRecorder) EquipmentTypes(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EquipmentTypes", reflect.TypeOf((*MockLicense)(nil).EquipmentTypes), arg0, arg1)
}

// GetProductInformation mocks base method
func (m *MockLicense) GetProductInformation(arg0 context.Context, arg1 string, arg2 []string) (*v1.ProductAdditionalInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductInformation", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.ProductAdditionalInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductInformation indicates an expected call of GetProductInformation
func (mr *MockLicenseMockRecorder) GetProductInformation(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductInformation", reflect.TypeOf((*MockLicense)(nil).GetProductInformation), arg0, arg1, arg2)
}

// ListMetricACS mocks base method
func (m *MockLicense) ListMetricACS(arg0 context.Context, arg1 []string) ([]*v1.MetricACS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricACS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricACS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricACS indicates an expected call of ListMetricACS
func (mr *MockLicenseMockRecorder) ListMetricACS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricACS", reflect.TypeOf((*MockLicense)(nil).ListMetricACS), arg0, arg1)
}

// ListMetricINM mocks base method
func (m *MockLicense) ListMetricINM(arg0 context.Context, arg1 []string) ([]*v1.MetricINM, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricINM", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricINM)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricINM indicates an expected call of ListMetricINM
func (mr *MockLicenseMockRecorder) ListMetricINM(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricINM", reflect.TypeOf((*MockLicense)(nil).ListMetricINM), arg0, arg1)
}

// ListMetricIPS mocks base method
func (m *MockLicense) ListMetricIPS(arg0 context.Context, arg1 []string) ([]*v1.MetricIPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricIPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricIPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricIPS indicates an expected call of ListMetricIPS
func (mr *MockLicenseMockRecorder) ListMetricIPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricIPS", reflect.TypeOf((*MockLicense)(nil).ListMetricIPS), arg0, arg1)
}

// ListMetricNUP mocks base method
func (m *MockLicense) ListMetricNUP(arg0 context.Context, arg1 []string) ([]*v1.MetricNUPOracle, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricNUP", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricNUPOracle)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricNUP indicates an expected call of ListMetricNUP
func (mr *MockLicenseMockRecorder) ListMetricNUP(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricNUP", reflect.TypeOf((*MockLicense)(nil).ListMetricNUP), arg0, arg1)
}

// ListMetricOPS mocks base method
func (m *MockLicense) ListMetricOPS(arg0 context.Context, arg1 []string) ([]*v1.MetricOPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricOPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricOPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricOPS indicates an expected call of ListMetricOPS
func (mr *MockLicenseMockRecorder) ListMetricOPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricOPS", reflect.TypeOf((*MockLicense)(nil).ListMetricOPS), arg0, arg1)
}

// ListMetricSPS mocks base method
func (m *MockLicense) ListMetricSPS(arg0 context.Context, arg1 []string) ([]*v1.MetricSPS, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetricSPS", arg0, arg1)
	ret0, _ := ret[0].([]*v1.MetricSPS)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetricSPS indicates an expected call of ListMetricSPS
func (mr *MockLicenseMockRecorder) ListMetricSPS(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetricSPS", reflect.TypeOf((*MockLicense)(nil).ListMetricSPS), arg0, arg1)
}

// ListMetrices mocks base method
func (m *MockLicense) ListMetrices(arg0 context.Context, arg1 []string) ([]*v1.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMetrices", arg0, arg1)
	ret0, _ := ret[0].([]*v1.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMetrices indicates an expected call of ListMetrices
func (mr *MockLicenseMockRecorder) ListMetrices(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMetrices", reflect.TypeOf((*MockLicense)(nil).ListMetrices), arg0, arg1)
}

// MetadataAllWithType mocks base method
func (m *MockLicense) MetadataAllWithType(arg0 context.Context, arg1 v1.MetadataType, arg2 []string) ([]*v1.Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetadataAllWithType", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*v1.Metadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetadataAllWithType indicates an expected call of MetadataAllWithType
func (mr *MockLicenseMockRecorder) MetadataAllWithType(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetadataAllWithType", reflect.TypeOf((*MockLicense)(nil).MetadataAllWithType), arg0, arg1, arg2)
}

// MetricACSComputedLicenses mocks base method
func (m *MockLicense) MetricACSComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricACSComputed, arg3 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricACSComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricACSComputedLicenses indicates an expected call of MetricACSComputedLicenses
func (mr *MockLicenseMockRecorder) MetricACSComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricACSComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricACSComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricACSComputedLicensesAgg mocks base method
func (m *MockLicense) MetricACSComputedLicensesAgg(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricACSComputed, arg4 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricACSComputedLicensesAgg", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricACSComputedLicensesAgg indicates an expected call of MetricACSComputedLicensesAgg
func (mr *MockLicenseMockRecorder) MetricACSComputedLicensesAgg(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricACSComputedLicensesAgg", reflect.TypeOf((*MockLicense)(nil).MetricACSComputedLicensesAgg), arg0, arg1, arg2, arg3, arg4)
}

// MetricINMComputedLicenses mocks base method
func (m *MockLicense) MetricINMComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricINMComputed, arg3 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricINMComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricINMComputedLicenses indicates an expected call of MetricINMComputedLicenses
func (mr *MockLicenseMockRecorder) MetricINMComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricINMComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricINMComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricIPSComputedLicenses mocks base method
func (m *MockLicense) MetricIPSComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricIPSComputed, arg3 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricIPSComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricIPSComputedLicenses indicates an expected call of MetricIPSComputedLicenses
func (mr *MockLicenseMockRecorder) MetricIPSComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricIPSComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricIPSComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricIPSComputedLicensesAgg mocks base method
func (m *MockLicense) MetricIPSComputedLicensesAgg(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricIPSComputed, arg4 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricIPSComputedLicensesAgg", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricIPSComputedLicensesAgg indicates an expected call of MetricIPSComputedLicensesAgg
func (mr *MockLicenseMockRecorder) MetricIPSComputedLicensesAgg(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricIPSComputedLicensesAgg", reflect.TypeOf((*MockLicense)(nil).MetricIPSComputedLicensesAgg), arg0, arg1, arg2, arg3, arg4)
}

// MetricNUPComputedLicenses mocks base method
func (m *MockLicense) MetricNUPComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricNUPComputed, arg3 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricNUPComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricNUPComputedLicenses indicates an expected call of MetricNUPComputedLicenses
func (mr *MockLicenseMockRecorder) MetricNUPComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricNUPComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricNUPComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricNUPComputedLicensesAgg mocks base method
func (m *MockLicense) MetricNUPComputedLicensesAgg(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricNUPComputed, arg4 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricNUPComputedLicensesAgg", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricNUPComputedLicensesAgg indicates an expected call of MetricNUPComputedLicensesAgg
func (mr *MockLicenseMockRecorder) MetricNUPComputedLicensesAgg(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricNUPComputedLicensesAgg", reflect.TypeOf((*MockLicense)(nil).MetricNUPComputedLicensesAgg), arg0, arg1, arg2, arg3, arg4)
}

// MetricOPSComputedLicenses mocks base method
func (m *MockLicense) MetricOPSComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricOPSComputed, arg3 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricOPSComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricOPSComputedLicenses indicates an expected call of MetricOPSComputedLicenses
func (mr *MockLicenseMockRecorder) MetricOPSComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricOPSComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricOPSComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricOPSComputedLicensesAgg mocks base method
func (m *MockLicense) MetricOPSComputedLicensesAgg(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricOPSComputed, arg4 []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricOPSComputedLicensesAgg", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MetricOPSComputedLicensesAgg indicates an expected call of MetricOPSComputedLicensesAgg
func (mr *MockLicenseMockRecorder) MetricOPSComputedLicensesAgg(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricOPSComputedLicensesAgg", reflect.TypeOf((*MockLicense)(nil).MetricOPSComputedLicensesAgg), arg0, arg1, arg2, arg3, arg4)
}

// MetricSPSComputedLicenses mocks base method
func (m *MockLicense) MetricSPSComputedLicenses(arg0 context.Context, arg1 string, arg2 *v1.MetricSPSComputed, arg3 []string) (uint64, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricSPSComputedLicenses", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// MetricSPSComputedLicenses indicates an expected call of MetricSPSComputedLicenses
func (mr *MockLicenseMockRecorder) MetricSPSComputedLicenses(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricSPSComputedLicenses", reflect.TypeOf((*MockLicense)(nil).MetricSPSComputedLicenses), arg0, arg1, arg2, arg3)
}

// MetricSPSComputedLicensesAgg mocks base method
func (m *MockLicense) MetricSPSComputedLicensesAgg(arg0 context.Context, arg1, arg2 string, arg3 *v1.MetricSPSComputed, arg4 []string) (uint64, uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MetricSPSComputedLicensesAgg", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(uint64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// MetricSPSComputedLicensesAgg indicates an expected call of MetricSPSComputedLicensesAgg
func (mr *MockLicenseMockRecorder) MetricSPSComputedLicensesAgg(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MetricSPSComputedLicensesAgg", reflect.TypeOf((*MockLicense)(nil).MetricSPSComputedLicensesAgg), arg0, arg1, arg2, arg3, arg4)
}

// ParentsHirerachyForEquipment mocks base method
func (m *MockLicense) ParentsHirerachyForEquipment(arg0 context.Context, arg1, arg2 string, arg3 byte, arg4 []string) (*v1.Equipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParentsHirerachyForEquipment", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*v1.Equipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ParentsHirerachyForEquipment indicates an expected call of ParentsHirerachyForEquipment
func (mr *MockLicenseMockRecorder) ParentsHirerachyForEquipment(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParentsHirerachyForEquipment", reflect.TypeOf((*MockLicense)(nil).ParentsHirerachyForEquipment), arg0, arg1, arg2, arg3, arg4)
}

// ProductAcquiredRights mocks base method
func (m *MockLicense) ProductAcquiredRights(arg0 context.Context, arg1 string, arg2 []string) (string, []*v1.ProductAcquiredRight, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductAcquiredRights", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].([]*v1.ProductAcquiredRight)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ProductAcquiredRights indicates an expected call of ProductAcquiredRights
func (mr *MockLicenseMockRecorder) ProductAcquiredRights(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductAcquiredRights", reflect.TypeOf((*MockLicense)(nil).ProductAcquiredRights), arg0, arg1, arg2)
}

// ProductAggregationDetails mocks base method
func (m *MockLicense) ProductAggregationDetails(arg0 context.Context, arg1 string, arg2 *v1.QueryProductAggregations, arg3 []string) (*v1.ProductAggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductAggregationDetails", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(*v1.ProductAggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductAggregationDetails indicates an expected call of ProductAggregationDetails
func (mr *MockLicenseMockRecorder) ProductAggregationDetails(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductAggregationDetails", reflect.TypeOf((*MockLicense)(nil).ProductAggregationDetails), arg0, arg1, arg2, arg3)
}

// ProductAggregationsByName mocks base method
func (m *MockLicense) ProductAggregationsByName(arg0 context.Context, arg1 string, arg2 []string) (*v1.ProductAggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductAggregationsByName", arg0, arg1, arg2)
	ret0, _ := ret[0].(*v1.ProductAggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductAggregationsByName indicates an expected call of ProductAggregationsByName
func (mr *MockLicenseMockRecorder) ProductAggregationsByName(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductAggregationsByName", reflect.TypeOf((*MockLicense)(nil).ProductAggregationsByName), arg0, arg1, arg2)
}

// ProductIDForSwidtag mocks base method
func (m *MockLicense) ProductIDForSwidtag(arg0 context.Context, arg1 string, arg2 *v1.QueryProducts, arg3 []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductIDForSwidtag", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductIDForSwidtag indicates an expected call of ProductIDForSwidtag
func (mr *MockLicenseMockRecorder) ProductIDForSwidtag(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductIDForSwidtag", reflect.TypeOf((*MockLicense)(nil).ProductIDForSwidtag), arg0, arg1, arg2, arg3)
}

// ProductsForEquipmentForMetricIPSStandard mocks base method
func (m *MockLicense) ProductsForEquipmentForMetricIPSStandard(arg0 context.Context, arg1, arg2 string, arg3 byte, arg4 *v1.MetricIPSComputed, arg5 []string) ([]*v1.ProductData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsForEquipmentForMetricIPSStandard", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].([]*v1.ProductData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsForEquipmentForMetricIPSStandard indicates an expected call of ProductsForEquipmentForMetricIPSStandard
func (mr *MockLicenseMockRecorder) ProductsForEquipmentForMetricIPSStandard(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsForEquipmentForMetricIPSStandard", reflect.TypeOf((*MockLicense)(nil).ProductsForEquipmentForMetricIPSStandard), arg0, arg1, arg2, arg3, arg4, arg5)
}

// ProductsForEquipmentForMetricOracleNUPStandard mocks base method
func (m *MockLicense) ProductsForEquipmentForMetricOracleNUPStandard(arg0 context.Context, arg1, arg2 string, arg3 byte, arg4 *v1.MetricNUPComputed, arg5 []string) ([]*v1.ProductData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsForEquipmentForMetricOracleNUPStandard", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].([]*v1.ProductData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsForEquipmentForMetricOracleNUPStandard indicates an expected call of ProductsForEquipmentForMetricOracleNUPStandard
func (mr *MockLicenseMockRecorder) ProductsForEquipmentForMetricOracleNUPStandard(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsForEquipmentForMetricOracleNUPStandard", reflect.TypeOf((*MockLicense)(nil).ProductsForEquipmentForMetricOracleNUPStandard), arg0, arg1, arg2, arg3, arg4, arg5)
}

// ProductsForEquipmentForMetricOracleProcessorStandard mocks base method
func (m *MockLicense) ProductsForEquipmentForMetricOracleProcessorStandard(arg0 context.Context, arg1, arg2 string, arg3 byte, arg4 *v1.MetricOPSComputed, arg5 []string) ([]*v1.ProductData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsForEquipmentForMetricOracleProcessorStandard", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].([]*v1.ProductData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsForEquipmentForMetricOracleProcessorStandard indicates an expected call of ProductsForEquipmentForMetricOracleProcessorStandard
func (mr *MockLicenseMockRecorder) ProductsForEquipmentForMetricOracleProcessorStandard(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsForEquipmentForMetricOracleProcessorStandard", reflect.TypeOf((*MockLicense)(nil).ProductsForEquipmentForMetricOracleProcessorStandard), arg0, arg1, arg2, arg3, arg4, arg5)
}

// ProductsForEquipmentForMetricSAGStandard mocks base method
func (m *MockLicense) ProductsForEquipmentForMetricSAGStandard(arg0 context.Context, arg1, arg2 string, arg3 byte, arg4 *v1.MetricSPSComputed, arg5 []string) ([]*v1.ProductData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductsForEquipmentForMetricSAGStandard", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].([]*v1.ProductData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductsForEquipmentForMetricSAGStandard indicates an expected call of ProductsForEquipmentForMetricSAGStandard
func (mr *MockLicenseMockRecorder) ProductsForEquipmentForMetricSAGStandard(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductsForEquipmentForMetricSAGStandard", reflect.TypeOf((*MockLicense)(nil).ProductsForEquipmentForMetricSAGStandard), arg0, arg1, arg2, arg3, arg4, arg5)
}

// UpdateProductAggregation mocks base method
func (m *MockLicense) UpdateProductAggregation(arg0 context.Context, arg1 string, arg2 *v1.UpdateProductAggregationRequest, arg3 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProductAggregation", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateProductAggregation indicates an expected call of UpdateProductAggregation
func (mr *MockLicenseMockRecorder) UpdateProductAggregation(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProductAggregation", reflect.TypeOf((*MockLicense)(nil).UpdateProductAggregation), arg0, arg1, arg2, arg3)
}

// UsersForEquipmentForMetricOracleNUP mocks base method
func (m *MockLicense) UsersForEquipmentForMetricOracleNUP(arg0 context.Context, arg1, arg2, arg3 string, arg4 byte, arg5 *v1.MetricNUPComputed, arg6 []string) ([]*v1.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UsersForEquipmentForMetricOracleNUP", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].([]*v1.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UsersForEquipmentForMetricOracleNUP indicates an expected call of UsersForEquipmentForMetricOracleNUP
func (mr *MockLicenseMockRecorder) UsersForEquipmentForMetricOracleNUP(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UsersForEquipmentForMetricOracleNUP", reflect.TypeOf((*MockLicense)(nil).UsersForEquipmentForMetricOracleNUP), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}
