// Code generated by MockGen. DO NOT EDIT.
// Source: gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/repository/v1 (interfaces: Application)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/api/v1"
	db "gitlab.tech.orange/optisam/optisam-it/optisam-services/application-service/pkg/repository/v1/postgres/db"
	reflect "reflect"
)

// MockApplication is a mock of Application interface
type MockApplication struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationMockRecorder
}

// MockApplicationMockRecorder is the mock recorder for MockApplication
type MockApplicationMockRecorder struct {
	mock *MockApplication
}

// NewMockApplication creates a new mock instance
func NewMockApplication(ctrl *gomock.Controller) *MockApplication {
	mock := &MockApplication{ctrl: ctrl}
	mock.recorder = &MockApplicationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockApplication) EXPECT() *MockApplicationMockRecorder {
	return m.recorder
}

// AddApplicationbsolescenceRisk mocks base method
func (m *MockApplication) AddApplicationbsolescenceRisk(arg0 context.Context, arg1 db.AddApplicationbsolescenceRiskParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddApplicationbsolescenceRisk", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddApplicationbsolescenceRisk indicates an expected call of AddApplicationbsolescenceRisk
func (mr *MockApplicationMockRecorder) AddApplicationbsolescenceRisk(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddApplicationbsolescenceRisk", reflect.TypeOf((*MockApplication)(nil).AddApplicationbsolescenceRisk), arg0, arg1)
}

// DeleteApplicationEquip mocks base method
func (m *MockApplication) DeleteApplicationEquip(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApplicationEquip", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteApplicationEquip indicates an expected call of DeleteApplicationEquip
func (mr *MockApplicationMockRecorder) DeleteApplicationEquip(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplicationEquip", reflect.TypeOf((*MockApplication)(nil).DeleteApplicationEquip), arg0, arg1)
}

// DeleteApplicationsByScope mocks base method
func (m *MockApplication) DeleteApplicationsByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApplicationsByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteApplicationsByScope indicates an expected call of DeleteApplicationsByScope
func (mr *MockApplicationMockRecorder) DeleteApplicationsByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApplicationsByScope", reflect.TypeOf((*MockApplication)(nil).DeleteApplicationsByScope), arg0, arg1)
}

// DeleteDomainCriticityByScope mocks base method
func (m *MockApplication) DeleteDomainCriticityByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteDomainCriticityByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteDomainCriticityByScope indicates an expected call of DeleteDomainCriticityByScope
func (mr *MockApplicationMockRecorder) DeleteDomainCriticityByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteDomainCriticityByScope", reflect.TypeOf((*MockApplication)(nil).DeleteDomainCriticityByScope), arg0, arg1)
}

// DeleteInstancesByScope mocks base method
func (m *MockApplication) DeleteInstancesByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteInstancesByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteInstancesByScope indicates an expected call of DeleteInstancesByScope
func (mr *MockApplicationMockRecorder) DeleteInstancesByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteInstancesByScope", reflect.TypeOf((*MockApplication)(nil).DeleteInstancesByScope), arg0, arg1)
}

// DeleteMaintenanceCirticityByScope mocks base method
func (m *MockApplication) DeleteMaintenanceCirticityByScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMaintenanceCirticityByScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMaintenanceCirticityByScope indicates an expected call of DeleteMaintenanceCirticityByScope
func (mr *MockApplicationMockRecorder) DeleteMaintenanceCirticityByScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMaintenanceCirticityByScope", reflect.TypeOf((*MockApplication)(nil).DeleteMaintenanceCirticityByScope), arg0, arg1)
}

// DeleteRiskMatricbyScope mocks base method
func (m *MockApplication) DeleteRiskMatricbyScope(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteRiskMatricbyScope", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteRiskMatricbyScope indicates an expected call of DeleteRiskMatricbyScope
func (mr *MockApplicationMockRecorder) DeleteRiskMatricbyScope(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteRiskMatricbyScope", reflect.TypeOf((*MockApplication)(nil).DeleteRiskMatricbyScope), arg0, arg1)
}

// DropApplicationDataTX mocks base method
func (m *MockApplication) DropApplicationDataTX(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropApplicationDataTX", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DropApplicationDataTX indicates an expected call of DropApplicationDataTX
func (mr *MockApplicationMockRecorder) DropApplicationDataTX(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropApplicationDataTX", reflect.TypeOf((*MockApplication)(nil).DropApplicationDataTX), arg0, arg1)
}

// DropObscolenscenceDataTX mocks base method
func (m *MockApplication) DropObscolenscenceDataTX(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DropObscolenscenceDataTX", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DropObscolenscenceDataTX indicates an expected call of DropObscolenscenceDataTX
func (mr *MockApplicationMockRecorder) DropObscolenscenceDataTX(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DropObscolenscenceDataTX", reflect.TypeOf((*MockApplication)(nil).DropObscolenscenceDataTX), arg0, arg1)
}

// GetApplicationDomains mocks base method
func (m *MockApplication) GetApplicationDomains(arg0 context.Context, arg1 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationDomains", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationDomains indicates an expected call of GetApplicationDomains
func (mr *MockApplicationMockRecorder) GetApplicationDomains(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationDomains", reflect.TypeOf((*MockApplication)(nil).GetApplicationDomains), arg0, arg1)
}

// GetApplicationEquip mocks base method
func (m *MockApplication) GetApplicationEquip(arg0 context.Context, arg1 db.GetApplicationEquipParams) ([]db.GetApplicationEquipRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationEquip", arg0, arg1)
	ret0, _ := ret[0].([]db.GetApplicationEquipRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationEquip indicates an expected call of GetApplicationEquip
func (mr *MockApplicationMockRecorder) GetApplicationEquip(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationEquip", reflect.TypeOf((*MockApplication)(nil).GetApplicationEquip), arg0, arg1)
}

// GetApplicationInstance mocks base method
func (m *MockApplication) GetApplicationInstance(arg0 context.Context, arg1 db.GetApplicationInstanceParams) (db.ApplicationsInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationInstance", arg0, arg1)
	ret0, _ := ret[0].(db.ApplicationsInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationInstance indicates an expected call of GetApplicationInstance
func (mr *MockApplicationMockRecorder) GetApplicationInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationInstance", reflect.TypeOf((*MockApplication)(nil).GetApplicationInstance), arg0, arg1)
}

// GetApplicationInstances mocks base method
func (m *MockApplication) GetApplicationInstances(arg0 context.Context, arg1 db.GetApplicationInstancesParams) ([]db.ApplicationsInstance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationInstances", arg0, arg1)
	ret0, _ := ret[0].([]db.ApplicationsInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationInstances indicates an expected call of GetApplicationInstances
func (mr *MockApplicationMockRecorder) GetApplicationInstances(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationInstances", reflect.TypeOf((*MockApplication)(nil).GetApplicationInstances), arg0, arg1)
}

// GetApplicationsByProduct mocks base method
func (m *MockApplication) GetApplicationsByProduct(arg0 context.Context, arg1 db.GetApplicationsByProductParams) ([]db.GetApplicationsByProductRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationsByProduct", arg0, arg1)
	ret0, _ := ret[0].([]db.GetApplicationsByProductRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationsByProduct indicates an expected call of GetApplicationsByProduct
func (mr *MockApplicationMockRecorder) GetApplicationsByProduct(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationsByProduct", reflect.TypeOf((*MockApplication)(nil).GetApplicationsByProduct), arg0, arg1)
}

// GetApplicationsDetails mocks base method
func (m *MockApplication) GetApplicationsDetails(arg0 context.Context) ([]db.GetApplicationsDetailsRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationsDetails", arg0)
	ret0, _ := ret[0].([]db.GetApplicationsDetailsRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationsDetails indicates an expected call of GetApplicationsDetails
func (mr *MockApplicationMockRecorder) GetApplicationsDetails(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationsDetails", reflect.TypeOf((*MockApplication)(nil).GetApplicationsDetails), arg0)
}

// GetApplicationsView mocks base method
func (m *MockApplication) GetApplicationsView(arg0 context.Context, arg1 db.GetApplicationsViewParams) ([]db.GetApplicationsViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApplicationsView", arg0, arg1)
	ret0, _ := ret[0].([]db.GetApplicationsViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetApplicationsView indicates an expected call of GetApplicationsView
func (mr *MockApplicationMockRecorder) GetApplicationsView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApplicationsView", reflect.TypeOf((*MockApplication)(nil).GetApplicationsView), arg0, arg1)
}

// GetDomainCriticity mocks base method
func (m *MockApplication) GetDomainCriticity(arg0 context.Context, arg1 string) ([]db.GetDomainCriticityRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainCriticity", arg0, arg1)
	ret0, _ := ret[0].([]db.GetDomainCriticityRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomainCriticity indicates an expected call of GetDomainCriticity
func (mr *MockApplicationMockRecorder) GetDomainCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainCriticity", reflect.TypeOf((*MockApplication)(nil).GetDomainCriticity), arg0, arg1)
}

// GetDomainCriticityByDomain mocks base method
func (m *MockApplication) GetDomainCriticityByDomain(arg0 context.Context, arg1 db.GetDomainCriticityByDomainParams) (db.DomainCriticityMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainCriticityByDomain", arg0, arg1)
	ret0, _ := ret[0].(db.DomainCriticityMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomainCriticityByDomain indicates an expected call of GetDomainCriticityByDomain
func (mr *MockApplicationMockRecorder) GetDomainCriticityByDomain(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainCriticityByDomain", reflect.TypeOf((*MockApplication)(nil).GetDomainCriticityByDomain), arg0, arg1)
}

// GetDomainCriticityMeta mocks base method
func (m *MockApplication) GetDomainCriticityMeta(arg0 context.Context) ([]db.DomainCriticityMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainCriticityMeta", arg0)
	ret0, _ := ret[0].([]db.DomainCriticityMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomainCriticityMeta indicates an expected call of GetDomainCriticityMeta
func (mr *MockApplicationMockRecorder) GetDomainCriticityMeta(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainCriticityMeta", reflect.TypeOf((*MockApplication)(nil).GetDomainCriticityMeta), arg0)
}

// GetDomainCriticityMetaIDs mocks base method
func (m *MockApplication) GetDomainCriticityMetaIDs(arg0 context.Context) ([]int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDomainCriticityMetaIDs", arg0)
	ret0, _ := ret[0].([]int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDomainCriticityMetaIDs indicates an expected call of GetDomainCriticityMetaIDs
func (mr *MockApplicationMockRecorder) GetDomainCriticityMetaIDs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDomainCriticityMetaIDs", reflect.TypeOf((*MockApplication)(nil).GetDomainCriticityMetaIDs), arg0)
}

// GetEquipmentsByApplicationID mocks base method
func (m *MockApplication) GetEquipmentsByApplicationID(arg0 context.Context, arg1 db.GetEquipmentsByApplicationIDParams) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEquipmentsByApplicationID", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipmentsByApplicationID indicates an expected call of GetEquipmentsByApplicationID
func (mr *MockApplicationMockRecorder) GetEquipmentsByApplicationID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipmentsByApplicationID", reflect.TypeOf((*MockApplication)(nil).GetEquipmentsByApplicationID), arg0, arg1)
}

// GetInstanceViewEquipments mocks base method
func (m *MockApplication) GetInstanceViewEquipments(arg0 context.Context, arg1 db.GetInstanceViewEquipmentsParams) ([]int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInstanceViewEquipments", arg0, arg1)
	ret0, _ := ret[0].([]int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInstanceViewEquipments indicates an expected call of GetInstanceViewEquipments
func (mr *MockApplicationMockRecorder) GetInstanceViewEquipments(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInstanceViewEquipments", reflect.TypeOf((*MockApplication)(nil).GetInstanceViewEquipments), arg0, arg1)
}

// GetInstancesView mocks base method
func (m *MockApplication) GetInstancesView(arg0 context.Context, arg1 db.GetInstancesViewParams) ([]db.GetInstancesViewRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInstancesView", arg0, arg1)
	ret0, _ := ret[0].([]db.GetInstancesViewRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInstancesView indicates an expected call of GetInstancesView
func (mr *MockApplicationMockRecorder) GetInstancesView(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInstancesView", reflect.TypeOf((*MockApplication)(nil).GetInstancesView), arg0, arg1)
}

// GetMaintenanceCricityMeta mocks base method
func (m *MockApplication) GetMaintenanceCricityMeta(arg0 context.Context) ([]db.MaintenanceLevelMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaintenanceCricityMeta", arg0)
	ret0, _ := ret[0].([]db.MaintenanceLevelMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaintenanceCricityMeta indicates an expected call of GetMaintenanceCricityMeta
func (mr *MockApplicationMockRecorder) GetMaintenanceCricityMeta(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaintenanceCricityMeta", reflect.TypeOf((*MockApplication)(nil).GetMaintenanceCricityMeta), arg0)
}

// GetMaintenanceCricityMetaIDs mocks base method
func (m *MockApplication) GetMaintenanceCricityMetaIDs(arg0 context.Context) ([]int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaintenanceCricityMetaIDs", arg0)
	ret0, _ := ret[0].([]int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaintenanceCricityMetaIDs indicates an expected call of GetMaintenanceCricityMetaIDs
func (mr *MockApplicationMockRecorder) GetMaintenanceCricityMetaIDs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaintenanceCricityMetaIDs", reflect.TypeOf((*MockApplication)(nil).GetMaintenanceCricityMetaIDs), arg0)
}

// GetMaintenanceLevelByMonth mocks base method
func (m *MockApplication) GetMaintenanceLevelByMonth(arg0 context.Context, arg1 db.GetMaintenanceLevelByMonthParams) (db.MaintenanceLevelMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaintenanceLevelByMonth", arg0, arg1)
	ret0, _ := ret[0].(db.MaintenanceLevelMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaintenanceLevelByMonth indicates an expected call of GetMaintenanceLevelByMonth
func (mr *MockApplicationMockRecorder) GetMaintenanceLevelByMonth(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaintenanceLevelByMonth", reflect.TypeOf((*MockApplication)(nil).GetMaintenanceLevelByMonth), arg0, arg1)
}

// GetMaintenanceLevelByMonthByName mocks base method
func (m *MockApplication) GetMaintenanceLevelByMonthByName(arg0 context.Context, arg1 string) (db.MaintenanceLevelMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaintenanceLevelByMonthByName", arg0, arg1)
	ret0, _ := ret[0].(db.MaintenanceLevelMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaintenanceLevelByMonthByName indicates an expected call of GetMaintenanceLevelByMonthByName
func (mr *MockApplicationMockRecorder) GetMaintenanceLevelByMonthByName(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaintenanceLevelByMonthByName", reflect.TypeOf((*MockApplication)(nil).GetMaintenanceLevelByMonthByName), arg0, arg1)
}

// GetMaintenanceTimeCriticity mocks base method
func (m *MockApplication) GetMaintenanceTimeCriticity(arg0 context.Context, arg1 string) ([]db.MaintenanceTimeCriticity, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaintenanceTimeCriticity", arg0, arg1)
	ret0, _ := ret[0].([]db.MaintenanceTimeCriticity)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaintenanceTimeCriticity indicates an expected call of GetMaintenanceTimeCriticity
func (mr *MockApplicationMockRecorder) GetMaintenanceTimeCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaintenanceTimeCriticity", reflect.TypeOf((*MockApplication)(nil).GetMaintenanceTimeCriticity), arg0, arg1)
}

// GetObsolescenceRiskForApplication mocks base method
func (m *MockApplication) GetObsolescenceRiskForApplication(arg0 context.Context, arg1 db.GetObsolescenceRiskForApplicationParams) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetObsolescenceRiskForApplication", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetObsolescenceRiskForApplication indicates an expected call of GetObsolescenceRiskForApplication
func (mr *MockApplicationMockRecorder) GetObsolescenceRiskForApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetObsolescenceRiskForApplication", reflect.TypeOf((*MockApplication)(nil).GetObsolescenceRiskForApplication), arg0, arg1)
}

// GetProductsByApplicationID mocks base method
func (m *MockApplication) GetProductsByApplicationID(arg0 context.Context, arg1 db.GetProductsByApplicationIDParams) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProductsByApplicationID", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProductsByApplicationID indicates an expected call of GetProductsByApplicationID
func (mr *MockApplicationMockRecorder) GetProductsByApplicationID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProductsByApplicationID", reflect.TypeOf((*MockApplication)(nil).GetProductsByApplicationID), arg0, arg1)
}

// GetRiskLevelMetaIDs mocks base method
func (m *MockApplication) GetRiskLevelMetaIDs(arg0 context.Context) ([]int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRiskLevelMetaIDs", arg0)
	ret0, _ := ret[0].([]int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRiskLevelMetaIDs indicates an expected call of GetRiskLevelMetaIDs
func (mr *MockApplicationMockRecorder) GetRiskLevelMetaIDs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRiskLevelMetaIDs", reflect.TypeOf((*MockApplication)(nil).GetRiskLevelMetaIDs), arg0)
}

// GetRiskMatrix mocks base method
func (m *MockApplication) GetRiskMatrix(arg0 context.Context) ([]db.RiskMatrix, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRiskMatrix", arg0)
	ret0, _ := ret[0].([]db.RiskMatrix)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRiskMatrix indicates an expected call of GetRiskMatrix
func (mr *MockApplicationMockRecorder) GetRiskMatrix(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRiskMatrix", reflect.TypeOf((*MockApplication)(nil).GetRiskMatrix), arg0)
}

// GetRiskMatrixConfig mocks base method
func (m *MockApplication) GetRiskMatrixConfig(arg0 context.Context, arg1 string) ([]db.GetRiskMatrixConfigRow, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRiskMatrixConfig", arg0, arg1)
	ret0, _ := ret[0].([]db.GetRiskMatrixConfigRow)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRiskMatrixConfig indicates an expected call of GetRiskMatrixConfig
func (mr *MockApplicationMockRecorder) GetRiskMatrixConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRiskMatrixConfig", reflect.TypeOf((*MockApplication)(nil).GetRiskMatrixConfig), arg0, arg1)
}

// GetRiskMeta mocks base method
func (m *MockApplication) GetRiskMeta(arg0 context.Context) ([]db.RiskMetum, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRiskMeta", arg0)
	ret0, _ := ret[0].([]db.RiskMetum)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRiskMeta indicates an expected call of GetRiskMeta
func (mr *MockApplicationMockRecorder) GetRiskMeta(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRiskMeta", reflect.TypeOf((*MockApplication)(nil).GetRiskMeta), arg0)
}

// InsertDomainCriticity mocks base method
func (m *MockApplication) InsertDomainCriticity(arg0 context.Context, arg1 db.InsertDomainCriticityParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertDomainCriticity", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertDomainCriticity indicates an expected call of InsertDomainCriticity
func (mr *MockApplicationMockRecorder) InsertDomainCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertDomainCriticity", reflect.TypeOf((*MockApplication)(nil).InsertDomainCriticity), arg0, arg1)
}

// InsertMaintenanceTimeCriticity mocks base method
func (m *MockApplication) InsertMaintenanceTimeCriticity(arg0 context.Context, arg1 db.InsertMaintenanceTimeCriticityParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertMaintenanceTimeCriticity", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertMaintenanceTimeCriticity indicates an expected call of InsertMaintenanceTimeCriticity
func (mr *MockApplicationMockRecorder) InsertMaintenanceTimeCriticity(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertMaintenanceTimeCriticity", reflect.TypeOf((*MockApplication)(nil).InsertMaintenanceTimeCriticity), arg0, arg1)
}

// InsertRiskMatrix mocks base method
func (m *MockApplication) InsertRiskMatrix(arg0 context.Context, arg1 db.InsertRiskMatrixParams) (int32, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertRiskMatrix", arg0, arg1)
	ret0, _ := ret[0].(int32)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertRiskMatrix indicates an expected call of InsertRiskMatrix
func (mr *MockApplicationMockRecorder) InsertRiskMatrix(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertRiskMatrix", reflect.TypeOf((*MockApplication)(nil).InsertRiskMatrix), arg0, arg1)
}

// InsertRiskMatrixConfig mocks base method
func (m *MockApplication) InsertRiskMatrixConfig(arg0 context.Context, arg1 db.InsertRiskMatrixConfigParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertRiskMatrixConfig", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertRiskMatrixConfig indicates an expected call of InsertRiskMatrixConfig
func (mr *MockApplicationMockRecorder) InsertRiskMatrixConfig(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertRiskMatrixConfig", reflect.TypeOf((*MockApplication)(nil).InsertRiskMatrixConfig), arg0, arg1)
}

// UpsertApplication mocks base method
func (m *MockApplication) UpsertApplication(arg0 context.Context, arg1 db.UpsertApplicationParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplication", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplication indicates an expected call of UpsertApplication
func (mr *MockApplicationMockRecorder) UpsertApplication(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplication", reflect.TypeOf((*MockApplication)(nil).UpsertApplication), arg0, arg1)
}

// UpsertApplicationEquip mocks base method
func (m *MockApplication) UpsertApplicationEquip(arg0 context.Context, arg1 db.UpsertApplicationEquipParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplicationEquip", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplicationEquip indicates an expected call of UpsertApplicationEquip
func (mr *MockApplicationMockRecorder) UpsertApplicationEquip(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationEquip", reflect.TypeOf((*MockApplication)(nil).UpsertApplicationEquip), arg0, arg1)
}

// UpsertApplicationEquipTx mocks base method
func (m *MockApplication) UpsertApplicationEquipTx(arg0 context.Context, arg1 *v1.UpsertApplicationEquipRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplicationEquipTx", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplicationEquipTx indicates an expected call of UpsertApplicationEquipTx
func (mr *MockApplicationMockRecorder) UpsertApplicationEquipTx(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationEquipTx", reflect.TypeOf((*MockApplication)(nil).UpsertApplicationEquipTx), arg0, arg1)
}

// UpsertApplicationInstance mocks base method
func (m *MockApplication) UpsertApplicationInstance(arg0 context.Context, arg1 db.UpsertApplicationInstanceParams) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertApplicationInstance", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertApplicationInstance indicates an expected call of UpsertApplicationInstance
func (mr *MockApplicationMockRecorder) UpsertApplicationInstance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertApplicationInstance", reflect.TypeOf((*MockApplication)(nil).UpsertApplicationInstance), arg0, arg1)
}

// UpsertInstanceTX mocks base method
func (m *MockApplication) UpsertInstanceTX(arg0 context.Context, arg1 *v1.UpsertInstanceRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpsertInstanceTX", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpsertInstanceTX indicates an expected call of UpsertInstanceTX
func (mr *MockApplicationMockRecorder) UpsertInstanceTX(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpsertInstanceTX", reflect.TypeOf((*MockApplication)(nil).UpsertInstanceTX), arg0, arg1)
}
