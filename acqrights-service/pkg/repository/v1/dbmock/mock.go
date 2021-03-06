// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/acqrights-service/pkg/repository/v1 (interfaces: AcqRights)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	db "optisam-backend/acqrights-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockAcqRights is a mock of AcqRights interface
type MockAcqRights struct {
	ctrl     *gomock.Controller
	recorder *MockAcqRightsMockRecorder
}

// MockAcqRightsMockRecorder is the mock recorder for MockAcqRights
type MockAcqRightsMockRecorder struct {
	mock *MockAcqRights
}

// NewMockAcqRights creates a new mock instance
func NewMockAcqRights(ctrl *gomock.Controller) *MockAcqRights {
	mock := &MockAcqRights{ctrl: ctrl}
	mock.recorder = &MockAcqRightsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAcqRights) EXPECT() *MockAcqRightsMockRecorder {
	return m.recorder
}

// DeleteAggregation mocks base method
func (m *MockAcqRights) DeleteAggregation(arg0 context.Context, arg1 db.DeleteAggregationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAggregation", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAggregation indicates an expected call of DeleteAggregation
func (mr *MockAcqRightsMockRecorder) DeleteAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAggregation", reflect.TypeOf((*MockAcqRights)(nil).DeleteAggregation), arg0, arg1)
}

// InsertAggregation mocks base method
func (m *MockAcqRights) InsertAggregation(arg0 context.Context, arg1 db.InsertAggregationParams) (db.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertAggregation", arg0, arg1)
	ret0, _ := ret[0].(db.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertAggregation indicates an expected call of InsertAggregation
func (mr *MockAcqRightsMockRecorder) InsertAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertAggregation", reflect.TypeOf((*MockAcqRights)(nil).InsertAggregation), arg0, arg1)
}

// ListAcqRightsAggregation mocks base method
func (m *MockAcqRights) ListAcqRightsAggregation(arg0 context.Context, arg1 db.ListAcqRightsAggregationParams) ([]db.ListAcqRightsAggregationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsAggregation", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsAggregationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsAggregation indicates an expected call of ListAcqRightsAggregation
func (mr *MockAcqRightsMockRecorder) ListAcqRightsAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsAggregation", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsAggregation), arg0, arg1)
}

// ListAcqRightsAggregationIndividual mocks base method
func (m *MockAcqRights) ListAcqRightsAggregationIndividual(arg0 context.Context, arg1 db.ListAcqRightsAggregationIndividualParams) ([]db.ListAcqRightsAggregationIndividualRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsAggregationIndividual", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsAggregationIndividualRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsAggregationIndividual indicates an expected call of ListAcqRightsAggregationIndividual
func (mr *MockAcqRightsMockRecorder) ListAcqRightsAggregationIndividual(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsAggregationIndividual", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsAggregationIndividual), arg0, arg1)
}

// ListAcqRightsEditors mocks base method
func (m *MockAcqRights) ListAcqRightsEditors(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsEditors", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsEditors indicates an expected call of ListAcqRightsEditors
func (mr *MockAcqRightsMockRecorder) ListAcqRightsEditors(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsEditors", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsEditors), arg0, arg1)
}

// ListAcqRightsIndividual mocks base method
func (m *MockAcqRights) ListAcqRightsIndividual(arg0 context.Context, arg1 db.ListAcqRightsIndividualParams) ([]db.ListAcqRightsIndividualRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsIndividual", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsIndividualRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsIndividual indicates an expected call of ListAcqRightsIndividual
func (mr *MockAcqRightsMockRecorder) ListAcqRightsIndividual(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsIndividual", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsIndividual), arg0, arg1)
}

// ListAcqRightsMetrics mocks base method
func (m *MockAcqRights) ListAcqRightsMetrics(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsMetrics", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsMetrics indicates an expected call of ListAcqRightsMetrics
func (mr *MockAcqRightsMockRecorder) ListAcqRightsMetrics(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsMetrics", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsMetrics), arg0, arg1)
}

// ListAcqRightsProducts mocks base method
func (m *MockAcqRights) ListAcqRightsProducts(arg0 context.Context, arg1 db.ListAcqRightsProductsParams) ([]db.ListAcqRightsProductsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsProducts", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAcqRightsProductsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsProducts indicates an expected call of ListAcqRightsProducts
func (mr *MockAcqRightsMockRecorder) ListAcqRightsProducts(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsProducts", reflect.TypeOf((*MockAcqRights)(nil).ListAcqRightsProducts), arg0, arg1)
}

// ListAggregation mocks base method
func (m *MockAcqRights) ListAggregation(arg0 context.Context, arg1 []string) ([]db.ListAggregationRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAggregation", arg0, arg1)
	ret0, _ := ret[0].([]db.ListAggregationRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAggregation indicates an expected call of ListAggregation
func (mr *MockAcqRightsMockRecorder) ListAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAggregation", reflect.TypeOf((*MockAcqRights)(nil).ListAggregation), arg0, arg1)
}

// UpdateAggregation mocks base method
func (m *MockAcqRights) UpdateAggregation(arg0 context.Context, arg1 db.UpdateAggregationParams) (db.Aggregation, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAggregation", arg0, arg1)
	ret0, _ := ret[0].(db.Aggregation)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateAggregation indicates an expected call of UpdateAggregation
func (mr *MockAcqRightsMockRecorder) UpdateAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAggregation", reflect.TypeOf((*MockAcqRights)(nil).UpdateAggregation), arg0, arg1)
}

// UpsertAcqRights mocks base method
func (m *MockAcqRights) UpsertAcqRights(arg0 context.Context, arg1 db.UpsertAcqRightsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertAcqRights", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertAcqRights indicates an expected call of UpsertAcqRights
func (mr *MockAcqRightsMockRecorder) UpsertAcqRights(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertAcqRights", reflect.TypeOf((*MockAcqRights)(nil).UpsertAcqRights), arg0, arg1)
}
