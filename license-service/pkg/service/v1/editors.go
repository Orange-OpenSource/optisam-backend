// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListEditors(ctx context.Context, req *v1.ListEditorsRequest) (*v1.ListEditorsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	repoEditors, err := s.licenseRepo.ListEditors(ctx, nil, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch product aggregations")
	}
	return &v1.ListEditorsResponse{
		Editors: convertRepoToSrvEditorsAll(repoEditors),
	}, nil
}

func convertRepoToSrvEditorsAll(editors []*repo.Editor) []*v1.Editor {
	srvEditors := make([]*v1.Editor, len(editors))
	for i := range editors {
		srvEditors[i] = convertRepoToSrvEditor(editors[i])
	}
	return srvEditors
}

func convertRepoToSrvEditor(proAgg *repo.Editor) *v1.Editor {
	return &v1.Editor{
		ID:   proAgg.ID,
		Name: proAgg.Name,
	}
}
