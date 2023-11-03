// Code generated by MockGen. DO NOT EDIT.
// Source: ../../thirdparty/license-service/pkg/api/v1/license_grpc.pb.go

// Package mock_v1 is a generated GoMock package.
package mock_v1

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/license-service/pkg/api/v1"
	grpc "google.golang.org/grpc"
	reflect "reflect"
)

// MockLicenseServiceClient is a mock of LicenseServiceClient interface
type MockLicenseServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockLicenseServiceClientMockRecorder
}

// MockLicenseServiceClientMockRecorder is the mock recorder for MockLicenseServiceClient
type MockLicenseServiceClientMockRecorder struct {
	mock *MockLicenseServiceClient
}

// NewMockLicenseServiceClient creates a new mock instance
func NewMockLicenseServiceClient(ctrl *gomock.Controller) *MockLicenseServiceClient {
	mock := &MockLicenseServiceClient{ctrl: ctrl}
	mock.recorder = &MockLicenseServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLicenseServiceClient) EXPECT() *MockLicenseServiceClientMockRecorder {
	return m.recorder
}

// GetOverAllCompliance mocks base method
func (m *MockLicenseServiceClient) GetOverAllCompliance(ctx context.Context, in *v1.GetOverAllComplianceRequest, opts ...grpc.CallOption) (*v1.GetOverAllComplianceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetOverAllCompliance", varargs...)
	ret0, _ := ret[0].(*v1.GetOverAllComplianceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOverAllCompliance indicates an expected call of GetOverAllCompliance
func (mr *MockLicenseServiceClientMockRecorder) GetOverAllCompliance(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOverAllCompliance", reflect.TypeOf((*MockLicenseServiceClient)(nil).GetOverAllCompliance), varargs...)
}

// ListAcqRightsForProduct mocks base method
func (m *MockLicenseServiceClient) ListAcqRightsForProduct(ctx context.Context, in *v1.ListAcquiredRightsForProductRequest, opts ...grpc.CallOption) (*v1.ListAcquiredRightsForProductResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListAcqRightsForProduct", varargs...)
	ret0, _ := ret[0].(*v1.ListAcquiredRightsForProductResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForProduct indicates an expected call of ListAcqRightsForProduct
func (mr *MockLicenseServiceClientMockRecorder) ListAcqRightsForProduct(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForProduct", reflect.TypeOf((*MockLicenseServiceClient)(nil).ListAcqRightsForProduct), varargs...)
}

// ListAcqRightsForApplicationsProduct mocks base method
func (m *MockLicenseServiceClient) ListAcqRightsForApplicationsProduct(ctx context.Context, in *v1.ListAcqRightsForApplicationsProductRequest, opts ...grpc.CallOption) (*v1.ListAcqRightsForApplicationsProductResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListAcqRightsForApplicationsProduct", varargs...)
	ret0, _ := ret[0].(*v1.ListAcqRightsForApplicationsProductResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForApplicationsProduct indicates an expected call of ListAcqRightsForApplicationsProduct
func (mr *MockLicenseServiceClientMockRecorder) ListAcqRightsForApplicationsProduct(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForApplicationsProduct", reflect.TypeOf((*MockLicenseServiceClient)(nil).ListAcqRightsForApplicationsProduct), varargs...)
}

// ListComputationDetails mocks base method
func (m *MockLicenseServiceClient) ListComputationDetails(ctx context.Context, in *v1.ListComputationDetailsRequest, opts ...grpc.CallOption) (*v1.ListComputationDetailsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListComputationDetails", varargs...)
	ret0, _ := ret[0].(*v1.ListComputationDetailsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListComputationDetails indicates an expected call of ListComputationDetails
func (mr *MockLicenseServiceClientMockRecorder) ListComputationDetails(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListComputationDetails", reflect.TypeOf((*MockLicenseServiceClient)(nil).ListComputationDetails), varargs...)
}

// ListAcqRightsForAggregation mocks base method
func (m *MockLicenseServiceClient) ListAcqRightsForAggregation(ctx context.Context, in *v1.ListAcqRightsForAggregationRequest, opts ...grpc.CallOption) (*v1.ListAcqRightsForAggregationResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListAcqRightsForAggregation", varargs...)
	ret0, _ := ret[0].(*v1.ListAcqRightsForAggregationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForAggregation indicates an expected call of ListAcqRightsForAggregation
func (mr *MockLicenseServiceClientMockRecorder) ListAcqRightsForAggregation(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForAggregation", reflect.TypeOf((*MockLicenseServiceClient)(nil).ListAcqRightsForAggregation), varargs...)
}

// ProductLicensesForMetric mocks base method
func (m *MockLicenseServiceClient) ProductLicensesForMetric(ctx context.Context, in *v1.ProductLicensesForMetricRequest, opts ...grpc.CallOption) (*v1.ProductLicensesForMetricResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ProductLicensesForMetric", varargs...)
	ret0, _ := ret[0].(*v1.ProductLicensesForMetricResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductLicensesForMetric indicates an expected call of ProductLicensesForMetric
func (mr *MockLicenseServiceClientMockRecorder) ProductLicensesForMetric(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductLicensesForMetric", reflect.TypeOf((*MockLicenseServiceClient)(nil).ProductLicensesForMetric), varargs...)
}

// LicensesForEquipAndMetric mocks base method
func (m *MockLicenseServiceClient) LicensesForEquipAndMetric(ctx context.Context, in *v1.LicensesForEquipAndMetricRequest, opts ...grpc.CallOption) (*v1.LicensesForEquipAndMetricResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LicensesForEquipAndMetric", varargs...)
	ret0, _ := ret[0].(*v1.LicensesForEquipAndMetricResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LicensesForEquipAndMetric indicates an expected call of LicensesForEquipAndMetric
func (mr *MockLicenseServiceClientMockRecorder) LicensesForEquipAndMetric(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LicensesForEquipAndMetric", reflect.TypeOf((*MockLicenseServiceClient)(nil).LicensesForEquipAndMetric), varargs...)
}

// MockLicenseServiceServer is a mock of LicenseServiceServer interface
type MockLicenseServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockLicenseServiceServerMockRecorder
}

// MockLicenseServiceServerMockRecorder is the mock recorder for MockLicenseServiceServer
type MockLicenseServiceServerMockRecorder struct {
	mock *MockLicenseServiceServer
}

// NewMockLicenseServiceServer creates a new mock instance
func NewMockLicenseServiceServer(ctrl *gomock.Controller) *MockLicenseServiceServer {
	mock := &MockLicenseServiceServer{ctrl: ctrl}
	mock.recorder = &MockLicenseServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLicenseServiceServer) EXPECT() *MockLicenseServiceServerMockRecorder {
	return m.recorder
}

// GetOverAllCompliance mocks base method
func (m *MockLicenseServiceServer) GetOverAllCompliance(arg0 context.Context, arg1 *v1.GetOverAllComplianceRequest) (*v1.GetOverAllComplianceResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOverAllCompliance", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetOverAllComplianceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOverAllCompliance indicates an expected call of GetOverAllCompliance
func (mr *MockLicenseServiceServerMockRecorder) GetOverAllCompliance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOverAllCompliance", reflect.TypeOf((*MockLicenseServiceServer)(nil).GetOverAllCompliance), arg0, arg1)
}

// ListAcqRightsForProduct mocks base method
func (m *MockLicenseServiceServer) ListAcqRightsForProduct(arg0 context.Context, arg1 *v1.ListAcquiredRightsForProductRequest) (*v1.ListAcquiredRightsForProductResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsForProduct", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListAcquiredRightsForProductResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForProduct indicates an expected call of ListAcqRightsForProduct
func (mr *MockLicenseServiceServerMockRecorder) ListAcqRightsForProduct(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForProduct", reflect.TypeOf((*MockLicenseServiceServer)(nil).ListAcqRightsForProduct), arg0, arg1)
}

// ListAcqRightsForApplicationsProduct mocks base method
func (m *MockLicenseServiceServer) ListAcqRightsForApplicationsProduct(arg0 context.Context, arg1 *v1.ListAcqRightsForApplicationsProductRequest) (*v1.ListAcqRightsForApplicationsProductResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsForApplicationsProduct", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListAcqRightsForApplicationsProductResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForApplicationsProduct indicates an expected call of ListAcqRightsForApplicationsProduct
func (mr *MockLicenseServiceServerMockRecorder) ListAcqRightsForApplicationsProduct(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForApplicationsProduct", reflect.TypeOf((*MockLicenseServiceServer)(nil).ListAcqRightsForApplicationsProduct), arg0, arg1)
}

// ListComputationDetails mocks base method
func (m *MockLicenseServiceServer) ListComputationDetails(arg0 context.Context, arg1 *v1.ListComputationDetailsRequest) (*v1.ListComputationDetailsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListComputationDetails", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListComputationDetailsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListComputationDetails indicates an expected call of ListComputationDetails
func (mr *MockLicenseServiceServerMockRecorder) ListComputationDetails(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListComputationDetails", reflect.TypeOf((*MockLicenseServiceServer)(nil).ListComputationDetails), arg0, arg1)
}

// ListAcqRightsForAggregation mocks base method
func (m *MockLicenseServiceServer) ListAcqRightsForAggregation(arg0 context.Context, arg1 *v1.ListAcqRightsForAggregationRequest) (*v1.ListAcqRightsForAggregationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAcqRightsForAggregation", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListAcqRightsForAggregationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListAcqRightsForAggregation indicates an expected call of ListAcqRightsForAggregation
func (mr *MockLicenseServiceServerMockRecorder) ListAcqRightsForAggregation(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAcqRightsForAggregation", reflect.TypeOf((*MockLicenseServiceServer)(nil).ListAcqRightsForAggregation), arg0, arg1)
}

// ProductLicensesForMetric mocks base method
func (m *MockLicenseServiceServer) ProductLicensesForMetric(arg0 context.Context, arg1 *v1.ProductLicensesForMetricRequest) (*v1.ProductLicensesForMetricResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProductLicensesForMetric", arg0, arg1)
	ret0, _ := ret[0].(*v1.ProductLicensesForMetricResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProductLicensesForMetric indicates an expected call of ProductLicensesForMetric
func (mr *MockLicenseServiceServerMockRecorder) ProductLicensesForMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProductLicensesForMetric", reflect.TypeOf((*MockLicenseServiceServer)(nil).ProductLicensesForMetric), arg0, arg1)
}

// LicensesForEquipAndMetric mocks base method
func (m *MockLicenseServiceServer) LicensesForEquipAndMetric(arg0 context.Context, arg1 *v1.LicensesForEquipAndMetricRequest) (*v1.LicensesForEquipAndMetricResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LicensesForEquipAndMetric", arg0, arg1)
	ret0, _ := ret[0].(*v1.LicensesForEquipAndMetricResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LicensesForEquipAndMetric indicates an expected call of LicensesForEquipAndMetric
func (mr *MockLicenseServiceServerMockRecorder) LicensesForEquipAndMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LicensesForEquipAndMetric", reflect.TypeOf((*MockLicenseServiceServer)(nil).LicensesForEquipAndMetric), arg0, arg1)
}

// MockUnsafeLicenseServiceServer is a mock of UnsafeLicenseServiceServer interface
type MockUnsafeLicenseServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeLicenseServiceServerMockRecorder
}

// MockUnsafeLicenseServiceServerMockRecorder is the mock recorder for MockUnsafeLicenseServiceServer
type MockUnsafeLicenseServiceServerMockRecorder struct {
	mock *MockUnsafeLicenseServiceServer
}

// NewMockUnsafeLicenseServiceServer creates a new mock instance
func NewMockUnsafeLicenseServiceServer(ctrl *gomock.Controller) *MockUnsafeLicenseServiceServer {
	mock := &MockUnsafeLicenseServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeLicenseServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUnsafeLicenseServiceServer) EXPECT() *MockUnsafeLicenseServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedLicenseServiceServer mocks base method
func (m *MockUnsafeLicenseServiceServer) mustEmbedUnimplementedLicenseServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedLicenseServiceServer")
}

// mustEmbedUnimplementedLicenseServiceServer indicates an expected call of mustEmbedUnimplementedLicenseServiceServer
func (mr *MockUnsafeLicenseServiceServerMockRecorder) mustEmbedUnimplementedLicenseServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedLicenseServiceServer", reflect.TypeOf((*MockUnsafeLicenseServiceServer)(nil).mustEmbedUnimplementedLicenseServiceServer))
}
