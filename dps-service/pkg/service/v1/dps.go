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
	"fmt"
	appV1 "optisam-backend/application-service/pkg/api/v1"
	equipV1 "optisam-backend/equipment-service/pkg/api/v1"
	prodV1 "optisam-backend/product-service/pkg/api/v1"
	"strings"

	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	worker "optisam-backend/common/optisam/workerqueue"
	job "optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	repo "optisam-backend/dps-service/pkg/repository/v1"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	"optisam-backend/dps-service/pkg/worker/constants"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type dpsServiceServer struct {
	dpsRepo     repo.Dps
	queue       worker.Queue
	application appV1.ApplicationServiceClient
	equipment   equipV1.EquipmentServiceClient
	product     prodV1.ProductServiceClient
}

// NewDpsServiceServer creates Application service
func NewDpsServiceServer(dpsRepo repo.Dps, queue worker.Queue, grpcServers map[string]*grpc.ClientConn) v1.DpsServiceServer {
	return &dpsServiceServer{
		dpsRepo:     dpsRepo,
		queue:       queue,
		application: appV1.NewApplicationServiceClient(grpcServers["application"]),
		equipment:   equipV1.NewEquipmentServiceClient(grpcServers["equipment"]),
		product:     prodV1.NewProductServiceClient(grpcServers["product"]),
	}
}

//ListFailedRecord
func (d *dpsServiceServer) ListFailedRecord(ctx context.Context, req *v1.ListFailedRequest) (*v1.ListFailedResponse, error) {
	if req.PageSize == 0 {
		req.PageSize = 50
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	r := json.RawMessage(fmt.Sprintf("%d", req.GetUploadId()))
	dbresp, err := d.dpsRepo.GetFailedRecord(ctx, db.GetFailedRecordParams{
		Data:   r,
		Limit:  req.PageSize,
		Offset: (req.PageNum - 1) * req.PageSize,
	})
	if err != nil {
		logger.Log.Error("Failed to fetch failed record from DB ", zap.Error(err))
		return &v1.ListFailedResponse{}, status.Error(codes.Internal, "InternalError")
	}
	totalRecords := int32(0)
	out := []*v1.FailedRecord{}
	for _, tmp := range dbresp {
		temp := &v1.FailedRecord{}
		resp := make(map[string]interface{})
		err := json.Unmarshal(tmp.Record.([]byte), &resp)
		if err != nil {
			logger.Log.Error("Failed to fetch failed record from DB ", zap.Error(err))
			continue
		}
		temp.Reason = tmp.Comments.String
		totalRecords = int32(tmp.Totalrecords)
		temp.Data = make(map[string]string)
		for k, v := range resp {
			val := ""
			switch v.(type) {
			case int:
				val = fmt.Sprintf("%d", v)
			case float64:
				val = fmt.Sprintf("%f", v)
			case string:
				val = v.(string)
			default:
				x, _ := json.Marshal(v)
				val = string(x)
			}
			temp.Data[k] = val
		}
		out = append(out, temp)
	}

	return &v1.ListFailedResponse{FailedRecords: out, TotalRecords: totalRecords}, nil
}

//NotifyUpload tells dps to process a batch of files of a scope
func (d *dpsServiceServer) NotifyUpload(ctx context.Context, req *v1.NotifyUploadRequest) (*v1.NotifyUploadResponse, error) {
	var isDeletionStarted bool
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var datatype db.DataType
	if req.GetType() == strings.ToLower(constants.METADATA) {
		datatype = db.DataTypeMETADATA
	} else if req.GetType() == strings.ToLower(constants.GLOBALDATA) {
		datatype = db.DataTypeGLOBALDATA
	} else {
		datatype = db.DataTypeDATA
	}
	for _, file := range req.GetFiles() {
		if strings.TrimSpace(file) == "" {
			continue
		}

		//TODO will go in import service in future also handle txn for multiple files if one fails
		dbresp, err := d.dpsRepo.InsertUploadedData(ctx, db.InsertUploadedDataParams{
			FileName:   file,
			DataType:   datatype,
			Scope:      req.GetScope(),
			UploadedBy: req.GetUploadedBy(),
		})
		if err != nil {
			logger.Log.Debug("Failed to insert file record in dps, err :", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}

		if datatype == db.DataTypeGLOBALDATA {
			if req.GetIsDeleteOldInventory() && !isDeletionStarted {
				logger.Log.Debug("delete inventory is called for ", zap.String("scope", req.Scope))
				_, err := d.DeleteInventory(ctx, &v1.DeleteInventoryRequest{Scope: req.GetScope()})
				if err != nil {
					logger.Log.Debug("delete inventory call failed for ", zap.String("scope", req.Scope), zap.Error(err))
					return nil, status.Error(codes.Internal, "InventoryDeletionFailed")
				}
				isDeletionStarted = true
			}
		} else {
			dataForJob, err := json.Marshal(dbresp)
			if err != nil {
				logger.Log.Debug("Failed to marshal notifyPayload data for file type job, err:", zap.Error(err))
				continue
			}
			job := job.Job{
				Type:   constants.FILETYPE,
				Data:   dataForJob,
				Status: job.JobStatusPENDING,
			}
			_, err = d.queue.PushJob(ctx, job, constants.FILEWORKER)
			if err != nil {
				logger.Log.Debug("Failed to push the job ", zap.String("file", file), zap.String("fileType", req.GetType()))
			}
		}
	}
	return &v1.NotifyUploadResponse{Success: true}, nil
}

func (d *dpsServiceServer) ListUploadData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	var err error
	var scopes []string
	scopes = append(scopes, req.GetScope())
	if !helper.Contains(userClaims.Socpes, scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := d.dpsRepo.ListUploadedDataFiles(ctx, db.ListUploadedDataFilesParams{
		Scope:          scopes,
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
		apiresp.Uploads[i].Comments = dbresp[i].Comments.String
	}
	return apiresp, nil
}

func (d *dpsServiceServer) ListUploadMetaData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	var err error
	var scopes []string
	scopes = append(scopes, req.GetScope())
	if !helper.Contains(userClaims.Socpes, scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := d.dpsRepo.ListUploadedMetaDataFiles(ctx, db.ListUploadedMetaDataFilesParams{
		Scope:          scopes,
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
		apiresp.Uploads[i].Comments = dbresp[i].Comments.String

	}
	return apiresp, nil
}

//TODO This is aysnc , will be converted into sync
func (d *dpsServiceServer) DeleteInventory(ctx context.Context, req *v1.DeleteInventoryRequest) (*v1.DeleteInventoryResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var g errgroup.Group
	//DropApplicationData
	g.Go(func() error {
		_, err := d.application.DropApplicationData(ctx, &appV1.DropApplicationDataRequest{Scope: req.Scope})
		return err
	})
	//DropEquipmentData
	g.Go(func() error {
		_, err := d.equipment.DropEquipmentData(ctx, &equipV1.DropEquipmentDataRequest{Scope: req.Scope})
		return err
	})
	//DropProductData
	g.Go(func() error {
		_, err := d.product.DropProductData(ctx, &prodV1.DropProductDataRequest{Scope: req.Scope})
		return err
	})
	if err := g.Wait(); err != nil {
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.Internal, "InternalError")
	}
	return &v1.DeleteInventoryResponse{
		Success: true,
	}, nil
}

func (d *dpsServiceServer) ListUploadGlobalData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	var err error
	var scopes []string
	scopes = append(scopes, req.GetScope())
	if !helper.Contains(userClaims.Socpes, scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := d.dpsRepo.ListUploadedGlobalDataFiles(ctx, db.ListUploadedGlobalDataFilesParams{
		Scope:          scopes,
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
			return &v1.ListUploadResponse{}, status.Error(codes.Unknown, "NoContent")
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
	}
	return apiresp, nil
}
