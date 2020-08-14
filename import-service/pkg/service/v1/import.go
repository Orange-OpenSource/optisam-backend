// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/import-service/pkg/config"
	"os"
	"path"
	"path/filepath"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type importServiceServer struct {
	// grpcServers map[string]*grpc.ClientConn
	dpsClient v1.DpsServiceClient
	config    *config.Config
}

type uploadType string

const (
	metadataload uploadType = "metadata"
	dataload     uploadType = "data"
)

// NewProductServiceServer creates Product service
func NewImportServiceServer(grpcServers map[string]*grpc.ClientConn, config *config.Config) ImportServiceServer {
	return &importServiceServer{config: config, dpsClient: v1.NewDpsServiceClient(grpcServers["dps"])}
}

func (i *importServiceServer) UploadDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	dataScope := req.FormValue("scope")
	if dataScope == "" {
		logger.Log.Error("No Scope for Data")
		return
	}
	userClaims, ok := ctxmanage.RetrieveClaims(req.Context())
	if !ok {
		logger.Log.Error("cannot find claims in context")
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}

	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("parse multi past form ", zap.Error(err))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			logger.Log.Info("Import File Handler", zap.String("File", hdr.Filename), zap.String("uploadedBy", uploadedBy))
			if !helper.RegexContains(i.config.Upload.DataFileAllowedRegex, hdr.Filename) {
				logger.Log.Error("Validation Error-File Not allowed", zap.String("File", hdr.Filename))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			// open uploaded
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("cannot open file directory", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			// open destination
			var outfile *os.File
			fn := filepath.Join(i.config.Upload.UploadDir, dataScope+"_"+hdr.Filename)

			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Error("cannot create file", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
				return
			}
			if _, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				// if all contents are not copied remove the files
				if err := os.Remove(fn); err != nil {
					logger.Log.Error("cannot remove", zap.Error(err))
					http.Error(res, "cannot upload Error", http.StatusInternalServerError)
					removeFiles(dataScope, i.config.Upload.UploadDir, dataload)
					return
				}
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				outfile.Close()
				return
			}
			outfile.Close()
			filenames = append(filenames, fmt.Sprintf("%s_%s", dataScope, hdr.Filename))
		}
		log.Println(" Sending data to dps service, scope [", dataScope, "] files [", filenames, "] by [", uploadedBy, "] len of files ", len(filenames))
		// Notify call to DPS
		resp, err := i.dpsClient.NotifyUpload(req.Context(), &v1.NotifyUploadRequest{
			Scope:      dataScope,
			Type:       "data",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
		}
		logger.Log.Info("Incoming response", zap.Any("response", resp))
		res.Write([]byte("Files Uploaded"))
	}
}

func (i *importServiceServer) UploadMetaDataHandler(res http.ResponseWriter, req *http.Request, param httprouter.Params) {
	userClaims, ok := ctxmanage.RetrieveClaims(req.Context())
	if !ok {
		logger.Log.Error("cannot find claims in context")
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}

	uploadedBy := userClaims.UserID
	// const _24K = (1 << 20) * 24
	if err := req.ParseMultipartForm(32 << 20); nil != err {
		logger.Log.Error("parse multi past form ", zap.Error(err))
		http.Error(res, "cannot store files", http.StatusInternalServerError)
		return
	}
	err := os.MkdirAll(i.config.Upload.UploadDir, os.ModePerm)
	if err != nil {
		logger.Log.Error("Cannot create Dir", zap.Error(err))
		http.Error(res, "cannot upload Error", http.StatusInternalServerError)
		return
	}
	var filenames []string
	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			logger.Log.Info("Import MetaData File Handler", zap.String("File", hdr.Filename), zap.String("uploadedBy", uploadedBy))
			if !helper.RegexContains(i.config.Upload.MetaDatafileAllowedRegex, hdr.Filename) {
				logger.Log.Error("Validation Error-File Not allowed", zap.Any("Regex", i.config.Upload.MetaDatafileAllowedRegex), zap.String("File", hdr.Filename))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			// open uploaded
			infile, err := hdr.Open()
			if err != nil {
				logger.Log.Error("cannot open file directory", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			// open destination
			var outfile *os.File
			fn := filepath.Join(i.config.Upload.UploadDir, ""+hdr.Filename)

			if outfile, err = os.Create(fn); nil != err {
				logger.Log.Error("cannot create file", zap.Error(err))
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				removeFiles("", i.config.Upload.UploadDir, metadataload)
				return
			}
			if _, err = io.Copy(outfile, infile); nil != err {
				logger.Log.Error("cannot copy content of files", zap.Error(err))
				// if all contents are not copied remove the files
				if err := os.Remove(fn); err != nil {
					logger.Log.Error("cannot remove", zap.Error(err))
					http.Error(res, "cannot upload Error", http.StatusInternalServerError)
					removeFiles("", i.config.Upload.UploadDir, metadataload)
					return
				}
				http.Error(res, "cannot upload Error", http.StatusInternalServerError)
				outfile.Close()
				return
			}
			outfile.Close()
			filenames = append(filenames, fmt.Sprintf("%s", hdr.Filename))
		}
		log.Println(" Sending data to dps service, scope [", "", "] files [", filenames, "] by [", uploadedBy, "] len of files ", len(filenames))
		// Notify call to DPS
		resp, err := i.dpsClient.NotifyUpload(req.Context(), &v1.NotifyUploadRequest{
			Scope:      "",
			Type:       "metadata",
			Files:      filenames,
			UploadedBy: uploadedBy,
		})
		if err != nil {
			logger.Log.Error("DPS call failed", zap.Error(err))
		}
		logger.Log.Info("Incoming response", zap.Any("response", resp))
		res.Write([]byte("Files Uploaded"))
	}
}

func removeFiles(scope string, dir string, datatype uploadType) {
	logger.Log.Info("Removing Files", zap.String("Scope", scope))
	var delFilesRegex string
	if datatype == "data" {
		delFilesRegex = scope + "_*"
	} else {
		delFilesRegex = "metadata_*"
	}
	files, err := filepath.Glob(path.Join(dir, delFilesRegex))
	if err != nil {
		logger.Log.Error("Failed to list files", zap.Error(err))
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			logger.Log.Error("Failed to list files", zap.Error(err))
		}
	}
}
