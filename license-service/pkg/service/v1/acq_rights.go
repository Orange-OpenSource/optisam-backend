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
	"sort"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (lr *licenseServiceServer) ListAcquiredRights(ctx context.Context, req *v1.ListAcquiredRightsRequest) (*v1.ListAcquiredRightsResponse, error) {
	// ctx, span := trace.StartSpan(ctx, "Service Layer")
	// defer span.End()
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	params := &repo.QueryAcquiredRights{
		PageSize:  req.PageSize,
		Offset:    offset(req.PageSize, req.PageNum),
		SortBy:    repo.AcquiredRightsSortBy(req.SortBy),
		SortOrder: sortOrder(req.SortOrder),
		Filter:    acqRightsFilter(req.SearchParams),
	}

	totalRecords, acqRights, err := lr.licenseRepo.AcquiredRights(ctx, params, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateEquipmentType - creating equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "failed to get AcquiredRights")
	}

	return &v1.ListAcquiredRightsResponse{
		TotalRecords:   totalRecords,
		AcquiredRights: repoAcqRightsToSrvAcqRightsAll(acqRights),
	}, nil
}

func repoAcqRightsToSrvAcqRightsAll(ars []*repo.AcquiredRights) []*v1.AcquiredRights {
	aqRights := make([]*v1.AcquiredRights, len(ars))
	for i := range ars {
		aqRights[i] = repoAcqRightsToSrvAcqRights(ars[i])
	}
	return aqRights
}

func repoAcqRightsToSrvAcqRights(ar *repo.AcquiredRights) *v1.AcquiredRights {
	return &v1.AcquiredRights{
		Entity:                         ar.Entity,
		SKU:                            ar.SKU,
		SwidTag:                        ar.SwidTag,
		ProductName:                    ar.ProductName,
		Editor:                         ar.Editor,
		Metric:                         ar.Metric,
		AcquiredLicensesNumber:         ar.AcquiredLicensesNumber,
		LicensesUnderMaintenanceNumber: ar.LicensesUnderMaintenanceNumber,
		AvgLicenesUnitPrice:            ar.AvgLicenesUnitPrice,
		AvgMaintenanceUnitPrice:        ar.AvgMaintenanceUnitPrice,
		TotalPurchaseCost:              ar.TotalPurchaseCost,
		TotalMaintenanceCost:           ar.TotalMaintenanceCost,
		TotalCost:                      ar.TotalCost,
	}
}

func acqRightsFilter(params *v1.AcquiredRightsSearchParams) *repo.AggregateFilter {
	if params == nil {
		return nil
	}
	aggFilter := new(repo.AggregateFilter)
	if params.SKU != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.SKU.FilteringOrder,
			FilterKey:         repo.AcquiredRightsSearchKeySKU.String(),
			FilterValue:       params.SKU.Filteringkey,
		})
	}
	if params.SwidTag != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.SwidTag.FilteringOrder,
			FilterKey:         repo.AcquiredRightsSearchKeySwidTag.String(),
			FilterValue:       params.SwidTag.Filteringkey,
		})
	}
	if params.ProductName != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.ProductName.FilteringOrder,
			FilterKey:         repo.AcquiredRightsSearchKeyProductName.String(),
			FilterValue:       params.ProductName.Filteringkey,
		})
	}
	if params.Editor != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Editor.FilteringOrder,
			FilterKey:         repo.AcquiredRightsSearchKeyEditor.String(),
			FilterValue:       params.Editor.Filteringkey,
		})
	}
	if params.Metric != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Metric.FilteringOrder,
			FilterKey:         repo.AcquiredRightsSearchKeyMetric.String(),
			FilterValue:       params.Metric.Filteringkey,
		})
	}
	sort.Sort(aggFilter)

	return aggFilter
}
