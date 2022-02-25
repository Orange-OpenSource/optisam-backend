package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"

	"github.com/golang/protobuf/ptypes"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func coreFactorCached() {
	mu.Lock()
	defer mu.Unlock()
	isCoreFactorStored = true
}

func coreFactorUncached() {
	mu.Lock()
	defer mu.Unlock()
	isCoreFactorStored = false
}

func isCoreFactorCached() bool {
	mu.Lock()
	defer mu.Unlock()
	return isCoreFactorStored
}

// StoreCoreFactorReference saves the core factor reference
func (d *dpsServiceServer) StoreCoreFactorReference(ctx context.Context, req *v1.StoreReferenceDataRequest) (*v1.StoreReferenceDataResponse, error) { //nolint
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "UnAuthorisedUser")
	}
	data := make(map[string]map[string]string)
	if err := json.Unmarshal(req.ReferenceData, &data); err != nil {
		logger.Log.Error("Failed to Unmarshal the reference Data ", zap.Error(err))
		return nil, status.Error(codes.Internal, "InternalServerError")
	}
	if err := d.dpsRepo.DeleteCoreFactorReference(ctx); err != nil {
		logger.Log.Error("Failed to delete old reference Data : DB Error", zap.Error(err))
		return nil, status.Error(codes.Internal, "InternalServerError")
	}
	if err := d.dpsRepo.StoreCoreFactorReferences(ctx, data); err != nil {
		logger.Log.Error("Failed to store reference Data : DB Error", zap.Error(err))
		return nil, status.Error(codes.Internal, "InternalServerError")
	}

	if err := d.dpsRepo.LogCoreFactor(ctx, req.Filename); err != nil {
		logger.Log.Error("Failed to log reference file : DB Error", zap.Error(err))
		return nil, status.Error(codes.Internal, "InternalServerError")
	}
	coreFactorUncached()

	return &v1.StoreReferenceDataResponse{Success: true}, nil
}

// ViewFactorReference tells dps to process a batch of files of a scope
func (d *dpsServiceServer) ViewFactorReference(ctx context.Context, req *v1.ViewReferenceDataRequest) (*v1.ViewReferenceDataResponse, error) { //nolint
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "UnAuthorisedUser")
	}
	dbresp, err := d.dpsRepo.GetCoreFactorReferences(ctx, db.GetCoreFactorReferencesParams{
		Limit:  req.PageSize,
		Offset: (req.PageNo - 1) * req.PageSize,
	})
	if err != nil {
		logger.Log.Error("Failed to get corefactor references", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	var references []*v1.CoreFactorReference // nolint
	for _, v := range dbresp {
		reference := &v1.CoreFactorReference{
			Manufacturer: v.Manufacturer,
			Model:        v.Model,
			Corefactor:   v.CoreFactor,
		}
		references = append(references, reference)

	}

	return &v1.ViewReferenceDataResponse{References: references, TotalRecord: int32(dbresp[0].TotalRecords)}, nil
}

// ViewFactorReference tells dps to process a batch of files of a scope
func (d *dpsServiceServer) ViewCoreFactorLogs(ctx context.Context, req *v1.ViewCoreFactorLogsRequest) (*v1.ViewCoreFactorLogsResponse, error) { //nolint
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "UnAuthorisedUser")
	}
	dbresp, err := d.dpsRepo.GetCoreFactorLogs(ctx)
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error("Failed to get core factor logs", zap.Error(err))
		return nil, status.Error(codes.PermissionDenied, "InternalError")
	}
	var output []*v1.CoreFactorlogs // nolint
	for _, v := range dbresp {
		out := &v1.CoreFactorlogs{}
		out.Filename = v.FileName
		out.UploadedOn, _ = ptypes.TimestampProto(v.UploadedOn)

		output = append(output, out)
	}
	return &v1.ViewCoreFactorLogsResponse{
		Corefactorlogs: output,
	}, nil
}
