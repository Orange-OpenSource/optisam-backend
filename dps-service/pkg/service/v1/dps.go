package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	appV1 "optisam-backend/application-service/pkg/api/v1"
	equipV1 "optisam-backend/equipment-service/pkg/api/v1"
	prodV1 "optisam-backend/product-service/pkg/api/v1"
	"os"
	"strconv"
	"strings"
	"time"

	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	job "optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/config"
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
	queue       workerqueue.Workerqueue
	application appV1.ApplicationServiceClient
	equipment   equipV1.EquipmentServiceClient
	product     prodV1.ProductServiceClient
	account     accv1.AccountServiceClient
}

// NewDpsServiceServer creates Application service
func NewDpsServiceServer(dpsRepo repo.Dps, queue workerqueue.Workerqueue, grpcServers map[string]*grpc.ClientConn) v1.DpsServiceServer {
	return &dpsServiceServer{
		dpsRepo:     dpsRepo,
		queue:       queue,
		application: appV1.NewApplicationServiceClient(grpcServers["application"]),
		equipment:   equipV1.NewEquipmentServiceClient(grpcServers["equipment"]),
		product:     prodV1.NewProductServiceClient(grpcServers["product"]),
		account:     accv1.NewAccountServiceClient(grpcServers["account"]),
	}
}

// TODO This is aysnc , will be converted into sync
func (d *dpsServiceServer) DropUploadedFileData(ctx context.Context, req *v1.DropUploadedFileDataRequest) (*v1.DropUploadedFileDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropUploadedFileDataResponse{
			Success: false,
		}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DropUploadedFileDataResponse{
			Success: false,
		}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DropUploadedFileDataResponse{Success: false}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	if err := d.dpsRepo.DropFileRecords(ctx, req.Scope); err != nil {
		logger.Log.Error("Failed to delete file records", zap.Error(err))
		return &v1.DropUploadedFileDataResponse{
			Success: false,
		}, status.Error(codes.Internal, err.Error())
	}
	return &v1.DropUploadedFileDataResponse{
		Success: true,
	}, nil
}

// ListFailedRecord
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
		for k, i := range resp {
			val := ""
			switch v := i.(type) {
			case int, int32, int64:
				val = fmt.Sprintf("%d", v)
			case float64, float32:
				val = fmt.Sprintf("%f", v)
			case string:
				val = v
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

// NotifyUpload tells dps to process a batch of files of a scope
func (d *dpsServiceServer) NotifyUpload(ctx context.Context, req *v1.NotifyUploadRequest) (*v1.NotifyUploadResponse, error) { //nolint
	out := v1.NotifyUploadResponse{Success: true}
	var activeGID int32
	var err error
	out.FileUploadId = make(map[string]int32)
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if _, ok := d.isDeletionActive(ctx, req.Scope, "NA", "NA", false); ok {
		return nil, status.Error(codes.FailedPrecondition, "Deletion is already running")
	}
	if req.GetUploadedBy() == "nifi" {
		activeGID, err = d.dpsRepo.GetActiveGID(ctx, req.Scope)
		if err != nil && err != sql.ErrNoRows {
			logger.Log.Error("Failed to get active GID", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		} else if activeGID == int32(0) {
			return nil, status.Error(codes.Internal, "UnLinkedTransformedDataFile")
		}
		logger.Log.Debug("Active ", zap.Any("GID", activeGID))
	} else if d.isInjectionActive(ctx, req.Scope) {
		return nil, status.Error(codes.FailedPrecondition, "Injection is already running")
	}
	fileStatus := db.UploadStatusPENDING
	var datatype db.DataType
	if req.GetType() == strings.ToLower(constants.METADATA) { // nolint: gocritic
		datatype = db.DataTypeMETADATA
	} else if req.GetType() == strings.ToLower(constants.GLOBALDATA) {
		datatype = db.DataTypeGLOBALDATA
		fileStatus = db.UploadStatusUPLOADED
	} else {
		datatype = db.DataTypeDATA
	}
	scopeType := db.ScopeTypesGENERIC
	if req.GetScopeType() != v1.NotifyUploadRequest_GENERIC {
		scopeType = db.ScopeTypesSPECIFIC
	}

	for _, file := range req.GetFiles() {
		if strings.TrimSpace(file) == "" {
			continue
		}
		var fileToSave string
		var isNifi bool
		var gid int32

		temp := strings.Split(file, constants.NifiFileDelimeter)
		if len(temp) == 3 {
			fileToSave = temp[2]
			isNifi = true
			val, _ := strconv.ParseInt(strings.Split(temp[1], "_")[0], 10, 32)
			gid = int32(val)
			logger.Log.Debug("This data file belongs to ", zap.Any("gid", gid), zap.Any("dataFile", file))
			if gid != activeGID {
				return nil, status.Error(codes.FailedPrecondition, "Injection is already running")
			}
		} else {
			fileToSave = file
		}

		// TODO will go in import service in future also handle txn for multiple files if one fails
		dbresp, err := d.dpsRepo.InsertUploadedData(ctx, db.InsertUploadedDataParams{
			FileName:   fileToSave,
			DataType:   datatype,
			Scope:      req.GetScope(),
			UploadedBy: req.GetUploadedBy(),
			Gid:        gid,
			Status:     fileStatus,
			ScopeType:  scopeType,
			ErrorFile:  sql.NullString{String: req.AnalyzedErrorFile, Valid: true},
		})
		if err != nil {
			logger.Log.Debug("Failed to insert file record in dps, err :", zap.Error(err), zap.Any("file", fileToSave))
			return nil, status.Error(codes.Internal, "DBError")
		}
		if isNifi {
			dbresp.FileName = file
		}
		if datatype != db.DataTypeGLOBALDATA {
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
		} else {
			out.FileUploadId[fileToSave] = dbresp.UploadID
		}
	}

	return &out, nil
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
		Gid:            req.GetGlobalFileId(),
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
		// API expect pagenum from 1 but the offset in DB starts with 0
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

// nolint: gocyclo
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
	scopeinfo, err := d.account.GetScope(ctx, &accv1.GetScopeRequest{Scope: req.Scope})
	if err != nil {
		logger.Log.Error("service/v1 - ListUploadMetaData - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "unable to fetch scope info")
	}
	if scopeinfo.ScopeType == accv1.ScopeType_GENERIC.String() {
		return nil, status.Error(codes.PermissionDenied, "can not fetch list of metadata uploaded for generic scope")
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
		// API expect pagenum from 1 but the offset in DB starts with 0
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

// TODO This is aysnc , will be converted into sync
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

	if d.isInjectionActive(ctx, req.Scope) {
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.FailedPrecondition, "Injection is already running")
	}

	deletionID, ok := d.isDeletionActive(ctx, req.Scope, req.DeletionType.String(), userClaims.UserID, true)
	if ok {
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.FailedPrecondition, "Deletion is already running")
	}

	var g errgroup.Group
	// DropApplicationData
	if req.DeletionType == v1.DeleteInventoryRequest_PARK || req.DeletionType == v1.DeleteInventoryRequest_FULL {
		g.Go(func() error {
			_, err := d.application.DropApplicationData(ctx, &appV1.DropApplicationDataRequest{Scope: req.Scope})
			return err
		})
		// DropEquipmentData
		g.Go(func() error {
			// grpc timeout increased for drop equipment
			ctx1, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
			defer cancel()
			_, err := d.equipment.DropEquipmentData(ctx1, &equipV1.DropEquipmentDataRequest{Scope: req.Scope})
			return err
		})
	}
	var dtype prodV1.DropProductDataRequestDeletionTypes
	if req.DeletionType == v1.DeleteInventoryRequest_PARK {
		dtype = prodV1.DropProductDataRequest_PARK
	} else if req.DeletionType == v1.DeleteInventoryRequest_ACQRIGHTS {
		dtype = prodV1.DropProductDataRequest_ACQRIGHTS
	} else if req.DeletionType == v1.DeleteInventoryRequest_FULL {
		dtype = prodV1.DropProductDataRequest_FULL
	}
	// DropProductData
	g.Go(func() error {
		_, err := d.product.DropProductData(ctx, &prodV1.DropProductDataRequest{Scope: req.Scope, DeletionType: dtype})
		return err
	})

	if err := g.Wait(); err != nil {
		if err = d.dpsRepo.UpdateDeletionStatus(ctx, db.UpdateDeletionStatusParams{
			Status: db.UploadStatusFAILED,
			Reason: sql.NullString{String: err.Error(), Valid: true},
			ID:     deletionID}); err != nil {
			logger.Log.Error("Failed to update deletion  status ", zap.Any("scope", req.Scope), zap.Error(err))
		}
		return &v1.DeleteInventoryResponse{
			Success: false,
		}, status.Error(codes.Internal, "InternalError")
	}

	if err := d.dpsRepo.UpdateDeletionStatus(ctx, db.UpdateDeletionStatusParams{
		Status: db.UploadStatusSUCCESS,
		Reason: sql.NullString{String: "", Valid: true},
		ID:     deletionID}); err != nil {
		logger.Log.Error("Failed to update deletion  status ", zap.Any("scope", req.Scope), zap.Error(err))
	}

	if req.DeletionType != v1.DeleteInventoryRequest_ACQRIGHTS {
		if resp, err := d.product.CreateDashboardUpdateJob(ctx, &prodV1.CreateDashboardUpdateJobRequest{Scope: req.Scope}); err != nil || !resp.Success {
			logger.Log.Error("Failed to create push job", zap.Error(err))
			return &v1.DeleteInventoryResponse{
				Success: false,
			}, status.Error(codes.Internal, "PushJobFailure")
		}
	}
	// Send api call for licCal job
	return &v1.DeleteInventoryResponse{
		Success: true,
	}, nil
}

// nolint: gocyclo
func (d *dpsServiceServer) ListUploadGlobalData(ctx context.Context, req *v1.ListUploadRequest) (*v1.ListUploadResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	dbresp, err := d.dpsRepo.ListUploadedGlobalDataFiles(ctx, db.ListUploadedGlobalDataFilesParams{
		Scope:          []string{req.Scope},
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
		// API expect pagenum from 1 but the offset in DB starts with 0
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
		apiresp.Uploads[i].FileName = dbresp[i].FileName

		if dbresp[i].ErrorFile.String != "" {
			errFile := fmt.Sprintf("%s/%s/errors/%s", config.GetConfig().RawdataLocation, dbresp[i].Scope, dbresp[i].ErrorFile.String)
			if _, err := os.Stat(errFile); err != nil {
				logger.Log.Error("Error File is not generated", zap.Any("uid", dbresp[i].UploadID), zap.Any("errfile", errFile), zap.Error(err))
			} else {
				apiresp.Uploads[i].ErrorFileApi = fmt.Sprintf("/api/v1/import/download?fileName=%s&downloadType=error&scope=%s", dbresp[i].ErrorFile.String, req.Scope)
			}
		}

		apiresp.Uploads[i].Status = string(dbresp[i].Status)
		apiresp.Uploads[i].UploadedBy = dbresp[i].UploadedBy
		apiresp.Uploads[i].UploadedOn, _ = ptypes.TimestampProto(dbresp[i].UploadedOn)
		apiresp.Uploads[i].Comments = dbresp[i].Comments.String
	}
	return apiresp, nil
}

func (d *dpsServiceServer) ListDeletionRecords(ctx context.Context, req *v1.ListDeletionRequest) (*v1.ListDeletionResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := d.dpsRepo.ListDeletionRecrods(ctx, db.ListDeletionRecrodsParams{
		Scope:            req.Scope,
		DeletionTypeAsc:  strings.Contains(req.GetSortBy().String(), "deletion_type") && strings.Contains(req.GetSortOrder().String(), "asc"),
		DeletionTypeDesc: strings.Contains(req.GetSortBy().String(), "deletion_type") && strings.Contains(req.GetSortOrder().String(), "desc"),
		StatusAsc:        strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "asc"),
		StatusDesc:       strings.Contains(req.GetSortBy().String(), "status") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CreatedByAsc:     strings.Contains(req.GetSortBy().String(), "created_by") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CreatedByDesc:    strings.Contains(req.GetSortBy().String(), "created_by") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CreatedOnAsc:     strings.Contains(req.GetSortBy().String(), "created_on") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CreatedOnDesc:    strings.Contains(req.GetSortBy().String(), "created_on") && strings.Contains(req.GetSortOrder().String(), "desc"),
		PageNum:          req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:         req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("Failed to get deleted records ", zap.Error(err))
		if err != sql.ErrNoRows {
			return &v1.ListDeletionResponse{}, status.Error(codes.Unknown, "NoContent")
		}
		return &v1.ListDeletionResponse{}, status.Error(codes.Unknown, "DBError")
	}

	apiresp := &v1.ListDeletionResponse{}
	apiresp.Deletions = make([]*v1.Deletion, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.Deletions[i] = &v1.Deletion{}
		if dbresp[i].DeletionType == db.DeletionTypeACQRIGHTS {
			apiresp.Deletions[i].DeletionType = "Acquired Rights"
		} else if dbresp[i].DeletionType == db.DeletionTypeINVENTORYPARK {
			apiresp.Deletions[i].DeletionType = "Inventory Park"
		} else if dbresp[i].DeletionType == db.DeletionTypeWHOLEINVENTORY {
			apiresp.Deletions[i].DeletionType = "Whole Inventory"
		}
		apiresp.Deletions[i].Status = string(dbresp[i].Status)
		apiresp.Deletions[i].CreatedBy = dbresp[i].CreatedBy
		apiresp.Deletions[i].CreatedOn, _ = ptypes.TimestampProto(dbresp[i].CreatedOn)
	}
	return apiresp, nil
}

func (d *dpsServiceServer) isDeletionActive(ctx context.Context, scope, deletionType, userID string, set bool) (int32, bool) {
	var id int32
	count, err := d.dpsRepo.GetDeletionStatus(ctx, scope)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error(" GetDeletionActive failed", zap.Error(err))
		return id, true
	} else if int(count) > 0 {
		return id, true
	} else if set {
		var dtype db.DeletionType
		if deletionType == v1.DeleteInventoryRequest_PARK.String() {
			dtype = db.DeletionTypeINVENTORYPARK
		} else if deletionType == v1.DeleteInventoryRequest_ACQRIGHTS.String() {
			dtype = db.DeletionTypeACQRIGHTS
		} else if deletionType == v1.DeleteInventoryRequest_FULL.String() {
			dtype = db.DeletionTypeWHOLEINVENTORY
		} else {
			logger.Log.Error("UnknownType received", zap.Any("type", deletionType))
			return id, true
		}
		if id, err = d.dpsRepo.SetDeletionActive(ctx, db.SetDeletionActiveParams{
			Scope:        scope,
			DeletionType: dtype,
			CreatedBy:    userID,
		}); err != nil {
			logger.Log.Error("SetDeletionActive Failed", zap.Error(err))
			return id, true
		}
	}
	return id, false
}

func (d *dpsServiceServer) isInjectionActive(ctx context.Context, scope string) bool {
	count, err := d.dpsRepo.GetInjectionStatus(ctx, scope)
	if err != nil {
		logger.Log.Error("isInjectionActive failed", zap.Error(err))
		return true
	} else if int(count) > 0 {
		return true
	}
	return false
}
