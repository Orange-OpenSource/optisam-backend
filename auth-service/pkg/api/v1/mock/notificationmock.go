// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1 (interfaces: NotificationServiceClient)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/auth-service/pkg/api/v1"
	reflect "reflect"
)

// MockNotificationServiceClient is a mock of NotificationServiceClient interface
type MockNotificationServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockNotificationServiceClientMockRecorder
}

// MockNotificationServiceClientMockRecorder is the mock recorder for MockNotificationServiceClient
type MockNotificationServiceClientMockRecorder struct {
	mock *MockNotificationServiceClient
}

// NewMockNotificationServiceClient creates a new mock instance
func NewMockNotificationServiceClient(ctrl *gomock.Controller) *MockNotificationServiceClient {
	mock := &MockNotificationServiceClient{ctrl: ctrl}
	mock.recorder = &MockNotificationServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNotificationServiceClient) EXPECT() *MockNotificationServiceClientMockRecorder {
	return m.recorder
}

// SendMail mocks base method
func (m *MockNotificationServiceClient) SendMail(arg0 context.Context, arg1 *v1.SendMailRequest, arg2 ...grpc.CallOption) (*v1.SendMailResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SendMail", varargs...)
	ret0, _ := ret[0].(*v1.SendMailResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendMail indicates an expected call of SendMail
func (mr *MockNotificationServiceClientMockRecorder) SendMail(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMail", reflect.TypeOf((*MockNotificationServiceClient)(nil).SendMail), varargs...)
}