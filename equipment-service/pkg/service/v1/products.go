// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"

	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *equipmentServiceServer) ListEquipmentsForProduct(ctx context.Context, req *v1.ListEquipmentsForProductRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, userClaims.Socpes)
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

	numOfrecords, equipments, err := s.equipmentRepo.ProductEquipments(ctx, req.SwidTag, eqType, queryParams, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsForProduct - ", zap.String("reason", err.Error()), zap.Any("request params", queryParams))
		return nil, status.Error(codes.Internal, "cannot fetch product equipments")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: numOfrecords,
		Equipments:   equipments,
	}, nil
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

func filterTypeRepo(filterType v1.StringFilter_Type) repo.Filtertype {

	switch filterType {
	case v1.StringFilter_REGEX:
		return repo.RegexFilter
	case v1.StringFilter_EQ:
		return repo.EqFilter
	default:
		return repo.RegexFilter
	}
}

func addFilter(priority int32, key string, value interface{}, values []string, filterType v1.StringFilter_Type) *repo.Filter {
	return &repo.Filter{
		FilteringPriority:   priority,
		FilterKey:           key,
		FilterValue:         value,
		FilterValueMultiple: stringToInterface(values),
		FilterMatchingType:  filterTypeRepo(filterType),
	}
}

func stringToInterface(vals []string) []interface{} {
	interfaceSlice := make([]interface{}, len(vals))
	for i := range vals {
		interfaceSlice[i] = vals[i]
	}
	return interfaceSlice
}

func scopesIsSubSlice(scopes []string, claimsScopes []string) bool {
	if len(scopes) > len(claimsScopes) {
		return false
	}
	for _, e := range scopes {
		if contains(claimsScopes, e) == -1 {
			return false
		}
	}
	return true
}
func contains(s []string, e string) int {
	for i, a := range s {
		if a == e {
			return i
		}
	}
	return -1
}
