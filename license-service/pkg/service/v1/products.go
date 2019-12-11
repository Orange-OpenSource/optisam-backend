// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/ctxmanage"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"sort"

	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	pageSize := req.GetPageSize()
	pageNum := req.GetPageNum()
	sortOrder := req.GetSortOrder()
	sortBy := req.GetSortBy()

	skip := pageSize * (pageNum - 1)

	params := repo.QueryProducts{}

	params.PageSize = pageSize
	params.Offset = skip
	params.SortBy = sortBy
	if sortOrder == "asc" {
		params.SortOrder = "orderasc"
	} else {
		params.SortOrder = "orderdesc"
	}
	if req.SearchParams != nil {
		params.Filter = productFilter(req.SearchParams)
		params.AcqFilter = productAcqRightFilter(req.SearchParams.AgFilter)
		params.AggFilter = productAggregateFilter(req.SearchParams.AgFilter)
	}
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	res, err := s.licenseRepo.GetProducts(ctx, &params, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
	}

	ListProdResponse := new(v1.ListProductsResponse)

	ListProdResponse.Products = make([]*v1.Product, len(res.Products))

	if len(res.NumOfRecords) > 0 {
		ListProdResponse.TotalRecords = res.NumOfRecords[0].TotalCnt
	}

	for i := range res.Products {
		ListProdResponse.Products[i] = &v1.Product{}
		ListProdResponse.Products[i].Name = res.Products[i].Name
		ListProdResponse.Products[i].Version = res.Products[i].Version
		ListProdResponse.Products[i].Category = res.Products[i].Category
		ListProdResponse.Products[i].Editor = res.Products[i].Editor
		ListProdResponse.Products[i].SwidTag = res.Products[i].Swidtag
		ListProdResponse.Products[i].NumofEquipments = res.Products[i].NumOfEquipments
		ListProdResponse.Products[i].NumOfApplications = res.Products[i].NumOfApplications

	}

	return ListProdResponse, nil
}

func (s *licenseServiceServer) GetProduct(ctx context.Context, req *v1.ProductRequest) (*v1.ProductResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	tag := req.GetSwidTag()
	res, err := s.licenseRepo.GetProductInformation(ctx, tag, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
	}

	ProdResponse := v1.ProductResponse{}

	if len(res.Products) < 1 {
		return nil, errors.New("No Product found")
	}
	if len(res.Products) > 1 {
		return nil, errors.New(" ")
	}

	prodInfo := new(v1.ProductInfo)
	prodOptions := new(v1.ProductOptions)

	for _, obj := range res.Products {

		prodInfo.SwidTag = obj.Swidtag
		prodInfo.Editor = obj.Editor
		prodInfo.Release = obj.Version
		prodInfo.NumOfApplications = obj.NumOfApplications
		prodInfo.NumofEquipments = obj.NumofEquipments
		prodOptions.NumOfOptions = obj.NumofOptions
		prodOptions.Optioninfo = make([]*v1.OptionInfo, len(obj.Child))

		for i := range obj.Child {
			prodOptions.Optioninfo[i] = &v1.OptionInfo{}
			prodOptions.Optioninfo[i].SwidTag = obj.Child[i].SwidTag
			prodOptions.Optioninfo[i].Name = obj.Child[i].Name
			prodOptions.Optioninfo[i].Editor = obj.Child[i].Editor
			prodOptions.Optioninfo[i].Version = obj.Child[i].Version
		}
		ProdResponse.ProductInfo = prodInfo
		ProdResponse.ProductOptions = prodOptions
	}
	return &ProdResponse, nil
}

func (s *licenseServiceServer) ListApplicationsForProduct(ctx context.Context, req *v1.ListApplicationsForProductRequest) (*v1.ListApplicationsForProductResponse, error) {

	errValidation := req.Validate()
	if errValidation != nil {
		return nil, errValidation
	}

	pageSize := req.GetPageSize()
	pageNum := req.GetPageNum()

	var sortorder repo.SortOrder

	switch req.GetSortOrder() {
	case "asc":
		sortorder = repo.SortASC
	case "desc":
		sortorder = repo.SortDESC
	default:
		logger.Log.Error("ListApplicationsForProduct - ", zap.Any("Sort Order", sortOrder))
		sortorder = repo.SortASC
	}

	params := &repo.QueryApplicationsForProduct{
		SwidTag:   req.GetSwidTag(),
		PageSize:  pageSize,
		Offset:    offset(pageSize, pageNum),
		SortBy:    req.GetSortBy(),
		SortOrder: sortorder,
	}

	if req.SearchParams != nil {
		params.Filter = applicationFilterForListApplicationsForProduct(req.SearchParams)
	}
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	resp, err := s.licenseRepo.GetApplicationsForProduct(ctx, params, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - GetApplicationsForProduct - ", zap.String("reason", err.Error()), zap.Any("request params", params))
		return nil, status.Error(codes.Unknown, "service/v1 - GetApplicationsForProduct - failed to get Instances-> "+err.Error())
	}

	appForProdResponse := &v1.ListApplicationsForProductResponse{
		Applications: make([]*v1.ApplicationForProduct, len(resp.Applications)),
		TotalRecords: resp.NumOfRecords[0].TotalCnt,
	}

	for i, app := range resp.Applications {
		appForProdResponse.Applications[i] = &v1.ApplicationForProduct{
			ApplicationId:   app.ApplicationID,
			Name:            app.Name,
			AppOwner:        app.Owner,
			NumofEquipments: app.NumOfEquipments,
			NumOfInstances:  app.NumOfInstances,
		}

	}
	return appForProdResponse, nil
}

func (s *licenseServiceServer) ListInstancesForApplicationsProduct(ctx context.Context, req *v1.ListInstancesForApplicationProductRequest) (*v1.ListInstancesForApplicationProductResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	errValidation := req.Validate()
	if errValidation != nil {
		return nil, errValidation
	}

	pageSize := req.GetPageSize()
	pageNum := req.GetPageNum()

	sortOdr := sortOrder(req.GetSortOrder())

	params := &repo.QueryInstancesForApplicationProduct{
		SwidTag:   req.GetSwidTag(),
		AppID:     req.GetApplicationId(),
		PageSize:  pageSize,
		Offset:    offset(pageSize, pageNum),
		SortBy:    int32(req.GetSortBy()),
		SortOrder: sortOdr,
	}

	application, err := s.licenseRepo.GetApplication(ctx, req.ApplicationId, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - GetApplication - ", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Unknown, "service/v1 - GetApplication - failed to get application-> "+err.Error())
	}

	resp, err := s.licenseRepo.GetInstancesForApplicationsProduct(ctx, params, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - GetInstancesForApplicationsProduct - ", zap.String("reason", err.Error()), zap.Any("request params", params))
		return nil, status.Error(codes.Unknown, "service/v1 - GetInstancesForApplicationsProduct - failed to get Instances-> "+err.Error())
	}

	instanceForAppProdResponse := &v1.ListInstancesForApplicationProductResponse{
		Instances:    make([]*v1.InstancesForApplicationProduct, len(resp.Instances)),
		TotalRecords: resp.NumOfRecords[0].TotalCnt,
	}

	for i, instance := range resp.Instances {
		if instance.Name != "" {
			instanceForAppProdResponse.Instances[i] = &v1.InstancesForApplicationProduct{
				Id:              instance.ID,
				Name:            instance.Name,
				Environment:     instance.Environment,
				NumofEquipments: instance.NumOfEquipments,
				NumofProducts:   instance.NumOfProducts,
			}
		} else {
			instanceForAppProdResponse.Instances[i] = &v1.InstancesForApplicationProduct{
				Id:              instance.ID,
				Name:            application.Name,
				Environment:     instance.Environment,
				NumofEquipments: instance.NumOfEquipments,
				NumofProducts:   instance.NumOfProducts,
			}
		}
	}
	return instanceForAppProdResponse, nil
}

// ListAcqRightsForProduct implements license service ListAcqRightsForProduct function
func (s *licenseServiceServer) ListAcqRightsForProduct(ctx context.Context, req *v1.ListAcquiredRightsForProductRequest) (*v1.ListAcquiredRightsForProductResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	ID, prodRights, err := s.licenseRepo.ProductAcquiredRights(ctx, req.SwidTag, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch product acquired rights")
	}
	res, err := s.licenseRepo.GetProductInformation(ctx, req.SwidTag, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Products -> "+err.Error())
	}
	numEquips := int32(0)
	if len(res.Products) != 0 {
		numEquips = res.Products[0].NumofEquipments
	}

	metrics, err := s.licenseRepo.ListMetrices(ctx, userClaims.Socpes)
	if err != nil && err != repo.ErrNoData {
		return nil, status.Error(codes.Internal, "cannot fetch metric OPS")

	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")

	}
	prodAcqRights := make([]*v1.ProductAcquiredRights, len(prodRights))
	ind := 0
	for i, acqRight := range prodRights {
		prodAcqRights[i] = &v1.ProductAcquiredRights{
			SKU:            acqRight.SKU,
			SwidTag:        req.SwidTag,
			Metric:         acqRight.Metric,
			NumAcqLicences: int32(acqRight.AcqLicenses),
			TotalCost:      acqRight.TotalCost,
		}
		// TODO; separate cases and log error messages
		if ind = metricNameExistsAll(metrics, acqRight.Metric); ind == -1 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric name doesnt exist - " + acqRight.Metric)

			continue
		}
		if numEquips == 0 {
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - no equipments linked with product")
			continue
		}

		computedLicenses := uint64(0)
		switch metrics[ind].Type {
		case repo.MetricOPSOracleProcessorStandard:
			computedLicenses, err = s.computedLicensesOPS(ctx, eqTypes, ID, metrics[ind].Name)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - ", zap.String("reason", err.Error()))
				continue
			}
		case repo.MetricSPSSagProcessorStandard:
			licensesProd, licensesNonProd, err := s.computedLicensesSPS(ctx, eqTypes, ID, metrics[ind].Name)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricSPSSagProcessorStandard ", zap.String("reason", err.Error()))
				continue
			}
			if licensesProd > licensesNonProd {
				computedLicenses = licensesProd
			} else {
				computedLicenses = licensesNonProd
			}
		case repo.MetricIPSIbmPvuStandard:
			computedLicenses, err = s.computedLicensesIPS(ctx, eqTypes, ID, metrics[ind].Name)
			if err != nil {
				logger.Log.Error("service/v1 - ListAcqRightsForProduct - MetricIPSIbmPvuStandard", zap.String("reason", err.Error()))
				continue
			}
		default:
			logger.Log.Error("service/v1 - ListAcqRightsForProduct - metric type doesnt match - " + string(metrics[ind].Type))
			continue
		}

		delta := int32(acqRight.AcqLicenses) - int32(computedLicenses)

		prodAcqRights[i].NumCptLicences = int32(computedLicenses)
		prodAcqRights[i].DeltaNumber = int32(delta)
		prodAcqRights[i].DeltaCost = acqRight.AvgUnitPrice * float64(delta)
	}

	return &v1.ListAcquiredRightsForProductResponse{
		AcqRights: prodAcqRights,
	}, nil
}

func (s *licenseServiceServer) ListEquipmentsForProduct(ctx context.Context, req *v1.ListEquipmentsForProductRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProduct - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.EqTypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProduct - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	idx := attributeIndexByName(req.SortBy, eqType.Attributes)
	if idx < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot find sort by attribute: %s", req.SortBy)
	}

	if !eqType.Attributes[idx].IsDisplayed {
		return nil, status.Errorf(codes.InvalidArgument, "cannot sort by attribute: %s is not displayable", req.SortBy)
	}

	filter, err := parseEquipmentQueryParam(req.SearchParams, eqType.Attributes)
	if err != nil {
		return nil, err
	}

	queryParams := &repo.QueryEquipments{
		PageSize:  req.PageSize,
		Offset:    offset(req.PageSize, req.PageNum),
		SortBy:    req.SortBy,
		SortOrder: sortOrder(req.SortOrder),
		Filter:    filter,
	}

	numOfrecords, equipments, err := s.licenseRepo.ProductEquipments(ctx, req.SwidTag, eqType, queryParams, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProduct - ", zap.String("reason", err.Error()), zap.Any("request params", queryParams))
		return nil, status.Error(codes.Internal, "cannot fetch product equipments")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: numOfrecords,
		Equipments:   equipments,
	}, nil
}

func productFilter(params *v1.ProductSearchParams) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if params.SwidTag != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.SwidTag.FilteringOrder,
			FilterKey:         "swidtag",
			FilterValue:       params.SwidTag.Filteringkey,
		})
	}
	if params.Name != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Name.FilteringOrder,
			FilterKey:         "name",
			FilterValue:       params.Name.Filteringkey,
		})
	}
	if params.Editor != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Editor.FilteringOrder,
			FilterKey:         "editor",
			FilterValue:       params.Editor.Filteringkey,
		})
	}
	if params.Edition != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Edition.FilteringOrder,
			FilterKey:         "edition",
			FilterValue:       params.Edition.Filteringkey,
		})
	}

	sort.Sort(aggFilter)

	return aggFilter
}

func productAcqRightFilter(agFilter *v1.AggregationFilter) *repo.AggregateFilter {
	if agFilter == nil {
		return nil
	}
	return &repo.AggregateFilter{
		Filters: []repo.Queryable{
			&repo.Filter{
				FilterKey:   repo.AcquiredRightsSearchKeyMetric.String(),
				FilterValue: agFilter.NotForMetric,
			},
		},
	}
}

func productAggregateFilter(agFilter *v1.AggregationFilter) *repo.AggregateFilter {
	if agFilter == nil {
		return nil
	}
	return &repo.AggregateFilter{
		Filters: []repo.Queryable{
			&repo.Filter{
				FilterKey:   repo.MetricSearchKeyName.String(),
				FilterValue: agFilter.NotForMetric,
			},
		},
	}
}

func equipmentProductFilter(params *v1.EquipmentProductSearchParams) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if params.SwidTag != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.SwidTag.FilteringOrder,
			FilterKey:         "swidtag",
			FilterValue:       params.SwidTag.Filteringkey,
		})
	}
	if params.Name != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Name.FilteringOrder,
			FilterKey:         "name",
			FilterValue:       params.Name.Filteringkey,
		})
	}
	if params.Editor != nil {
		aggFilter.Filters = append(aggFilter.Filters, &repo.Filter{
			FilteringPriority: params.Editor.FilteringOrder,
			FilterKey:         "editor",
			FilterValue:       params.Editor.Filteringkey,
		})
	}
	sort.Sort(aggFilter)

	return aggFilter
}

func offset(pageSize, pageNum int32) int32 {
	return pageSize * (pageNum - 1)
}

func sortOrder(sortOdr v1.SortOrder) repo.SortOrder {

	switch sortOdr {
	case v1.SortOrder_ASC:
		return repo.SortASC
	case v1.SortOrder_DESC:
		return repo.SortDESC
	default:
		// we always condider asc order
		return repo.SortASC
	}

}
