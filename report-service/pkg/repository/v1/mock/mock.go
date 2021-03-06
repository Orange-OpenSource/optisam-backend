// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/report-service/pkg/repository/v1 (interfaces: Report)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	json "encoding/json"
	gomock "github.com/golang/mock/gomock"
	db "optisam-backend/report-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockReport is a mock of Report interface
type MockReport struct {
	ctrl     *gomock.Controller
	recorder *MockReportMockRecorder
}

// MockReportMockRecorder is the mock recorder for MockReport
type MockReportMockRecorder struct {
	mock *MockReport
}

// NewMockReport creates a new mock instance
func NewMockReport(ctrl *gomock.Controller) *MockReport {
	mock := &MockReport{ctrl: ctrl}
	mock.recorder = &MockReportMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReport) EXPECT() *MockReportMockRecorder {
	return m.recorder
}

// DownloadReport mocks base method
func (m *MockReport) DownloadReport(arg0 context.Context, arg1 db.DownloadReportParams) (json.RawMessage, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadReport", arg0, arg1)
	ret0, _ := ret[0].(json.RawMessage)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadReport indicates an expected call of DownloadReport
func (mr *MockReportMockRecorder) DownloadReport(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadReport", reflect.TypeOf((*MockReport)(nil).DownloadReport), arg0, arg1)
}

// GetReport mocks base method
func (m *MockReport) GetReport(arg0 context.Context, arg1 db.GetReportParams) ([]db.GetReportRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReport", arg0, arg1)
	ret0, _ := ret[0].([]db.GetReportRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReport indicates an expected call of GetReport
func (mr *MockReportMockRecorder) GetReport(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReport", reflect.TypeOf((*MockReport)(nil).GetReport), arg0, arg1)
}

// GetReportType mocks base method
func (m *MockReport) GetReportType(arg0 context.Context, arg1 int32) (db.ReportType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReportType", arg0, arg1)
	ret0, _ := ret[0].(db.ReportType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReportType indicates an expected call of GetReportType
func (mr *MockReportMockRecorder) GetReportType(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReportType", reflect.TypeOf((*MockReport)(nil).GetReportType), arg0, arg1)
}

// GetReportTypes mocks base method
func (m *MockReport) GetReportTypes(arg0 context.Context) ([]db.ReportType, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReportTypes", arg0)
	ret0, _ := ret[0].([]db.ReportType)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetReportTypes indicates an expected call of GetReportTypes
func (mr *MockReportMockRecorder) GetReportTypes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReportTypes", reflect.TypeOf((*MockReport)(nil).GetReportTypes), arg0)
}

// InsertReportData mocks base method
func (m *MockReport) InsertReportData(arg0 context.Context, arg1 db.InsertReportDataParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertReportData", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertReportData indicates an expected call of InsertReportData
func (mr *MockReportMockRecorder) InsertReportData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertReportData", reflect.TypeOf((*MockReport)(nil).InsertReportData), arg0, arg1)
}

// SubmitReport mocks base method
func (m *MockReport) SubmitReport(arg0 context.Context, arg1 db.SubmitReportParams) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SubmitReport", arg0, arg1)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SubmitReport indicates an expected call of SubmitReport
func (mr *MockReportMockRecorder) SubmitReport(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SubmitReport", reflect.TypeOf((*MockReport)(nil).SubmitReport), arg0, arg1)
}

// UpdateReportStatus mocks base method
func (m *MockReport) UpdateReportStatus(arg0 context.Context, arg1 db.UpdateReportStatusParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateReportStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateReportStatus indicates an expected call of UpdateReportStatus
func (mr *MockReportMockRecorder) UpdateReportStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateReportStatus", reflect.TypeOf((*MockReport)(nil).UpdateReportStatus), arg0, arg1)
}
