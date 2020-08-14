// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/jsonpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// equipmentServiceServer is implementation of v1.authServiceServer proto interface
type equipmentServiceServer struct {
	equipmentRepo repo.Equipment
}

// custom json unmarshal
type customType string

func (p *customType) UnmarshalJSON(data []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Log.Error("Failed to Unmarshal", zap.Error(err))
		return err
	}
	switch v := tmp.(type) {
	case int:
		*p = customType(strconv.Itoa(v))
	case float64:
		*p = customType(strconv.FormatFloat(v, 'f', -1, 64))
	case string:
		*p = customType(string(v))
	default:
		logger.Log.Info("Data doenot maches any type", zap.Any("type", reflect.TypeOf(v)))
	}
	return nil
}

// NewEquipmentServiceServer creates License service
func NewEquipmentServiceServer(equipmentRepo repo.Equipment) v1.EquipmentServiceServer {
	return &equipmentServiceServer{equipmentRepo: equipmentRepo}
}

func (s *equipmentServiceServer) UpsertMetadata(ctx context.Context, req *v1.UpsertMetadataRequest) (*v1.UpsertMetadataResponse, error) {
	err := s.equipmentRepo.UpsertMetadata(ctx, &repo.Metadata{
		MetadataType: req.GetMetadataType(),
		Source:       req.GetMetadataSource(),
		Attributes:   req.GetMetadataAttributes(),
	})
	if err != nil {
		logger.Log.Error("Failed to upser metadat in dgraph", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot upsert metadata")
	}

	return &v1.UpsertMetadataResponse{Success: true}, nil
}

//UpsertEquipment to load equipment data
//uses reflection heavily
func (s *equipmentServiceServer) UpsertEquipment(ctx context.Context, req *v1.UpsertEquipmentRequest) (*v1.UpsertEquipmentResponse, error) {
	validate := validator.New()
	eqTypedata, err := s.equipmentRepo.EquipmentTypeByType(ctx, req.GetEqType())
	if err != nil {
		logger.Log.Error("Failed to get Metadata from dgraph", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot get metadata")
	}
	//logger.Log.Info("", zap.Any("eqTypeAttr", eqTypedata))

	var reqEqDataJSON bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	err = marshaler.Marshal(&reqEqDataJSON, req.GetEqData())
	if err != nil {
		logger.Log.Error("unable to marshal to json", zap.Error(err))
	}
	// logger.Log.Info("", zap.Any("eqTypeData", reqEqDataJSON.String()))
	dynamicStructFieldsReq := []reflect.StructField{}
	for _, attr := range eqTypedata.Attributes {

		var t interface{}
		switch helper.GetTypeInstance(attr.Type.String()).(type) {
		case int:
			t = 0
		case bool:
			t = false
		case string:
			t = ""
		case float64:
			t = 0.0
		default:
			t = ""
		}
		var regTag string
		var reqType reflect.Type
		switch {
		case attr.IsIdentifier:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.id" validate:"required"`
			reqType = reflect.TypeOf(customType(""))
		case attr.IsParentIdentifier:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.parent"`
			reqType = reflect.TypeOf(customType(""))
		default:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.` + req.GetEqType() + `.` + attr.Name + `"`
			reqType = reflect.TypeOf(t)
		}
		dynamicStructFieldsReq = append(dynamicStructFieldsReq, reflect.StructField{
			Name: strings.Title(attr.Name),
			Type: reqType,
			Tag:  reflect.StructTag(regTag),
		})

	}
	instanceReq := reflect.New(reflect.StructOf(dynamicStructFieldsReq)).Interface()
	err = json.Unmarshal(reqEqDataJSON.Bytes(), &instanceReq)
	if err != nil {
		logger.Log.Error("Equipment Data Unmarshal Error", zap.Error(err))
	}
	logger.Log.Info("", zap.Any("Dynamic Struct Req", instanceReq))
	err = validate.Struct(instanceReq)
	if err != nil {
		logger.Log.Error("Validation Error", zap.Error(err))
		return &v1.UpsertEquipmentResponse{Success: false}, status.Error(codes.Unknown, "ValidationError")
	}

	// type conversion
	// reflect.ValueOf(instanceReq).Convert(reflect.TypeOf(instanceDB))
	// logger.Log.Info("", zap.Any("Dynamic Struct valueOf", reflect.ValueOf(instanceReq)))

	err = s.equipmentRepo.UpsertEquipment(ctx, req.GetScope(), req.GetEqType(), eqTypedata.ParentType, instanceReq)
	if err != nil {
		logger.Log.Error("UpsertEquipment Failed", zap.Error(err))
		return &v1.UpsertEquipmentResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}
	return &v1.UpsertEquipmentResponse{Success: true}, nil
}

func (s *equipmentServiceServer) ListEquipments(ctx context.Context, req *v1.ListEquipmentsRequest) (*v1.ListEquipmentsResponse, error) {
	// TODO: fetch only the required equipment type
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		//logger.Log.Error("service/v1 - ListEquipments - fetching equipments", zap.String("reason", err.Error()))
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
	if req.Filter != nil {
		queryParams.ProductFilter = equipProductFilter(req.Filter)
		queryParams.ApplicationFilter = equipApplicationFilter(req.Filter)
	}

	numOfrecords, equipments, err := s.equipmentRepo.Equipments(ctx, eqType, queryParams, userClaims.Socpes)
	if err != nil {
		// TODO log error
		return nil, status.Error(codes.Internal, "cannot get equipments")
	}

	return &v1.ListEquipmentsResponse{
		TotalRecords: numOfrecords,
		Equipments:   equipments,
	}, nil
}

func equipProductFilter(proFilter *v1.EquipFilter) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if proFilter.ProductId != nil {
		aggFilter.Filters = append(aggFilter.Filters, addFilter(proFilter.ProductId.FilteringOrder, repo.ProductSearchKeySwidTag.ToString(), proFilter.ProductId.Filteringkey, nil, 0))
	}
	return aggFilter
}

func equipApplicationFilter(proFilter *v1.EquipFilter) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if proFilter.ApplicationId != nil {
		aggFilter.Filters = append(aggFilter.Filters, addFilter(proFilter.ApplicationId.FilteringOrder, repo.ApplicationSearchKeyID.String(), proFilter.ApplicationId.Filteringkey, nil, 0))
	}
	return aggFilter
}
func (s *equipmentServiceServer) GetEquipment(ctx context.Context, req *v1.GetEquipmentRequest) (*v1.GetEquipmentResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, userClaims.Socpes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	resp, err := s.equipmentRepo.Equipment(ctx, eqType, req.EquipId, userClaims.Socpes)
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

func (s *equipmentServiceServer) ListEquipmentParents(ctx context.Context, req *v1.ListEquipmentParentsRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, userClaims.Socpes)
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

	records, resp, err := s.equipmentRepo.EquipmentParents(ctx, eqType, equipParent, req.EquipId, userClaims.Socpes)
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

func (s *equipmentServiceServer) ListEquipmentChildren(ctx context.Context, req *v1.ListEquipmentChildrenRequest) (*v1.ListEquipmentsResponse, error) {
	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, userClaims.Socpes)
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
	records, resp, err := s.equipmentRepo.EquipmentChildren(ctx, equip, equipChild, req.EquipId, queryParams, userClaims.Socpes)
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
			aggregateFilter.Filters = append(aggregateFilter.Filters, addFilter(0, key, val[0], nil, 0))
		case repo.DataTypeInt:
			v, err := strconv.ParseInt(val[0], 10, 64)
			if err != nil {
				// TODO log the error
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for int type attribute cannot be parsed", key, val[0])
			}
			aggregateFilter.Filters = append(aggregateFilter.Filters, addFilter(0, key, v, nil, 0))
		case repo.DataTypeFloat:
			v, err := strconv.ParseFloat(val[0], 10)
			if err != nil {
				// TODO log the error
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for int type attribute cannot be parsed", key, val[0])
			}
			aggregateFilter.Filters = append(aggregateFilter.Filters, addFilter(0, key, v, nil, 0))
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
