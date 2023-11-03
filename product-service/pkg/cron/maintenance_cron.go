package cron

import (
	"database/sql"
	"encoding/json"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"go.uber.org/zap"
)

func MaintenanceJob() {
	logger.Log.Debug("cron job started...")
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Sugar().Infof("Panic recovered from cron job", r)
		}
	}()
	cronCtx, err := createSharedContext(AuthAPI)
	if err != nil {
		logger.Log.Error("couldnt fetch token, will try next time when cron will execute", zap.Any("error", err))
	}

	if cronCtx != nil {
		cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, VerifyKey, APIKey)
		if err != nil {
			logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
		}

		jobID, err := Queue.PushJob(cronAPIKeyCtx, job.Job{
			Type:   sql.NullString{String: "mw"},
			Status: job.JobStatusPENDING,
			Data:   json.RawMessage(`{"updatedBy":"cron"}`),
		}, "mw")
		if err != nil {
			logger.Log.Info("Error from job", zap.Int32("jobId", jobID))
		}
		logger.Log.Info("Successfully pushed job by cron", zap.Int32("jobId", jobID))
	}
}
