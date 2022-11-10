package cron

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	Queue     workerqueue.Queue
	AuthAPI   string
	VerifyKey *rsa.PublicKey
	APIKey    string
)

func JobConfigInit(q workerqueue.Queue, authapi string, key *rsa.PublicKey, apiKey string) {
	Queue = q
	AuthAPI = authapi
	VerifyKey = key
	APIKey = apiKey
}

// Thiw Job will be executed by cron
func Job() {
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
			Type:   sql.NullString{String: "ob"},
			Status: job.JobStatusPENDING,
			Data:   json.RawMessage(`{}`),
		}, "ob")
		if err != nil {
			logger.Log.Info("Error from job", zap.Int32("jobId", jobID))
		}
		logger.Log.Info("Successfully pushed job by cron", zap.Int32("jobId", jobID))
	}
}

func createSharedContext(api string) (*context.Context, error) {
	ctx := context.Background()
	respMap := make(map[string]interface{})
	data := url.Values{
		"username":   {"admin@test.com"},
		"password":   {"Welcome@123"},
		"grant_type": {"password"},
	}

	resp, err := http.PostForm(api, data) // nolint: gosec
	if err != nil {
		log.Println("Failed to get user claims  ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		log.Println("failed to unmarshal byte data", err)
		return nil, err
	}
	authStr := fmt.Sprintf("Bearer %s", respMap["access_token"].(string))
	md := metadata.Pairs("Authorization", authStr)
	// for debug
	// md := metadata.Pairs("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJhZG1pbkB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IlN1cGVyQWRtaW4iLCJTb2NwZXMiOlsiREVNIiwiVFNUIiwiVEVTIiwiT1JOIl0sImV4cCI6MTYwNTcwMzU1MiwiaWF0IjoxNjA1Njk2MzUyLCJpc3MiOiJPcmFuZ2UiLCJzdWIiOiJBY2Nlc3MgVG9rZW4ifQ.fYR3qqjjhm56xPRam0_VQz-e00QWBexmev1gUmerCvx5MClUXmtujMMewr2cBNjzAuQNgour83AS4Es0RRXhTnAH7YPoYZIfmkvyRvKXDdT-MoLm0_Uh2kUSOLxz02e6-6Xlue3aECRtiCXZwphyORmtv-Suc1hlEuik_Y0W4PoEOTuL0cbWd3qian_zgtGS1xb4BQn8xsmZI35Fh13bvGYc9zO2B3mwYViXY0EHIlT3VFCvT95Qy0355xsyZEuCm2FlEsaUDqFaiDij9d5RbIQ2Fu910EtqjkGR04xKt6uI5ldqZPjOoeEWd5g2CaZumZTimSdzmcNPgT4AaalMbQ")
	ctx = metadata.NewIncomingContext(ctx, md)

	return &ctx, nil
}
