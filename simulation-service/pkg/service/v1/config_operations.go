package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	repo "optisam-backend/simulation-service/pkg/repository/v1"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
	"strings"
	"time"

	pTypes "github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteConfig will delete configuration and its all data
func (hcs *SimulationService) DeleteConfig(ctx context.Context, req *v1.DeleteConfigRequest) (*v1.DeleteConfigResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "user does not have access to delete config data")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		// TODO :Transaction Handling - Dharmjit Sir
		err := hcs.repo.DeleteConfig(ctx, db.DeleteConfigParams{
			Status: 2,
			ID:     req.ConfigId,
			Scope:  req.Scope,
		})
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - DeleteConfig - Repo - DeleteConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		err = hcs.repo.DeleteConfigData(ctx, req.ConfigId)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - DeleteConfig - Repo - DeleteConfigData", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		return &v1.DeleteConfigResponse{}, nil
	default:
		logger.Log.Sugar().Errorf("service/v1 - SimulationConfiguration - DeleteConfigData - Unknown Role %v", userClaims.Role)
		return nil, status.Errorf(codes.PermissionDenied, "unknown role : %v", userClaims.Role)
	}
}

// ListConfig lists all the configuration with its attributes
func (hcs *SimulationService) ListConfig(ctx context.Context, req *v1.ListConfigRequest) (*v1.ListConfigResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	// Check if the equipment type is present.
	isEquipType := true
	if req.EquipmentType == "" {
		isEquipType = false
	}
	// Call List Configuration Function
	configs, err := hcs.repo.ListConfig(ctx, db.ListConfigParams{
		IsEquipType:   isEquipType,
		EquipmentType: req.EquipmentType,
		Status:        1,
		Scope:         req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - SimulationConfiguration - ListConfig - Repo - ListConfig", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}
	if len(configs) == 0 {
		return &v1.ListConfigResponse{
			Configurations: []*v1.Configuration{},
		}, nil
	}
	res := make([]*v1.Configuration, 0, len(configs))
	for _, config := range configs {
		metadata, err := hcs.repo.GetMetadatabyConfigID(ctx, config.ID)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - ListConfig - Repo - GetMetadatabyConfigID", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		protoTime, err := pTypes.TimestampProto(config.CreatedOn)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - ListConfig  - timestampProto", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		res = append(res, repoToServConfigs(config, metadata, protoTime))
	}

	return &v1.ListConfigResponse{
		Configurations: res,
	}, nil
}

// CreateConfig will add config values into database
func (hcs *SimulationService) CreateConfig(ctx context.Context, req *v1.CreateConfigRequest) (*v1.CreateConfigResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "User do not have access to create config")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		configs, err := hcs.repo.ListConfig(ctx, db.ListConfigParams{
			IsEquipType:   false,
			EquipmentType: "",
			Status:        1,
			Scope:         req.Scope,
		})
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - CreateConfig - Repo - ListConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}

		// Check if the configuration of the same name exists
		index := configByName(configs, req.ConfigName)
		if index != -1 {
			logger.Log.Error("service/v1 - SimulationConfiguration - CreateConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Configuration with same name already exists")
		}

		err = hcs.repo.CreateConfig(ctx, servToRepoMasterData(userClaims.UserID, req.ConfigName, req.EquipmentType), servToRepoConfigDataAll(req.Data), req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - CreateConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal error")
		}
		return &v1.CreateConfigResponse{}, nil
	default:
		logger.Log.Sugar().Errorf("service/v1 - SimulationConfiguration - CreateConfig - Unknown Role %v", userClaims.Role)
		return nil, status.Errorf(codes.PermissionDenied, "unknown role: %v", userClaims.Role)
	}
}

// UpdateConfig updated the config data
func (hcs *SimulationService) UpdateConfig(ctx context.Context, req *v1.UpdateConfigRequest) (*v1.UpdateConfigResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	switch userClaims.Role {
	case claims.RoleUser:
		return nil, status.Error(codes.PermissionDenied, "User do not have access to update configuration")
	case claims.RoleAdmin, claims.RoleSuperAdmin:
		// Call List Configuration Function
		configs, err := hcs.repo.ListConfig(ctx, db.ListConfigParams{
			IsEquipType:   false,
			EquipmentType: "",
			Status:        1,
			Scope:         req.Scope,
		})
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - UpdateConfig - Repo - ListConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		// Check if the configuration of the given id exists
		index := configByID(configs, req.ConfigId)
		if index == -1 {
			logger.Log.Error("service/v1 - SimulationConfiguration - UpdateConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Configuration not found.")
		}

		// Get Metadata using configID
		metadata, err := hcs.repo.GetMetadatabyConfigID(ctx, req.ConfigId)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - UpdateConfig - Repo - GetMetadatabyConfigID", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal Error")
		}
		// Check if the deletedmetadataIDs are the part of config or not
		ok := checkIfAlreadyConfigured(req.Data, metadata, req.DeletedMetadataIds)
		if !ok {
			logger.Log.Error("service/v1 - SimulationConfiguration - UpdateConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "One or more attribute are already configured")
		}
		// Calling database function to insert data in master table
		err = hcs.repo.UpdateConfig(ctx, req.ConfigId, configs[index].EquipmentType, userClaims.UserID, req.DeletedMetadataIds, servToRepoConfigDataAll(req.Data), req.Scope)
		if err != nil {
			logger.Log.Error("service/v1 - SimulationConfiguration - UpdateConfig", zap.Error(err))
			return nil, status.Error(codes.Internal, "Could not update configuration")
		}
		return &v1.UpdateConfigResponse{}, nil
	default:
		logger.Log.Sugar().Errorf("service/v1 - SimulationConfiguration - UpdateConfig - Unknown Role %v", userClaims.Role)
		return nil, status.Errorf(codes.PermissionDenied, "unknown role: %v", userClaims.Role)
	}
}

// GetConfigData sends the config data back per metadataID
func (hcs *SimulationService) GetConfigData(ctx context.Context, req *v1.GetConfigDataRequest) (*v1.GetConfigDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	// Call List Configuration Function
	configs, err := hcs.repo.ListConfig(ctx, db.ListConfigParams{
		IsEquipType:   false,
		EquipmentType: "",
		Status:        1,
		Scope:         req.Scope,
	})
	if err != nil {
		logger.Log.Error("service/v1 - SimulationConfiguration - GetConfigData - Repo - ListConfig", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}
	// Check if the configuration of the given id exists
	index := configByID(configs, req.ConfigId)
	if index == -1 {
		return nil, status.Error(codes.Internal, "ConfigurationID not found.")
	}

	// Get Metadata using configID
	metadata, err := hcs.repo.GetMetadatabyConfigID(ctx, req.ConfigId)
	if err != nil {
		logger.Log.Error("service/v1 - SimulationConfiguration - GetConfigData - Repo - GetMetadatabyConfigID", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	// Check if the metadata is associated with the given config ID, Why? Because may be the metadataID
	// Can be associated with any other config and we can get the wrong data back
	index = checkMetadataID(req.MetadataId, metadata)
	if index == -1 {
		return nil, status.Error(codes.Internal, "Metadata ID is not associated with given config ID")
	}

	// Call Databasefunction
	configData, err := hcs.repo.GetDataByMetadataID(ctx, req.MetadataId)
	if err != nil {
		logger.Log.Error("service/v1 - SimulationConfiguration - GetConfigData - Repo - GetDataByMetadataID", zap.Error(err))
		return nil, status.Error(codes.Internal, "Internal Error")
	}

	jsonArray := make([]string, 0)

	for _, cd := range configData {
		jsonArray = append(jsonArray, string(cd.JsonData))
	}

	resultant := "[" + strings.Join(jsonArray, ",") + "]"

	return &v1.GetConfigDataResponse{
		Data: []byte(resultant),
	}, nil

}

func checkMetadataID(id int32, realMetadata []db.GetMetadatabyConfigIDRow) int {
	for i, metadata := range realMetadata {
		if metadata.ID == id {
			return i
		}
	}
	return -1
}

func checkIfAlreadyConfigured(data []*v1.Data, metadata []db.GetMetadatabyConfigIDRow, deletedIDS []int32) bool {
	for _, d := range data {
		if i := checkMetadataName(d.Metadata.AttributeName, metadata); i != -1 {
			if j := checkDeletedID(metadata[i].ID, deletedIDS); j == -1 {
				return false
			}
		}
	}
	return true
}

func checkDeletedID(id int32, deletedIds []int32) int {
	for i, rid := range deletedIds {
		if rid == id {
			return i
		}
	}
	return -1
}

func checkMetadataName(name string, realMetadata []db.GetMetadatabyConfigIDRow) int {
	for i, metadata := range realMetadata {
		if metadata.AttributeName == name {
			return i
		}
	}
	return -1
}

func servToRepoMetadata(metadata *v1.Metadata) *repo.Metadata {
	return &repo.Metadata{
		AttributeName:  metadata.AttributeName,
		ConfigFileName: metadata.ConfigFilename,
	}
}

func repoToServMetadata(metadata db.GetMetadatabyConfigIDRow) *v1.Attribute {
	return &v1.Attribute{
		AttributeId:    metadata.ID,
		AttributeName:  metadata.AttributeName,
		ConfigFilename: metadata.ConfigFilename,
	}
}

func servToRepoConfigValueAll(configValues []*v1.ConfigValue) []*repo.ConfigValue {
	repoConfigValues := make([]*repo.ConfigValue, 0)
	for _, configValue := range configValues {
		repoConfigValues = append(repoConfigValues, servToRepoConfigValue(configValue))
	}
	return repoConfigValues
}

func servToRepoConfigValue(configValue *v1.ConfigValue) *repo.ConfigValue {
	return &repo.ConfigValue{
		Key:   configValue.Key,
		Value: configValue.Value,
	}
}

func configByName(configs []db.ListConfigRow, configName string) int {
	for i, config := range configs {
		if config.Name == configName {
			return i
		}
	}
	return -1
}

func servToRepoMasterData(userID, configName, equipType string) *repo.MasterData {
	return &repo.MasterData{
		Name:          configName,
		EquipmentType: equipType,
		Status:        1,
		CreatedBy:     userID,
		CreatedOn:     time.Now().UTC(),
		UpdatedBy:     userID,
		UpdatedOn:     time.Now().UTC(),
	}
}

func configByID(configs []db.ListConfigRow, configID int32) int {
	for i, config := range configs {
		if config.ID == configID {
			return i
		}
	}
	return -1
}

func servToRepoConfigDataAll(data []*v1.Data) []*repo.ConfigData {
	result := make([]*repo.ConfigData, 0)

	for _, d := range data {
		resMetadata := servToRepoMetadata(d.Metadata)
		resValues := servToRepoConfigValueAll(d.Values)

		result = append(result, &repo.ConfigData{
			ConfigMetadata: resMetadata,
			ConfigValues:   resValues,
		})
	}
	return result
}

func repoToServConfigs(config db.ListConfigRow, metadata []db.GetMetadatabyConfigIDRow, createdOn *tspb.Timestamp) *v1.Configuration {
	res := &v1.Configuration{
		ConfigId:      config.ID,
		ConfigName:    config.Name,
		EquipmentType: config.EquipmentType,
		CreatedBy:     config.CreatedBy,
		CreatedOn:     createdOn,
	}
	attributes := make([]*v1.Attribute, 0, len(metadata))

	for _, metadataValue := range metadata {
		attributes = append(attributes, repoToServMetadata(metadataValue))
	}
	res.ConfigAttributes = attributes
	return res
}
