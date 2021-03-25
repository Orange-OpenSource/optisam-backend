// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/common/optisam/workerqueue/repository (interfaces: Workerqueue)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	db "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
	reflect "reflect"
)

// MockWorkerqueue is a mock of Workerqueue interface
type MockWorkerqueue struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerqueueMockRecorder
}

// MockWorkerqueueMockRecorder is the mock recorder for MockWorkerqueue
type MockWorkerqueueMockRecorder struct {
	mock *MockWorkerqueue
}

// NewMockWorkerqueue creates a new mock instance
func NewMockWorkerqueue(ctrl *gomock.Controller) *MockWorkerqueue {
	mock := &MockWorkerqueue{ctrl: ctrl}
	mock.recorder = &MockWorkerqueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkerqueue) EXPECT() *MockWorkerqueueMockRecorder {
	return m.recorder
}

// CreateJob mocks base method
func (m *MockWorkerqueue) CreateJob(arg0 context.Context, arg1 db.CreateJobParams) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateJob", arg0, arg1)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateJob indicates an expected call of CreateJob
func (mr *MockWorkerqueueMockRecorder) CreateJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateJob", reflect.TypeOf((*MockWorkerqueue)(nil).CreateJob), arg0, arg1)
}

// GetJob mocks base method
func (m *MockWorkerqueue) GetJob(arg0 context.Context, arg1 int32) (db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJob", arg0, arg1)
	ret0, _ := ret[0].(db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJob indicates an expected call of GetJob
func (mr *MockWorkerqueueMockRecorder) GetJob(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJob", reflect.TypeOf((*MockWorkerqueue)(nil).GetJob), arg0, arg1)
}

// GetJobs mocks base method
func (m *MockWorkerqueue) GetJobs(arg0 context.Context) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobs", arg0)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobs indicates an expected call of GetJobs
func (mr *MockWorkerqueueMockRecorder) GetJobs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobs", reflect.TypeOf((*MockWorkerqueue)(nil).GetJobs), arg0)
}

// GetJobsForRetry mocks base method
func (m *MockWorkerqueue) GetJobsForRetry(arg0 context.Context) ([]db.Job, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetJobsForRetry", arg0)
	ret0, _ := ret[0].([]db.Job)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetJobsForRetry indicates an expected call of GetJobsForRetry
func (mr *MockWorkerqueueMockRecorder) GetJobsForRetry(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetJobsForRetry", reflect.TypeOf((*MockWorkerqueue)(nil).GetJobsForRetry), arg0)
}

// UpdateJobStatusCompleted mocks base method
func (m *MockWorkerqueue) UpdateJobStatusCompleted(arg0 context.Context, arg1 db.UpdateJobStatusCompletedParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobStatusCompleted", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobStatusCompleted indicates an expected call of UpdateJobStatusCompleted
func (mr *MockWorkerqueueMockRecorder) UpdateJobStatusCompleted(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobStatusCompleted", reflect.TypeOf((*MockWorkerqueue)(nil).UpdateJobStatusCompleted), arg0, arg1)
}

// UpdateJobStatusFailed mocks base method
func (m *MockWorkerqueue) UpdateJobStatusFailed(arg0 context.Context, arg1 db.UpdateJobStatusFailedParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobStatusFailed", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobStatusFailed indicates an expected call of UpdateJobStatusFailed
func (mr *MockWorkerqueueMockRecorder) UpdateJobStatusFailed(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobStatusFailed", reflect.TypeOf((*MockWorkerqueue)(nil).UpdateJobStatusFailed), arg0, arg1)
}

// UpdateJobStatusRetry mocks base method
func (m *MockWorkerqueue) UpdateJobStatusRetry(arg0 context.Context, arg1 db.UpdateJobStatusRetryParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobStatusRetry", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobStatusRetry indicates an expected call of UpdateJobStatusRetry
func (mr *MockWorkerqueueMockRecorder) UpdateJobStatusRetry(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobStatusRetry", reflect.TypeOf((*MockWorkerqueue)(nil).UpdateJobStatusRetry), arg0, arg1)
}

// UpdateJobStatusRunning mocks base method
func (m *MockWorkerqueue) UpdateJobStatusRunning(arg0 context.Context, arg1 db.UpdateJobStatusRunningParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateJobStatusRunning", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateJobStatusRunning indicates an expected call of UpdateJobStatusRunning
func (mr *MockWorkerqueueMockRecorder) UpdateJobStatusRunning(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateJobStatusRunning", reflect.TypeOf((*MockWorkerqueue)(nil).UpdateJobStatusRunning), arg0, arg1)
}
