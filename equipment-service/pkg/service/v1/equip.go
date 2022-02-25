package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"reflect"

	// "regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/jsonpb" // nolint: staticcheck
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// equipmentServiceServer is implementation of v1.authServiceServer proto interface
type equipmentServiceServer struct {
	equipmentRepo repo.Equipment
	account       accv1.AccountServiceClient
}

// custom json unmarshal
type customTypeFloat float64

func (p *customTypeFloat) UnmarshalJSON(data []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Log.Error("Failed to Unmarshal", zap.Error(err))
		return err
	}
	switch v := tmp.(type) {
	case string:
		floatData, _ := strconv.ParseFloat(v, 64)
		*p = customTypeFloat(floatData)
	default:
	}
	return nil
}

type customTypeInt int64

func (p *customTypeInt) UnmarshalJSON(data []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		logger.Log.Error("Failed to Unmarshal", zap.Error(err))
		return err
	}
	switch v := tmp.(type) {
	case string:
		intData, _ := strconv.ParseInt(v, 10, 64)
		*p = customTypeInt(intData)
	default:
	}
	return nil
}

// NewEquipmentServiceServer creates License service
func NewEquipmentServiceServer(equipmentRepo repo.Equipment, grpcServers map[string]*grpc.ClientConn) v1.EquipmentServiceServer {
	return &equipmentServiceServer{
		equipmentRepo: equipmentRepo,
		account:       accv1.NewAccountServiceClient(grpcServers["account"]),
	}
}

func (s *equipmentServiceServer) DropMetaData(ctx context.Context, req *v1.DropMetaDataRequest) (*v1.DropMetaDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropMetaDataResponse{Success: false}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DropMetaDataResponse{Success: false}, status.Error(codes.InvalidArgument, "scope is not owned by user")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DropMetaDataResponse{Success: false}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	if err := s.equipmentRepo.DropMetaData(ctx, req.Scope); err != nil {
		logger.Log.Error("Failed to delete equipment resource", zap.Error(err))
		return &v1.DropMetaDataResponse{Success: false}, err
	}
	return &v1.DropMetaDataResponse{Success: true}, nil
}

func (s *equipmentServiceServer) UpsertMetadata(ctx context.Context, req *v1.UpsertMetadataRequest) (*v1.UpsertMetadataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.InvalidArgument, "scope is not owned by user")
	}
	if _, err := s.equipmentRepo.UpsertMetadata(ctx, &repo.Metadata{
		MetadataType: req.GetMetadataType(),
		Source:       req.GetMetadataSource(),
		Attributes:   req.GetMetadataAttributes(),
		Scope:        req.GetScope(),
	}); err != nil {
		logger.Log.Error("Failed to upser metadat in dgraph", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot upsert metadata")
	}
	return &v1.UpsertMetadataResponse{Success: true}, nil
}

// UpsertEquipment to load equipment data
// uses reflection heavily
func (s *equipmentServiceServer) UpsertEquipment(ctx context.Context, req *v1.UpsertEquipmentRequest) (*v1.UpsertEquipmentResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "scope is not owned by user")
	}
	validate := validator.New()
	eqTypedata, err := s.equipmentRepo.EquipmentTypeByType(ctx, req.GetEqType(), []string{req.GetScope()})
	if err != nil {
		logger.Log.Error("Failed to get Metadata from dgraph", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot get metadata")
	}

	var reqEqDataJSON bytes.Buffer
	marshaler := &jsonpb.Marshaler{}
	err = marshaler.Marshal(&reqEqDataJSON, req.GetEqData())
	if err != nil {
		logger.Log.Error("unable to marshal to json", zap.Error(err))
	}
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
			reqType = reflect.TypeOf(t)
		case attr.IsParentIdentifier:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.parent"`
			reqType = reflect.TypeOf(t)
		case attr.Type == repo.DataTypeFloat:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.` + req.GetEqType() + `.` + attr.Name + `"`
			reqType = reflect.TypeOf(customTypeFloat(0.0))
		case attr.Type == repo.DataTypeInt:
			regTag = `json:"` + attr.MappedTo + `,omitempty" dbname:"equipment.` + req.GetEqType() + `.` + attr.Name + `"`
			reqType = reflect.TypeOf(customTypeInt(0))
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
	err = validate.Struct(instanceReq)
	if err != nil {
		logger.Log.Error("Validation Error", zap.Error(err))
		return &v1.UpsertEquipmentResponse{Success: false}, status.Error(codes.Unknown, "ValidationError")
	}

	// type conversion
	// reflect.ValueOf(instanceReq).Convert(reflect.TypeOf(instanceDB))

	err = s.equipmentRepo.UpsertEquipment(ctx, req.GetScope(), req.GetEqType(), eqTypedata.ParentType, instanceReq)
	if err != nil {
		logger.Log.Error("UpsertEquipment Failed", zap.Error(err))
		return &v1.UpsertEquipmentResponse{Success: false}, status.Error(codes.Unknown, "DBError")
	}
	return &v1.UpsertEquipmentResponse{Success: true}, nil
}

func (s *equipmentServiceServer) ListEquipments(ctx context.Context, req *v1.ListEquipmentsRequest) (*v1.ListEquipmentsResponse, error) {
	// TODO: fetch only the required equipment type
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "scope is not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
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
		queryParams.InstanceFilter = equipInstanceFilter(req.Filter)
	}

	numOfrecords, equipments, err := s.equipmentRepo.Equipments(ctx, eqType, queryParams, req.Scopes)
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

func equipApplicationFilter(appFilter *v1.EquipFilter) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if appFilter.ApplicationId != nil {
		aggFilter.Filters = append(aggFilter.Filters, addFilter(appFilter.ApplicationId.FilteringOrder, repo.ApplicationSearchKeyID.String(), appFilter.ApplicationId.Filteringkey, nil, 0))
	}
	return aggFilter
}

func equipInstanceFilter(insFilter *v1.EquipFilter) *repo.AggregateFilter {
	aggFilter := new(repo.AggregateFilter)
	//	filter := make(map[int32]repo.Queryable)
	if insFilter.InstanceId != nil {
		aggFilter.Filters = append(aggFilter.Filters, addFilter(insFilter.InstanceId.FilteringOrder, repo.InstanceSearchKeyID.String(), insFilter.InstanceId.Filteringkey, nil, 0))
	}
	return aggFilter
}

func (s *equipmentServiceServer) GetEquipment(ctx context.Context, req *v1.GetEquipmentRequest) (*v1.GetEquipmentResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment types", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	eqType, err := equipmentTypeExistsByID(req.TypeId, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipment - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}

	resp, err := s.equipmentRepo.Equipment(ctx, eqType, req.EquipId, req.Scopes)
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
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "scope is not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
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

	records, resp, err := s.equipmentRepo.EquipmentParents(ctx, eqType, equipParent, req.EquipId, req.Scopes)
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
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "scope is not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
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
		filter, error := parseEquipmentQueryParam(req.SearchParams, equipChild.Attributes)
		if error != nil {
			return nil, error
		}
		queryParams.Filter = filter
	}
	records, resp, err := s.equipmentRepo.EquipmentChildren(ctx, equip, equipChild, req.EquipId, queryParams, req.Scopes)
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

func (s *equipmentServiceServer) DropEquipmentData(ctx context.Context, req *v1.DropEquipmentDataRequest) (*v1.DropEquipmentDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropEquipmentDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.DropEquipmentDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.equipmentRepo.DeleteEquipments(ctx, req.Scope); err != nil {
		return &v1.DropEquipmentDataResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	return &v1.DropEquipmentDataResponse{Success: true}, nil
}

func parseEquipmentQueryParam(query string, attributes []*repo.Attribute) (*repo.AggregateFilter, error) {
	query = strings.Replace(query, ",", "&", -1) // nolint: gocritic
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
			if len(val[0]) < 1 {
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for string type attributes should have at least 1 characters", key, val[0])
			}
			// val[0] = strings.Replace(regexp.QuoteMeta(val[0]), "/", "\\/", -1) // nolint: gocritic
			aggregateFilter.Filters = append(aggregateFilter.Filters, addFilter(0, key, val[0], nil, 0))
		case repo.DataTypeInt:
			v, err := strconv.ParseInt(val[0], 10, 64)
			if err != nil {
				// TODO log the error
				return nil, status.Errorf(codes.InvalidArgument, "attribute: %s cannot be not searched as provided value: %s for int type attribute cannot be parsed", key, val[0])
			}
			aggregateFilter.Filters = append(aggregateFilter.Filters, addFilter(0, key, v, nil, 0))
		case repo.DataTypeFloat:
			v, err := strconv.ParseFloat(val[0], 10) //nolint
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
