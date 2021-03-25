// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package cron

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	//Queue ...
	Queue workerqueue.Queue
	//AuthAPI ...
	AuthAPI string
)

//CronJobConfigInit ...
func CronJobConfigInit(q workerqueue.Queue, authapi string) {
	Queue = q
	AuthAPI = authapi
}

//Job will be executed by cron
func Job() {

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
		jobID, err := Queue.PushJob(*cronCtx, job.Job{
			Type:   sql.NullString{String: "lcalw"},
			Status: job.JobStatusPENDING,
			Data:   json.RawMessage(`{}`),
		}, "lcalw")
		if err != nil {
			logger.Log.Info("Error from job", zap.Int32("jobId", jobID))
		}
		logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	}
}

func createSharedContext(api string) (*context.Context, error) {
	ctx := context.Background()
	respMap := make(map[string]interface{})
	data := url.Values{
		"username":   {"admin@test.com"},
		"password":   {"admin"},
		"grant_type": {"password"},
	}

	resp, err := http.PostForm(api, data)
	if err != nil {
		log.Println("Failed to get user claims  ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	log.Println(" Token Data received", string(bodyBytes))
	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		log.Println("failed to unmarshal byte data", err)
		return nil, err
	}
	authStr := fmt.Sprintf("Bearer %s", respMap["access_token"].(string))
	md := metadata.Pairs("Authorization", authStr)
	// for debug
	//md := metadata.Pairs("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJhZG1pbkB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IlN1cGVyQWRtaW4iLCJTb2NwZXMiOlsiT0ZSIiwiT1NQIiwiSU5EIiwiT0JTIiwiT1JOIiwiVFNUIiwiT0xOIiwiVEVTIiwiVEVUIiwiREVNIiwiRE1PIiwiVFNBIiwiQVRFIiwiQVVUIiwiRlNUIiwiT0tUIiwiTVZQIiwiT1NOIiwiWlpZIiwiTkVXIiwiWlpaIl0sImV4cCI6MTYwOTg2MDE1MCwiaWF0IjoxNjA5ODUyOTUwLCJpc3MiOiJPcmFuZ2UiLCJzdWIiOiJBY2Nlc3MgVG9rZW4ifQ.eZgy0yLF1zsCM42_vkAZrT4RMKOh8tRpI92l_ObNXm5D6Ax94tGzji-tDFku3_XLVHYUDE41W0xJhVy5SrmbY676GeCgNUYhVXxWU2JwfLyFuxi1DVxhA_SG9xbsIDKLHlIyqOAF-KDnrJMRBsbMF4Fat4zULlAA31v_px_0zChL6MmijIGt9pcpqM9AL9V5iq9tbRHIPqkPV8dUgkdYEQiXoJoQLtxlHaFpEGy_0YIlj0r4y1tWSZ_oZxymVcMvZOaCHR7ZfCWZ7rzI8r-E72Dwn9sGEPMDVmR5-KCoa3DfypgvWu6-z10r6if7SOw8NyGtP12eigmc4g8NcS5a4Q")
	ctx = metadata.NewIncomingContext(ctx, md)

	return &ctx, nil
}