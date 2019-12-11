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
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) CreateProductAggregation(ctx context.Context, req *v1.ProductAggregation) (*v1.ProductAggregation, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to create product aggregation")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		_, err := s.licenseRepo.ProductAggregationsByName(ctx, req.Name, userClaims.Socpes)
		if err != nil && err != repo.ErrNodeNotFound {
			logger.Log.Error("service/v1 - CreateProductAggregation - ProductAggregationsByName", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot get product aggregation")
		} else if err == nil {
			return nil, status.Error(codes.AlreadyExists, "product aggregation node already exists")
		}
		var metricID string
		var productIDs []string

		metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
		if err != nil {
			logger.Log.Error("service/v1 - CreateProductAggregation - ListMetrices", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot fetch metrics")
		}
		for _, met := range metrics {
			if met.Name == req.Metric {
				metricID = met.ID
				break
			}
		}
		if metricID == "" {
			return nil, status.Error(codes.NotFound, "metric does not exist")
		}
		agFilter := &v1.AggregationFilter{
			NotForMetric: req.Metric,
		}

		params := &repo.QueryProducts{
			Filter: &repo.AggregateFilter{
				Filters: []repo.Queryable{
					&repo.Filter{
						FilterKey:   "name",
						FilterValue: req.Product,
					},
					&repo.Filter{
						FilterKey:   "editor",
						FilterValue: req.Editor,
					},
				},
			},

			AcqFilter: productAcqRightFilter(agFilter),
			AggFilter: productAggregateFilter(agFilter),
		}
		for _, proSwid := range req.Products {

			proID, err := s.licenseRepo.ProductIDForSwidtag(ctx, proSwid, params, userClaims.Socpes)
			if err != nil {
				logger.Log.Error("service/v1 - CreateProductAggregation - ProductIDForSwidtag", zap.Error(err))
				return nil, status.Error(codes.NotFound, "cannot get product id for swid tag")
			}
			productIDs = append(productIDs, proID)
		}
		repoProAgg := &repo.ProductAggregation{
			Name:     req.Name,
			Editor:   req.Editor,
			Product:  req.Product,
			Metric:   metricID,
			Products: productIDs,
		}
		repoProAgg, err = s.licenseRepo.CreateProductAggregation(ctx, repoProAgg, userClaims.Socpes)
		if err != nil {
			logger.Log.Error("service/v1 - CreateProductAggregation - CreateProductAggregation", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot create product aggregation")
		}
		return convertRepoToSrvProAgg(repoProAgg), nil
	default:
		logger.Log.Error("service/v1 - ProductAggregation - CreateProductAggregation")
		return nil, status.Error(codes.PermissionDenied, "unknown role")
	}
}

func (s *licenseServiceServer) ListProductAggregation(ctx context.Context, req *v1.ListProductAggregationRequest) (*v1.ListProductAggregationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	repoProAggs, err := s.licenseRepo.ListProductAggregations(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListProductAggregations - ListProductAggregations", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch product aggregations")
	}
	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateProductAggregation - ListMetrices", zap.Error(err))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}
	for _, agg := range repoProAggs {
		for _, met := range metrics {
			if met.ID == agg.Metric {
				agg.Metric = met.Name
				break
			}
		}
	}
	return &v1.ListProductAggregationResponse{
		Aggregations: convertRepoToSrvProAggAll(repoProAggs),
	}, nil
}

func (s *licenseServiceServer) DeleteProductAggregation(ctx context.Context, req *v1.DeleteProductAggregationRequest) (*v1.ListProductAggregationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user doesnot have access to delete product aggregation")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		repoProAggs, err := s.licenseRepo.DeleteProductAggregation(ctx, req.ID, userClaims.Socpes)
		if err != nil {
			logger.Log.Error("service/v1 - DeleteProductAggregation - DeleteProductAggregation", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot delete product aggregation")
		}
		return &v1.ListProductAggregationResponse{
			Aggregations: convertRepoToSrvProAggAll(repoProAggs),
		}, nil
	default:
		logger.Log.Error("service/v1 - ProductAggregation - DeleteProductAggregation")
		return nil, status.Error(codes.PermissionDenied, "unknown role")
	}
}

func convertRepoToSrvProAggAll(proAggs []*repo.ProductAggregation) []*v1.ProductAggregation {
	srvProAggs := make([]*v1.ProductAggregation, len(proAggs))
	for i := range proAggs {
		srvProAggs[i] = convertRepoToSrvProAgg(proAggs[i])
	}
	return srvProAggs
}

func convertRepoToSrvProAgg(proAgg *repo.ProductAggregation) *v1.ProductAggregation {
	return &v1.ProductAggregation{
		ID:       proAgg.ID,
		Name:     proAgg.Name,
		Editor:   proAgg.Editor,
		Product:  proAgg.Product,
		Metric:   proAgg.Metric,
		Products: proAgg.Products,
	}
}
