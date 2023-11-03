package job

import (
	"database/sql"
	"encoding/json"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	dbgen "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/repository/postgres/db"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type JobStatus string // nolint: golint

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

// Job shows data model for queue
type Job struct {
	JobID      int32           `json:"job_id"`
	Type       sql.NullString  `json:"type"`
	Status     JobStatus       `json:"status"`
	Data       json.RawMessage `json:"data"`
	Comments   sql.NullString  `json:"comments"`
	StartTime  sql.NullTime    `json:"start_time"`
	EndTime    sql.NullTime    `json:"end_time"`
	CreatedAt  sql.NullTime    `json:"created_at"`
	RetryCount sql.NullInt32   `json:"retry_count"`
	MetaData   metadata.MD     `json:"meta_data"`
	PPID       string          `json:"pp_id"`
}

// ToRepoJob handles data modelling from queue job to repo job
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
		Ppid:       sql.NullString{String: j.PPID},
	}
}

// FromRepoJob handles data modelling from repo job to queue job
func FromRepoJob(j *dbgen.Job) *Job {
	md := metadata.MD{}
	err := json.Unmarshal(j.MetaData, &md)
	if err != nil {
		logger.Log.Error("Error unmarshling meta data %s", zap.Error(err))
	}
	return &Job{JobID: j.JobID,
		Type:       sql.NullString{String: j.Type},
		Comments:   j.Comments,
		Status:     JobStatus(j.Status),
		Data:       j.Data,
		CreatedAt:  sql.NullTime{Time: j.CreatedAt},
		StartTime:  j.StartTime,
		EndTime:    j.EndTime,
		RetryCount: j.RetryCount,
		MetaData:   md,
		PPID:       j.Ppid.String,
	}
}
