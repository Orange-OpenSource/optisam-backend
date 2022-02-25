// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"context"
)

type Querier interface {
	AddApplicationbsolescenceRisk(ctx context.Context, arg AddApplicationbsolescenceRiskParams) error
	DeleteApplicationsByScope(ctx context.Context, scope string) error
	DeleteDomainCriticityByScope(ctx context.Context, scope string) error
	DeleteInstancesByScope(ctx context.Context, scope string) error
	DeleteMaintenanceCirticityByScope(ctx context.Context, scope string) error
	DeleteRiskMatricbyScope(ctx context.Context, scope string) error
	GetApplicationDomains(ctx context.Context, scope string) ([]string, error)
	GetApplicationInstance(ctx context.Context, instanceID string) (ApplicationsInstance, error)
	GetApplicationInstances(ctx context.Context, arg GetApplicationInstancesParams) ([]ApplicationsInstance, error)
	GetApplicationsByProduct(ctx context.Context, arg GetApplicationsByProductParams) ([]GetApplicationsByProductRow, error)
	GetApplicationsDetails(ctx context.Context) ([]GetApplicationsDetailsRow, error)
	GetApplicationsView(ctx context.Context, arg GetApplicationsViewParams) ([]GetApplicationsViewRow, error)
	GetDomainCriticity(ctx context.Context, scope string) ([]GetDomainCriticityRow, error)
	GetDomainCriticityByDomain(ctx context.Context, arg GetDomainCriticityByDomainParams) (DomainCriticityMetum, error)
	GetDomainCriticityMeta(ctx context.Context) ([]DomainCriticityMetum, error)
	GetDomainCriticityMetaIDs(ctx context.Context) ([]int32, error)
	GetEquipmentsByApplicationID(ctx context.Context, arg GetEquipmentsByApplicationIDParams) ([]string, error)
	GetInstanceViewEquipments(ctx context.Context, arg GetInstanceViewEquipmentsParams) ([]int64, error)
	GetInstancesView(ctx context.Context, arg GetInstancesViewParams) ([]GetInstancesViewRow, error)
	GetMaintenanceCricityMeta(ctx context.Context) ([]MaintenanceLevelMetum, error)
	GetMaintenanceCricityMetaIDs(ctx context.Context) ([]int32, error)
	GetMaintenanceLevelByMonth(ctx context.Context, arg GetMaintenanceLevelByMonthParams) (MaintenanceLevelMetum, error)
	GetMaintenanceLevelByMonthByName(ctx context.Context, levelname string) (MaintenanceLevelMetum, error)
	GetMaintenanceTimeCriticity(ctx context.Context, scope string) ([]MaintenanceTimeCriticity, error)
	GetObsolescenceRiskForApplication(ctx context.Context, arg GetObsolescenceRiskForApplicationParams) (string, error)
	GetRiskLevelMetaIDs(ctx context.Context) ([]int32, error)
	GetRiskMatrix(ctx context.Context) ([]RiskMatrix, error)
	GetRiskMatrixConfig(ctx context.Context, scope string) ([]GetRiskMatrixConfigRow, error)
	GetRiskMeta(ctx context.Context) ([]RiskMetum, error)
	InsertDomainCriticity(ctx context.Context, arg InsertDomainCriticityParams) error
	InsertMaintenanceTimeCriticity(ctx context.Context, arg InsertMaintenanceTimeCriticityParams) error
	InsertRiskMatrix(ctx context.Context, arg InsertRiskMatrixParams) (int32, error)
	InsertRiskMatrixConfig(ctx context.Context, arg InsertRiskMatrixConfigParams) error
	UpsertApplication(ctx context.Context, arg UpsertApplicationParams) error
	UpsertApplicationInstance(ctx context.Context, arg UpsertApplicationInstanceParams) error
}

var _ Querier = (*Queries)(nil)
