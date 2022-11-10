// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type JobStatus string

const (
	JobStatusPENDING   JobStatus = "PENDING"
	JobStatusCOMPLETED JobStatus = "COMPLETED"
	JobStatusFAILED    JobStatus = "FAILED"
	JobStatusRETRY     JobStatus = "RETRY"
	JobStatusRUNNING   JobStatus = "RUNNING"
)

func (e *JobStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = JobStatus(s)
	case string:
		*e = JobStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for JobStatus: %T", src)
	}
	return nil
}

type Application struct {
	ApplicationID          string         `json:"application_id"`
	ApplicationName        string         `json:"application_name"`
	ApplicationVersion     string         `json:"application_version"`
	ApplicationOwner       string         `json:"application_owner"`
	ApplicationEnvironment string         `json:"application_environment"`
	ApplicationDomain      string         `json:"application_domain"`
	Scope                  string         `json:"scope"`
	ObsolescenceRisk       sql.NullString `json:"obsolescence_risk"`
	CreatedOn              time.Time      `json:"created_on"`
}

type ApplicationsEquipment struct {
	ApplicationID string `json:"application_id"`
	EquipmentID   string `json:"equipment_id"`
	Scope         string `json:"scope"`
}

type ApplicationsInstance struct {
	ApplicationID       string   `json:"application_id"`
	InstanceID          string   `json:"instance_id"`
	InstanceEnvironment string   `json:"instance_environment"`
	Products            []string `json:"products"`
	Equipments          []string `json:"equipments"`
	Scope               string   `json:"scope"`
}

type DomainCriticity struct {
	CriticID       int32        `json:"critic_id"`
	Scope          string       `json:"scope"`
	DomainCriticID int32        `json:"domain_critic_id"`
	Domains        []string     `json:"domains"`
	CreatedBy      string       `json:"created_by"`
	CreatedOn      sql.NullTime `json:"created_on"`
}

type DomainCriticityMetum struct {
	DomainCriticID   int32  `json:"domain_critic_id"`
	DomainCriticName string `json:"domain_critic_name"`
}

type Job struct {
	JobID      int32           `json:"job_id"`
	Type       string          `json:"type"`
	Status     JobStatus       `json:"status"`
	Data       json.RawMessage `json:"data"`
	Comments   sql.NullString  `json:"comments"`
	StartTime  sql.NullTime    `json:"start_time"`
	EndTime    sql.NullTime    `json:"end_time"`
	CreatedAt  time.Time       `json:"created_at"`
	RetryCount sql.NullInt32   `json:"retry_count"`
	MetaData   json.RawMessage `json:"meta_data"`
}

type MaintenanceLevelMetum struct {
	MaintenanceLevelID   int32  `json:"maintenance_level_id"`
	MaintenanceLevelName string `json:"maintenance_level_name"`
}

type MaintenanceTimeCriticity struct {
	MaintenanceCriticID int32        `json:"maintenance_critic_id"`
	Scope               string       `json:"scope"`
	LevelID             int32        `json:"level_id"`
	StartMonth          int32        `json:"start_month"`
	EndMonth            int32        `json:"end_month"`
	CreatedBy           string       `json:"created_by"`
	CreatedOn           sql.NullTime `json:"created_on"`
}

type RiskMatrix struct {
	ConfigurationID int32     `json:"configuration_id"`
	Scope           string    `json:"scope"`
	CreatedBy       string    `json:"created_by"`
	CreatedOn       time.Time `json:"created_on"`
}

type RiskMatrixConfig struct {
	ConfigurationID    int32 `json:"configuration_id"`
	DomainCriticID     int32 `json:"domain_critic_id"`
	MaintenanceLevelID int32 `json:"maintenance_level_id"`
	RiskID             int32 `json:"risk_id"`
}

type RiskMetum struct {
	RiskID   int32  `json:"risk_id"`
	RiskName string `json:"risk_name"`
}
