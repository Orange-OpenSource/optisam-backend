// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/report-service/pkg/api/v1"
	repo "optisam-backend/report-service/pkg/repository/v1"
	"optisam-backend/report-service/pkg/repository/v1/postgres/db"
	"optisam-backend/report-service/pkg/worker"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReportServiceServer struct {
	reportRepo repo.Report
	queue      workerqueue.Workerqueue
}

// NewReportServiceServer creates Auth service
func NewReportServiceServer(reportRepo repo.Report, queue workerqueue.Workerqueue) v1.ReportServiceServer {
	return &ReportServiceServer{reportRepo: reportRepo, queue: queue}
}

func (r *ReportServiceServer) ListReportType(ctx context.Context, req *v1.ListReportTypeRequest) (*v1.ListReportTypeResponse, error) {
	dbresp, err := r.reportRepo.GetReportTypes(ctx)
	if err != nil {
		logger.Log.Error("Failed to fetch report types", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "DB Error")
	}
	apiresp := &v1.ListReportTypeResponse{}
	apiresp.ReportType = make([]*v1.ReportType, len(dbresp))
	for i := range dbresp {
		apiresp.ReportType[i] = &v1.ReportType{ReportTypeId: dbresp[i].ReportTypeID, ReportTypeName: dbresp[i].ReportTypeName}
	}
	return apiresp, nil
}

func (r *ReportServiceServer) SubmitReport(ctx context.Context, req *v1.SubmitReportRequest) (*v1.SubmitReportResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return &v1.SubmitReportResponse{Success: false}, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}

	reportType, err := r.reportRepo.GetReportType(ctx, req.GetReportTypeId())
	if err != nil {
		logger.Log.Error("Service/SubmitReport - Error in db/GetReportType", zap.Error(err))
		return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
	}

	switch reportType.ReportTypeID {
	case int32(1):
		_, ok := req.ReportMetadata.(*v1.SubmitReportRequest_AcqrightsReport)
		if !ok {
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.InvalidArgument, "Bad Report Request")
		}
		var j bytes.Buffer
		marshaler := &jsonpb.Marshaler{}
		err := marshaler.Marshal(&j, req.GetAcqrightsReport())
		if err != nil {
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.InvalidArgument, "Json Marshal Error")
		}
		reportID, err := r.reportRepo.SubmitReport(ctx, db.SubmitReportParams{
			Scope:          req.GetScope(),
			ReportStatus:   db.ReportStatusPENDING,
			CreatedBy:      userClaims.UserID,
			ReportTypeID:   req.GetReportTypeId(),
			ReportMetadata: j.Bytes(),
		})
		if err != nil {
			logger.Log.Error("Service/SubmitReport - Error in db/SubmitReport", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		e := worker.Envelope{Type: worker.AcqRightsReport, Scope: req.GetScope(), JSON: j.Bytes(), ReportID: reportID}
		envolveData, err := json.Marshal(e)
		if err != nil {
			logger.Log.Error("Failed to do json marshalling of worker envelope", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		jobID, err := r.queue.PushJob(ctx, job.Job{
			Type:   sql.NullString{String: "rw"},
			Status: job.JobStatusPENDING,
			Data:   envolveData,
		}, "rw")
		if err != nil {
			logger.Log.Error("Failed to push job to the queue", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	case int32(2):
		_, ok := req.ReportMetadata.(*v1.SubmitReportRequest_ProductEquipmentsReport)
		if !ok {
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.InvalidArgument, "Bad Report Request")
		}
		var j bytes.Buffer
		marshaler := &jsonpb.Marshaler{}
		err := marshaler.Marshal(&j, req.GetProductEquipmentsReport())
		if err != nil {
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.InvalidArgument, "Json Marshal Error")
		}
		reportID, err := r.reportRepo.SubmitReport(ctx, db.SubmitReportParams{
			Scope:          req.GetScope(),
			ReportStatus:   db.ReportStatusPENDING,
			CreatedBy:      userClaims.UserID,
			ReportTypeID:   req.GetReportTypeId(),
			ReportMetadata: j.Bytes(),
		})
		if err != nil {
			logger.Log.Error("Service/SubmitReport - Error in db/SubmitReport", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		e := worker.Envelope{Type: worker.ProductEquipmentsReport, Scope: req.GetScope(), JSON: j.Bytes(), ReportID: reportID}

		envolveData, err := json.Marshal(e)
		if err != nil {
			logger.Log.Error("Failed to do json marshalling of worker envelope", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		jobID, err := r.queue.PushJob(ctx, job.Job{
			Type:   sql.NullString{String: "rw"},
			Status: job.JobStatusPENDING,
			Data:   envolveData,
		}, "rw")
		if err != nil {
			logger.Log.Error("Failed to push job to the queue", zap.Error(err))
			return &v1.SubmitReportResponse{Success: false}, status.Error(codes.Internal, "Internal Server Error")
		}
		logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	default:
		return &v1.SubmitReportResponse{Success: false}, status.Error(codes.InvalidArgument, "Wrong ReportID sent")

	}
	return &v1.SubmitReportResponse{Success: true}, nil
}

func (r *ReportServiceServer) ListReport(ctx context.Context, req *v1.ListReportRequest) (*v1.ListReportResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := r.reportRepo.GetReport(ctx, db.GetReportParams{
		Scope:              userClaims.Socpes,
		ReportIDAsc:        strings.Contains(req.GetSortBy(), "report_id") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ReportIDDesc:       strings.Contains(req.GetSortBy(), "report_id") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ReportStatusAsc:    strings.Contains(req.GetSortBy(), "report_status") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ReportStatusDesc:   strings.Contains(req.GetSortBy(), "report_status") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ReportTypeNameAsc:  strings.Contains(req.GetSortBy(), "report_type") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ReportTypeNameDesc: strings.Contains(req.GetSortBy(), "report_type") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CreatedByAsc:       strings.Contains(req.GetSortBy(), "created_by") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CreatedByDesc:      strings.Contains(req.GetSortBy(), "created_by") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CreatedOnAsc:       strings.Contains(req.GetSortBy(), "created_on") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CreatedOnDesc:      strings.Contains(req.GetSortBy(), "created_on") && strings.Contains(req.GetSortOrder().String(), "desc"),
		PageNum:            req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:           req.GetPageSize(),
	})

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Reports-> "+err.Error())
	}

	apiresp := v1.ListReportResponse{}
	apiresp.Reports = make([]*v1.Report, len(dbresp))

	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		created_on, _ := ptypes.TimestampProto(dbresp[i].CreatedOn)
		apiresp.Reports[i] = &v1.Report{}
		apiresp.Reports[i].ReportId = dbresp[i].ReportID
		apiresp.Reports[i].ReportType = dbresp[i].ReportTypeName
		apiresp.Reports[i].ReportStatus = string(dbresp[i].ReportStatus)
		apiresp.Reports[i].CreatedBy = dbresp[i].CreatedBy
		apiresp.Reports[i].CreatedOn = created_on
	}
	return &apiresp, nil

}

func (r *ReportServiceServer) DownloadReport(ctx context.Context, req *v1.DownloadReportRequest) (*v1.DownloadReportResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	dbresp, err := r.reportRepo.DownloadReport(ctx, db.DownloadReportParams{ReportID: req.ReportID, Scope: userClaims.Socpes})

	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Reports-> "+err.Error())
	}
	dbrespBytes := dbresp

	apiresp := v1.DownloadReportResponse{ReportData: dbrespBytes}
	return &apiresp, nil
}
