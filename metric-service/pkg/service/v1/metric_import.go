package v1

import (
	"context"
	"fmt"

	equipv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/thirdparty/equipment-service/pkg/api/v1"
	"go.uber.org/zap"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *metricServiceServer) CreateMetricImport(ctx context.Context, req *v1.MetricImportRequest) (*v1.MetricImportResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	if userClaims.Role == claims.RoleUser {
		return nil, status.Error(codes.PermissionDenied, "only superadmin and Admin user can import metrics")
	}
	eqtypes, er := s.equipments.EquipmentsTypes(ctx, &equipv1.EquipmentTypesRequest{
		Scopes: []string{req.Scope},
	})

	if er != nil {
		logger.Log.Sugar().Errorw("MetricService/v1 - EquipmentTypes - Error while getting equipments data ",
			"scope", req.Scope,
			"status", codes.Internal,
			"reason", er.Error(),
		)
		return nil, status.Error(codes.Internal, "unable to get equipment data")
	}

	m, e := s.EquipmentIDs(ctx, eqtypes)
	if e != nil {
		if e != repo.ErrNoData {
			logger.Log.Sugar().Errorw("metricService/pkg/service/v1/import.go- error while creating maps for IDs",
				"scope", req.Scope,
				"status", codes.Internal,
				"reason", e.Error(),
			)
			return &v1.MetricImportResponse{
				Success: false,
			}, status.Error(codes.Internal, "cannot put Equipment and there Attribute Ids in a map")
		}
	}
	var metExist []string
	metrics, err := s.metricRepo.ListMetrices(ctx, req.Scope)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 -CreateMetricInstanceNumberStandard - ListMetrices", zap.String("reason", err.Error()))
		return nil, status.Error(codes.Internal, "cannot fetch metrics")
	}

	metadata := repo.GlobalMetricMetadata(req.Scope)
	for _, met := range req.Metric {
		for i, val := range metadata {
			if i == met {
				if i == "oracle.processor.standard" {
					MetricOps := &v1.MetricOPS{
						ID:                    val.MetadataOPS.ID,
						Name:                  val.MetadataOPS.Name,
						NumCoreAttrId:         m[val.MetadataOPS.Num_core_attr_id],
						NumCPUAttrId:          m[val.MetadataOPS.NumCPU_attr_id],
						CoreFactorAttrId:      m[val.MetadataOPS.Core_factor_attr_id],
						StartEqTypeId:         m[val.MetadataOPS.Start_eq_type_id],
						BaseEqTypeId:          m[val.MetadataOPS.Base_eq_type_id],
						AggerateLevelEqTypeId: m[val.MetadataOPS.AggerateLevel_eq_type_id],
						EndEqTypeId:           m[val.MetadataOPS.End_eq_type_id],
						Scopes:                val.Scopes,
						Default:               true,
					}
					if metricNameExistsAll(metrics, MetricOps.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricOracleProcessorStandard(ctx, MetricOps)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricOracleProcessorStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "oracle.processor.standard")
						//return nil, status.Error(codes.Internal, "cannot create metric ops")
					}
				} else if i == "oracle.nup.standard" {
					MetricNup := &v1.MetricNUP{
						ID:                    val.MetadataNUP.ID,
						Name:                  val.MetadataNUP.Name,
						NumCoreAttrId:         m[val.MetadataNUP.Num_core_attr_id],
						NumCPUAttrId:          m[val.MetadataNUP.NumCPU_attr_id],
						CoreFactorAttrId:      m[val.MetadataNUP.Core_factor_attr_id],
						StartEqTypeId:         m[val.MetadataNUP.Start_eq_type_id],
						BaseEqTypeId:          m[val.MetadataNUP.Base_eq_type_id],
						AggerateLevelEqTypeId: m[val.MetadataNUP.AggerateLevel_eq_type_id],
						EndEqTypeId:           m[val.MetadataNUP.End_eq_type_id],
						NumberOfUsers:         val.MetadataNUP.Number_of_users,
						Transform:             val.MetadataNUP.Transform,
						TransformMetricName:   val.MetadataNUP.Transform_metric_name,
						Scopes:                val.Scopes,
						Default:               true,
					}
					if metricNameExistsAll(metrics, MetricNup.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricOracleNUPStandard(ctx, MetricNup)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricOracleNUPStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "oracle.nup.standard")
						// return nil, status.Error(codes.Internal, "cannot create metric nup")
					}
				} else if i == "instance.number.standard" {
					MetricInm := &v1.MetricINM{
						ID:               val.MetadataINM.ID,
						Name:             val.MetadataINM.Name,
						NumOfDeployments: val.MetadataINM.Num_Of_Deployments,
						Scopes:           val.Scopes,
						Default:          true,
					}
					if metricNameExistsAll(metrics, MetricInm.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricInstanceNumberStandard(ctx, MetricInm)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricInstanceNumberStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "instance.number.standard")
						// return nil, status.Error(codes.Internal, "cannot create metric inm")
					}
				} else if i == "user.sum.standard" {
					MetricUss := &v1.MetricUSS{
						ID:      val.MetadataUSS.ID,
						Name:    val.MetadataUSS.Name,
						Scopes:  val.Scopes,
						Default: true,
					}
					if metricNameExistsAll(metrics, MetricUss.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricUserSumStandard(ctx, MetricUss)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricUserSumStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "user.sum.standard")
						// return nil, status.Error(codes.Internal, "cannot create metric uss")
					}
				} else if i == "sag.processor.standard" {
					MetricSps := &v1.MetricSPS{
						ID:               val.MetadataSPS.ID,
						Name:             val.MetadataSPS.Name,
						NumCoreAttrId:    m[val.MetadataSPS.Num_core_attr_id],
						NumCPUAttrId:     m[val.MetadataSPS.NumCPU_attr_id],
						CoreFactorAttrId: m[val.MetadataSPS.Core_factor_attr_id],
						BaseEqTypeId:     m[val.MetadataSPS.Base_eq_type_id],
						Scopes:           val.Scopes,
						Default:          true,
					}
					if metricNameExistsAll(metrics, MetricSps.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricSAGProcessorStandard(ctx, MetricSps)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricSAGProcessorStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "sag.processor.standard")
						// return nil, status.Error(codes.Internal, "cannot create metric sps")
					}
				} else if i == "ibm.pvu.standard" {
					MetricIps := &v1.MetricIPS{
						ID:               val.MetadataSPS.ID,
						Name:             val.MetadataSPS.Name,
						NumCoreAttrId:    m[val.MetadataSPS.Num_core_attr_id],
						NumCPUAttrId:     m[val.MetadataSPS.NumCPU_attr_id],
						CoreFactorAttrId: m[val.MetadataSPS.Core_factor_attr_id],
						BaseEqTypeId:     m[val.MetadataSPS.Base_eq_type_id],
						Scopes:           val.Scopes,
						Default:          true,
					}
					if metricNameExistsAll(metrics, MetricIps.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricIBMPvuStandard(ctx, MetricIps)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricIBMPvuStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "ibm.pvu.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric ips")
					}
				} else if i == "microsoft.sql.standard" {
					MetricSQL := &v1.MetricScopeSQL{
						ID:         val.MetadataSQL.ID,
						MetricType: "microsoft.sql.standard",
						MetricName: val.MetadataSQL.MetricName,
						Reference:  val.MetadataSQL.Reference,
						Core:       val.MetadataSQL.Core,
						CPU:        val.MetadataSQL.CPU,
						Default:    val.MetadataSQL.Default,
						Scopes:     val.Scopes,
					}
					if metricNameExistsAll(metrics, MetricSQL.MetricName) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricSQLStandard(ctx, MetricSQL)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricSQLStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "microsoft.sql.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric sql_standard")
					}
				} else if i == "microsoft.sql.enterprise" {
					reqScope := &v1.CreateScopeMetricRequest{
						Scope: req.Scope,
						Type:  "microsoft.sql.enterprise",
					}
					if metricNameExistsAll(metrics, "microsoft.sql.enterprise.2019") != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateScopeMetric(ctx, reqScope)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateScopeMetric",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "microsoft.sql.enterprise")
						//	return nil, status.Error(codes.Internal, "cannot create metric Mse ")
					}
				} else if i == "windows.server.datacenter" {
					reqScope := &v1.CreateScopeMetricRequest{
						Scope: req.Scope,
						Type:  "windows.server.datacenter",
					}
					if metricNameExistsAll(metrics, "windows.server.datacenter.2016") != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateScopeMetric(ctx, reqScope)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateScopeMetric",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "windows.server.datacenter")
						//	return nil, status.Error(codes.Internal, "cannot create metric Data-center")
					}
				} else if i == "user.nominative.standard" {
					MetricUns := &v1.MetricUNS{
						ID:      val.MetadataUNS.ID,
						Name:    val.MetadataUNS.Name,
						Profile: val.MetadataUNS.Profile,
						Scopes:  val.Scopes,
						Default: true,
					}
					if metricNameExistsAll(metrics, MetricUns.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricUserNominativeStandard(ctx, MetricUns)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricUserNominativeStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "user.nominative.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric uns")
					}
				} else if i == "user.concurrent.standard" {
					MetricUcs := &v1.MetricUCS{
						ID:      val.MetadataUNS.ID,
						Name:    val.MetadataUNS.Name,
						Profile: val.MetadataUNS.Profile,
						Scopes:  val.Scopes,
						Default: true,
					}
					if metricNameExistsAll(metrics, MetricUcs.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricUserConcurentStandard(ctx, MetricUcs)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricUserConcurentStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "user.concurrent.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric ucs")
					}
				} else if i == "equipment.attribute.standard" {
					count := 0
					for ite, data := range val.MetadataEquipAttr {
						MetricEquipAtt := &v1.MetricEquipAtt{
							ID:            data.ID,
							Name:          data.Name,
							EqType:        data.Eq_type,
							AttributeName: data.Attribute_name,
							Environment:   data.Environment,
							Value:         data.Value,
							Scopes:        data.Scopes,
							Default:       true,
						}
						if metricNameExistsAll(metrics, MetricEquipAtt.Name) != -1 {
							return nil, status.Error(codes.InvalidArgument, "metric name already exists")
						}
						_, err := s.CreateMetricEquipAttrStandard(ctx, MetricEquipAtt)
						if err != nil {
							logger.Log.Sugar().Errorw("service/v1 - CreateMetricEquipAttrStandard",
								"iteration", ite,
								"scope", req.Scope,
								"status", codes.Internal,
								"reason", err.Error(),
							)
							count += 1
							//	return nil, status.Error(codes.Internal, "cannot create metric equip_attr")
						}
					}
					if count != 0 {
						metExist = append(metExist, "equipment.attribute.standard")
					}
				} else if i == "windows.server.standard" {
					MetricWSS := &v1.MetricWSS{
						ID:         val.MetadataSQL.ID,
						MetricType: "windows.server.standard",
						MetricName: val.MetadataSQL.MetricName,
						Reference:  val.MetadataSQL.Reference,
						Core:       val.MetadataSQL.Core,
						CPU:        val.MetadataSQL.CPU,
						Default:    val.MetadataSQL.Default,
						Scopes:     val.Scopes,
					}
					if metricNameExistsAll(metrics, MetricWSS.MetricName) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricWindowServerStandard(ctx, MetricWSS)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricWindowServerStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "windows.server.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric wss_standard")
					}
				} else if i == "static.standard" {
					MetricSS := &v1.MetricSS{
						ID:             val.MetadataSS.ID,
						Name:           val.MetadataSS.Name,
						ReferenceValue: int32(val.MetadataSS.Reference),
						Default:        true,
						Scopes:         val.Scopes,
					}
					if metricNameExistsAll(metrics, MetricSS.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricStaticStandard(ctx, MetricSS)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricStaticStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "static.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric static.standard")
					}
				} else if i == "attribute.sum.standard" {
					MetricAttrSum := &v1.MetricAttrSum{
						ID:             val.MetadataAttrSum.ID,
						Name:           val.MetadataAttrSum.Name,
						EqType:         val.MetadataAttrSum.Eq_type,
						AttributeName:  val.MetadataAttrSum.Attribute_name,
						ReferenceValue: val.MetadataAttrSum.ReferenceValue,
						Default:        true,
						Scopes:         val.Scopes,
					}
					if metricNameExistsAll(metrics, MetricAttrSum.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricAttrSumStandard(ctx, MetricAttrSum)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricAttrSumStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "attribute.sum.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric attribute.sum.standard")
					}
				} else if i == "attribute.counter.standard" {
					MetricACS := &v1.MetricACS{
						ID:            val.MetadataACS.ID,
						Name:          val.MetadataACS.Name,
						EqType:        val.MetadataACS.Eq_type,
						AttributeName: val.MetadataACS.Attribute_name,
						Value:         val.MetadataACS.Value,
						Default:       true,
						Scopes:        val.Scopes,
					}
					if metricNameExistsAll(metrics, MetricACS.Name) != -1 {
						return nil, status.Error(codes.InvalidArgument, "metric name already exists")
					}
					_, err := s.CreateMetricAttrCounterStandard(ctx, MetricACS)
					if err != nil {
						logger.Log.Sugar().Errorw("service/v1 - CreateMetricAttrCounterStandard",
							"scope", req.Scope,
							"status", codes.Internal,
							"reason", err.Error(),
						)
						metExist = append(metExist, "attribute.counter.standard")
						//	return nil, status.Error(codes.Internal, "cannot create metric attribute.counter.standard")
					}
				} else {
					return nil, status.Error(codes.Internal, "invalid metric")
				}
			}
		}
	}
	if len(metExist) >= 1 {
		errMessage := fmt.Sprintf("cannot create metric: %v", metExist)
		return nil, status.Error(codes.Internal, errMessage)
	}
	return &v1.MetricImportResponse{
		Success: true,
	}, nil
}

func (s *metricServiceServer) EquipmentIDs(ctx context.Context, eqType *equipv1.EquipmentTypesResponse) (map[string]string, error) {
	_, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	myMap := make(map[string]string)

	for _, val := range eqType.EquipmentTypes {
		myMap[val.Type] = val.ID
		for _, data := range val.Attributes {
			myMap[data.Name] = data.ID
		}
	}
	return myMap, nil
}
