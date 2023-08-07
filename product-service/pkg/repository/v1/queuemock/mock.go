// Code generated by MockGen. DO NOT EDIT.
// Source: optisam-backend/common/optisam/workerqueue (interfaces: Workerqueue)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	workerqueue "optisam-backend/common/optisam/workerqueue"
	job "optisam-backend/common/optisam/workerqueue/job"
	worker "optisam-backend/common/optisam/workerqueue/worker"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockWorkerqueue is a mock of Workerqueue interface.
type MockWorkerqueue struct {
	ctrl     *gomock.Controller
	recorder *MockWorkerqueueMockRecorder
}

// MockWorkerqueueMockRecorder is the mock recorder for MockWorkerqueue.
type MockWorkerqueueMockRecorder struct {
	mock *MockWorkerqueue
}

// NewMockWorkerqueue creates a new mock instance.
func NewMockWorkerqueue(ctrl *gomock.Controller) *MockWorkerqueue {
	mock := &MockWorkerqueue{ctrl: ctrl}
	mock.recorder = &MockWorkerqueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWorkerqueue) EXPECT() *MockWorkerqueueMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockWorkerqueue) Close(arg0 context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close", arg0)
}

// Close indicates an expected call of Close.
func (mr *MockWorkerqueueMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockWorkerqueue)(nil).Close), arg0)
}

// GetCapacity mocks base method.
func (m *MockWorkerqueue) GetCapacity() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCapacity")
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetCapacity indicates an expected call of GetCapacity.
func (mr *MockWorkerqueueMockRecorder) GetCapacity() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCapacity", reflect.TypeOf((*MockWorkerqueue)(nil).GetCapacity))
}

// GetIthLength mocks base method.
func (m *MockWorkerqueue) GetIthLength(arg0 int) int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIthLength", arg0)
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetIthLength indicates an expected call of GetIthLength.
func (mr *MockWorkerqueueMockRecorder) GetIthLength(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIthLength", reflect.TypeOf((*MockWorkerqueue)(nil).GetIthLength), arg0)
}

// GetLength mocks base method.
func (m *MockWorkerqueue) GetLength() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLength")
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetLength indicates an expected call of GetLength.
func (mr *MockWorkerqueueMockRecorder) GetLength() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLength", reflect.TypeOf((*MockWorkerqueue)(nil).GetLength))
}

// GetRetries mocks base method.
func (m *MockWorkerqueue) GetRetries() int32 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRetries")
	ret0, _ := ret[0].(int32)
	return ret0
}

// GetRetries indicates an expected call of GetRetries.
func (mr *MockWorkerqueueMockRecorder) GetRetries() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRetries", reflect.TypeOf((*MockWorkerqueue)(nil).GetRetries))
}

// Grow mocks base method.
func (m *MockWorkerqueue) Grow() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Grow")
}

// Grow indicates an expected call of Grow.
func (mr *MockWorkerqueueMockRecorder) Grow() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Grow", reflect.TypeOf((*MockWorkerqueue)(nil).Grow))
}

// PopJob mocks base method.
func (m *MockWorkerqueue) PopJob() workerqueue.JobChan {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PopJob")
	ret0, _ := ret[0].(workerqueue.JobChan)
	return ret0
}

// PopJob indicates an expected call of PopJob.
func (mr *MockWorkerqueueMockRecorder) PopJob() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PopJob", reflect.TypeOf((*MockWorkerqueue)(nil).PopJob))
}

// PushJob mocks base method.
func (m *MockWorkerqueue) PushJob(arg0 context.Context, arg1 job.Job, arg2 string) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PushJob", arg0, arg1, arg2)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PushJob indicates an expected call of PushJob.
func (mr *MockWorkerqueueMockRecorder) PushJob(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushJob", reflect.TypeOf((*MockWorkerqueue)(nil).PushJob), arg0, arg1, arg2)
}

// RegisterWorker mocks base method.
func (m *MockWorkerqueue) RegisterWorker(arg0 context.Context, arg1 worker.Worker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterWorker", arg0, arg1)
}

// RegisterWorker indicates an expected call of RegisterWorker.
func (mr *MockWorkerqueueMockRecorder) RegisterWorker(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterWorker", reflect.TypeOf((*MockWorkerqueue)(nil).RegisterWorker), arg0, arg1)
}

// ResumePendingJobs mocks base method.
func (m *MockWorkerqueue) ResumePendingJobs(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResumePendingJobs", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResumePendingJobs indicates an expected call of ResumePendingJobs.
func (mr *MockWorkerqueueMockRecorder) ResumePendingJobs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResumePendingJobs", reflect.TypeOf((*MockWorkerqueue)(nil).ResumePendingJobs), arg0)
}

// Shrink mocks base method.
func (m *MockWorkerqueue) Shrink() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Shrink")
}

// Shrink indicates an expected call of Shrink.
func (mr *MockWorkerqueueMockRecorder) Shrink() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Shrink", reflect.TypeOf((*MockWorkerqueue)(nil).Shrink))
}
