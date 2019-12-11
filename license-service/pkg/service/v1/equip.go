// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"net/url"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) ListEquipments(ctx context.Context, req *v1.ListEquipmentsRequest) (*v1.ListEquipmentsResponse, error) {
	// TODO: fetch only the required equipment type
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipments - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	idx := -1
	for i := range eqTypes {
		if eqTypes[i].ID == req.TypeId {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, status.Errorf(codes.NotFound, "equipment type doesnot exist, typeID %s", req.TypeId)
	}

	eqType := eqTypes[idx]
	idx = attributeIndexByName(req.SortBy, eqType.Attributes)
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

	numOfrecords, equipments, err := s.licenseRepo.Equipments(ctx, eqType, queryParams, userClaims.Socpes)
	if err != nil {
		// TODO log error
		return nil, status.Error(codes.Internal, "cannot get equipments")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: numOfrecords,
		Equipments:   equipments,
	}, nil
}

func (s *licenseServiceServer) GetEquipment(ctx context.Context, req *v1.GetEquipmentRequest) (*v1.GetEquipmentResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	resp, err := s.licenseRepo.Equipment(ctx, eqType, req.EquipId, userClaims.Socpes)
	if err != nil {
		switch err {
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment data")
		case repo.ErrNodeNotFound:
			return nil, status.Error(codes.NotFound, "Equipment doesn't exists")
		}
		logger.Log.Error("service/v1 -GetEquipment - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment")
	}

	return &v1.GetEquipmentResponse{
		Equipment: string(resp),
	}, nil
}

func (s *licenseServiceServer) ListEquipmentParents(ctx context.Context, req *v1.ListEquipmentParentsRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentParents - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentParents - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	equipParent, err := equipmentTypeExistsByID(eqType.ParentID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentParents - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	records, resp, err := s.licenseRepo.EquipmentParents(ctx, eqType, equipParent, req.EquipId, userClaims.Socpes)
	if err != nil {
		switch err {
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment parents")
		case repo.ErrNodeNotFound:
			return nil, status.Error(codes.NotFound, "Equipment Parent doesn't exists")
		}
		logger.Log.Error("service/v1 -ListEquipmentParents - fetching equipment parents", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment parents")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: records,
		Equipments:   resp,
	}, nil
}

func (s *licenseServiceServer) ListEquipmentChildren(ctx context.Context, req *v1.ListEquipmentChildrenRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentChildren - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	equip, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentChildren - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type of given Id")
	}

	equipChild, err := equipmentTypeExistsByID(req.ChildrenTypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentChildren - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	if equipChild.ParentID != equip.ID {
		return nil, status.Error(codes.InvalidArgument, "Child of given type is not valid")
	}
	queryParams := &repo.QueryEquipments{
		PageSize:  req.PageSize,
		Offset:    offset(req.PageSize, req.PageNum),
		SortBy:    req.SortBy,
		SortOrder: sortOrder(req.SortOrder),
	}
	if req.SearchParams != "" {
		filter, err := parseEquipmentQueryParam(req.SearchParams, equipChild.Attributes)
		if err != nil {
			return nil, err
		}
		queryParams.Filter = filter
	}
	records, resp, err := s.licenseRepo.EquipmentChildren(ctx, equip, equipChild, req.EquipId, queryParams, userClaims.Socpes)
	if err != nil {
		switch err {
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment children")
		case repo.ErrNodeNotFound:
			return nil, status.Error(codes.NotFound, "Equipment children do not exists")
		}
		logger.Log.Error("service/v1 -ListEquipmentChildren - cannot fetch equipment children", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment children")
	}
	return &v1.ListEquipmentsResponse{
		TotalRecords: records,
		Equipments:   resp,
	}, nil
}

func (s *licenseServiceServer) ListEquipmentProducts(ctx context.Context, req *v1.ListEquipmentProductsRequest) (*v1.ListEquipmentProductsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.licenseRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentProducts - fetching equipments types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	equip, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentProducts - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}
	queryParams := &repo.QueryEquipmentProduct{
		PageSize:  req.PageSize,
		Offset:    offset(req.PageSize, req.PageNum),
		SortBy:    repo.EquipmentProductSortBy(req.SortBy),
		SortOrder: sortOrder(req.SortOrder),
	}

	if req.SearchParams != nil {
		queryParams.Filter = equipmentProductFilter(req.SearchParams)
	}

	records, resp, err := s.licenseRepo.EquipmentProducts(ctx, equip, req.EquipId, queryParams, userClaims.Socpes)
	if err != nil {
		switch err {
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment products")
		case repo.ErrNodeNotFound:
			return nil, status.Error(codes.NotFound, "Equipment Products do not exists")
		}
		logger.Log.Error("service/v1 -ListEquipmentProducts - fetching equipment products", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment products")
	}
	return &v1.ListEquipmentProductsResponse{
		TotalRecords: records,
		Products:     repoEquipProductToServiceEquipProductAll(resp),
	}, nil
}

func parseEquipmentQueryParam(query string, attributes []*repo.Attribute) (*repo.AggregateFilter, error) {
	query = strings.Replace(query, ",", "&", -1)
	values, err := url.ParseQuery(query)
	if err != nil {
		// TODO log error
		return nil, status.Error(codes.InvalidArgument, "proper format of query is search_params=attr1=val1,attr2=val2,attr3=val3")
	}

	aggregateFilter := &repo.AggregateFilter{}

	for key, val := range values {
		idx := attributeIndexByName(key, attributes)
		if idx == -1 {
			return nil, status.Errorf(codes.InvalidArgument, "attribute: %s not found ", key)
		}

		if !attributes[idx].IsDisplayed {
			return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched not dispalayable", key)
		}

		if !attributes[idx].IsSearchable {
			return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched not searchable", key)
		}

		if val[0] == "" {
			return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value for attribute is empty", key)
		}

		switch attributes[idx].Type {
		case repo.DataTypeString:
			if len(val[0]) < 3 {
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for string type attributes should have at least 3 characters", key, val[0])
			}
			val[0] = strings.Replace(regexp.QuoteMeta(val[0]), "/", "\\/", -1)
			aggregateFilter.Filters = append(aggregateFilter.Filters, &repo.Filter{
				FilterKey:   key,
				FilterValue: val[0],
			})
		case repo.DataTypeInt:
			v, err := strconv.ParseInt(val[0], 10, 64)
			if err != nil {
				// TODO log the error
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for int type attribute cannot be parsed", key, val[0])
			}
			aggregateFilter.Filters = append(aggregateFilter.Filters, &repo.Filter{
				FilterKey:   key,
				FilterValue: v,
			})
		case repo.DataTypeFloat:
			v, err := strconv.ParseFloat(val[0], 10)
			if err != nil {
				// TODO log the error
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for int type attribute cannot be parsed", key, val[0])
			}
			aggregateFilter.Filters = append(aggregateFilter.Filters, &repo.Filter{
				FilterKey:   key,
				FilterValue: v,
			})
		default:
			return nil, status.Errorf(codes.Internal, "attribute: %s cannot be not searched unsupported data type for attribute", key)
			// TODO: log here that we have unknown data type
		}
	}
	return aggregateFilter, nil
}

func attributeIndexByName(name string, attrs []*repo.Attribute) int {
	for i := range attrs {
		if attrs[i].Name == name {
			return i
		}
	}
	return -1
}
