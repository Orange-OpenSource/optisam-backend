// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package cron

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/worker/constants"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	Queue     workerqueue.Queue
	AuthAPI   string
	SourceDir string
	Obj       v1.DpsServiceServer
	VerifyKey *rsa.PublicKey
)

func Init(q workerqueue.Queue, authapi, sourceDir string, obj v1.DpsServiceServer, key *rsa.PublicKey) {
	Queue = q
	AuthAPI = authapi
	SourceDir = sourceDir
	Obj = obj
	VerifyKey = key
}

//Thiw Job will be executed by cron
func Job() {
	logger.Log.Debug("cron job started...")
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Debug("Panic recovered from cron job", zap.Any("recover", r))
		}
	}()
	cronCtx, err := createSharedContext(AuthAPI)
	if err != nil {
		logger.Log.Debug("couldnt fetch token, will try next time when cron will execute", zap.Any("error", err))
		return
	}
	if cronCtx != nil {
		*cronCtx, err = grpc.AddClaimsInContext(*cronCtx, VerifyKey)
		fileScopeMapping := make(map[string][]string)
		//Read Dir , if found create the job
		files, er := ioutil.ReadDir(SourceDir)
		if er != nil {
			logger.Log.Debug("Failed to read the dirctory/files", zap.Any("directory", SourceDir), zap.Error(er))
			return
		}
		for _, fileInfo := range files {
			temp := strings.Split(fileInfo.Name(), constants.SCOPE_DELIMETER)
			if len(temp) == 0 {
				continue
			}
			//data["TST"]= []{"f1.csv","f2.csv","f3.csv"}, map is because if multiple files come
			fileScopeMapping[temp[0]] = append(fileScopeMapping[temp[0]], fileInfo.Name())
		}

		for scope, files := range fileScopeMapping {
			resp, err := Obj.NotifyUpload(*cronCtx, &v1.NotifyUploadRequest{
				Scope:      scope,
				Type:       "data",
				UploadedBy: "Nifi",
				Files:      files})
			if err != nil || (resp != nil && !resp.Success) {
				logger.Log.Debug("failed to upload the transformed files", zap.Error(err))
			}
		}
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
		logger.Log.Debug("Failed to get user claims  ", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &respMap)
	if err != nil {
		logger.Log.Debug("failed to unmarshal byte data", zap.Error(err))
		return nil, err
	}
	authStr := fmt.Sprintf("Bearer %s", respMap["access_token"].(string))
	md := metadata.Pairs("Authorization", authStr)
	// for debug
	//md := metadata.Pairs("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJhZG1pbkB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IlN1cGVyQWRtaW4iLCJTb2NwZXMiOlsiT0ZSIiwiT1NQIiwiT0JTIiwiT1JOIiwiVFNUIiwiT0xOIiwiRlNUIiwiT1NOIl0sImV4cCI6MTYxMzU2MTY0OCwiaWF0IjoxNjEzNTU0NDQ4LCJpc3MiOiJPcmFuZ2UiLCJzdWIiOiJBY2Nlc3MgVG9rZW4ifQ.f7RZgV8Imj2s8MlfzY2TlALUQTaYWFIggd7II7T34VP6whhOkRulF9ud51TdL1dkQN9Nke_4v6qry2ClcXzmHPq9uXfkbzqBZGyIYyCTmlibK-8MpbvdiN51PsO5EUBGZqgtLB7sRQ5XwmJozG2b7QN-ORPAChFX3RehbJeJbw_NrxT2Wz_DsElXTVUU3LxWCuvFdA_nv3FC6xhRnPifhqbcsPwuetI_2CQTHQa43Aj1w6zjvHQ3c4yNvqKQWv1c-HnZ2uh482s-G319oaIFed5xrxyx4sWsp4yUdgYHPvQOo1rEIIzmY96vSqMR3oxOQo2E4C2vUKZ7l50mNftTVg")
	ctx = metadata.NewIncomingContext(ctx, md)

	return &ctx, nil
}
