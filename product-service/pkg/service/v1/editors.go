package v1

import (
	"context"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ProductServiceServer) ListEditors(ctx context.Context, req *v1.ListEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.ListEditors(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListEditorsResponse{Editors: dbresp}, nil
}

func (s *ProductServiceServer) ListEditorProducts(ctx context.Context, req *v1.ListEditorProductsRequest) (*v1.ListEditorProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	apiresp := v1.ListEditorProductsResponse{}
	if req.ParkInventory {
		dbresp, err := s.ProductRepo.GetProductsByEditorScope(ctx, db.GetProductsByEditorScopeParams{ProductEditor: req.Editor, Scopes: req.Scopes})
		if err != nil {
			logger.Log.Error("service/v1 - ListEditorProducts - ListEditorProducts", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		apiresp.Products = make([]*v1.Product, len(dbresp))
		for i := range dbresp {
			apiresp.Products[i] = &v1.Product{}
			apiresp.Products[i].SwidTag = dbresp[i].Swidtag
			apiresp.Products[i].Name = dbresp[i].ProductName
			apiresp.Products[i].Version = dbresp[i].ProductVersion
		}
	} else {
		dbresp, err := s.ProductRepo.GetProductsByEditor(ctx, db.GetProductsByEditorParams{ProductEditor: req.Editor, Scopes: req.Scopes})
		if err != nil {
			logger.Log.Error("service/v1 - ListEditorProducts - ListEditorProducts", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		apiresp.Products = make([]*v1.Product, len(dbresp))
		for i := range dbresp {
			apiresp.Products[i] = &v1.Product{}
			apiresp.Products[i].SwidTag = dbresp[i].Swidtag
			apiresp.Products[i].Name = dbresp[i].ProductName
			apiresp.Products[i].Version = dbresp[i].ProductVersion
		}
	}

	return &apiresp, nil

}

func (s *ProductServiceServer) ListDeployedAndAcquiredEditors(ctx context.Context, req *v1.ListDeployedAndAcquiredEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.ListDeployedAndAcquiredEditors(ctx, req.Scope)
	if err != nil {
		logger.Log.Error("service/v1 - ListDeployedAndAcquiredEditors - db/ListDeployedAndAcquiredEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.ListEditorsResponse{Editors: dbresp}, nil
}

func (s *ProductServiceServer) GetRightsInfoByEditor(ctx context.Context, req *v1.GetRightsInfoByEditorRequest) (*v1.GetRightsInfoByEditorResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetAcqRightsByEditor(ctx, db.GetAcqRightsByEditorParams{
		ProductEditor: req.Editor,
		Scope:         req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetRightsInfoByEditor - GetAcqRightsByEditor", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	dbresp1, err := s.ProductRepo.GetAggregationByEditor(ctx, db.GetAggregationByEditorParams{
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

func (s *ProductServiceServer) GetAllEditorsCatalog(ctx context.Context, req *v1.GetAllEditorsCatalogRequest) (*v1.GetAllEditorsCatalogResponse, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}

	dbresp, err := s.ProductRepo.GetEditor(ctx)
	if err != nil {
		logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	return &v1.GetAllEditorsCatalogResponse{EditorName: dbresp}, nil
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

func (s *ProductServiceServer) GetEditorExpensesByScope(ctx context.Context, req *v1.EditorExpensesByScopeRequest) (*v1.EditorExpensesByScopeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	dbresp, err := s.ProductRepo.GetEditorExpensesByScopeData(ctx, []string{req.Scope})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - GetEditorExpensesByScope - GetEditorExpensesByScopeData",
			"error", err.Error(),
			"scope", req.Scope,
			"status", codes.Internal,
		)
		return nil, status.Error(codes.Internal, "DBError")
	}
	dbrespocl, err := s.ProductRepo.GetComputedCostEditors(ctx, []string{req.Scope})
	if err != nil {
		logger.Log.Sugar().Errorw("service/v1 - GetEditorExpensesByScope - GetComputedCostEditorsData",
			"error", err.Error(),
			"scope", req.Scope,
			"status", codes.Internal,
		)
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.EditorExpensesByScopeResponse{}
	apiresp.EditorExpensesByScope = make([]*v1.EditorExpensesByScopeData, len(dbresp))
	for i := range dbresp {
		apiresp.EditorExpensesByScope[i] = &v1.EditorExpensesByScopeData{}
		apiresp.EditorExpensesByScope[i].EditorName = dbresp[i].Editor
		apiresp.EditorExpensesByScope[i].TotalPurchaseCost = dbresp[i].TotalPurchaseCost
		apiresp.EditorExpensesByScope[i].TotalMaintenanceCost = dbresp[i].TotalMaintenanceCost
		apiresp.EditorExpensesByScope[i].TotalCost = dbresp[i].TotalCost
	}
	for _, oclrow := range dbrespocl {
		for i, apires := range apiresp.EditorExpensesByScope {
			if oclrow.Editor == apires.EditorName {
				apiresp.EditorExpensesByScope[i].TotalComputedCost = oclrow.Cost
			}
		}
	}
	return &apiresp, nil
}
