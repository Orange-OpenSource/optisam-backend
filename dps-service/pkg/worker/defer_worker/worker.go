package deferworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	apiworker "optisam-backend/dps-service/pkg/worker/api_worker"
	"optisam-backend/dps-service/pkg/worker/constants"
	"optisam-backend/dps-service/pkg/worker/models"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
	grpcServers map[string]*grpc.ClientConn
}

// NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB, conn map[string]*grpc.ClientConn) *worker { // nolint: golint
	return &worker{id: id, Queue: queue, Queries: gendb.New(db), grpcServers: conn}
}

func (w *worker) ID() string {
	return w.id
}

func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	time.Sleep(10 * time.Second)
	var data models.Envlope
	var fileNewStatus gendb.UploadStatus
	err := json.Unmarshal(j.Data, &data)
	if err != nil {
		logger.Log.Error("Failed to unmarshal the defer job", zap.Error(err))
		return err
	}

	if j.Comments.String == constants.FailedData { // nolint: gocritic
		dataCount := apiworker.GetDataCountInPayload(data.Data, data.TargetRPC)
		resp, dberr := w.Queries.UpdateFileFailedRecord(ctx, gendb.UpdateFileFailedRecordParams{
			UploadID:      data.UploadID,
			FileName:      data.FileName,
			FailedRecords: dataCount,
		})
		if resp.Issuccess { // nolint: gocritic
			fileNewStatus = gendb.UploadStatusSUCCESS
		} else if resp.Isfailed {
			fileNewStatus = gendb.UploadStatusFAILED
		} else if resp.Ispartial {
			fileNewStatus = gendb.UploadStatusPARTIAL
		}
		err = dberr
		logger.Log.Debug("UpdateFileFailedRecord", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", fileNewStatus), zap.Any("DBError", err))
	} else if j.Comments.String == constants.SuccessData {
		dataCount := apiworker.GetDataCountInPayload(data.Data, data.TargetRPC)
		resp, dberr := w.Queries.UpdateFileSuccessRecord(ctx, gendb.UpdateFileSuccessRecordParams{
			UploadID:       data.UploadID,
			FileName:       data.FileName,
			SuccessRecords: dataCount,
		})
		if resp.Issuccess { // nolint: gocritic
			fileNewStatus = gendb.UploadStatusSUCCESS
		} else if resp.Isfailed {
			fileNewStatus = gendb.UploadStatusFAILED
		} else if resp.Ispartial {
			fileNewStatus = gendb.UploadStatusPARTIAL
		}
		err = dberr
		logger.Log.Debug("UpdateFileSuccessRecord", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", fileNewStatus), zap.Any("DBError", err))
	} else {
		switch j.Comments.String {
		case constants.SUCCESS:
			fileNewStatus = gendb.UploadStatusSUCCESS
		case constants.FAILED:
			fileNewStatus = gendb.UploadStatusFAILED
		case constants.PARTIAL:
			fileNewStatus = gendb.UploadStatusPARTIAL
		}
	}
	if err == nil && fileNewStatus != "" {
		err = apiworker.HandleDataFileStatus(ctx, w.Queries, fileNewStatus, data, w.grpcServers["product"])
	}

	if err != nil {
		logger.Log.Error("Failed to update records in db ", zap.Any("defer action", j.Comments.String), zap.Error(err))
		w.Queue.PushJob(ctx, *j, constants.DEFERWORKER) // nolint: errcheck
		return err
	}
	return nil
}
