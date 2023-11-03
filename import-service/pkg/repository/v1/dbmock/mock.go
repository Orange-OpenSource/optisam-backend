// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1 (interfaces: Import)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	sql "database/sql"
	gomock "github.com/golang/mock/gomock"
	db "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockImport is a mock of Import interface
type MockImport struct {
	ctrl     *gomock.Controller
	recorder *MockImportMockRecorder
}

// MockImportMockRecorder is the mock recorder for MockImport
type MockImportMockRecorder struct {
	mock *MockImport
}

// NewMockImport creates a new mock instance
func NewMockImport(ctrl *gomock.Controller) *MockImport {
	mock := &MockImport{ctrl: ctrl}
	mock.recorder = &MockImportMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockImport) EXPECT() *MockImportMockRecorder {
	return m.recorder
}

// GetDgraphCompletedBatches mocks base method
func (m *MockImport) GetDgraphCompletedBatches(arg0 context.Context, arg1 string) (sql.NullInt32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDgraphCompletedBatches", arg0, arg1)
	ret0, _ := ret[0].(sql.NullInt32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDgraphCompletedBatches indicates an expected call of GetDgraphCompletedBatches
func (mr *MockImportMockRecorder) GetDgraphCompletedBatches(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDgraphCompletedBatches", reflect.TypeOf((*MockImport)(nil).GetDgraphCompletedBatches), arg0, arg1)
}

// InsertNominativeUserRequest mocks base method
func (m *MockImport) InsertNominativeUserRequest(arg0 context.Context, arg1 db.InsertNominativeUserRequestParams) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertNominativeUserRequest", arg0, arg1)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertNominativeUserRequest indicates an expected call of InsertNominativeUserRequest
func (mr *MockImportMockRecorder) InsertNominativeUserRequest(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertNominativeUserRequest", reflect.TypeOf((*MockImport)(nil).InsertNominativeUserRequest), arg0, arg1)
}

// InsertNominativeUserRequestDetails mocks base method
func (m *MockImport) InsertNominativeUserRequestDetails(arg0 context.Context, arg1 db.InsertNominativeUserRequestDetailsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertNominativeUserRequestDetails", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertNominativeUserRequestDetails indicates an expected call of InsertNominativeUserRequestDetails
func (mr *MockImportMockRecorder) InsertNominativeUserRequestDetails(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertNominativeUserRequestDetails", reflect.TypeOf((*MockImport)(nil).InsertNominativeUserRequestDetails), arg0, arg1)
}

// InsertNominativeUserRequestTx mocks base method
func (m *MockImport) InsertNominativeUserRequestTx(arg0 context.Context, arg1 db.InsertNominativeUserRequestParams, arg2 db.InsertNominativeUserRequestDetailsParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertNominativeUserRequestTx", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertNominativeUserRequestTx indicates an expected call of InsertNominativeUserRequestTx
func (mr *MockImportMockRecorder) InsertNominativeUserRequestTx(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertNominativeUserRequestTx", reflect.TypeOf((*MockImport)(nil).InsertNominativeUserRequestTx), arg0, arg1, arg2)
}

// ListNominativeUsersUploadedFiles mocks base method
func (m *MockImport) ListNominativeUsersUploadedFiles(arg0 context.Context, arg1 db.ListNominativeUsersUploadedFilesParams) ([]db.ListNominativeUsersUploadedFilesRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListNominativeUsersUploadedFiles", arg0, arg1)
	ret0, _ := ret[0].([]db.ListNominativeUsersUploadedFilesRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListNominativeUsersUploadedFiles indicates an expected call of ListNominativeUsersUploadedFiles
func (mr *MockImportMockRecorder) ListNominativeUsersUploadedFiles(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListNominativeUsersUploadedFiles", reflect.TypeOf((*MockImport)(nil).ListNominativeUsersUploadedFiles), arg0, arg1)
}

// UpdateNominativeUserDetailsRequestAnalysis mocks base method
func (m *MockImport) UpdateNominativeUserDetailsRequestAnalysis(arg0 context.Context, arg1 db.UpdateNominativeUserDetailsRequestAnalysisParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserDetailsRequestAnalysis", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNominativeUserDetailsRequestAnalysis indicates an expected call of UpdateNominativeUserDetailsRequestAnalysis
func (mr *MockImportMockRecorder) UpdateNominativeUserDetailsRequestAnalysis(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserDetailsRequestAnalysis", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserDetailsRequestAnalysis), arg0, arg1)
}

// UpdateNominativeUserRequestAnalysis mocks base method
func (m *MockImport) UpdateNominativeUserRequestAnalysis(arg0 context.Context, arg1 db.UpdateNominativeUserRequestAnalysisParams) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestAnalysis", arg0, arg1)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNominativeUserRequestAnalysis indicates an expected call of UpdateNominativeUserRequestAnalysis
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestAnalysis(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestAnalysis", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestAnalysis), arg0, arg1)
}

// UpdateNominativeUserRequestAnalysisTx mocks base method
func (m *MockImport) UpdateNominativeUserRequestAnalysisTx(arg0 context.Context, arg1 db.UpdateNominativeUserRequestAnalysisParams, arg2 db.UpdateNominativeUserDetailsRequestAnalysisParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestAnalysisTx", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNominativeUserRequestAnalysisTx indicates an expected call of UpdateNominativeUserRequestAnalysisTx
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestAnalysisTx(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestAnalysisTx", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestAnalysisTx), arg0, arg1, arg2)
}

// UpdateNominativeUserRequestDgraphBatchSuccess mocks base method
func (m *MockImport) UpdateNominativeUserRequestDgraphBatchSuccess(arg0 context.Context, arg1 string) (db.UpdateNominativeUserRequestDgraphBatchSuccessRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestDgraphBatchSuccess", arg0, arg1)
	ret0, _ := ret[0].(db.UpdateNominativeUserRequestDgraphBatchSuccessRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNominativeUserRequestDgraphBatchSuccess indicates an expected call of UpdateNominativeUserRequestDgraphBatchSuccess
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestDgraphBatchSuccess(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestDgraphBatchSuccess", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestDgraphBatchSuccess), arg0, arg1)
}

// UpdateNominativeUserRequestDgraphSuccess mocks base method
func (m *MockImport) UpdateNominativeUserRequestDgraphSuccess(arg0 context.Context, arg1 db.UpdateNominativeUserRequestDgraphSuccessParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestDgraphSuccess", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNominativeUserRequestDgraphSuccess indicates an expected call of UpdateNominativeUserRequestDgraphSuccess
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestDgraphSuccess(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestDgraphSuccess", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestDgraphSuccess), arg0, arg1)
}

// UpdateNominativeUserRequestPostgresSuccess mocks base method
func (m *MockImport) UpdateNominativeUserRequestPostgresSuccess(arg0 context.Context, arg1 db.UpdateNominativeUserRequestPostgresSuccessParams) (db.UpdateNominativeUserRequestPostgresSuccessRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestPostgresSuccess", arg0, arg1)
	ret0, _ := ret[0].(db.UpdateNominativeUserRequestPostgresSuccessRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNominativeUserRequestPostgresSuccess indicates an expected call of UpdateNominativeUserRequestPostgresSuccess
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestPostgresSuccess(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestPostgresSuccess", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestPostgresSuccess), arg0, arg1)
}

// UpdateNominativeUserRequestSuccess mocks base method
func (m *MockImport) UpdateNominativeUserRequestSuccess(arg0 context.Context, arg1 db.UpdateNominativeUserRequestSuccessParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNominativeUserRequestSuccess", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNominativeUserRequestSuccess indicates an expected call of UpdateNominativeUserRequestSuccess
func (mr *MockImportMockRecorder) UpdateNominativeUserRequestSuccess(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNominativeUserRequestSuccess", reflect.TypeOf((*MockImport)(nil).UpdateNominativeUserRequestSuccess), arg0, arg1)
}
