// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"fmt"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) CreateProductAggregation(ctx context.Context, req *v1.ProductAggregation) (*v1.ProductAggregation, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
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

		metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes...)
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
		fmt.Println(req.ProductNames)
		params := &repo.QueryProducts{
			Filter: &repo.AggregateFilter{
				Filters: []repo.Queryable{
					&repo.Filter{
						FilterMatchingType:  repo.EqFilter,
						FilterKey:           "name",
						FilterValueMultiple: stringToInterface(req.ProductNames),
					},
					&repo.Filter{
						FilterMatchingType: repo.EqFilter,
						FilterKey:          "editor",
						FilterValue:        req.Editor,
					},
				},
			},

			AcqFilter: productAcqRightFilter(req.Metric),
			AggFilter: productAggregateFilter(req.Metric),
		}
		for _, proSwid := range req.Products {

			proID, err := s.licenseRepo.ProductIDForSwidtag(ctx, proSwid, params, userClaims.Socpes...)
			if err != nil {
				logger.Log.Error("service/v1 - CreateProductAggregation - ProductIDForSwidtag", zap.Error(err))
				return nil, status.Error(codes.NotFound, "cannot get product id for swid tag")
			}
			productIDs = append(productIDs, proID)
		}
		repoProAgg := &repo.ProductAggregation{
			Name:     req.Name,
			Editor:   req.Editor,
			Product:  strings.Join(req.ProductNames, ","),
			Metric:   metricID,
			Products: productIDs,
		}
		repoProAgg, err = s.licenseRepo.CreateProductAggregation(ctx, repoProAgg, userClaims.Socpes)
		if err != nil {
			logger.Log.Error("service/v1 - CreateProductAggregation - CreateProductAggregation", zap.Error(err))
			return nil, status.Error(codes.Internal, "cannot create product aggregation")
		}
		req.ID = repoProAgg.ID
		return req, nil
	default:
		logger.Log.Error("service/v1 - ProductAggregation - CreateProductAggregation")
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
		ID:           proAgg.ID,
		Name:         proAgg.Name,
		Editor:       proAgg.Editor,
		ProductNames: strings.Split(proAgg.Product, ","),
		Metric:       proAgg.MetricName,
		Products:     proAgg.Products,
		ProductsFull: convertRepoToSrvProductAll(proAgg.ProductsFull),
	}
}

func productAggregateFilter(notForMetric string) *repo.AggregateFilter {
	return &repo.AggregateFilter{
		Filters: []repo.Queryable{
			&repo.Filter{
				FilterMatchingType: repo.EqFilter,
				FilterKey:          repo.MetricSearchKeyName.String(),
				FilterValue:        notForMetric,
			},
		},
	}
}
