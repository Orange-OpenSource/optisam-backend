package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServiceServer) ListEditors(ctx context.Context, req *v1.ListEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.ListEditors(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListEditorsResponse{Editors: dbresp}, nil
}

func (s *productServiceServer) ListEditorProducts(ctx context.Context, req *v1.ListEditorProductsRequest) (*v1.ListEditorProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetProductsByEditor(ctx, db.GetProductsByEditorParams{ProductEditor: req.Editor, Scopes: req.Scopes})
	if err != nil {
		logger.Log.Error("service/v1 - ListEditorProducts - ListEditorProducts", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.ListEditorProductsResponse{}
	apiresp.Products = make([]*v1.Product, len(dbresp))
	for i := range dbresp {
		apiresp.Products[i] = &v1.Product{}
		apiresp.Products[i].SwidTag = dbresp[i].Swidtag
		apiresp.Products[i].Name = dbresp[i].ProductName
		apiresp.Products[i].Version = dbresp[i].ProductVersion
	}
	return &apiresp, nil

}
