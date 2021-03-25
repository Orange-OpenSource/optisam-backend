// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type DataType string

const (
	DataTypeDATA       DataType = "DATA"
	DataTypeMETADATA   DataType = "METADATA"
	DataTypeGLOBALDATA DataType = "GLOBALDATA"
)

func (e *DataType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = DataType(s)
	case string:
		*e = DataType(s)
	default:
		return fmt.Errorf("unsupported scan type for DataType: %T", src)
	}
	return nil
}

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

type UploadStatus string

const (
	UploadStatusPENDING    UploadStatus = "PENDING"
	UploadStatusCOMPLETED  UploadStatus = "COMPLETED"
	UploadStatusFAILED     UploadStatus = "FAILED"
	UploadStatusINPROGRESS UploadStatus = "INPROGRESS"
)

func (e *UploadStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UploadStatus(s)
	case string:
		*e = UploadStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for UploadStatus: %T", src)
	}
	return nil
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
}

type UploadedDataFile struct {
	UploadID       int32          `json:"upload_id"`
	Scope          string         `json:"scope"`
	DataType       DataType       `json:"data_type"`
	FileName       string         `json:"file_name"`
	Status         UploadStatus   `json:"status"`
	UploadedBy     string         `json:"uploaded_by"`
	UploadedOn     time.Time      `json:"uploaded_on"`
	TotalRecords   int32          `json:"total_records"`
	SuccessRecords int32          `json:"success_records"`
	FailedRecords  int32          `json:"failed_records"`
	Comments       sql.NullString `json:"comments"`
}
