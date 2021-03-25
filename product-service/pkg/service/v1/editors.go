// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServiceServer) ListEditors(ctx context.Context, req *v1.ListEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	var scopes []string
	heystack := strings.Join(userClaims.Socpes, "")
	for _, scope := range req.Scopes {
		if strings.Contains(heystack, scope) == true {
			scopes = append(scopes, scope)
		}
	}

	dbresp, err := s.productRepo.ListEditors(ctx, scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	apiresp := v1.ListEditorsResponse{}
	apiresp.Editors = dbresp

	return &apiresp, nil
}

func (s *productServiceServer) ListEditorProducts(ctx context.Context, req *v1.ListEditorProductsRequest) (*v1.ListEditorProductsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	var scopes []string
	heystack := strings.Join(userClaims.Socpes, "/")
	for _, scope := range req.Scopes {
		if strings.Contains(heystack, scope) == true {
			scopes = append(scopes, scope)
		}
	}

	dbresp, err := s.productRepo.GetProductsByEditor(ctx, db.GetProductsByEditorParams{ProductEditor: req.GetEditor(), Scopes: scopes})
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
	}
	return &apiresp, nil

}
