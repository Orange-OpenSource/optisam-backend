// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/dps-service/pkg/repository/v1 (interfaces: Dps)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	db "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockDps is a mock of Dps interface
type MockDps struct {
	ctrl     *gomock.Controller
	recorder *MockDpsMockRecorder
}

// MockDpsMockRecorder is the mock recorder for MockDps
type MockDpsMockRecorder struct {
	mock *MockDps
}

// NewMockDps creates a new mock instance
func NewMockDps(ctrl *gomock.Controller) *MockDps {
	mock := &MockDps{ctrl: ctrl}
	mock.recorder = &MockDpsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockDps) EXPECT() *MockDpsMockRecorder {
	return m.recorder
}

// GetDataFileRecords mocks base method
func (m *MockDps) GetDataFileRecords(arg0 context.Context, arg1 db.GetDataFileRecordsParams) (db.GetDataFileRecordsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataFileRecords", arg0, arg1)
	ret0, _ := ret[0].(db.GetDataFileRecordsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataFileRecords indicates an expected call of GetDataFileRecords
func (mr *MockDpsMockRecorder) GetDataFileRecords(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataFileRecords", reflect.TypeOf((*MockDps)(nil).GetDataFileRecords), arg0, arg1)
}

// GetEntityMonthWise mocks base method
func (m *MockDps) GetEntityMonthWise(arg0 context.Context, arg1 db.GetEntityMonthWiseParams) ([]db.GetEntityMonthWiseRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEntityMonthWise", arg0, arg1)
	ret0, _ := ret[0].([]db.GetEntityMonthWiseRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEntityMonthWise indicates an expected call of GetEntityMonthWise
func (mr *MockDpsMockRecorder) GetEntityMonthWise(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEntityMonthWise", reflect.TypeOf((*MockDps)(nil).GetEntityMonthWise), arg0, arg1)
}

// GetFailedRecord mocks base method
func (m *MockDps) GetFailedRecord(arg0 context.Context, arg1 db.GetFailedRecordParams) ([]db.GetFailedRecordRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFailedRecord", arg0, arg1)
	ret0, _ := ret[0].([]db.GetFailedRecordRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFailedRecord indicates an expected call of GetFailedRecord
func (mr *MockDpsMockRecorder) GetFailedRecord(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFailedRecord", reflect.TypeOf((*MockDps)(nil).GetFailedRecord), arg0, arg1)
}

// GetFailureReasons mocks base method
func (m *MockDps) GetFailureReasons(arg0 context.Context, arg1 db.GetFailureReasonsParams) ([]db.GetFailureReasonsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFailureReasons", arg0, arg1)
	ret0, _ := ret[0].([]db.GetFailureReasonsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFailureReasons indicates an expected call of GetFailureReasons
func (mr *MockDpsMockRecorder) GetFailureReasons(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFailureReasons", reflect.TypeOf((*MockDps)(nil).GetFailureReasons), arg0, arg1)
}

// GetFileStatus mocks base method
func (m *MockDps) GetFileStatus(arg0 context.Context, arg1 db.GetFileStatusParams) (db.UploadStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFileStatus", arg0, arg1)
	ret0, _ := ret[0].(db.UploadStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFileStatus indicates an expected call of GetFileStatus
func (mr *MockDpsMockRecorder) GetFileStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFileStatus", reflect.TypeOf((*MockDps)(nil).GetFileStatus), arg0, arg1)
}

// InsertUploadedData mocks base method
func (m *MockDps) InsertUploadedData(arg0 context.Context, arg1 db.InsertUploadedDataParams) (db.UploadedDataFile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertUploadedData", arg0, arg1)
	ret0, _ := ret[0].(db.UploadedDataFile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertUploadedData indicates an expected call of InsertUploadedData
func (mr *MockDpsMockRecorder) InsertUploadedData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertUploadedData", reflect.TypeOf((*MockDps)(nil).InsertUploadedData), arg0, arg1)
}

// InsertUploadedMetaData mocks base method
func (m *MockDps) InsertUploadedMetaData(arg0 context.Context, arg1 db.InsertUploadedMetaDataParams) (db.UploadedDataFile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertUploadedMetaData", arg0, arg1)
	ret0, _ := ret[0].(db.UploadedDataFile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertUploadedMetaData indicates an expected call of InsertUploadedMetaData
func (mr *MockDpsMockRecorder) InsertUploadedMetaData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertUploadedMetaData", reflect.TypeOf((*MockDps)(nil).InsertUploadedMetaData), arg0, arg1)
}

// ListUploadedDataFiles mocks base method
func (m *MockDps) ListUploadedDataFiles(arg0 context.Context, arg1 db.ListUploadedDataFilesParams) ([]db.ListUploadedDataFilesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadedDataFiles", arg0, arg1)
	ret0, _ := ret[0].([]db.ListUploadedDataFilesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadedDataFiles indicates an expected call of ListUploadedDataFiles
func (mr *MockDpsMockRecorder) ListUploadedDataFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadedDataFiles", reflect.TypeOf((*MockDps)(nil).ListUploadedDataFiles), arg0, arg1)
}

// ListUploadedGlobalDataFiles mocks base method
func (m *MockDps) ListUploadedGlobalDataFiles(arg0 context.Context, arg1 db.ListUploadedGlobalDataFilesParams) ([]db.ListUploadedGlobalDataFilesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadedGlobalDataFiles", arg0, arg1)
	ret0, _ := ret[0].([]db.ListUploadedGlobalDataFilesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadedGlobalDataFiles indicates an expected call of ListUploadedGlobalDataFiles
func (mr *MockDpsMockRecorder) ListUploadedGlobalDataFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadedGlobalDataFiles", reflect.TypeOf((*MockDps)(nil).ListUploadedGlobalDataFiles), arg0, arg1)
}

// ListUploadedMetaDataFiles mocks base method
func (m *MockDps) ListUploadedMetaDataFiles(arg0 context.Context, arg1 db.ListUploadedMetaDataFilesParams) ([]db.ListUploadedMetaDataFilesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUploadedMetaDataFiles", arg0, arg1)
	ret0, _ := ret[0].([]db.ListUploadedMetaDataFilesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUploadedMetaDataFiles indicates an expected call of ListUploadedMetaDataFiles
func (mr *MockDpsMockRecorder) ListUploadedMetaDataFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUploadedMetaDataFiles", reflect.TypeOf((*MockDps)(nil).ListUploadedMetaDataFiles), arg0, arg1)
}

// UpdateFileFailedRecord mocks base method
func (m *MockDps) UpdateFileFailedRecord(arg0 context.Context, arg1 db.UpdateFileFailedRecordParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileFailedRecord", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileFailedRecord indicates an expected call of UpdateFileFailedRecord
func (mr *MockDpsMockRecorder) UpdateFileFailedRecord(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileFailedRecord", reflect.TypeOf((*MockDps)(nil).UpdateFileFailedRecord), arg0, arg1)
}

// UpdateFileFailure mocks base method
func (m *MockDps) UpdateFileFailure(arg0 context.Context, arg1 db.UpdateFileFailureParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileFailure", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileFailure indicates an expected call of UpdateFileFailure
func (mr *MockDpsMockRecorder) UpdateFileFailure(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileFailure", reflect.TypeOf((*MockDps)(nil).UpdateFileFailure), arg0, arg1)
}

// UpdateFileStatus mocks base method
func (m *MockDps) UpdateFileStatus(arg0 context.Context, arg1 db.UpdateFileStatusParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileStatus", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileStatus indicates an expected call of UpdateFileStatus
func (mr *MockDpsMockRecorder) UpdateFileStatus(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileStatus", reflect.TypeOf((*MockDps)(nil).UpdateFileStatus), arg0, arg1)
}

// UpdateFileSuccessRecord mocks base method
func (m *MockDps) UpdateFileSuccessRecord(arg0 context.Context, arg1 db.UpdateFileSuccessRecordParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileSuccessRecord", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileSuccessRecord indicates an expected call of UpdateFileSuccessRecord
func (mr *MockDpsMockRecorder) UpdateFileSuccessRecord(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileSuccessRecord", reflect.TypeOf((*MockDps)(nil).UpdateFileSuccessRecord), arg0, arg1)
}

// UpdateFileTotalRecord mocks base method
func (m *MockDps) UpdateFileTotalRecord(arg0 context.Context, arg1 db.UpdateFileTotalRecordParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateFileTotalRecord", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateFileTotalRecord indicates an expected call of UpdateFileTotalRecord
func (mr *MockDpsMockRecorder) UpdateFileTotalRecord(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFileTotalRecord", reflect.TypeOf((*MockDps)(nil).UpdateFileTotalRecord), arg0, arg1)
}