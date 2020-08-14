// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"strings"

	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	worker "optisam-backend/common/optisam/workerqueue"
	job "optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	repo "optisam-backend/dps-service/pkg/repository/v1"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dpsServiceServer struct {
	dpsRepo repo.Dps
	queue   worker.Queue
}

// NewDpsServiceServer creates Application service
func NewDpsServiceServer(dpsRepo repo.Dps, queue worker.Queue) v1.DpsServiceServer {
	return &dpsServiceServer{dpsRepo: dpsRepo, queue: queue}
}

//NotifyUpload tells dps to process a batch of files of a scope
func (d *dpsServiceServer) NotifyUpload(ctx context.Context, req *v1.NotifyUploadRequest) (*v1.NotifyUploadResponse, error) {

	var datatype db.DataType
	if req.GetType() == strings.ToLower(constants.METADATA) {
		datatype = db.DataTypeMETADATA
	} else {
		datatype = db.DataTypeDATA
	}
	for _, file := range req.GetFiles() {
		if strings.TrimSpace(file) == "" {
			continue
		}

		//TODO will go in import service in future
		dbresp, err := d.dpsRepo.InsertUploadedData(ctx, db.InsertUploadedDataParams{
			FileName:   file,
			DataType:   datatype,
			Scope:      req.GetScope(),
			UploadedBy: req.GetUploadedBy(),
		})
		if err != nil {
			log.Println("Failed to insert file record in dps, err :", err)
			continue
		}
		logger.Log.Info("service", zap.Int32("uploadID", dbresp.UploadID))

		//Async Job Submission
		dataForJob, err := json.Marshal(dbresp)
		if err != nil {
			log.Println("Failed to marshal notifyPayload data for file type job, err:", err)
			continue
		}
		job := job.Job{
			Type:   constants.FILETYPE,
			Data:   dataForJob,
			Status: job.JobStatusPENDING,
		}
		_, err = d.queue.PushJob(ctx, job, constants.FILEWORKER)
		if err != nil {
			log.Println("Failed to push job for file :", file, " for scope ", req.GetScope(), " err : ", err)
			continue
		}

	}
	return &v1.NotifyUploadResponse{Success: true}, nil
}

func (d *dpsServiceServer) ListUploadData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	var err error
	dbresp, err := d.dpsRepo.ListUploadedDataFiles(ctx, db.ListUploadedDataFilesParams{
		Scope:          userClaims.Socpes,
		UploadIDAsc:    strings.Contains(req.GetSortBy().String(), "upload_id") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadIDDesc:   strings.Contains(req.GetSortBy().String(), "upload_id") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ScopeAsc:       strings.Contains(req.GetSortBy().String(), "scope") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ScopeDesc:      strings.Contains(req.GetSortBy().String(), "scope") && strings.Contains(req.GetSortOrder().String(), "desc"),
		FileNameAsc:    strings.Contains(req.GetSortBy().String(), "file_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		FileNameDesc:   strings.Contains(req.GetSortBy().String(), "file_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		StatusAsc:      strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "asc"),
		StatusDesc:     strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "desc"),
		UploadedByAsc:  strings.Contains(req.GetSortBy().String(), "uploaded_by") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadedByDesc: strings.Contains(req.GetSortBy().String(), "uploaded_by") && strings.Contains(req.GetSortOrder().String(), "desc"),
		UploadedOnAsc:  strings.Contains(req.GetSortBy().String(), "uploaded_on") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadedOnDesc: strings.Contains(req.GetSortBy().String(), "uploaded_on") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListUploadResponse{}, nil
		}
		return &v1.ListUploadResponse{}, status.Error(codes.Unknown, "DBError")
	}

	apiresp := &v1.ListUploadResponse{}
	apiresp.Uploads = make([]*v1.Upload, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Uploads[i] = &v1.Upload{}
		apiresp.Uploads[i].UploadId = dbresp[i].UploadID
		apiresp.Uploads[i].Scope = dbresp[i].Scope
		apiresp.Uploads[i].Status = string(dbresp[i].Status)
		apiresp.Uploads[i].FileName = dbresp[i].FileName
		apiresp.Uploads[i].UploadedBy = dbresp[i].UploadedBy
		apiresp.Uploads[i].UploadedOn, _ = ptypes.TimestampProto(dbresp[i].UploadedOn)
		apiresp.Uploads[i].FailedRecords = dbresp[i].FailedRecords
		apiresp.Uploads[i].SuccessRecords = dbresp[i].SuccessRecords
		apiresp.Uploads[i].TotalRecords = dbresp[i].TotalRecords
	}
	return apiresp, nil
}

func (d *dpsServiceServer) ListUploadMetaData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {

	var err error
	dbresp, err := d.dpsRepo.ListUploadedMetaDataFiles(ctx, db.ListUploadedMetaDataFilesParams{
		UploadIDAsc:    strings.Contains(req.GetSortBy().String(), "upload_id") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadIDDesc:   strings.Contains(req.GetSortBy().String(), "upload_id") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ScopeAsc:       strings.Contains(req.GetSortBy().String(), "scope") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ScopeDesc:      strings.Contains(req.GetSortBy().String(), "scope") && strings.Contains(req.GetSortOrder().String(), "desc"),
		FileNameAsc:    strings.Contains(req.GetSortBy().String(), "file_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		FileNameDesc:   strings.Contains(req.GetSortBy().String(), "file_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		StatusAsc:      strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "asc"),
		StatusDesc:     strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "desc"),
		UploadedByAsc:  strings.Contains(req.GetSortBy().String(), "uploaded_by") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadedByDesc: strings.Contains(req.GetSortBy().String(), "uploaded_by") && strings.Contains(req.GetSortOrder().String(), "desc"),
		UploadedOnAsc:  strings.Contains(req.GetSortBy().String(), "uploaded_on") && strings.Contains(req.GetSortOrder().String(), "asc"),
		UploadedOnDesc: strings.Contains(req.GetSortBy().String(), "uploaded_on") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return &v1.ListUploadResponse{}, nil
		}
		return &v1.ListUploadResponse{}, status.Error(codes.Unknown, "DBError")
	}

	apiresp := &v1.ListUploadResponse{}
	apiresp.Uploads = make([]*v1.Upload, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Uploads[i] = &v1.Upload{}
		apiresp.Uploads[i].UploadId = dbresp[i].UploadID
		apiresp.Uploads[i].Scope = dbresp[i].Scope
		apiresp.Uploads[i].Status = string(dbresp[i].Status)
		apiresp.Uploads[i].FileName = dbresp[i].FileName
		apiresp.Uploads[i].UploadedBy = dbresp[i].UploadedBy
		apiresp.Uploads[i].UploadedOn, _ = ptypes.TimestampProto(dbresp[i].UploadedOn)
		apiresp.Uploads[i].FailedRecords = dbresp[i].FailedRecords
		apiresp.Uploads[i].SuccessRecords = dbresp[i].SuccessRecords
		apiresp.Uploads[i].TotalRecords = dbresp[i].TotalRecords
	}
	return apiresp, nil
}
