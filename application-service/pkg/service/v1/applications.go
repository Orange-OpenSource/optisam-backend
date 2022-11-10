package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/application-service/pkg/api/v1"
	repo "optisam-backend/application-service/pkg/repository/v1"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	dgworker "optisam-backend/application-service/pkg/worker/dgraph"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	prov1 "optisam-backend/product-service/pkg/api/v1"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type applicationServiceServer struct {
	applicationRepo repo.Application
	product         prov1.ProductServiceClient
	queue           workerqueue.Workerqueue
}

// NewApplicationServiceServer creates Application service
func NewApplicationServiceServer(applicationRepo repo.Application, queue workerqueue.Workerqueue, grpcServers map[string]*grpc.ClientConn) v1.ApplicationServiceServer {
	return &applicationServiceServer{
		applicationRepo: applicationRepo,
		queue:           queue,
		product:         prov1.NewProductServiceClient(grpcServers["product"]),
	}
}

func (s *applicationServiceServer) DropObscolenscenceData(ctx context.Context, req *v1.DropObscolenscenceDataRequest) (*v1.DropObscolenscenceDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropObscolenscenceDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return &v1.DropObscolenscenceDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if userClaims.Role != claims.RoleSuperAdmin {
		return &v1.DropObscolenscenceDataResponse{Success: false}, status.Error(codes.PermissionDenied, "RoleValidationError")
	}

	if err := s.applicationRepo.DropObscolenscenceDataTX(ctx, req.Scope); err != nil {
		logger.Log.Error("Failed To delete obscolenscene resource", zap.Error(err))
		return &v1.DropObscolenscenceDataResponse{Success: false}, status.Error(codes.PermissionDenied, "DBError")
	}
	return &v1.DropObscolenscenceDataResponse{Success: true}, nil
}

// UpsertApplication create or update Application Resource
// Initialize a new job for license dgworker
func (s *applicationServiceServer) UpsertApplication(ctx context.Context, req *v1.UpsertApplicationRequest) (*v1.UpsertApplicationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	appDomain := "Not specified"
	if req.GetDomain() != "" {
		appDomain = req.GetDomain()
	}
	err := s.applicationRepo.UpsertApplication(ctx, db.UpsertApplicationParams{
		ApplicationID:   req.GetApplicationId(),
		ApplicationName: req.GetName(),
		// ApplicationOwner:   req.GetOwner(),
		// ApplicationVersion: req.GetVersion(),
		ApplicationEnvironment: req.GetEnvironment(),
		ApplicationDomain:      appDomain,
		Scope:                  req.GetScope(),
	})

	if err != nil {
		logger.Log.Error("UpsertApplication", zap.Error(err))
		return &v1.UpsertApplicationResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertApplicationRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "lw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}

	return &v1.UpsertApplicationResponse{Success: true}, nil
}

func (s *applicationServiceServer) UpsertApplicationEquip(ctx context.Context, req *v1.UpsertApplicationEquipRequest) (*v1.UpsertApplicationEquipResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	err := s.applicationRepo.UpsertApplicationEquipTx(ctx, req)
	if err != nil {
		logger.Log.Error("UpsertApplicationEquip Failed", zap.Error(err))
		return &v1.UpsertApplicationEquipResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// For dgworker Queue
	// jsonData, err := json.Marshal(req)
	// if err != nil {
	// 	logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	// }
	// e := dgworker.Envelope{Type: dgworker.UpsertApplicationEquipRequest, JSON: jsonData}

	// envolveData, err := json.Marshal(e)
	// if err != nil {
	// 	logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	// }

	// _, err = s.queue.PushJob(ctx, job.Job{
	// 	Type:   sql.NullString{String: "lw"},
	// 	Status: job.JobStatusPENDING,
	// 	Data:   envolveData,
	// }, "lw")
	// if err != nil {
	// 	logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	// }

	return &v1.UpsertApplicationEquipResponse{Success: true}, nil
}

func (s *applicationServiceServer) DeleteApplication(ctx context.Context, req *v1.DeleteApplicationRequest) (*v1.DeleteApplicationResponse, error) {
	return nil, nil
}

func (s *applicationServiceServer) DropApplicationData(ctx context.Context, req *v1.DropApplicationDataRequest) (*v1.DropApplicationDataResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return &v1.DropApplicationDataResponse{Success: false}, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return &v1.DropApplicationDataResponse{Success: false}, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	if err := s.applicationRepo.DropApplicationDataTX(ctx, req.Scope); err != nil {
		return &v1.DropApplicationDataResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DropApplicationDataRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "lw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.DropApplicationDataResponse{Success: true}, nil
}

func (s *applicationServiceServer) UpsertInstance(ctx context.Context, req *v1.UpsertInstanceRequest) (*v1.UpsertInstanceResponse, error) {
	err := s.applicationRepo.UpsertInstanceTX(ctx, req)
	if err != nil {
		logger.Log.Error("service/v1 - UpsertInstance - UpsertApplicationInstance", zap.Error(err))
		return &v1.UpsertInstanceResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For dgworker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertInstanceRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "lw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}

	return &v1.UpsertInstanceResponse{Success: true}, nil
}

func (s *applicationServiceServer) DeleteInstance(ctx context.Context, req *v1.DeleteInstanceRequest) (*v1.DeleteInstanceResponse, error) {
	return nil, nil
}

// nolint: gocyclo
func (s *applicationServiceServer) ListApplications(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if req.SearchParams != nil {
		if req.SearchParams.ProductId != nil {
			if len(req.SearchParams.ProductId.Filteringkey) != 0 {
				apiresp, err := s.listApplicationsByProductSwidtags(ctx, req)
				if err != nil {
					return nil, err
				}
				return apiresp, nil
			}
		}
	}
	apiresp, err := s.listApplicationsView(ctx, req)
	if err != nil {
		return nil, err
	}
	return apiresp, nil
}

func (s *applicationServiceServer) ListInstances(ctx context.Context, req *v1.ListInstancesRequest) (*v1.ListInstancesResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()...) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	equipmentFilter := []string{}
	if req.SearchParams != nil {
		if req.SearchParams.ProductId != nil {
			if req.SearchParams.ProductId.Filteringkey != "" {
				prodEquipments, err := s.product.GetEquipmentsByProduct(ctx, &prov1.GetEquipmentsByProductRequest{
					Scope:   req.Scopes[0],
					SwidTag: req.SearchParams.ProductId.Filteringkey,
				})
				if err != nil {
					logger.Log.Error("service/v1 - ListInstances - product/GetEquipmentsByProduct", zap.Error(err))
					return nil, status.Error(codes.Internal, "ServiceError")
				}
				if prodEquipments != nil {
					equipmentFilter = prodEquipments.EquipmentId
				}
			}
		}
	}
	resp, err := s.applicationRepo.GetInstancesView(ctx, db.GetInstancesViewParams{
		Scope:                   req.GetScopes(),
		ProductID:               req.GetSearchParams().GetProductId().GetFilteringkey(),
		ApplicationID:           req.GetSearchParams().GetApplicationId().GetFilteringkey(),
		IsApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
		IsProductID:             req.GetSearchParams().GetProductId().GetFilteringkey() != "",
		InstanceIDAsc:           strings.Contains(req.GetSortBy().String(), "instance_id") && (req.SortOrder == v1.SortOrder_asc),
		InstanceIDDesc:          strings.Contains(req.GetSortBy().String(), "instance_id") && (req.SortOrder == v1.SortOrder_desc),
		InstanceEnvironmentAsc:  strings.Contains(req.GetSortBy().String(), "instance_environment") && (req.SortOrder == v1.SortOrder_asc),
		InstanceEnvironmentDesc: strings.Contains(req.GetSortBy().String(), "instance_environment") && (req.SortOrder == v1.SortOrder_desc),
		NumOfProductsAsc:        strings.Contains(req.GetSortBy().String(), "num_of_products") && (req.SortOrder == v1.SortOrder_asc),
		NumOfProductsDesc:       strings.Contains(req.GetSortBy().String(), "num_of_products") && (req.SortOrder == v1.SortOrder_desc),
		PageNum:                 req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize:                req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListInstances - repo/GetInstancesView", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get Instances")
	}

	listInsResponse := &v1.ListInstancesResponse{}

	if len(resp) > 0 {
		listInsResponse.TotalRecords = int32(resp[0].Totalrecords)
	}
	instances := []*v1.Instance{}
	for _, dbins := range resp {
		ins := &v1.Instance{
			Id:            dbins.InstanceID,
			Environment:   dbins.InstanceEnvironment,
			NumOfProducts: dbins.NumOfProducts,
		}
		numEquip, err := s.applicationRepo.GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
			EquipmentIds:    equipmentFilter,
			Scope:           req.Scopes[0],
			InstanceID:      dbins.InstanceID,
			ProductID:       req.GetSearchParams().GetProductId().GetFilteringkey(),
			ApplicationID:   req.GetSearchParams().GetApplicationId().GetFilteringkey(),
			IsApplicationID: req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
			IsProductID:     req.GetSearchParams().GetProductId().GetFilteringkey() != "",
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get num of equipments")
		}
		ins.NumOfEquipments = int32(numEquip[0])
		instances = append(instances, ins)
	}
	listInsResponse.Instances = instances
	return listInsResponse, nil
}

// func (s *applicationServiceServer) ListProductsForApplication(ctx context.Context, req *v1.ListProductsForApplicationRequest) (*v1.ListProductsForApplicationResponse, error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "cannot find claims in context")
// 	}
// 	applicationID := req.GetApplicationId()
// 	resp, err := s.applicationRepo.GetProductsForApplication(ctx, applicationID, userClaims.Socpes)
// 	if err != nil {
// 		return nil, status.Error(codes.Unknown, "failed to get Products-> "+err.Error())
// 	}

// 	prodForAppResponse := &v1.ListProductsForApplicationResponse{}

// 	prodForAppResponse.Products = make([]*v1.ProductForApplication, len(resp.Products))

// 	prodForAppResponse.TotalRecords = resp.NumOfRecords[0].TotalCnt

// 	for i, prod := range resp.Products {
// 		prodForAppResponse.Products[i] = &v1.ProductForApplication{
// 			SwidTag:         prod.SwidTag,
// 			Name:            prod.Name,
// 			Editor:          prod.Editor,
// 			Version:         prod.Version,
// 			NumofEquipments: prod.NumOfEquipments,
// 			NumOfInstances:  prod.NumOfInstances,
// 			TotalCost:       prod.TotalCost,
// 		}

// 	}
// 	return prodForAppResponse, nil
// }

// func addFilter(priority int32, key string, value interface{}, values []string, filterType v1.StringFilter_Type) *repo.Filter {
// 	return &repo.Filter{
// 		FilteringPriority:   priority,
// 		FilterKey:           key,
// 		FilterValue:         value,
// 		FilterValueMultiple: utils.StringToInterface(values),
// 		FilterMatchingType:  filterTypev1(filterType),
// 	}
// }

// func filterTypev1(filterType v1.StringFilter_Type) repo.Filtertype {

// 	switch filterType {
// 	case v1.StringFilter_REGEX:
// 		return repo.RegexFilter
// 	case v1.StringFilter_EQ:
// 		return repo.EqFilter
// 	default:
// 		return repo.RegexFilter
// 	}
// }

// func applicationFilter(params *v1.ApplicationSearchParams) *repo.AggregateFilter {
// 	aggFilter := new(repo.AggregateFilter)
// 	//	filter := make(map[int32]v1.Queryable)
// 	if params.Name != nil {
// 		aggFilter.Filters = append(aggFilter.Filters, addFilter(params.Name.FilteringOrder, "name", params.Name.Filteringkey, nil, 0))
// 	}
// 	if params.ApplicationOwner != nil {
// 		aggFilter.Filters = append(aggFilter.Filters, addFilter(params.ApplicationOwner.FilteringOrder, "application_owner", params.ApplicationOwner.Filteringkey, nil, 0))
// 	}
// 	sort.Sort(aggFilter)

// 	return aggFilter
// }

func (s *applicationServiceServer) GetEquipmentsByApplication(ctx context.Context, req *v1.GetEquipmentsByApplicationRequest) (*v1.GetEquipmentsByApplicationResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFound")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		logger.Log.Error("GetEquipmentsByApplication - Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	application, err := s.applicationRepo.GetApplicationEquip(ctx, db.GetApplicationEquipParams{
		Scope:         req.Scope,
		ApplicationID: req.ApplicationId,
	})
	if err != nil {
		logger.Log.Error("service/v1 - GetEquipmentsByApplication - error from repo/GetApplicationEquip", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	resp := v1.GetEquipmentsByApplicationResponse{}

	if len(application) > 0 {
		resp.TotalRecords = int32(application[0].Totalrecords)
	}
	for i := range application {
		resp.EquipmentId[i] = application[i].EquipmentID
	}
	return &resp, nil
}

// func (s *applicationServiceServer) GetProductsByApplication(ctx context.Context, req *v1.GetProductsByApplicationRequest) (*v1.GetProductsByApplicationResponse, error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "ClaimsNotFound")
// 	}
// 	if !helper.Contains(userClaims.Socpes, req.Scope) {
// 		logger.Log.Error("GetProductsByApplication - Permission Error", zap.Any("Scopes", userClaims.Socpes), zap.String("Requested Scope", req.GetScope()))
// 		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
// 	}
// 	products, err := s.applicationRepo.GetProductsByApplicationID(ctx, db.GetProductsByApplicationIDParams{
// 		Scope:         req.Scope,
// 		ApplicationID: req.ApplicationId,
// 	})
// 	if err != nil {
// 		logger.Log.Error("service/v1 - GetProductsByApplication - error from repo/GetProductsByApplicationID", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "DBError")
// 	}
// 	return &v1.GetProductsByApplicationResponse{ProductId: products}, nil
// }

// nolint: gocyclo
func (s *applicationServiceServer) listApplicationsView(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	resp, err := s.applicationRepo.GetApplicationsView(ctx, db.GetApplicationsViewParams{
		Scope:                      req.GetScopes(),
		ApplicationEnvironment:     req.GetSearchParams().GetEnvironment(),
		ApplicationDomain:          req.GetSearchParams().GetDomain().GetFilteringkey(),
		ApplicationName:            req.GetSearchParams().GetName().GetFilteringkey(),
		ObsolescenceRisk:           req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey(),
		IsApplicationName:          req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		IsApplicationDomain:        req.GetSearchParams().GetDomain().GetFilterType() && req.GetSearchParams().GetDomain().GetFilteringkey() != "",
		IsObsolescenceRisk:         req.GetSearchParams().GetObsolescenceRisk().GetFilterType() && req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey() != "",
		LkApplicationName:          !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkApplicationDomain:        !req.GetSearchParams().GetDomain().GetFilterType() && req.GetSearchParams().GetDomain().GetFilteringkey() != "",
		LkObsolescenceRisk:         !req.GetSearchParams().GetObsolescenceRisk().GetFilterType() && req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey() != "",
		ApplicationOwner:           req.GetSearchParams().GetOwner().GetFilteringkey(),
		IsApplicationOwner:         req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		LkApplicationOwner:         !req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		ApplicationNameAsc:         strings.Contains(req.GetSortBy().String(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationNameDesc:        strings.Contains(req.GetSortBy().String(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationOwnerAsc:        strings.Contains(req.GetSortBy().String(), "owner") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationOwnerDesc:       strings.Contains(req.GetSortBy().String(), "owner") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:         strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:        strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationDomainAsc:       strings.Contains(req.GetSortBy().String(), "domain") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationDomainDesc:      strings.Contains(req.GetSortBy().String(), "domain") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationEnvironmentAsc:  strings.Contains(req.GetSortBy().String(), "environment") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationEnvironmentDesc: strings.Contains(req.GetSortBy().String(), "environment") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ObsolescenceRiskAsc:        strings.Contains(req.GetSortBy().String(), "obsolescence_risk") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ObsolescenceRiskDesc:       strings.Contains(req.GetSortBy().String(), "obsolescence_risk") && strings.Contains(req.GetSortOrder().String(), "desc"),
		// API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListApplications - GetApplicationsView", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}

	app, err := s.product.GetProductCountByApp(ctx, &prov1.GetProductCountByAppRequest{
		Scope: req.Scopes[0],
	})
	if err != nil {
		logger.Log.Error("service/v1 - listApplicationsView - product/GetProductCountByApp", zap.Error(err))
		return nil, status.Error(codes.Internal, "ServiceError")
	}

	ListAppResponse := v1.ListApplicationsResponse{}

	ListAppResponse.Applications = make([]*v1.Application, len(resp))

	if len(resp) > 0 {
		ListAppResponse.TotalRecords = int32(resp[0].Totalrecords)
	}

	for i := range resp {
		ListAppResponse.Applications[i] = &v1.Application{}
		ListAppResponse.Applications[i].Name = resp[i].ApplicationName
		ListAppResponse.Applications[i].ApplicationId = resp[i].ApplicationID
		// ListAppResponse.Applications[i].Owner = resp[i].ApplicationOwner
		// ListAppResponse.Applications[i].NumOfInstances = resp[i].NumOfInstances
		for j := range resp {
			if resp[i].ApplicationID == app.AppData[j].ApplicationId {
				ListAppResponse.Applications[i].NumOfProducts = int32(app.AppData[j].NumOfProducts)
				break
			}
		}
		ListAppResponse.Applications[i].Environment = resp[i].ApplicationEnvironment
		ListAppResponse.Applications[i].Domain = resp[i].ApplicationDomain
		ListAppResponse.Applications[i].ObsolescenceRisk = resp[i].ObsolescenceRisk.String
		ListAppResponse.Applications[i].NumOfEquipments = resp[i].NumOfEquipments
	}
	return &ListAppResponse, nil
}

// nolint: gocyclo
func (s *applicationServiceServer) listApplicationsByProductSwidtags(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	app, err := s.product.GetApplicationsByProduct(ctx, &prov1.GetApplicationsByProductRequest{
		Scope:   req.Scopes[0],
		Swidtag: req.SearchParams.ProductId.Filteringkey,
	})
	if err != nil {
		logger.Log.Error("service/v1 - listApplicationsByProductSwidtags - product/GetApplicationsByProduct", zap.Error(err))
		return nil, status.Error(codes.Internal, "ServiceError")
	}
	resp, err := s.applicationRepo.GetApplicationsByProduct(ctx, db.GetApplicationsByProductParams{
		Scope:                 req.Scopes,
		ApplicationID:         app.ApplicationId,
		ApplicationDomain:     req.GetSearchParams().GetDomain().GetFilteringkey(),
		ApplicationName:       req.GetSearchParams().GetName().GetFilteringkey(),
		ObsolescenceRisk:      req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey(),
		IsApplicationName:     req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		IsApplicationDomain:   req.GetSearchParams().GetDomain().GetFilterType() && req.GetSearchParams().GetDomain().GetFilteringkey() != "",
		IsObsolescenceRisk:    req.GetSearchParams().GetObsolescenceRisk().GetFilterType() && req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey() != "",
		LkApplicationName:     !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkApplicationDomain:   !req.GetSearchParams().GetDomain().GetFilterType() && req.GetSearchParams().GetDomain().GetFilteringkey() != "",
		LkObsolescenceRisk:    !req.GetSearchParams().GetObsolescenceRisk().GetFilterType() && req.GetSearchParams().GetObsolescenceRisk().GetFilteringkey() != "",
		ApplicationOwner:      req.GetSearchParams().GetOwner().GetFilteringkey(),
		IsApplicationOwner:    req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		LkApplicationOwner:    !req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		ApplicationNameAsc:    strings.Contains(req.GetSortBy().String(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationNameDesc:   strings.Contains(req.GetSortBy().String(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationOwnerAsc:   strings.Contains(req.GetSortBy().String(), "owner") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationOwnerDesc:  strings.Contains(req.GetSortBy().String(), "owner") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:    strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:   strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationDomainAsc:  strings.Contains(req.GetSortBy().String(), "domain") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationDomainDesc: strings.Contains(req.GetSortBy().String(), "domain") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ObsolescenceRiskAsc:   strings.Contains(req.GetSortBy().String(), "obsolescence_risk") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ObsolescenceRiskDesc:  strings.Contains(req.GetSortBy().String(), "obsolescence_risk") && strings.Contains(req.GetSortOrder().String(), "desc"),
		// API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListApplications - GetApplicationsView", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	listAppResponse := &v1.ListApplicationsResponse{}
	listAppResponse.Applications = make([]*v1.Application, len(resp))
	if len(resp) > 0 {
		listAppResponse.TotalRecords = int32(resp[0].Totalrecords)
	}
	for i := range resp {
		listAppResponse.Applications[i] = &v1.Application{}
		listAppResponse.Applications[i].Name = resp[i].ApplicationName
		listAppResponse.Applications[i].ApplicationId = resp[i].ApplicationID
		listAppResponse.Applications[i].Owner = resp[i].ApplicationOwner
		listAppResponse.Applications[i].Domain = resp[i].ApplicationDomain
		listAppResponse.Applications[i].ObsolescenceRisk = resp[i].ObsolescenceRisk.String
		listAppResponse.Applications[i].NumOfEquipments = resp[i].NumOfEquipments
	}
	return listAppResponse, nil
}
