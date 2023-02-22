package v1

import (
	"context"
	"database/sql"
	v1 "optisam-backend/catalog-service/pkg/api/v1"
	"optisam-backend/catalog-service/pkg/repository/v1/postgres/db"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"

	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *productCatalogServer) BulkFileUpload(ctx context.Context, req *v1.UploadRecords) (res *v1.UploadResponse, err error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.UploadResponse{Message: "ClaimsNotFound"}, status.Error(codes.Internal, "ClaimsNotFound")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.UploadResponse{Message: "RoleValidationError"}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}
	msg, err := p.productRepo.InsertRecordsTx(ctx, req)

	_ = p.productRepo.CreateUploadFileLog(ctx, db.CreateUploadFileLogParams{
		FileName: req.FileName,
		Message:  sql.NullString{String: msg, Valid: true},
	})

	return &v1.UploadResponse{Message: msg}, err
}

func (p *productCatalogServer) BulkFileUploadLogs(ctx context.Context, req *v1.UploadCatalogDataLogsRequest) (*v1.UploadCatalogDataLogsResponse, error) { //nolint
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "UnAuthorisedUser")
	}
	dbresp, err := p.productRepo.GetUploadFileLogs(ctx)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error("Failed to get core factor logs", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "InternalError")
	}
	var output []*v1.UploadCatalogDataLogs // nolint
	for _, v := range dbresp {
		out := &v1.UploadCatalogDataLogs{}
		out.Filename = v.FileName
		out.UploadedOn, _ = ptypes.TimestampProto(v.UploadedOn)

		output = append(output, out)
	}
	return &v1.UploadCatalogDataLogsResponse{
		UploadCatalogDataLogs: output,
	}, nil
}
