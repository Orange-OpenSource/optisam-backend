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

	apiworker "optisam-backend/dps-service/pkg/worker/api_worker"

	"go.uber.org/zap"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
}

// NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB) *worker { // nolint: golint
	return &worker{id: id, Queue: queue, Queries: gendb.New(db)}
}

// ID gives unique id of worker
func (w *worker) ID() string {
	return w.id
}

// DoWork tell the functionality of worker
func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop()
	dataFromJob := gendb.UploadedDataFile{}
	var data models.FileData
	var err error
	var jobs []job.Job
	defer func(error, gendb.UploadedDataFile, worker) {
		fileToArchive := dataFromJob.FileName
		id := dataFromJob.UploadID
		// Archiving the file when 1.There is no error or 2.When retries exceeded or 3. When no retries set
		if err != nil && (j.RetryCount.Int32 >= w.Queue.GetRetries() || w.Queue.GetRetries() == 0) {
			if dataFromJob.Gid > int32(0) {
				if apiworker.HandleGlobalFileStatus(ctx, w.Queries, dataFromJob.Gid, nil) != nil {
					logger.Log.Error("Failed to handle global file status", zap.Any("gid", dataFromJob.Gid))
				}
				id = dataFromJob.Gid
			}
			// When file not found, nothing to archive
			if fileToArchive == "" {
				return
			}
			error := archiveFile(fileToArchive, id)
			if error != nil {
				logger.Log.Error("Failed to archive file", zap.Error(error), zap.Any("file", fileToArchive), zap.Any("id", id))
			}
		}
	}(err, dataFromJob, *w)
	err = json.Unmarshal(j.Data, &dataFromJob)
	if err != nil {
		logger.Log.Debug("Failed to unmarshal the file type job data , err :", zap.Error(err))
		return err
	}

	fileNameInDB := getFileName(dataFromJob.FileName)

	err = w.Queries.UpdateFileStatus(ctx, gendb.UpdateFileStatusParams{
		UploadID: dataFromJob.UploadID,
		FileName: fileNameInDB,
		Status:   gendb.UploadStatusINPROGRESS,
	})
	if err != nil {
		logger.Log.Debug("Failed to update status , err ", zap.Error(err))
		return err
	}

	data, err = fileProcessing(dataFromJob)
	// to store nifi transformed files as we keep scope_data ,not with global
	if dataFromJob.FileName != fileNameInDB {
		data.TransfromedFileName = dataFromJob.FileName
	}
	if err != nil {
		logger.Log.Debug("Failed to process the file ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
		er := w.Queries.UpdateFileFailure(ctx, gendb.UpdateFileFailureParams{
			Status:   gendb.UploadStatusFAILED,
			Comments: sql.NullString{String: data.FileFailureReason, Valid: true},
			UploadID: dataFromJob.UploadID,
			FileName: fileNameInDB,
		})
		if er != nil {
			logger.Log.Debug("Failed to update file status ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
			return er
		}
		return errors.New(data.FileFailureReason)
	}

	logger.Log.Debug("file reading complete, stats:", zap.Any("TransformedFile", data.TransfromedFileName), zap.Any("file", data.FileName), zap.Any("totalRecord", data.TotalCount), zap.Any("Invalidrecords", data.InvalidCount), zap.Any("duplicateRecords", len(data.DuplicateRecords)))

	err = w.Queries.UpdateFileTotalRecord(ctx, gendb.UpdateFileTotalRecordParams{
		FileName:      fileNameInDB,
		UploadID:      dataFromJob.UploadID,
		TotalRecords:  data.TotalCount,
		FailedRecords: data.InvalidCount + int32(len(data.DuplicateRecords)),
	})
	if err != nil {
		logger.Log.Debug("Failed to update total Records in DB for file ", zap.Any("filename", dataFromJob.FileName), zap.Error(err))
		return err
	}

	oldName := data.FileName
	data.FileName = fileNameInDB
	jobs, err = createAPITypeJobs(data)
	data.FileName = oldName

	for _, job := range jobs {
		// Will implement through workerpool
		_, err := w.Queue.PushJob(ctx, job, constants.APIWORKER)
		if err != nil {
			logger.Log.Error("Job not pushed Successfully:", zap.Int32("job", job.JobID), zap.Error(err))
		}
	}

	setInvalidRecords(ctx, w, data, dataFromJob.UploadID, fileNameInDB)
	setDuplicateRecords(ctx, w, data, dataFromJob.UploadID, fileNameInDB)

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

type DuplicateRecord struct {
	Data     interface{}
	UploadID int32
	FileName string
	Scope    string
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

func setDuplicateRecords(ctx context.Context, w *worker, data models.FileData, id int32, fileName string) {

	for i := 0; i < len(data.DuplicateRecords); i++ {
		e := DuplicateRecord{
			Data:     data.DuplicateRecords[i],
			UploadID: id,
			FileName: fileName,
			Scope:    data.Scope,
		}
		dataToPush, err := json.Marshal(e)
		if err != nil {
			logger.Log.Error("Failed tp marshal the duplicate data, err ", zap.Error(err))
			continue
		}
		j := job.Job{
			Status:   job.JobStatusFAILED,
			Comments: sql.NullString{String: "DuplicateRecord", Valid: true},
			Data:     dataToPush,
			Type:     sql.NullString{String: constants.APIWORKER, Valid: true},
		}
		_, err = w.Queue.PushJob(ctx, j, constants.APIWORKER)
		if err != nil {
			logger.Log.Error("Failed to upsert duplicate-failed records, err ", zap.Error(err))
		}
	}
}
