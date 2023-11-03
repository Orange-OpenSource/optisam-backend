package apiworker

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	product "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/product-service/pkg/api/v1"

	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/postgres/db"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/worker/constants"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/worker/models"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type worker struct {
	id string
	*workerqueue.Queue
	*gendb.Queries
	grpcServers map[string]*grpc.ClientConn
	t           time.Duration
}

var sourceDir, archiveDir string

// NewWorker give worker object
func NewWorker(id string, queue *workerqueue.Queue, db *sql.DB, conn map[string]*grpc.ClientConn, t time.Duration, sdir, adir string) *worker {
	setRpcTimeOut(t)
	sourceDir = sdir
	archiveDir = adir
	return &worker{id: id, Queue: queue, Queries: gendb.New(db), grpcServers: conn}
}

func (w *worker) ID() string {
	return w.id
}

func (w *worker) DoWork(ctx context.Context, j *job.Job) error {
	var data models.Envlope
	err := json.Unmarshal(j.Data, &data)
	if err != nil {
		logger.Log.Error("Failed to unmarshal job data ", zap.Error(err))
		return err
	}
	err = dataToRPCMappings[data.TargetRPC][data.TargetAction](ctx, data, w.grpcServers[data.TargetService])
	if data.TargetAction != constants.DROP {
		if err != nil {
			if j.RetryCount.Int32 == w.Queue.GetRetries() {
				if dberr := w.setFailedOrSuccessRecords(ctx, data, constants.FailedData, j.JobID); dberr != nil {
					return dberr
				}
			}
		} else {
			if dberr := w.setFailedOrSuccessRecords(ctx, data, constants.SuccessData, j.JobID); dberr != nil {
				return dberr
			}
		}
	}
	return err
}

func (w *worker) setFailedOrSuccessRecords(ctx context.Context, data models.Envlope, action string, jobId int32) error {
	var fileNewStatus gendb.UploadStatus
	var dbErr error
	dataCount := GetDataCountInPayload(data.Data, data.TargetRPC)
	if action == constants.FailedData {
		resp, err := w.Queries.UpdateFileFailedRecord(ctx, gendb.UpdateFileFailedRecordParams{
			UploadID:      data.UploadID,
			FileName:      data.FileName,
			FailedRecords: dataCount,
		})
		if resp.Issuccess {
			fileNewStatus = gendb.UploadStatusSUCCESS
		} else if resp.Isfailed {
			fileNewStatus = gendb.UploadStatusFAILED
		} else if resp.Ispartial {
			fileNewStatus = gendb.UploadStatusPARTIAL
		}
		dbErr = err
		logger.Log.Debug("UpdateFailedRecord", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", fileNewStatus), zap.Any("DBError", dbErr))
	} else if action == constants.SuccessData {
		resp, err := w.Queries.UpdateFileSuccessRecord(ctx, gendb.UpdateFileSuccessRecordParams{
			UploadID:       data.UploadID,
			FileName:       data.FileName,
			SuccessRecords: dataCount,
		})
		if resp.Issuccess {
			fileNewStatus = gendb.UploadStatusSUCCESS
		} else if resp.Isfailed {
			fileNewStatus = gendb.UploadStatusFAILED
		} else if resp.Ispartial {
			fileNewStatus = gendb.UploadStatusPARTIAL
		}
		dbErr = err
		logger.Log.Debug("UpdateSuccessRecord", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", fileNewStatus), zap.Any("DBError", dbErr))
	} else {
		return errors.New("Unknown Action on records")
	}

	if dbErr == nil && fileNewStatus != "" {
		dbErr = HandleDataFileStatus(ctx, w.Queries, fileNewStatus, data, w.grpcServers["product"])
		action = string(fileNewStatus)
	}
	if dbErr != nil {
		logger.Log.Error("UpdateFileStatus", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", fileNewStatus), zap.Any("DBError", dbErr))
		dJob := job.Job{
			Type:     constants.DEFERTYPE,
			Data:     data.Data,
			Comments: sql.NullString{String: action, Valid: true},
			Status:   job.JobStatusPENDING,
		}
		w.Queue.PushJob(ctx, dJob, constants.DEFERWORKER)
		return dbErr
	}
	return nil
}

func HandleDataFileStatus(ctx context.Context, dbObj *gendb.Queries, dataFileStatus gendb.UploadStatus, data models.Envlope, prod grpc.ClientConnInterface) (err error) {

	oldName := fmt.Sprintf("%s/%s", sourceDir, data.FileName)
	newName := fmt.Sprintf("%s/%d_%s", archiveDir, data.UploadID, data.FileName)
	//  dataFile status update
	err = dbObj.UpdateFileStatus(ctx, gendb.UpdateFileStatusParams{
		Status:   dataFileStatus,
		UploadID: data.UploadID,
		FileName: data.FileName,
	})
	if err != nil {
		logger.Log.Error("UpdateFileStatus", zap.Any("uid", data.UploadID), zap.Any("file", data.FileName), zap.Any("newfilestatus", dataFileStatus), zap.Error(err))
		return
	}
	logger.Log.Debug("UpdateFileStatus", zap.Any("uid", data.UploadID), zap.Any("gid", data.GlobalFileID), zap.Any("file", data.FileName), zap.Any("newfilestatus", dataFileStatus))

	// global file status update
	if data.GlobalFileID > int32(0) {
		err = HandleGlobalFileStatus(ctx, dbObj, data.GlobalFileID, prod)
		oldName = fmt.Sprintf("%s/%s", sourceDir, data.TransfromedFileName)
		newName = fmt.Sprintf("%s/%s", archiveDir, data.TransfromedFileName)
	}

	// file archive
	osErr := os.Rename(oldName, newName)
	if osErr != nil {
		logger.Log.Error("Failed to archive the file", zap.Any("uid", data.UploadID), zap.Any("gid", data.GlobalFileID), zap.Error(osErr))
		return
	}
	logger.Log.Error("File Archived", zap.Any("old", oldName), zap.Any("new", newName))
	return
}

func HandleGlobalFileStatus(ctx context.Context, dbObj *gendb.Queries, gid int32, prod grpc.ClientConnInterface) (err error) {
	var dstatus []gendb.UploadStatus
	gstatus := gendb.UploadStatusCOMPLETED //defau
	fileRegx := fmt.Sprintf("%s/%d_*.csv", sourceDir, gid)
	files, _ := filepath.Glob(fileRegx)
	if files != nil && len(files) > 0 {
		logger.Log.Sugar().Errorf("More transformed file need to be processed", "gid", gid)
		return nil
	}

	dstatus, err = dbObj.GetAllDataFileStatusByGID(ctx, gid)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to get all data file status ", "gid", gid, "err", err.Error())
		return
	}
	for _, val := range dstatus {
		if val == gendb.UploadStatusPENDING || val == gendb.UploadStatusINPROGRESS {
			logger.Log.Sugar().Errorf("Few nifi transformed files are still in pending/progress", "gid", gid)
			return
		}
		if val == gendb.UploadStatusFAILED || val == gendb.UploadStatusPARTIAL {
			gstatus = gendb.UploadStatusPARTIAL
			break
		}
	}

	scope, err := dbObj.UpdateGlobalFileStatus(ctx, gendb.UpdateGlobalFileStatusParams{
		UploadID: gid,
		Column2:  gstatus,
	})
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to update global file status", "err", err.Error(), "gid", gid, "status", gstatus)
		return err
	}
	logger.Log.Sugar().Infof("Global file status update ", "gid", gid, "status", gstatus)

	if prod != nil {
		logger.Log.Sugar().Infof("Calling CreateDashboardUpdateJob........")
		if resp, err := product.NewProductServiceClient(prod).CreateDashboardUpdateJob(ctx, &product.CreateDashboardUpdateJobRequest{Scope: scope, Ppid: fmt.Sprintf("%v", gid)}); err != nil || !resp.Success {
			logger.Log.Sugar().Errorf("Failed to create licences calculation job", "err", err)
			return err
		}
	}
	return nil
}
