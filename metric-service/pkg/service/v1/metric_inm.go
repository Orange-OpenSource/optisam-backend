// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"

	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/api/v1"
	repo "optisam-backend/metric-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricInstanceNumberStandard(ctx context.Context, req *v1.CreateINM) (*v1.CreateINM, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	metrics, err := s.metricRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricInstanceNumberStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")

	}
	if metricNameExistsAll(metrics, req.Name) != -1 {
		return nil, status.Error(codes.InvalidArgument, "metric name already exists")
	}

	met, err := s.metricRepo.CreateMetricInstanceNumberStandard(ctx, serverToRepoINM(req), userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateMetricInstanceNumberStandard  in repo", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create metric inm")
	}

	return repoToServerINM(met), nil
}

func serverToRepoINM(met *v1.CreateINM) *repo.MetricINM {
	return &repo.MetricINM{
		Name:        met.Name,
		Coefficient: met.Coefficient}
}

func repoToServerINM(met *repo.MetricINM) *v1.CreateINM {
	return &v1.CreateINM{
		Name:        met.Name,
		ID:          met.ID,
		Coefficient: met.Coefficient}
}
