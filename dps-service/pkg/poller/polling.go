package cron

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"optisam-backend/common/optisam/config"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"os"
	"path/filepath"
	"strings"

	repo "optisam-backend/dps-service/pkg/repository/v1"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	Queue          workerqueue.Queue
	AuthAPI        string
	SourceDir      string
	ArchieveDir    string
	RawdataDir     string
	Obj            v1.DpsServiceServer
	VerifyKey      *rsa.PublicKey
	APIKey         string
	dbObj          repo.Dps
	WaitLimitCount int
	AppConfig      config.Application
)

const (
	PROCESSING        string = "PROCESSING"
	NIFIIsDown        string = "NIFIIsDown"
	NIFIInternalError string = "NIFIInternalError"
)

func Init(q workerqueue.Queue, authapi, sourceDir, archieveDir, rawdataDir string, obj v1.DpsServiceServer, key *rsa.PublicKey, apiKey string, db repo.Dps, waitLimitCount int, config config.Application) {
	Queue = q
	RawdataDir = rawdataDir
	AuthAPI = authapi
	SourceDir = sourceDir
	ArchieveDir = archieveDir
	Obj = obj
	VerifyKey = key
	APIKey = apiKey
	dbObj = db
	WaitLimitCount = 3
	if waitLimitCount > 0 {
		WaitLimitCount = waitLimitCount
	}
	AppConfig = config

}

var (
	nonProcessedFileRecord = make(map[string]int)
)

// Thiw Job will be executed by cron
func Job() { //nolint
	logger.Log.Info("cron job started...")
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Panic recovered from cron job", zap.Any("recover", r))
		}
	}()
	cronCtx, err := createSharedContext(AuthAPI, AppConfig)
	if err != nil {
		logger.Log.Error("couldnt fetch token, will try next time when cron will execute", zap.Any("error", err))
		return
	}
	if cronCtx != nil {
		cronAPIKeyCtx, err := grpc.AddClaimsInContext(*cronCtx, VerifyKey, APIKey)
		if err != nil {
			logger.Log.Error("Cron AddClaims Failed", zap.Error(err))
			return
		} else if cronAPIKeyCtx == nil {
			logger.Log.Error("Failed to get context. nil pointer ctx")
			return
		}
		resp, err := dbObj.GetTransformedGlobalFileInfo(cronAPIKeyCtx)
		if err != nil {
			logger.Log.Error("Failed to get unprocessed global files info", zap.Error(err))
			return
		}
		logger.Log.Debug("Processing global file ", zap.Any("data", resp))
		for _, global := range resp {
			fileNameWithoutExt := ""
			if strings.Contains(global.FileName, ".xlsx") {
				fileNameWithoutExt = strings.TrimSuffix(global.FileName, ".xlsx")
			} else {
				fileNameWithoutExt = strings.TrimSuffix(global.FileName, ".csv")
			}
			globalFileDir := "GEN"
			if global.ScopeType == db.ScopeTypesSPECIFIC {
				globalFileDir = global.Scope
			}
			errFileRegax := fmt.Sprintf("%s/%s/error/%d_%s_error_ft_%s.zip", RawdataDir, globalFileDir, int(global.UploadID), global.Scope, fileNameWithoutExt)
			errFiles, err := filepath.Glob(errFileRegax)
			if err != nil {
				logger.Log.Error("Failed to read error dir", zap.Any("filepath", errFileRegax), zap.Error(err))
				continue
			}

			// File Type Error , global file should mark failed
			if errFiles != nil && len(errFiles) > 0 {
				if err = dbObj.UpdateFileStatus(cronAPIKeyCtx, db.UpdateFileStatusParams{
					Status:   db.UploadStatusFAILED,
					UploadID: global.UploadID,
					FileName: global.FileName,
					Comments: sql.NullString{String: NIFIInternalError, Valid: true},
				}); err != nil {
					logger.Log.Error("Failed to update the status", zap.Any("uid", global.UploadID), zap.Any("scope", global.Scope), zap.Error(err))
				}
				continue
			} else if global.Status == db.UploadStatusUPLOADED {
				err = handleNifiErrors(cronAPIKeyCtx, global.Scope, globalFileDir, global.FileName, int(global.UploadID))
				if err != nil {
					logger.Log.Error("Failed to handle nifi error", zap.Error(err))
					continue
				}
			}

			// get data files
			dataFileRegex := fmt.Sprintf("%s/%d_*.csv", SourceDir, global.UploadID)
			dataFiles, err := filepath.Glob(dataFileRegex)
			if err != nil {
				logger.Log.Error("Failed to read data dir", zap.Error(err))
				continue
			} else if dataFiles == nil {
				continue
			}
			logger.Log.Debug("Global Id transformed data files ", zap.Any("gid", global.UploadID), zap.Any("dataFiles", dataFiles))

			if _, err = dbObj.UpdateGlobalFileStatus(cronAPIKeyCtx, db.UpdateGlobalFileStatusParams{
				Column2:  db.UploadStatusPROCESSED,
				UploadID: global.UploadID,
			}); err != nil {
				logger.Log.Error("Failed to update the status", zap.Any("uid", global.UploadID), zap.Any("scope", global.Scope), zap.Error(err))
			}

			var filesToSend []string
			for _, val := range dataFiles {
				_, df := filepath.Split(val)
				newFile := fmt.Sprintf("%s#%s", PROCESSING, df)
				if err = os.Rename(fmt.Sprintf("%s/%s", SourceDir, df), fmt.Sprintf("%s/%s", SourceDir, newFile)); err != nil {
					logger.Log.Error("Failed to mark processing the global_data_file", zap.Any("oldFile", df), zap.Any("newFileName", newFile), zap.Error(err))
					continue
				}
				filesToSend = append(filesToSend, newFile)
			}
			scopeType := v1.NotifyUploadRequest_GENERIC
			if global.ScopeType != db.ScopeTypesGENERIC {
				scopeType = v1.NotifyUploadRequest_SPECIFIC
			}
			notifyResp, err := Obj.NotifyUpload(cronAPIKeyCtx, &v1.NotifyUploadRequest{
				Scope:      global.Scope,
				Type:       "data",
				UploadedBy: "nifi",
				ScopeType:  scopeType,
				Files:      filesToSend})
			if err != nil || (notifyResp != nil && !notifyResp.Success) {
				logger.Log.Error("Notify uplaod failed for nifi transformed files", zap.Error(err))
				revertProcessingFilesName(filesToSend)
			}
		}
	}
}

func handleNifiErrors(ctx context.Context, scope, globalFileDir, fileName string, id int) error {
	globalFile := fmt.Sprintf("%s/%s/%d_%s", RawdataDir, globalFileDir, id, fileName)
	res, _ := filepath.Glob(globalFile)
	if len(res) > 0 {
		if nonProcessedFileRecord[globalFile] < WaitLimitCount {
			nonProcessedFileRecord[globalFile]++
		} else {
			archivedFile := fmt.Sprintf("%s/%s/archive/%d_*", RawdataDir, globalFileDir, id)
			res, _ = filepath.Glob(archivedFile)
			fmt.Println("ARCHIVE ", res, archivedFile)
			errComment := NIFIIsDown
			if len(res) > 0 {
				errComment = NIFIInternalError
			}
			delete(nonProcessedFileRecord, globalFile)
			os.Remove(globalFile)
			if err := dbObj.UpdateFileStatus(ctx, db.UpdateFileStatusParams{
				Status:   db.UploadStatusFAILED,
				UploadID: int32(id),
				FileName: fileName,
				Comments: sql.NullString{String: errComment, Valid: true},
			}); err != nil {
				logger.Log.Error("Failed to update the status", zap.Any("uid", id), zap.Any("scope", scope), zap.Error(err))
				return err
			}
		}
	} else {
		dataFileRegex := fmt.Sprintf("%s/%d_*.csv", SourceDir, id)
		res, _ := filepath.Glob(dataFileRegex)
		if len(res) == 0 {
			if nonProcessedFileRecord[globalFile] < WaitLimitCount {
				nonProcessedFileRecord[globalFile]++
			} else {
				delete(nonProcessedFileRecord, globalFile)
				if err := dbObj.UpdateFileStatus(ctx, db.UpdateFileStatusParams{
					Status:   db.UploadStatusFAILED,
					UploadID: int32(id),
					FileName: fileName,
					Comments: sql.NullString{String: NIFIInternalError, Valid: true},
				}); err != nil {
					logger.Log.Error("Failed to update the status", zap.Any("uid", id), zap.Any("scope", scope), zap.Error(err))
					return err
				}
			}
		} else {
			delete(nonProcessedFileRecord, globalFile)
		}
	}
	return nil
}

func revertProcessingFilesName(files []string) {
	for _, file := range files {
		oldFile := fmt.Sprintf("%s/%s", SourceDir, strings.Split(file, fmt.Sprintf("%s#", PROCESSING))[1])
		if err := os.Rename(fmt.Sprintf("%s/%s", SourceDir, file), oldFile); err != nil {
			logger.Log.Error("Failed to revert the processing file", zap.Error(err))
		}
	}
}

func createSharedContext(api string, appcred config.Application) (*context.Context, error) {
	ctx := context.Background()
	respMap := make(map[string]interface{})
	data := url.Values{
		"username":   {appcred.UserNameSuperAdmin},
		"password":   {appcred.PasswordSuperAdmin},
		"grant_type": {"password"},
	}

	resp, err := http.PostForm(api, data) // nolint: gosec
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
	// md := metadata.Pairs("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOiJhZG1pbkB0ZXN0LmNvbSIsIkxvY2FsZSI6ImVuIiwiUm9sZSI6IlN1cGVyQWRtaW4iLCJTb2NwZXMiOlsiQVVUIiwiT0ZSIiwiR0VOIiwiT0xOIiwiREVWIiwiQ0xSIiwiRE1PIiwiUFNUIiwiT1NOIiwiS0VTIl0sImV4cCI6MTYyMzE5NTk0NiwiaWF0IjoxNjIzMTg4NzQ2LCJpc3MiOiJPcmFuZ2UiLCJzdWIiOiJBY2Nlc3MgVG9rZW4ifQ.vcJDBPMENrSqjtt3VW4qDFO2fH_MtIk45ZHrIikbmtF6Ske7h5THteSLF2AX711NUOsHZksFy-anlUquKH2OHTNqP9GEZe8dsibDskGFgvBIQ2d24abwV6pI0REgqDPJrXuINQ0gFXTHZZ4bg7FukUK50fbxETJy-0LARa6OsKgoXJ5G-NIkmb65661P2pBQYX5hlA6y4ke1LqmDzYZyjEng5QlIs0nkQDVoW74vPUJBNoAV9pX410rb-vaCy1JXAt9axiqqNdgW6UytPUy2G9DAa6SfF_f6hnYURDQZ8ahxY68yA_HtlDjV8DQr76ZLFG9Tq9icJA3OL89XpDYvbA")

	ctx = metadata.NewIncomingContext(ctx, md)

	return &ctx, nil
}
