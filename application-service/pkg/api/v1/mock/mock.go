// Code generated by MockGen. DO NOT EDIT.
// Source: ../../pkg/api/v1/application_grpc.pb.go

// Package mock_v1 is a generated GoMock package.
package mock_v1

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/api/v1"
	grpc "google.golang.org/grpc"
)

// MockApplicationServiceClient is a mock of ApplicationServiceClient interface.
type MockApplicationServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationServiceClientMockRecorder
}

// MockApplicationServiceClientMockRecorder is the mock recorder for MockApplicationServiceClient.
type MockApplicationServiceClientMockRecorder struct {
	mock *MockApplicationServiceClient
}

// NewMockApplicationServiceClient creates a new mock instance.
func NewMockApplicationServiceClient(ctrl *gomock.Controller) *MockApplicationServiceClient {
	mock := &MockApplicationServiceClient{ctrl: ctrl}
	mock.recorder = &MockApplicationServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationServiceClient) EXPECT() *MockApplicationServiceClientMockRecorder {
	return m.recorder
}

// ApplicationDomains mocks base method.
func (m *MockApplicationServiceClient) ApplicationDomains(ctx context.Context, in *v1.ApplicationDomainsRequest, opts ...grpc.CallOption) (*v1.ApplicationDomainsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ApplicationDomains", varargs...)
	ret0, _ := ret[0].(*v1.ApplicationDomainsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ApplicationDomains indicates an expected call of ApplicationDomains.
func (mr *MockApplicationServiceClientMockRecorder) ApplicationDomains(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplicationDomains", reflect.TypeOf((*MockApplicationServiceClient)(nil).ApplicationDomains), varargs...)
}

// DeleteApplication mocks base method.
func (m *MockApplicationServiceClient) DeleteApplication(ctx context.Context, in *v1.DeleteApplicationRequest, opts ...grpc.CallOption) (*v1.DeleteApplicationResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteApplication", varargs...)
	ret0, _ := ret[0].(*v1.DeleteApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteApplication indicates an expected call of DeleteApplication.
func (mr *MockApplicationServiceClientMockRecorder) DeleteApplication(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplication", reflect.TypeOf((*MockApplicationServiceClient)(nil).DeleteApplication), varargs...)
}

// DeleteInstance mocks base method.
func (m *MockApplicationServiceClient) DeleteInstance(ctx context.Context, in *v1.DeleteInstanceRequest, opts ...grpc.CallOption) (*v1.DeleteInstanceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteInstance", varargs...)
	ret0, _ := ret[0].(*v1.DeleteInstanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteInstance indicates an expected call of DeleteInstance.
func (mr *MockApplicationServiceClientMockRecorder) DeleteInstance(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInstance", reflect.TypeOf((*MockApplicationServiceClient)(nil).DeleteInstance), varargs...)
}

// DropApplicationData mocks base method.
func (m *MockApplicationServiceClient) DropApplicationData(ctx context.Context, in *v1.DropApplicationDataRequest, opts ...grpc.CallOption) (*v1.DropApplicationDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DropApplicationData", varargs...)
	ret0, _ := ret[0].(*v1.DropApplicationDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropApplicationData indicates an expected call of DropApplicationData.
func (mr *MockApplicationServiceClientMockRecorder) DropApplicationData(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropApplicationData", reflect.TypeOf((*MockApplicationServiceClient)(nil).DropApplicationData), varargs...)
}

// DropObscolenscenceData mocks base method.
func (m *MockApplicationServiceClient) DropObscolenscenceData(ctx context.Context, in *v1.DropObscolenscenceDataRequest, opts ...grpc.CallOption) (*v1.DropObscolenscenceDataResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DropObscolenscenceData", varargs...)
	ret0, _ := ret[0].(*v1.DropObscolenscenceDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropObscolenscenceData indicates an expected call of DropObscolenscenceData.
func (mr *MockApplicationServiceClientMockRecorder) DropObscolenscenceData(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropObscolenscenceData", reflect.TypeOf((*MockApplicationServiceClient)(nil).DropObscolenscenceData), varargs...)
}

// GetEquipmentsByApplication mocks base method.
func (m *MockApplicationServiceClient) GetEquipmentsByApplication(ctx context.Context, in *v1.GetEquipmentsByApplicationRequest, opts ...grpc.CallOption) (*v1.GetEquipmentsByApplicationResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetEquipmentsByApplication", varargs...)
	ret0, _ := ret[0].(*v1.GetEquipmentsByApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipmentsByApplication indicates an expected call of GetEquipmentsByApplication.
func (mr *MockApplicationServiceClientMockRecorder) GetEquipmentsByApplication(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipmentsByApplication", reflect.TypeOf((*MockApplicationServiceClient)(nil).GetEquipmentsByApplication), varargs...)
}

// ListApplications mocks base method.
func (m *MockApplicationServiceClient) ListApplications(ctx context.Context, in *v1.ListApplicationsRequest, opts ...grpc.CallOption) (*v1.ListApplicationsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListApplications", varargs...)
	ret0, _ := ret[0].(*v1.ListApplicationsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications.
func (mr *MockApplicationServiceClientMockRecorder) ListApplications(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockApplicationServiceClient)(nil).ListApplications), varargs...)
}

// ListInstances mocks base method.
func (m *MockApplicationServiceClient) ListInstances(ctx context.Context, in *v1.ListInstancesRequest, opts ...grpc.CallOption) (*v1.ListInstancesResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ListInstances", varargs...)
	ret0, _ := ret[0].(*v1.ListInstancesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListInstances indicates an expected call of ListInstances.
func (mr *MockApplicationServiceClientMockRecorder) ListInstances(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListInstances", reflect.TypeOf((*MockApplicationServiceClient)(nil).ListInstances), varargs...)
}

// ObsolescenceDomainCriticity mocks base method.
func (m *MockApplicationServiceClient) ObsolescenceDomainCriticity(ctx context.Context, in *v1.DomainCriticityRequest, opts ...grpc.CallOption) (*v1.DomainCriticityResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenceDomainCriticity", varargs...)
	ret0, _ := ret[0].(*v1.DomainCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceDomainCriticity indicates an expected call of ObsolescenceDomainCriticity.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenceDomainCriticity(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceDomainCriticity", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenceDomainCriticity), varargs...)
}

// ObsolescenceDomainCriticityMeta mocks base method.
func (m *MockApplicationServiceClient) ObsolescenceDomainCriticityMeta(ctx context.Context, in *v1.DomainCriticityMetaRequest, opts ...grpc.CallOption) (*v1.DomainCriticityMetaResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenceDomainCriticityMeta", varargs...)
	ret0, _ := ret[0].(*v1.DomainCriticityMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceDomainCriticityMeta indicates an expected call of ObsolescenceDomainCriticityMeta.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenceDomainCriticityMeta(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceDomainCriticityMeta", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenceDomainCriticityMeta), varargs...)
}

// ObsolescenceMaintenanceCriticityMeta mocks base method.
func (m *MockApplicationServiceClient) ObsolescenceMaintenanceCriticityMeta(ctx context.Context, in *v1.MaintenanceCriticityMetaRequest, opts ...grpc.CallOption) (*v1.MaintenanceCriticityMetaResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenceMaintenanceCriticityMeta", varargs...)
	ret0, _ := ret[0].(*v1.MaintenanceCriticityMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceMaintenanceCriticityMeta indicates an expected call of ObsolescenceMaintenanceCriticityMeta.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenceMaintenanceCriticityMeta(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceMaintenanceCriticityMeta", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenceMaintenanceCriticityMeta), varargs...)
}

// ObsolescenceRiskMeta mocks base method.
func (m *MockApplicationServiceClient) ObsolescenceRiskMeta(ctx context.Context, in *v1.RiskMetaRequest, opts ...grpc.CallOption) (*v1.RiskMetaResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenceRiskMeta", varargs...)
	ret0, _ := ret[0].(*v1.RiskMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceRiskMeta indicates an expected call of ObsolescenceRiskMeta.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenceRiskMeta(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceRiskMeta", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenceRiskMeta), varargs...)
}

// ObsolescenseMaintenanceCriticity mocks base method.
func (m *MockApplicationServiceClient) ObsolescenseMaintenanceCriticity(ctx context.Context, in *v1.MaintenanceCriticityRequest, opts ...grpc.CallOption) (*v1.MaintenanceCriticityResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenseMaintenanceCriticity", varargs...)
	ret0, _ := ret[0].(*v1.MaintenanceCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenseMaintenanceCriticity indicates an expected call of ObsolescenseMaintenanceCriticity.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenseMaintenanceCriticity(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenseMaintenanceCriticity", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenseMaintenanceCriticity), varargs...)
}

// ObsolescenseRiskMatrix mocks base method.
func (m *MockApplicationServiceClient) ObsolescenseRiskMatrix(ctx context.Context, in *v1.RiskMatrixRequest, opts ...grpc.CallOption) (*v1.RiskMatrixResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "ObsolescenseRiskMatrix", varargs...)
	ret0, _ := ret[0].(*v1.RiskMatrixResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenseRiskMatrix indicates an expected call of ObsolescenseRiskMatrix.
func (mr *MockApplicationServiceClientMockRecorder) ObsolescenseRiskMatrix(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenseRiskMatrix", reflect.TypeOf((*MockApplicationServiceClient)(nil).ObsolescenseRiskMatrix), varargs...)
}

// PostObsolescenceDomainCriticity mocks base method.
func (m *MockApplicationServiceClient) PostObsolescenceDomainCriticity(ctx context.Context, in *v1.PostDomainCriticityRequest, opts ...grpc.CallOption) (*v1.PostDomainCriticityResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PostObsolescenceDomainCriticity", varargs...)
	ret0, _ := ret[0].(*v1.PostDomainCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenceDomainCriticity indicates an expected call of PostObsolescenceDomainCriticity.
func (mr *MockApplicationServiceClientMockRecorder) PostObsolescenceDomainCriticity(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenceDomainCriticity", reflect.TypeOf((*MockApplicationServiceClient)(nil).PostObsolescenceDomainCriticity), varargs...)
}

// PostObsolescenseMaintenanceCriticity mocks base method.
func (m *MockApplicationServiceClient) PostObsolescenseMaintenanceCriticity(ctx context.Context, in *v1.PostMaintenanceCriticityRequest, opts ...grpc.CallOption) (*v1.PostMaintenanceCriticityResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PostObsolescenseMaintenanceCriticity", varargs...)
	ret0, _ := ret[0].(*v1.PostMaintenanceCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenseMaintenanceCriticity indicates an expected call of PostObsolescenseMaintenanceCriticity.
func (mr *MockApplicationServiceClientMockRecorder) PostObsolescenseMaintenanceCriticity(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenseMaintenanceCriticity", reflect.TypeOf((*MockApplicationServiceClient)(nil).PostObsolescenseMaintenanceCriticity), varargs...)
}

// PostObsolescenseRiskMatrix mocks base method.
func (m *MockApplicationServiceClient) PostObsolescenseRiskMatrix(ctx context.Context, in *v1.PostRiskMatrixRequest, opts ...grpc.CallOption) (*v1.PostRiskMatrixResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PostObsolescenseRiskMatrix", varargs...)
	ret0, _ := ret[0].(*v1.PostRiskMatrixResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenseRiskMatrix indicates an expected call of PostObsolescenseRiskMatrix.
func (mr *MockApplicationServiceClientMockRecorder) PostObsolescenseRiskMatrix(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenseRiskMatrix", reflect.TypeOf((*MockApplicationServiceClient)(nil).PostObsolescenseRiskMatrix), varargs...)
}

// UpsertApplication mocks base method.
func (m *MockApplicationServiceClient) UpsertApplication(ctx context.Context, in *v1.UpsertApplicationRequest, opts ...grpc.CallOption) (*v1.UpsertApplicationResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertApplication", varargs...)
	ret0, _ := ret[0].(*v1.UpsertApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertApplication indicates an expected call of UpsertApplication.
func (mr *MockApplicationServiceClientMockRecorder) UpsertApplication(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplication", reflect.TypeOf((*MockApplicationServiceClient)(nil).UpsertApplication), varargs...)
}

// UpsertApplicationEquip mocks base method.
func (m *MockApplicationServiceClient) UpsertApplicationEquip(ctx context.Context, in *v1.UpsertApplicationEquipRequest, opts ...grpc.CallOption) (*v1.UpsertApplicationEquipResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertApplicationEquip", varargs...)
	ret0, _ := ret[0].(*v1.UpsertApplicationEquipResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertApplicationEquip indicates an expected call of UpsertApplicationEquip.
func (mr *MockApplicationServiceClientMockRecorder) UpsertApplicationEquip(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationEquip", reflect.TypeOf((*MockApplicationServiceClient)(nil).UpsertApplicationEquip), varargs...)
}

// UpsertInstance mocks base method.
func (m *MockApplicationServiceClient) UpsertInstance(ctx context.Context, in *v1.UpsertInstanceRequest, opts ...grpc.CallOption) (*v1.UpsertInstanceResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpsertInstance", varargs...)
	ret0, _ := ret[0].(*v1.UpsertInstanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertInstance indicates an expected call of UpsertInstance.
func (mr *MockApplicationServiceClientMockRecorder) UpsertInstance(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertInstance", reflect.TypeOf((*MockApplicationServiceClient)(nil).UpsertInstance), varargs...)
}

// MockApplicationServiceServer is a mock of ApplicationServiceServer interface.
type MockApplicationServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationServiceServerMockRecorder
}

// MockApplicationServiceServerMockRecorder is the mock recorder for MockApplicationServiceServer.
type MockApplicationServiceServerMockRecorder struct {
	mock *MockApplicationServiceServer
}

// NewMockApplicationServiceServer creates a new mock instance.
func NewMockApplicationServiceServer(ctrl *gomock.Controller) *MockApplicationServiceServer {
	mock := &MockApplicationServiceServer{ctrl: ctrl}
	mock.recorder = &MockApplicationServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationServiceServer) EXPECT() *MockApplicationServiceServerMockRecorder {
	return m.recorder
}

// ApplicationDomains mocks base method.
func (m *MockApplicationServiceServer) ApplicationDomains(arg0 context.Context, arg1 *v1.ApplicationDomainsRequest) (*v1.ApplicationDomainsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplicationDomains", arg0, arg1)
	ret0, _ := ret[0].(*v1.ApplicationDomainsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ApplicationDomains indicates an expected call of ApplicationDomains.
func (mr *MockApplicationServiceServerMockRecorder) ApplicationDomains(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplicationDomains", reflect.TypeOf((*MockApplicationServiceServer)(nil).ApplicationDomains), arg0, arg1)
}

// DeleteApplication mocks base method.
func (m *MockApplicationServiceServer) DeleteApplication(arg0 context.Context, arg1 *v1.DeleteApplicationRequest) (*v1.DeleteApplicationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApplication", arg0, arg1)
	ret0, _ := ret[0].(*v1.DeleteApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteApplication indicates an expected call of DeleteApplication.
func (mr *MockApplicationServiceServerMockRecorder) DeleteApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplication", reflect.TypeOf((*MockApplicationServiceServer)(nil).DeleteApplication), arg0, arg1)
}

// DeleteInstance mocks base method.
func (m *MockApplicationServiceServer) DeleteInstance(arg0 context.Context, arg1 *v1.DeleteInstanceRequest) (*v1.DeleteInstanceResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteInstance", arg0, arg1)
	ret0, _ := ret[0].(*v1.DeleteInstanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteInstance indicates an expected call of DeleteInstance.
func (mr *MockApplicationServiceServerMockRecorder) DeleteInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInstance", reflect.TypeOf((*MockApplicationServiceServer)(nil).DeleteInstance), arg0, arg1)
}

// DropApplicationData mocks base method.
func (m *MockApplicationServiceServer) DropApplicationData(arg0 context.Context, arg1 *v1.DropApplicationDataRequest) (*v1.DropApplicationDataResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropApplicationData", arg0, arg1)
	ret0, _ := ret[0].(*v1.DropApplicationDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropApplicationData indicates an expected call of DropApplicationData.
func (mr *MockApplicationServiceServerMockRecorder) DropApplicationData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropApplicationData", reflect.TypeOf((*MockApplicationServiceServer)(nil).DropApplicationData), arg0, arg1)
}

// DropObscolenscenceData mocks base method.
func (m *MockApplicationServiceServer) DropObscolenscenceData(arg0 context.Context, arg1 *v1.DropObscolenscenceDataRequest) (*v1.DropObscolenscenceDataResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropObscolenscenceData", arg0, arg1)
	ret0, _ := ret[0].(*v1.DropObscolenscenceDataResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DropObscolenscenceData indicates an expected call of DropObscolenscenceData.
func (mr *MockApplicationServiceServerMockRecorder) DropObscolenscenceData(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropObscolenscenceData", reflect.TypeOf((*MockApplicationServiceServer)(nil).DropObscolenscenceData), arg0, arg1)
}

// GetEquipmentsByApplication mocks base method.
func (m *MockApplicationServiceServer) GetEquipmentsByApplication(arg0 context.Context, arg1 *v1.GetEquipmentsByApplicationRequest) (*v1.GetEquipmentsByApplicationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEquipmentsByApplication", arg0, arg1)
	ret0, _ := ret[0].(*v1.GetEquipmentsByApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipmentsByApplication indicates an expected call of GetEquipmentsByApplication.
func (mr *MockApplicationServiceServerMockRecorder) GetEquipmentsByApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipmentsByApplication", reflect.TypeOf((*MockApplicationServiceServer)(nil).GetEquipmentsByApplication), arg0, arg1)
}

// ListApplications mocks base method.
func (m *MockApplicationServiceServer) ListApplications(arg0 context.Context, arg1 *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListApplications", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListApplicationsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListApplications indicates an expected call of ListApplications.
func (mr *MockApplicationServiceServerMockRecorder) ListApplications(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListApplications", reflect.TypeOf((*MockApplicationServiceServer)(nil).ListApplications), arg0, arg1)
}

// ListInstances mocks base method.
func (m *MockApplicationServiceServer) ListInstances(arg0 context.Context, arg1 *v1.ListInstancesRequest) (*v1.ListInstancesResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListInstances", arg0, arg1)
	ret0, _ := ret[0].(*v1.ListInstancesResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListInstances indicates an expected call of ListInstances.
func (mr *MockApplicationServiceServerMockRecorder) ListInstances(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListInstances", reflect.TypeOf((*MockApplicationServiceServer)(nil).ListInstances), arg0, arg1)
}

// ObsolescenceDomainCriticity mocks base method.
func (m *MockApplicationServiceServer) ObsolescenceDomainCriticity(arg0 context.Context, arg1 *v1.DomainCriticityRequest) (*v1.DomainCriticityResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenceDomainCriticity", arg0, arg1)
	ret0, _ := ret[0].(*v1.DomainCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceDomainCriticity indicates an expected call of ObsolescenceDomainCriticity.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenceDomainCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceDomainCriticity", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenceDomainCriticity), arg0, arg1)
}

// ObsolescenceDomainCriticityMeta mocks base method.
func (m *MockApplicationServiceServer) ObsolescenceDomainCriticityMeta(arg0 context.Context, arg1 *v1.DomainCriticityMetaRequest) (*v1.DomainCriticityMetaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenceDomainCriticityMeta", arg0, arg1)
	ret0, _ := ret[0].(*v1.DomainCriticityMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceDomainCriticityMeta indicates an expected call of ObsolescenceDomainCriticityMeta.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenceDomainCriticityMeta(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceDomainCriticityMeta", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenceDomainCriticityMeta), arg0, arg1)
}

// ObsolescenceMaintenanceCriticityMeta mocks base method.
func (m *MockApplicationServiceServer) ObsolescenceMaintenanceCriticityMeta(arg0 context.Context, arg1 *v1.MaintenanceCriticityMetaRequest) (*v1.MaintenanceCriticityMetaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenceMaintenanceCriticityMeta", arg0, arg1)
	ret0, _ := ret[0].(*v1.MaintenanceCriticityMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceMaintenanceCriticityMeta indicates an expected call of ObsolescenceMaintenanceCriticityMeta.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenceMaintenanceCriticityMeta(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceMaintenanceCriticityMeta", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenceMaintenanceCriticityMeta), arg0, arg1)
}

// ObsolescenceRiskMeta mocks base method.
func (m *MockApplicationServiceServer) ObsolescenceRiskMeta(arg0 context.Context, arg1 *v1.RiskMetaRequest) (*v1.RiskMetaResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenceRiskMeta", arg0, arg1)
	ret0, _ := ret[0].(*v1.RiskMetaResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenceRiskMeta indicates an expected call of ObsolescenceRiskMeta.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenceRiskMeta(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenceRiskMeta", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenceRiskMeta), arg0, arg1)
}

// ObsolescenseMaintenanceCriticity mocks base method.
func (m *MockApplicationServiceServer) ObsolescenseMaintenanceCriticity(arg0 context.Context, arg1 *v1.MaintenanceCriticityRequest) (*v1.MaintenanceCriticityResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenseMaintenanceCriticity", arg0, arg1)
	ret0, _ := ret[0].(*v1.MaintenanceCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenseMaintenanceCriticity indicates an expected call of ObsolescenseMaintenanceCriticity.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenseMaintenanceCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenseMaintenanceCriticity", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenseMaintenanceCriticity), arg0, arg1)
}

// ObsolescenseRiskMatrix mocks base method.
func (m *MockApplicationServiceServer) ObsolescenseRiskMatrix(arg0 context.Context, arg1 *v1.RiskMatrixRequest) (*v1.RiskMatrixResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ObsolescenseRiskMatrix", arg0, arg1)
	ret0, _ := ret[0].(*v1.RiskMatrixResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ObsolescenseRiskMatrix indicates an expected call of ObsolescenseRiskMatrix.
func (mr *MockApplicationServiceServerMockRecorder) ObsolescenseRiskMatrix(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ObsolescenseRiskMatrix", reflect.TypeOf((*MockApplicationServiceServer)(nil).ObsolescenseRiskMatrix), arg0, arg1)
}

// PostObsolescenceDomainCriticity mocks base method.
func (m *MockApplicationServiceServer) PostObsolescenceDomainCriticity(arg0 context.Context, arg1 *v1.PostDomainCriticityRequest) (*v1.PostDomainCriticityResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostObsolescenceDomainCriticity", arg0, arg1)
	ret0, _ := ret[0].(*v1.PostDomainCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenceDomainCriticity indicates an expected call of PostObsolescenceDomainCriticity.
func (mr *MockApplicationServiceServerMockRecorder) PostObsolescenceDomainCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenceDomainCriticity", reflect.TypeOf((*MockApplicationServiceServer)(nil).PostObsolescenceDomainCriticity), arg0, arg1)
}

// PostObsolescenseMaintenanceCriticity mocks base method.
func (m *MockApplicationServiceServer) PostObsolescenseMaintenanceCriticity(arg0 context.Context, arg1 *v1.PostMaintenanceCriticityRequest) (*v1.PostMaintenanceCriticityResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostObsolescenseMaintenanceCriticity", arg0, arg1)
	ret0, _ := ret[0].(*v1.PostMaintenanceCriticityResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenseMaintenanceCriticity indicates an expected call of PostObsolescenseMaintenanceCriticity.
func (mr *MockApplicationServiceServerMockRecorder) PostObsolescenseMaintenanceCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenseMaintenanceCriticity", reflect.TypeOf((*MockApplicationServiceServer)(nil).PostObsolescenseMaintenanceCriticity), arg0, arg1)
}

// PostObsolescenseRiskMatrix mocks base method.
func (m *MockApplicationServiceServer) PostObsolescenseRiskMatrix(arg0 context.Context, arg1 *v1.PostRiskMatrixRequest) (*v1.PostRiskMatrixResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostObsolescenseRiskMatrix", arg0, arg1)
	ret0, _ := ret[0].(*v1.PostRiskMatrixResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PostObsolescenseRiskMatrix indicates an expected call of PostObsolescenseRiskMatrix.
func (mr *MockApplicationServiceServerMockRecorder) PostObsolescenseRiskMatrix(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostObsolescenseRiskMatrix", reflect.TypeOf((*MockApplicationServiceServer)(nil).PostObsolescenseRiskMatrix), arg0, arg1)
}

// UpsertApplication mocks base method.
func (m *MockApplicationServiceServer) UpsertApplication(arg0 context.Context, arg1 *v1.UpsertApplicationRequest) (*v1.UpsertApplicationResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplication", arg0, arg1)
	ret0, _ := ret[0].(*v1.UpsertApplicationResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertApplication indicates an expected call of UpsertApplication.
func (mr *MockApplicationServiceServerMockRecorder) UpsertApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplication", reflect.TypeOf((*MockApplicationServiceServer)(nil).UpsertApplication), arg0, arg1)
}

// UpsertApplicationEquip mocks base method.
func (m *MockApplicationServiceServer) UpsertApplicationEquip(arg0 context.Context, arg1 *v1.UpsertApplicationEquipRequest) (*v1.UpsertApplicationEquipResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplicationEquip", arg0, arg1)
	ret0, _ := ret[0].(*v1.UpsertApplicationEquipResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertApplicationEquip indicates an expected call of UpsertApplicationEquip.
func (mr *MockApplicationServiceServerMockRecorder) UpsertApplicationEquip(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationEquip", reflect.TypeOf((*MockApplicationServiceServer)(nil).UpsertApplicationEquip), arg0, arg1)
}

// UpsertInstance mocks base method.
func (m *MockApplicationServiceServer) UpsertInstance(arg0 context.Context, arg1 *v1.UpsertInstanceRequest) (*v1.UpsertInstanceResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertInstance", arg0, arg1)
	ret0, _ := ret[0].(*v1.UpsertInstanceResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpsertInstance indicates an expected call of UpsertInstance.
func (mr *MockApplicationServiceServerMockRecorder) UpsertInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertInstance", reflect.TypeOf((*MockApplicationServiceServer)(nil).UpsertInstance), arg0, arg1)
}

// MockUnsafeApplicationServiceServer is a mock of UnsafeApplicationServiceServer interface.
type MockUnsafeApplicationServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeApplicationServiceServerMockRecorder
}

// MockUnsafeApplicationServiceServerMockRecorder is the mock recorder for MockUnsafeApplicationServiceServer.
type MockUnsafeApplicationServiceServerMockRecorder struct {
	mock *MockUnsafeApplicationServiceServer
}

// NewMockUnsafeApplicationServiceServer creates a new mock instance.
func NewMockUnsafeApplicationServiceServer(ctrl *gomock.Controller) *MockUnsafeApplicationServiceServer {
	mock := &MockUnsafeApplicationServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeApplicationServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeApplicationServiceServer) EXPECT() *MockUnsafeApplicationServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedApplicationServiceServer mocks base method.
func (m *MockUnsafeApplicationServiceServer) mustEmbedUnimplementedApplicationServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedApplicationServiceServer")
}

// mustEmbedUnimplementedApplicationServiceServer indicates an expected call of mustEmbedUnimplementedApplicationServiceServer.
func (mr *MockUnsafeApplicationServiceServerMockRecorder) mustEmbedUnimplementedApplicationServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedApplicationServiceServer", reflect.TypeOf((*MockUnsafeApplicationServiceServer)(nil).mustEmbedUnimplementedApplicationServiceServer))
}
