// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package fileworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"

	"go.uber.org/zap"
	//"github.com/pkg/profile"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
}

//NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB) *worker {
	return &worker{id: id, Queue: queue, Queries: gendb.New(db)}
}

//ID gives unique id of worker
func (w *worker) ID() string {
	return w.id
}

//DoWork tell the functionality of worker
func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	//defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	dataFromJob := gendb.UploadedDataFile{}
	var data models.FileData
	var err error
	var jobs []job.Job
	defer func(error, job.Job, worker) {
		//Archiving the file when 1.There is no error or 2.When retries exceeded or 3. When no retries set
		if err == nil || j.RetryCount.Int32 >= w.Queue.GetRetries() || w.Queue.GetRetries() == 0 {
			archiveFile(dataFromJob.FileName, dataFromJob.UploadID)
		}
	}(err, *j, *w)
	err = json.Unmarshal(j.Data, &dataFromJob)
	if err != nil {
		logger.Log.Debug("Failed to unmarshal the file type job data , err :", zap.Error(err))
		return err
	}

	dataToUpdate := gendb.UpdateFileStatusParams{
		UploadID: dataFromJob.UploadID,
		FileName: dataFromJob.FileName,
		Status:   gendb.UploadStatusINPROGRESS,
	}

	err = w.Queries.UpdateFileStatus(ctx, dataToUpdate)
	if err != nil {
		logger.Log.Debug("Failed to update status , err ", zap.Error(err))
		return err
	}

	data, err = fileProcessing(dataFromJob)
	if err != nil {
		logger.Log.Debug("Failed to process the file ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
		er := w.Queries.UpdateFileFailure(ctx, gendb.UpdateFileFailureParams{
			Status:   gendb.UploadStatusFAILED,
			Comments: sql.NullString{String: data.FileFailureReason, Valid: true},
			UploadID: dataFromJob.UploadID,
			FileName: dataFromJob.FileName,
		})
		if er != nil {
			logger.Log.Debug("Failed to update file status ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
			return er
		}
		return errors.New(data.FileFailureReason)
	}

	logger.Log.Debug("proccessed ", zap.Any("file", data.FileName), zap.Any("totalRecord", data.TotalCount))

	err = w.Queries.UpdateFileTotalRecord(ctx, gendb.UpdateFileTotalRecordParams{
		FileName:      dataFromJob.FileName,
		UploadID:      dataFromJob.UploadID,
		TotalRecords:  data.TotalCount,
		FailedRecords: data.InvalidCount,
	})
	if err != nil {
		logger.Log.Debug("Failed to update total Records in DB for file ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
		return err
	}
	jobs, err = createAPITypeJobs(data)

	for _, job := range jobs {
		//Will implement through workerpool
		w.Queue.PushJob(ctx, job, constants.APIWORKER)
	}
	dataToUpdate.Status = gendb.UploadStatusCOMPLETED
	err = w.Queries.UpdateFileStatus(ctx, dataToUpdate)
	if err != nil {
		logger.Log.Debug("Failed to update status , err ", zap.Error(err))
		return err
	}
	setInvalidRecords(ctx, w, data, dataFromJob.UploadID, dataFromJob.FileName)

	return nil
}

type InvalidRecord struct {
	Data struct {
		AtLineNo string
	}
	UploadID int32
	FileName string
	Scope    string `json:"scope"`
}

func setInvalidRecords(ctx context.Context, w *worker, data models.FileData, id int32, fileName string) {

	for i := 0; i < int(data.InvalidCount); i++ {
		e := InvalidRecord{
			Data:     struct{ AtLineNo string }{fmt.Sprintf("%d", data.InvalidDataRowNum[i])},
			UploadID: id,
			FileName: fileName,
			Scope:    data.Scope,
		}
		dataToPush, err := json.Marshal(e)
		if err != nil {
			log.Println("Failed tp marshal the invalid data, err ", err)
			continue
		}
		j := job.Job{
			Status:   job.JobStatusFAILED,
			Comments: sql.NullString{String: "InsufficentData", Valid: true},
			Data:     dataToPush,
			Type:     sql.NullString{String: constants.APIWORKER, Valid: true},
		}
		_, err = w.Queue.PushJob(ctx, j, constants.APIWORKER)
		if err != nil {
			log.Println("Failed to upsert invalid-failed records, err ", err)
		}
	}
}
