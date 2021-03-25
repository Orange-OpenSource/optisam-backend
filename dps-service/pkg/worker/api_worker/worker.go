// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package apiworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"
	"time"

	"google.golang.org/grpc"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
	grpcServers map[string]*grpc.ClientConn
	t           time.Duration
}

//NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB, conn map[string]*grpc.ClientConn, t time.Duration) *worker {
	setRpcTimeOut(t)
	return &worker{id: id, Queue: queue, Queries: gendb.New(db), grpcServers: conn}
}

func (w *worker) ID() string {
	return w.id
}

func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	var data models.Envlope
	err := json.Unmarshal(j.Data, &data)
	if err != nil {
		log.Println("Failed to get data from job, err : ", err)
		return err
	}
	var dataCount int32
	if data.TargetAction != constants.DROP {
		dataCount = GetDataCountInPayload(data.Data, data.TargetRPC)
	}
	err = dataToRPCMappings[data.TargetRPC][data.TargetAction](ctx, data, w.grpcServers[data.TargetService])
	log.Println(" DEBUG: RPC call repsone , Err[", err, "] No of Data Sent [", dataCount, "] retries left [", j.RetryCount.Int32, "] maxretries [", w.Queue.GetRetries(), "]")
	if err != nil {
		if data.TargetAction != constants.DROP {
			if j.RetryCount.Int32 == w.Queue.GetRetries() {
				dbErr := w.Queries.UpdateFileFailedRecord(ctx, gendb.UpdateFileFailedRecordParams{
					UploadID:      data.UploadID,
					FileName:      data.FileName,
					FailedRecords: dataCount,
				})
				if dbErr != nil {
					log.Println("Failed to update failedrecord in db ,err [", err, "] ,requeued for defer worker for jobId ", j.JobID, "]")
					dJob := job.Job{
						Type:     constants.DEFERTYPE,
						Data:     data.Data,
						Comments: sql.NullString{String: "FAILED", Valid: true},
						Status:   job.JobStatusPENDING,
					}
					w.Queue.PushJob(ctx, dJob, constants.DEFERWORKER)
					return dbErr
				}
			}
		}
		return err
	}
	if data.TargetAction != constants.DROP {
		err = w.Queries.UpdateFileSuccessRecord(ctx, gendb.UpdateFileSuccessRecordParams{
			UploadID:       data.UploadID,
			FileName:       data.FileName,
			SuccessRecords: dataCount,
		})
		if err != nil {
			log.Println("Failed to update success record in db , err [", err, "] requeued for defer worker for jobId ", j.JobID, "]")
			dJob := job.Job{
				Type:     constants.DEFERTYPE,
				Data:     data.Data,
				Comments: sql.NullString{String: "SUCCESS", Valid: true},
				Status:   job.JobStatusPENDING,
			}
			w.Queue.PushJob(ctx, dJob, constants.DEFERWORKER)
			return err
		}
	}
	return nil
}
