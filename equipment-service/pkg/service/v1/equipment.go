package v1

import (
	"context"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	"strings"

	"go.uber.org/zap"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *equipmentServiceServer) CreateGenericScopeEquipmentTypes(ctx context.Context, req *v1.CreateGenericScopeEquipmentTypesRequest) (*v1.CreateGenericScopeEquipmentTypesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		return nil, status.Error(codes.PermissionDenied, "only superadmin user can create equipments")
	}
	metadata := repo.GetGenericScopeMetadata(req.Scope)
	eqTypes := repo.GetGenericScopeEquipmentTypes(req.Scope)
	eqTypeIds := make(map[string]string)
	for _, val := range metadata {
		var uid string
		var err error
		var respEqType *repo.EquipmentType
		if uid, err = s.equipmentRepo.UpsertMetadata(ctx, &val); err != nil { //nolint
			logger.Log.Error("Failed to upser metadata in dgraph", zap.String("reason", err.Error()), zap.Any("scope", req.Scope))
			return nil, status.Error(codes.Internal, "cannot upsert metadata")
		}
		eqTypes[val.Source].SourceID = uid
		eqTypes[val.Source].ParentID = eqTypeIds[eqTypes[val.Source].ParentType]
		if respEqType, err = s.equipmentRepo.CreateEquipmentType(ctx, eqTypes[val.Source], eqTypes[val.Source].Scopes); err != nil {
			logger.Log.Error("Failed to create  eqtype in dgraph", zap.String("reason", err.Error()), zap.Any("scope", req.Scope))
			return nil, status.Error(codes.Internal, "cannot create  eqType")
		}
		eqTypeIds[respEqType.Type] = respEqType.ID
	}

	return &v1.CreateGenericScopeEquipmentTypesResponse{
		Success: true,
	}, nil
}

func (s *equipmentServiceServer) ListEquipmentsMetadata(ctx context.Context, req *v1.ListEquipmentMetadataRequest) (*v1.ListEquipmentMetadataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - ListEquipmentsMetadata - query parameter", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	res, err := s.equipmentRepo.MetadataAllWithType(ctx, repo.MetadataTypeEquipment, req.Scopes)
	if err != nil {
		switch err { // nolint: gocritic
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment metadata")
		}
		logger.Log.Error("service/v1 - ListEquipmentsMetadata - query parameter", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment metadata")
	}

	servMetadata := make([]*v1.EquipmentMetadata, 0, len(res))

	switch req.Type {
	case v1.ListEquipmentMetadataRequest_ALL:
		servMetadata = repoMetadataToSrvMetadataAll(res)
	case v1.ListEquipmentMetadataRequest_UN_MAPPED:
		for _, metadata := range res {
			if metadataSourceUsed(metadata.ID, eqTypes) >= 0 {
				continue
			}
			servMetadata = append(servMetadata, repoMetadataToSrvMetadata(metadata))
		}
	case v1.ListEquipmentMetadataRequest_MAPPED:
		for _, metadata := range res {
			if metadataSourceUsed(metadata.ID, eqTypes) >= 0 {
				servMetadata = append(servMetadata, repoMetadataToSrvMetadata(metadata))
			}
		}
	default:
		logger.Log.Error("service/v1 - ListEquipmentsMetadata - query parameter", zap.String("Type", req.Type.String()))
		return nil, status.Error(codes.Internal, "unknown parameter in request.Type")
	}

	return &v1.ListEquipmentMetadataResponse{
		Metadata: servMetadata,
	}, nil
}

func (s *equipmentServiceServer) EquipmentsTypes(ctx context.Context, req *v1.EquipmentTypesRequest) (*v1.EquipmentTypesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	res, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - EquipmentsTypes - query parameter", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	return &v1.EquipmentTypesResponse{
		EquipmentTypes: repoEquipTypeToServiceTypeAll(res),
	}, nil
}

func (s *equipmentServiceServer) DeleteEquipmentType(ctx context.Context, req *v1.DeleteEquipmentTypeRequest) (*v1.DeleteEquipmentTypeResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	// check if equipment type exists
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, []string{req.Scope})
	if err != nil {
		logger.Log.Debug("service/v1 - DeleteEquipmentType - repo/EquipmentTypes -", zap.String("reason", err.Error()))
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	idx := equipmentTypeExistsByType(req.EquipType, eqTypes)
	if idx == -1 {
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.NotFound, "equipment type does not exist")
	}
	// check if equipment type has children
	_, err = s.equipmentRepo.EquipmentTypeChildren(ctx, eqTypes[idx].ID, len(eqTypes), []string{req.Scope})
	if err != nil {
		if err != repo.ErrNoData {
			logger.Log.Debug("service/v1 - DeleteEquipmentType - repo/quipmentTypeChildren - ", zap.String("reason", err.Error()))
			return &v1.DeleteEquipmentTypeResponse{
				Success: false,
			}, status.Error(codes.Internal, "cannot fetch equipment type children")
		}
	} else {
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "equipment type has children")
	}
	// check if equipments data exists
	numEquipments, _, err := s.equipmentRepo.Equipments(ctx, eqTypes[idx], &repo.QueryEquipments{
		PageSize:  50,
		Offset:    offset(50, 1),
		SortOrder: sortOrder(v1.SortOrder_ASC),
	}, []string{req.Scope})
	if err != nil {
		if err != repo.ErrNoData {
			logger.Log.Debug("service/v1 - DeleteEquipmentType - repo/Equipments -", zap.String("reason", err.Error()))
			return &v1.DeleteEquipmentTypeResponse{
				Success: false,
			}, status.Error(codes.Internal, "cannot fetch equipments")
		}
	}
	if numEquipments != 0 {
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.InvalidArgument, "equipment type contains equipments data")
	}
	if err := s.equipmentRepo.DeleteEquipmentType(ctx, req.EquipType, req.Scope); err != nil {
		logger.Log.Debug("service/v1 - DeleteEquipmentType - repo/DeleteEquipmentType - ", zap.String("reason", err.Error()))
		return &v1.DeleteEquipmentTypeResponse{
			Success: false,
		}, status.Error(codes.Internal, "cannot delete equipment type")
	}
	return &v1.DeleteEquipmentTypeResponse{
		Success: true,
	}, nil
}

func (s *equipmentServiceServer) CreateEquipmentType(ctx context.Context, req *v1.EquipmentType) (*v1.EquipmentType, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	scopeinfo, err := s.account.GetScope(ctx, &accv1.GetScopeRequest{Scope: req.Scopes[0]})
	if err != nil {
		logger.Log.Error("service/v1 - CreateEquipmentType - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "unable to fetch scope info")
	}
	if scopeinfo.ScopeType == accv1.ScopeType_GENERIC.String() {
		return nil, status.Error(codes.PermissionDenied, "can not create equipment type for generic scope")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateEquipmentType - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}

	if metadataSourceUsed(req.MetadataId, eqTypes) >= 0 {
		return nil, status.Error(codes.InvalidArgument, "data source is already consumed by another equipment type")
	}

	// check if type name is available or not
	for _, eqt := range eqTypes {
		if strcomp.CompareStrings(eqt.Type, req.Type) {
			return nil, status.Errorf(codes.InvalidArgument, "type name: %v is not available", req.Type)
		}
	}

	metadata, err := s.equipmentRepo.MetadataWithID(ctx, req.MetadataId, req.Scopes)
	if err != nil {
		switch err { // nolint: gocritic
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment metadata")
		}

		logger.Log.Error("service/v1 - CreateEquipmentType - fetching metadata with id", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment metadata")
	}

	if error := validateEquipCreation(metadata.Attributes, eqTypes, req); error != nil {
		return nil, error
	}

	resp, err := s.equipmentRepo.CreateEquipmentType(ctx, servEquipTypeToRepoType(req), req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - CreateEquipmentType - creating equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot create equipment type")
	}
	return repoEquipTypeToServiceType(resp), nil
}

// nolint: gocyclo
func (s *equipmentServiceServer) UpdateEquipmentType(ctx context.Context, req *v1.UpdateEquipmentTypeRequest) (*v1.EquipmentType, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	if userClaims.Role != claims.RoleSuperAdmin {
		scopeinfo, err := s.account.GetScope(ctx, &accv1.GetScopeRequest{Scope: req.Scopes[0]})
		if err != nil {
			logger.Log.Error("service/v1 - UpdateEquipmentType - account/GetScope - fetching scope info", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "unable to fetch scope info")
		}
		if scopeinfo.ScopeType == accv1.ScopeType_GENERIC.String() {
			return nil, status.Error(codes.PermissionDenied, "can not update equipment type for generic scope")
		}
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateEquipmentType - fetching equipments", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	equip, err := equipmentTypeExistsByID(req.Id, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - UpdateEquipmentType - fetching equipment", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment with given Id")
	}

	metadata, err := s.equipmentRepo.MetadataWithID(ctx, equip.SourceID, req.Scopes)
	if err != nil {
		switch err { // nolint: gocritic
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment metadata")
		}

		logger.Log.Error("service/v1 - UpdateEquipmentType - fetching metadata with id", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment metadata")
	}
	if req.ParentId != "" && req.ParentId != equip.ParentID {
		if req.ParentId == req.Id {
			return nil, status.Error(codes.InvalidArgument, "equipment type cannot be parent of itself")
		}
		// check if parent exists ot not
		_, error := equipmentTypeExistsByID(req.ParentId, eqTypes)
		if error != nil {
			return nil, status.Error(codes.InvalidArgument, "parent not found")
		}
		// check if parent is any of the children of equip
		equipChildren, error := s.equipmentRepo.EquipmentTypeChildren(ctx, req.Id, len(eqTypes), req.Scopes)
		if error != nil {
			if error != repo.ErrNoData {
				logger.Log.Error("service/v1 - UpdateEquipmentType - EquipmentTypeChildren - fetching equipment type children", zap.String("reason", error.Error()))
				return nil, status.Error(codes.Internal, "cannot fetch equipment type children")
			}
		} else {
			_, error = equipmentTypeExistsByID(req.ParentId, equipChildren)
			if error == nil {
				return nil, status.Error(codes.InvalidArgument, "child can not be parent")
			}
		}
		// if parent id already exits
		if equip.ParentID != "" {
			// check if data exists
			numEquipments, _, error := s.equipmentRepo.Equipments(ctx, equip, &repo.QueryEquipments{
				PageSize:  50,
				Offset:    offset(50, 1),
				SortOrder: sortOrder(v1.SortOrder_ASC),
			}, req.Scopes)
			if error != nil {
				if error != repo.ErrNoData {
					logger.Log.Error("service/v1 - UpdateEquipmentType - Equipments - fetching equipments for eqType", zap.String("reason", error.Error()))
					return nil, status.Error(codes.Internal, "cannot fetch equipments")
				}
			}
			if numEquipments != 0 {
				return nil, status.Error(codes.InvalidArgument, "equipment type contains equipments data")
			}
		}
	}
	if error := validateEquipUpdation(metadata.Attributes, equip, req.ParentId, req.Attributes); error != nil {
		return nil, error
	}
	repoUpdateRequest := &repo.UpdateEquipmentRequest{
		ParentID: req.ParentId,
		Attr:     servAttrToRepoAttrAll(req.Attributes),
	}
	resp, err := s.equipmentRepo.UpdateEquipmentType(ctx, equip.ID, equip.Type, equip.ParentID, repoUpdateRequest, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 -UpdateEquipmentType - updating equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot update equipment type")
	}

	if req.ParentId != "" {
		equip.ParentID = req.ParentId
	}
	equip.Attributes = append(equip.Attributes, resp...)
	return repoEquipTypeToServiceType(equip), nil
}

func (s *equipmentServiceServer) GetEquipmentMetadata(ctx context.Context, req *v1.EquipmentMetadataRequest) (*v1.EquipmentMetadata, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scopes...) {
		return nil, status.Error(codes.InvalidArgument, "some claims are not owned by user")
	}
	metadata, err := s.equipmentRepo.MetadataWithID(ctx, req.ID, req.Scopes)
	if err != nil {
		switch err { // nolint: gocritic
		case repo.ErrNoData:
			return nil, status.Error(codes.NotFound, "cannot fetch equipment metadata")
		}

		logger.Log.Error("service/v1 -GetEquipmentMetadata - fetching metadata with id", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment metadata")
	}
	eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipmentMetadata - query parameter", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch equipment types")
	}
	ind := metadataSourceUsed(metadata.ID, eqTypes)
	if ind == -1 {
		return repoMetadataToSrvMetadata(metadata), nil
	}
	metadataAttr := make([]string, 0, len(metadata.Attributes))
	switch req.Attributes {
	case v1.EquipmentMetadataRequest_All:
		return repoMetadataToSrvMetadata(metadata), nil
	case v1.EquipmentMetadataRequest_Mapped:
		for _, attr := range metadata.Attributes {
			if attributeUsed(attr, eqTypes[ind].Attributes) {
				metadataAttr = append(metadataAttr, attr)
			}
		}
	case v1.EquipmentMetadataRequest_Unmapped:
		for _, attr := range metadata.Attributes {
			if attributeUsed(attr, eqTypes[ind].Attributes) {
				continue
			}
			metadataAttr = append(metadataAttr, attr)
		}
	}
	metadata.Attributes = metadataAttr
	return repoMetadataToSrvMetadata(metadata), nil
}

func equipmentTypeExistsByID(id string, eqTypes []*repo.EquipmentType) (*repo.EquipmentType, error) {
	for _, eqt := range eqTypes {
		if eqt.ID == id {
			return eqt, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "equipment not exists")
}

func equipmentTypeExistsByType(eqType string, eqTypes []*repo.EquipmentType) int {
	for i := 0; i < len(eqTypes); i++ {
		if eqTypes[i].Type == eqType {
			return i
		}
	}
	return -1
}

func attributeUsed(name string, attr []*repo.Attribute) bool {
	for _, attrMap := range attr {
		if name == attrMap.MappedTo {
			return true
		}
	}
	return false
}

func validateEquipUpdation(mappedTo []string, equip *repo.EquipmentType, parentID string, newAttr []*v1.Attribute) error {
	countParentKey := 0
	for _, attr := range newAttr {
		if attr.PrimaryKey {
			return status.Error(codes.InvalidArgument, "primary key not required")
		}
		if attr.ParentIdentifier {
			countParentKey++
			if attr.DataType != v1.DataTypes_STRING {
				return status.Error(codes.InvalidArgument, "only string data type is allowed for parent identifier")
			}
		}
	}
	if countParentKey == 0 { // nolint: gocritic
		if equip.ParentID == "" && parentID != "" {
			return status.Error(codes.InvalidArgument, "one parent identifier required")
		}
	} else if countParentKey == 1 {
		if equip.ParentID != "" {
			return status.Error(codes.InvalidArgument, "no parent identifier required when parent is already present")
		}
		if parentID == "" {
			return status.Error(codes.InvalidArgument, "parent is not selected for equipment type")
		}
	} else {
		return status.Errorf(codes.InvalidArgument, "multiple parent keys are found")
	}
	return validateNewAttributes(mappedTo, equip.Attributes, newAttr)
}

func validateNewAttributes(mappedTo []string, oldAttr []*repo.Attribute, newAttr []*v1.Attribute) error {
	names := make(map[string]struct{})
	mappings := make(map[string]string)

	for _, attr := range oldAttr {
		name := strings.ToUpper(attr.Name)
		names[name] = struct{}{}
		mappings[attr.MappedTo] = name
	}
	// vaidations on attributes
	for _, attr := range newAttr {
		// check if name if unique or not
		name := strings.ToUpper(attr.Name)
		_, ok := names[name]
		if ok {
			// we arlready have this name for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute name: %v, is already given to some other attribute", attr.Name)
		}

		// atttribute name does not exist before
		// make an entry
		names[name] = struct{}{}
		// check if mapping of equipment exists
		mappingFound := false
		for _, mapping := range mappedTo {
			if mapping == attr.MappedTo {
				mappingFound = true
				break
			}
		}

		if !mappingFound {
			return status.Errorf(codes.InvalidArgument, "mapping:%v does not exist", attr.MappedTo)
		}

		attrName, ok := mappings[attr.MappedTo]
		if ok {
			// mapping is already assigned to some other attributes for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute mapping: %v, is already given to attribte: %v", attr.MappedTo, attrName)
		}

		// atttribute mapping does not exist before
		// make an entry
		mappings[attr.MappedTo] = attr.Name

		if attr.Searchable {
			if !attr.Displayed {
				return status.Error(codes.InvalidArgument, "searchable attribute should always be displayable")
			}
		}
	}
	return nil
}

func repoEquipTypeToServiceTypeAll(eqTypes []*repo.EquipmentType) []*v1.EquipmentType {
	servEqTypes := make([]*v1.EquipmentType, len(eqTypes))
	for i := range eqTypes {
		servEqTypes[i] = repoEquipTypeToServiceType(eqTypes[i])
	}
	return servEqTypes
}

func servEquipTypeToRepoType(eqType *v1.EquipmentType) *repo.EquipmentType {
	return &repo.EquipmentType{
		ID:         eqType.ID,
		Type:       eqType.Type,
		ParentID:   eqType.ParentId,
		SourceID:   eqType.MetadataId,
		Scopes:     eqType.Scopes,
		Attributes: servAttrToRepoAttrAll(eqType.Attributes),
	}
}

func repoEquipTypeToServiceType(eqType *repo.EquipmentType) *v1.EquipmentType {
	return &v1.EquipmentType{
		ID:             eqType.ID,
		Type:           eqType.Type,
		ParentId:       eqType.ParentID,
		ParentType:     eqType.ParentType,
		MetadataId:     eqType.SourceID,
		MetadataSource: eqType.SourceName,
		Scopes:         eqType.Scopes,
		Attributes:     repoAttrToServiceAttrAll(eqType.Attributes),
	}
}

func servAttrToRepoAttrAll(attrs []*v1.Attribute) []*repo.Attribute {
	servAttrs := make([]*repo.Attribute, len(attrs))
	for i := range attrs {
		servAttrs[i] = servAttrToRepoAttr(attrs[i])
	}
	return servAttrs
}

func servAttrToRepoAttr(attr *v1.Attribute) *repo.Attribute {
	repoAttr := &repo.Attribute{
		ID:                 attr.ID,
		Name:               attr.Name,
		Type:               repo.DataType(attr.DataType),
		IsIdentifier:       attr.PrimaryKey,
		IsSearchable:       attr.Searchable,
		IsDisplayed:        attr.Displayed,
		IsParentIdentifier: attr.ParentIdentifier,
		MappedTo:           attr.MappedTo,
		IsSimulated:        attr.Simulated,
	}

	switch attr.DataType {
	case v1.DataTypes_INT:
		repoAttr.IntVal = int(attr.GetIntVal())
		repoAttr.IntValOld = int(attr.GetIntValOld())
	case v1.DataTypes_FLOAT:
		repoAttr.FloatVal = attr.GetFloatVal()
		repoAttr.FloatValOld = attr.GetFloatValOld()
	case v1.DataTypes_STRING:
		repoAttr.StringVal = attr.GetStringVal()
		repoAttr.StringValOld = attr.GetStringValOld()
	}

	return repoAttr

}

func repoAttrToServiceAttrAll(attrs []*repo.Attribute) []*v1.Attribute {
	servAttrs := make([]*v1.Attribute, len(attrs))
	for i := range attrs {
		servAttrs[i] = repoAttrToServiceAttr(attrs[i])
	}
	return servAttrs
}

func repoAttrToServiceAttr(attr *repo.Attribute) *v1.Attribute {
	return &v1.Attribute{
		ID:               attr.ID,
		Name:             attr.Name,
		DataType:         v1.DataTypes(attr.Type),
		PrimaryKey:       attr.IsIdentifier,
		Searchable:       attr.IsSearchable,
		Displayed:        attr.IsDisplayed,
		ParentIdentifier: attr.IsParentIdentifier,
		MappedTo:         attr.MappedTo,
	}
}
func validateEquipCreation(mappedTo []string, eqTypes []*repo.EquipmentType, eqType *v1.EquipmentType) error {
	// valibate if we have a valid parent or not
	// Parent Found
	if eqType.ParentId != "" {
		parentExists := false
		for _, eqt := range eqTypes {
			if eqt.ID == eqType.ParentId {
				parentExists = true
				break
			}
		}
		if !parentExists {
			return status.Errorf(codes.InvalidArgument, "parent with ID: %v is not found", eqType.ParentId)
		}
	}

	// ensure that we have a single primary key
	countPK := 0
	countParentKey := 0
	for _, attr := range eqType.Attributes {
		if attr.PrimaryKey {
			countPK++
		}
		if attr.ParentIdentifier {
			countParentKey++
		}
	}

	switch {
	case countPK == 0:
		return status.Error(codes.InvalidArgument, "one of attributes must be of primary key type")
	case countPK > 1:
		return status.Errorf(codes.InvalidArgument, "multiple primary keys:%v are found in attributes only one primary key is allowed", countPK)
	}

	if eqType.ParentId == "" && countParentKey > 0 {
		return status.Error(codes.InvalidArgument, "parent key is not required when parent is not selected for equipment type ")
	}

	if countParentKey > 1 {
		return status.Errorf(codes.InvalidArgument, "multiple parent keys:%v are found in attributes only one parent key is allowed", countParentKey)
	}

	return validateAttribute(mappedTo, eqType)
}

func validateAttribute(mappedTo []string, eqType *v1.EquipmentType) error {
	names := make(map[string]struct{})
	mappings := make(map[string]string)
	// vaidations on attributes
	for _, attr := range eqType.Attributes {
		// check if name if unique or not
		name := strings.ToUpper(attr.Name)
		_, ok := names[name]
		if ok {
			// we arlready have this name for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute name: %v, is already given to some other attribte", attr.Name)
		}

		// atttribute name does not exist before
		// make an entry
		names[name] = struct{}{}
		// check if mapping of equipment exists
		mappingFound := false
		for _, mapping := range mappedTo {
			if mapping == attr.MappedTo {
				mappingFound = true
				break
			}
		}

		if !mappingFound {
			return status.Errorf(codes.InvalidArgument, "mapping:%v does not exist", attr.MappedTo)
		}

		attrName, ok := mappings[attr.MappedTo]
		if ok {
			// mapping is already assigned to some other attributes for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute mapping: %v, is already given to attribte: %v", attr.MappedTo, attrName)
		}

		// atttribute mapping does not exist before
		// make an entry
		mappings[attr.MappedTo] = attr.Name

		if attr.PrimaryKey && attr.ParentIdentifier {
			return status.Error(codes.InvalidArgument, "atrritbute can be either primary key or parent key")
		}

		if attr.PrimaryKey {
			// type of primary key should be string only
			if attr.DataType != v1.DataTypes_STRING {
				return status.Error(codes.InvalidArgument, "only string data type is allowed for primary key")
			}
			if !attr.Displayed {
				return status.Error(codes.InvalidArgument, "primary key should always be displayable")
			}
		}

		if attr.ParentIdentifier {
			// type of primary key should be string only
			if attr.DataType != v1.DataTypes_STRING {
				return status.Error(codes.InvalidArgument, "only string data type is allowed for parent key")
			}
		}

		if attr.Searchable {
			if !attr.Displayed {
				return status.Error(codes.InvalidArgument, "searchable attribute should always be displayable")
			}
		}

	}
	return nil
}

func metadataSourceUsed(sourceID string, eqTypes []*repo.EquipmentType) int {
	for i, eqType := range eqTypes {
		if sourceID == eqType.SourceID {
			return i
		}
	}
	return -1
}

func repoMetadataToSrvMetadata(metadata *repo.Metadata) *v1.EquipmentMetadata {
	return &v1.EquipmentMetadata{
		ID:         metadata.ID,
		Name:       metadata.Source,
		Attributes: metadata.Attributes,
		Scopes:     []string{metadata.Scope},
	}
}

func repoMetadataToSrvMetadataAll(metadata []*repo.Metadata) []*v1.EquipmentMetadata {
	servMetadata := make([]*v1.EquipmentMetadata, 0, len(metadata))
	for _, mtdata := range metadata {
		servMetadata = append(servMetadata, repoMetadataToSrvMetadata(mtdata))
	}
	return servMetadata
}
