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

func (s *productServiceServer) ListDeployedAndAcquiredEditors(ctx context.Context, req *v1.ListDeployedAndAcquiredEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.ListDeployedAndAcquiredEditors(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ListDeployedAndAcquiredEditors - db/ListDeployedAndAcquiredEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListEditorsResponse{Editors: dbresp}, nil
}

func (s *productServiceServer) GetRightsInfoByEditor(ctx context.Context, req *v1.GetRightsInfoByEditorRequest) (*v1.GetRightsInfoByEditorResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.productRepo.GetAcqRightsByEditor(ctx, db.GetAcqRightsByEditorParams{
		ProductEditor: req.Editor,
		Scope:         req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetRightsInfoByEditor - GetAcqRightsByEditor", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	dbresp1, err := s.productRepo.GetAggregationByEditor(ctx, db.GetAggregationByEditorParams{
		ProductEditor: req.Editor,
		Scope:         req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetRightsInfoByEditor -GetAggregationByEditor", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetRightsInfoByEditorResponse{
		EditorRights: dbRightsInfoToSrvRightsInfoAll(dbresp, dbresp1),
	}, nil
}

func dbRightsInfoToSrvRightsInfoAll(indRightsInfo []db.GetAcqRightsByEditorRow, aggRightsInfo []db.GetAggregationByEditorRow) []*v1.RightsInfoByEditor {
	servRightsInfo := make([]*v1.RightsInfoByEditor, 0, len(indRightsInfo)+len(aggRightsInfo))
	for _, ri := range indRightsInfo {
		servRightsInfo = append(servRightsInfo, dbIndRightsInfoToSrvRightsInfo(ri))
	}
	for _, ri := range aggRightsInfo {
		servRightsInfo = append(servRightsInfo, dbAggRightsInfoToSrvRightsInfo(ri))
	}
	return servRightsInfo
}

func dbIndRightsInfoToSrvRightsInfo(rightsInfo db.GetAcqRightsByEditorRow) *v1.RightsInfoByEditor {
	resp := &v1.RightsInfoByEditor{
		Sku:                 rightsInfo.Sku,
		Swidtag:             rightsInfo.Swidtag,
		MetricName:          rightsInfo.Metric,
		AvgUnitPrice:        rightsInfo.AvgUnitPrice,
		NumLicensesAcquired: rightsInfo.NumLicensesAcquired,
	}
	return resp
}

func dbAggRightsInfoToSrvRightsInfo(rightsInfo db.GetAggregationByEditorRow) *v1.RightsInfoByEditor {
	resp := &v1.RightsInfoByEditor{
		Sku:                 rightsInfo.Sku,
		Swidtag:             rightsInfo.Swidtags,
		MetricName:          rightsInfo.Metric,
		AvgUnitPrice:        rightsInfo.AvgUnitPrice,
		NumLicensesAcquired: rightsInfo.NumLicensesAcquired,
		AggregationName:     rightsInfo.AggregationName,
	}
	return resp
}
