// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package job

import (
	"database/sql"
	"encoding/json"
	dbgen "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
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
	*e = JobStatus(src.([]byte))
	return nil
}

//Job shows data model for queue
type Job struct {
	JobID      int32           `json:"job_id"`
	Type       sql.NullString  `json:"type"`
	Status     JobStatus       `json:"status"`
	Data       json.RawMessage `json:"data"` //this is byte data
	Comments   sql.NullString  `json:"comments"`
	StartTime  sql.NullTime    `json:"start_time"`
	EndTime    sql.NullTime    `json:"end_time"`
	CreatedAt  sql.NullTime    `json:"created_at"`
	RetryCount sql.NullInt32   `json:"retry_count"`
}

//ToRepoJob handles data modelling from queue job to repo job
func ToRepoJob(j *Job) *dbgen.Job {
	return &dbgen.Job{JobID: j.JobID,
		Type:       j.Type.String,
		Comments:   j.Comments,
		Status:     dbgen.JobStatus(j.Status),
		Data:       j.Data,
		CreatedAt:  j.CreatedAt.Time,
		StartTime:  j.StartTime,
		EndTime:    j.EndTime,
		RetryCount: j.RetryCount,
	}
}

//FromRepoJob handles data modelling from repo job to queue job
func FromRepoJob(j *dbgen.Job) *Job {
	return &Job{JobID: j.JobID,
		Type:       sql.NullString{String: j.Type},
		Comments:   j.Comments,
		Status:     JobStatus(j.Status),
		Data:       j.Data,
		CreatedAt:  sql.NullTime{Time: j.CreatedAt},
		StartTime:  j.StartTime,
		EndTime:    j.EndTime,
		RetryCount: j.RetryCount,
	}
}
