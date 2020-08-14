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
	"optisam-backend/dps-service/pkg/worker/models"

	"google.golang.org/grpc"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
	grpcServers map[string]*grpc.ClientConn
}

//NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB, conn map[string]*grpc.ClientConn) *worker {
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
	dataCount := getDataCountInPayload(data.Data, data.TargetRPC)
	err = dataToRPCMappings[data.TargetRPC][data.TargetAction](ctx, data, w.grpcServers[data.TargetService])
	if err != nil {
		log.Println("Failed RPC request , err : ", err)
		if j.RetryCount.Int32 == w.Queue.GetRetries() {
			dbErr := w.Queries.UpdateFileFailedRecord(ctx, gendb.UpdateFileFailedRecordParams{
				UploadID:      data.UploadID,
				FileName:      data.FileName,
				FailedRecords: dataCount,
			})
			if dbErr != nil {
				log.Println("Failed to update failedrecord in db , err :", err)
				return dbErr
			}
		}
		return err
	}
	err = w.Queries.UpdateFileSuccessRecord(ctx, gendb.UpdateFileSuccessRecordParams{
		UploadID:       data.UploadID,
		FileName:       data.FileName,
		SuccessRecords: dataCount,
	})
	if err != nil {
		log.Println("Failed to update success record in db , err :", err)
		return err
	}
	return nil
}
