package v1

import (
	"context"
	"encoding/json"
	"fmt"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/strcomp"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/equipment-service/pkg/api/v1"
	repo "optisam-backend/equipment-service/pkg/repository/v1"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	mrepo "optisam-backend/metric-service/pkg/repository/v1"
	prdv1 "optisam-backend/product-service/pkg/api/v1"

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

// This is function is implemented only for oracle.procerssor and oracle.nup metric ony
/* 1. list all the metrics to verify the requested metric exist or not
2. only run the case for orcale.processor and oracle.nup metirc
3. fetch the metric configuration to get the start and end equipment id's and type by hitting byid to true and false params
4. fetch all the equipment type within the scope to verify start and end equipment type exist in scope
5. logically get the partent hierarchy passing the equipment type and start equipment id
6. fetch depth of equipment parent child hierarchy
7. from response get the end equipment id or type
8. Get All equipment passing the end equipment type and id travering from top to bottom to fetch all the equipment in parent child hierarchy
9. Get All equipments where a specific product is deployed with swidtag
10. filter out all the equipments where product is deployed with all the equipment exist in parent hierarchy
11. Create MetricAllocation type looping through for all the valid equipments.
12. Since Create and update has different behaviour and one event supplies the uid while other doesn't as a part of response.
	To avoid the hussle simply fetch all the metric allocation type after create and update.
	Hence  fetch all uid for MetricAllocation type
13. On basis of the product swidtag attach the metic allocation uids to the product to establish linking
14. update all the metricallocation and equipment_user to the postgres DB
*/
// nolint: golint, gosec, funlen, govet, gocyclo
func (s *equipmentServiceServer) UpsertEquipmentAllocatedMetric(ctx context.Context, req *v1.UpsertEquipmentAllocatedMetricRequest) (*v1.UpsertEquipmentResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "scope is not owned by user")
	}

	// First List all the metric to identify only oracle.nup and oracle.processro metircs is passed.
	metrics, err := s.metricClient.ListMetrices(ctx, &metv1.ListMetricRequest{
		Scopes: []string{req.Scope},
	})

	if err != nil {
		logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot upsert allocated metric")
	}
	if metrics == nil || len(metrics.Metrices) == 0 {
		logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", "MetricNotExists"))
		return nil, status.Error(codes.Internal, "MetricNotExists")
	}

	idx := validateMetrics(metrics.Metrices, req.AllocatedMetrics)
	if idx == -1 {
		return nil, status.Error(codes.InvalidArgument, "metric does not exist")
	}

	switch metrics.Metrices[idx].Type {
	case mrepo.MetricOPSOracleProcessorStandard.String(), mrepo.MetricOracleNUPStandard.String():

		metConfigRequestWithID := metv1.GetMetricConfigurationRequest{
			GetID: true,
			MetricInfo: &metv1.Metric{
				Name: req.AllocatedMetrics,
				Type: metrics.Metrices[idx].Type,
			},
			Scopes: []string{req.Scope},
		}

		metConfigRequestWithoutID := metv1.GetMetricConfigurationRequest{
			GetID: false,
			MetricInfo: &metv1.Metric{
				Name: req.AllocatedMetrics,
				Type: metrics.Metrices[idx].Type,
			},
			Scopes: []string{req.Scope},
		}

		metricConfigWithID, err := s.metricClient.GetMetricConfiguration(ctx, &metConfigRequestWithID)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in getting metric configuration")
		}

		metricConfigWithoutID, err := s.metricClient.GetMetricConfiguration(ctx, &metConfigRequestWithoutID)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in getting metric configuration")
		}

		var metConfigWithID map[string]interface{}
		var metConfigWithoutID map[string]interface{}

		err = json.Unmarshal([]byte(metricConfigWithID.MetricConfig), &metConfigWithID)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in unmarshaling metric configuration")
		}

		err = json.Unmarshal([]byte(metricConfigWithoutID.MetricConfig), &metConfigWithoutID)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in unmarshaling metric configuration")
		}

		startEqID := metConfigWithID["StartEqTypeID"]
		EndEqType := metConfigWithoutID["EndEqType"]

		eqTypes, err := s.equipmentRepo.EquipmentTypes(ctx, []string{req.Scope})
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error getting equipment types")
		}

		pHierarchy, err := parentHierarchy(eqTypes, fmt.Sprintf("%v", startEqID))
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error getting parentHierarchy")
		}

		eqpParentHierarchy, err := s.equipmentRepo.ParentsHirerachyForEquipment(ctx, req.EquipmentId, req.EqType, uint8(len(pHierarchy)), req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error getting ParentsHirerachyForEquipment")
		}

		equip := eqpParentHierarchy[0]
		metricAllocationEquipments := []*repo.MetricAllocationRequest{}
		if equip.Parent != nil {
			endEqpFound := false
			endEquID := ""
			for equip != nil {
				if strings.EqualFold(fmt.Sprintf("%v", EndEqType), equip.Type) {
					endEqpFound = true
					endEquID = equip.ID
					break
				}

				for _, v := range *equip.Parent {
					equip = &v
				}
			}

			if !endEqpFound || strings.EqualFold(endEquID, "") {
				logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "error getting end equipment in parent hierarchy")
			}

			query, err := s.equipmentRepo.GetAllEquipmentsInHierarchy(ctx, fmt.Sprintf("%v", EndEqType), endEquID, req.Scope)
			if err != nil {
				logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "error getting GetAllEquipmentsInHierarchy")
			}

			eqmap := make(map[string]*repo.EquipmentParent)

			for _, k := range query.ServerEquipments {

				for _, v := range k.EquipmentParent {
					eq := &repo.EquipmentParent{
						UID:           v.UID,
						EquipmentID:   v.EquipmentID,
						EquipmentType: v.EquipmentType,
					}
					eqmap[v.EquipmentID] = eq
				}
			}

			for _, k := range query.SoftPartitionEquipments {

				for _, v := range k.EquipmentParent {
					eq := &repo.EquipmentParent{
						UID:           v.UID,
						EquipmentID:   v.EquipmentID,
						EquipmentType: v.EquipmentType,
					}
					eqmap[v.EquipmentID] = eq
				}
			}

			dproducts, err := s.equipmentRepo.GetAllEquipmentForSpecifiedProduct(ctx, req.Swidtag, req.Scope)
			if err != nil {
				logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "error getting all equipments for specific deployed product")
			}

			for _, v := range dproducts.Products {
				for _, d := range v.ProductEquipment {
					if _, ok := eqmap[d.EquipmentID]; ok {
						m := &repo.MetricAllocationRequest{
							AllocationMetric: req.AllocatedMetrics,
							Swidtag:          req.Swidtag,
							EquipmentID:      d.EquipmentID,
						}
						metricAllocationEquipments = append(metricAllocationEquipments, m)
					}
				}
			}
		} else {
			m := &repo.MetricAllocationRequest{
				AllocationMetric: req.AllocatedMetrics,
				Swidtag:          req.Swidtag,
				EquipmentID:      equip.EquipID,
			}
			metricAllocationEquipments = append(metricAllocationEquipments, m)
		}
		for _, v := range metricAllocationEquipments {
			err = s.equipmentRepo.UpsertAllocateMetricInEquipmentHierarchy(ctx, v, req.Scope)
			if err != nil {
				logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "error in getting all equipments for specific deployed product")
			}
		}

		allocatedMetrics, err := s.equipmentRepo.GetAllocatedMetrics(ctx, req.Swidtag, req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in getting allocated metrics nodes")
		}

		metAllocationString := []string{}
		for _, v := range allocatedMetrics.AllocatedMetricList {
			metAllocationString = append(metAllocationString, v.UID)
		}

		err = s.equipmentRepo.UpsertAllocateMetricInProduct(ctx, req.Swidtag, metAllocationString, req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
			return nil, status.Error(codes.Internal, "error in linking allocated metric in product")
		}

		for _, vl := range metricAllocationEquipments {
			_, err = s.productClient.UpsertAllocatedMetricEquipment(ctx, &prdv1.UpsertAllocateMetricEquipementRequest{
				Scope:            req.Scope,
				Swidtag:          req.Swidtag,
				EquipmentId:      vl.EquipmentID,
				AllocatedMetrics: req.AllocatedMetrics,
				EquipmentUser:    req.EquipmentUser,
			})
			if err != nil {
				logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
				return nil, status.Error(codes.Internal, "error in getting all equipments for specific deployed product")
			}
		}

	default:
		logger.Log.Error("service/v1 - UpsertEquipmentAllocatedMetric", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "requested metric is not allowed")
	}

	return &v1.UpsertEquipmentResponse{Success: true}, nil
}

func validateMetrics(metrics []*metv1.Metric, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func parentHierarchy(eqTypes []*repo.EquipmentType, startID string) ([]*repo.EquipmentType, error) {
	equip, err := equipmentTypeExistsByID(startID, eqTypes)
	if err != nil {
		logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.NotFound, "cannot fetch equipment type with given Id")
	}
	ancestors := []*repo.EquipmentType{}
	ancestors = append(ancestors, equip)
	parID := equip.ParentID
	for parID != "" {
		equipAnc, err := equipmentTypeExistsByID(parID, eqTypes)
		if err != nil {
			logger.Log.Error("service/v1 - parentHierarchy - fetching equipment type", zap.String("reason", err.Error()))
			return nil, status.Error(codes.NotFound, "parent hierarchy not found")
		}
		ancestors = append(ancestors, equipAnc)
		parID = equipAnc.ParentID
	}
	return ancestors, nil
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

	if error := validateEquipUpdation(metadata.Attributes, equip, req.ParentId, req.Attributes, req.Updattr); error != nil {
		return nil, error
	}
	repoUpdateRequest := &repo.UpdateEquipmentRequest{
		ParentID:   req.ParentId,
		AddAttr:    servAttrToRepoAttrAll(req.Attributes),
		UpdateAttr: servUpdAttrToRepoAttrAll(req.Updattr),
	}
	resp, err := s.equipmentRepo.UpdateEquipmentType(ctx, equip.ID, equip.Type, equip.ParentID, repoUpdateRequest, req.Scopes)
	if err != nil {
		logger.Log.Error("service/v1 -UpdateEquipmentType - updating equipment type", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot update equipment type")
	}

	for _, attr := range req.Updattr {
		for _, eqattr := range equip.Attributes {
			if eqattr.Name == attr.Name {
				eqattr.IsDisplayed = attr.Displayed
				eqattr.IsSearchable = attr.Searchable
				break
			}
		}
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

func validateEquipUpdation(mappedTo []string, equip *repo.EquipmentType, parentID string, newAttr []*v1.Attribute, updAttr []*v1.UpdAttribute) error {
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
	err := validateUpdateAttributes(equip.Attributes, updAttr)
	if err != nil {
		return err
	}
	return validateNewAttributes(mappedTo, equip.Attributes, newAttr)
}

func validateUpdateAttributes(oldAttr []*repo.Attribute, updAttr []*v1.UpdAttribute) error {
	names := make(map[string]struct{})
	mappings := make(map[string]string)

	for _, attr := range oldAttr {
		name := strings.ToUpper(attr.Name)
		names[name] = struct{}{}
		mappings[attr.MappedTo] = name
	}
	// vaidations on attributes
	for _, attr := range updAttr {
		// check if name if unique or not
		name := strings.ToUpper(attr.Name)
		_, ok := names[name]
		if !ok {
			// we arlready have this name for some other attribute
			return status.Errorf(codes.InvalidArgument, "attribute name: %v does not exists", attr.Name)
		}

		if attr.Searchable {
			if !attr.Displayed {
				return status.Error(codes.InvalidArgument, "searchable attribute should always be displayable")
			}
		}
	}
	return nil
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

func servUpdAttrToRepoAttrAll(attrs []*v1.UpdAttribute) []*repo.Attribute {
	servAttrs := make([]*repo.Attribute, len(attrs))
	for i := range attrs {
		servAttrs[i] = servUpdAttrToRepoAttr(attrs[i])
	}
	return servAttrs
}

func servUpdAttrToRepoAttr(attr *v1.UpdAttribute) *repo.Attribute {
	repoAttr := &repo.Attribute{
		ID:           attr.ID,
		Name:         attr.Name,
		IsSearchable: attr.Searchable,
		IsDisplayed:  attr.Displayed,
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
