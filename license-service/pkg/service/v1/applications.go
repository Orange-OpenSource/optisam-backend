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
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"sort"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListApplications(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	err := req.Validate()
	if err != nil {
		return nil, err
	}

	pageSize := req.GetPageSize()
	pageNum := req.GetPageNum()
	sortOrder := req.GetSortOrder()
	sortBy := req.GetSortBy()

	skip := pageSize * (pageNum - 1)

	params := repo.QueryApplications{}

	params.PageSize = pageSize
	params.Offset = skip
	params.SortBy = sortBy

	if sortOrder == "asc" {
		params.SortOrder = "orderasc"
	} else {
		params.SortOrder = "orderdesc"
	}

	if req.SearchParams != nil {
		params.Filter = applicationFilter(req.SearchParams)
	}

	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	resp, err := s.licenseRepo.GetApplications(ctx, &params, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Applications-> "+err.Error())
	}

	ListAppResponse := v1.ListApplicationsResponse{}

	ListAppResponse.Applications = make([]*v1.Application, len(resp.Applications))

	if len(resp.NumOfRecords) > 0 {
		ListAppResponse.TotalRecords = resp.NumOfRecords[0].TotalCnt
	}

	for i := range resp.Applications {
		ListAppResponse.Applications[i] = &v1.Application{}
		ListAppResponse.Applications[i].Name = resp.Applications[i].Name
		ListAppResponse.Applications[i].ApplicationId = resp.Applications[i].ApplicationID
		ListAppResponse.Applications[i].ApplicationOwner = resp.Applications[i].ApplicationOwner
		ListAppResponse.Applications[i].NumOfInstances = resp.Applications[i].NumOfInstances
		ListAppResponse.Applications[i].NumofProducts = resp.Applications[i].NumOfProducts

	}

	return &ListAppResponse, nil
}

func (s *licenseServiceServer) ListProductsForApplication(ctx context.Context, req *v1.ListProductsForApplicationRequest) (*v1.ListProductsForApplicationResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	applicationID := req.GetApplicationId()
	resp, err := s.licenseRepo.GetProductsForApplication(ctx, applicationID, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
	}

	prodForAppResponse := &v1.ListProductsForApplicationResponse{}

	prodForAppResponse.Products = make([]*v1.ProductForApplication, len(resp.Products))

	prodForAppResponse.TotalRecords = resp.NumOfRecords[0].TotalCnt

	for i, prod := range resp.Products {
		prodForAppResponse.Products[i] = &v1.ProductForApplication{
			SwidTag:         prod.SwidTag,
			Name:            prod.Name,
			Editor:          prod.Editor,
			Version:         prod.Version,
			NumofEquipments: prod.NumOfEquipments,
			NumOfInstances:  prod.NumOfInstances,
		}

	}
	return prodForAppResponse, nil
}

func applicationFilter(params *v1.ApplicationSearchParams) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if params.Name != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Name.FilteringOrder,
			FilterKey:         "name",
			FilterValue:       params.Name.Filteringkey,
		})
	}
	if params.ApplicationOwner != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.ApplicationOwner.FilteringOrder,
			FilterKey:         "application_owner",
			FilterValue:       params.ApplicationOwner.Filteringkey,
		})
	}
	sort.Sort(aggFilter)

	return aggFilter
}

func applicationFilterForListApplicationsForProduct(params *v1.ApplicationSearchParams) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if params.Name != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Name.FilteringOrder,
			FilterKey:         "name",
			FilterValue:       params.Name.Filteringkey,
		})
	}
	if params.ApplicationOwner != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.ApplicationOwner.FilteringOrder,
			FilterKey:         "application_owner",
			FilterValue:       params.ApplicationOwner.Filteringkey,
		})
	}
	sort.Sort(aggFilter)

	return aggFilter
}
