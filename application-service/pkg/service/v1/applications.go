// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/application-service/pkg/api/v1"
	repo "optisam-backend/application-service/pkg/repository/v1"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	"optisam-backend/application-service/pkg/worker"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type applicationServiceServer struct {
	applicationRepo repo.Application
	queue           workerqueue.Workerqueue
}

// NewApplicationServiceServer creates Application service
func NewApplicationServiceServer(applicationRepo repo.Application, queue workerqueue.Workerqueue) v1.ApplicationServiceServer {
	return &applicationServiceServer{applicationRepo: applicationRepo, queue: queue}
}

//UpsertApplication create or update Application Resource
//Initialize a new job for license worker
func (s *applicationServiceServer) UpsertApplication(ctx context.Context, req *v1.UpsertApplicationRequest) (*v1.UpsertApplicationResponse, error) {
	logger.Log.Info("Service", zap.Any("UpsertApplication", req))
	err := s.applicationRepo.UpsertApplication(ctx, db.UpsertApplicationParams{
		ApplicationID:      req.GetApplicationId(),
		ApplicationName:    req.GetName(),
		ApplicationOwner:   req.GetOwner(),
		ApplicationVersion: req.GetVersion(),
		Scope:              req.GetScope(),
	})

	if err != nil {
		logger.Log.Error("UpsertApplication", zap.Error(err))
		return &v1.UpsertApplicationResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}

	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.UpsertApplicationRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "lw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))
	return &v1.UpsertApplicationResponse{Success: true}, nil
}

func (s *applicationServiceServer) DeleteApplication(ctx context.Context, req *v1.DeleteApplicationRequest) (*v1.DeleteApplicationResponse, error) {
	return nil, nil
}

func (s *applicationServiceServer) UpsertInstance(ctx context.Context, req *v1.UpsertInstanceRequest) (*v1.UpsertInstanceResponse, error) {
	logger.Log.Info("Service", zap.Any("UpsertInstance", req))
	err := s.applicationRepo.UpsertInstanceTX(ctx, req)
	if err != nil {
		logger.Log.Error("service/v1 - UpsertInstance - UpsertApplicationInstance", zap.Error(err))
		return &v1.UpsertInstanceResponse{Success: false}, status.Error(codes.Internal, "DBError")
	}
	// For Worker Queue
	jsonData, err := json.Marshal(req)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := worker.Envelope{Type: worker.UpsertInstanceRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	jobID, err := s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "lw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "lw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	logger.Log.Info("Successfully pushed job", zap.Int32("jobId", jobID))

	return &v1.UpsertInstanceResponse{Success: true}, nil
}

func (s *applicationServiceServer) DeleteInstance(ctx context.Context, req *v1.DeleteInstanceRequest) (*v1.DeleteInstanceResponse, error) {
	return nil, nil
}

func (s *applicationServiceServer) ListApplications(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {

	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	resp, err := s.applicationRepo.GetApplicationsView(ctx, db.GetApplicationsViewParams{
		Scope:                userClaims.Socpes,
		ProductID:            req.GetSearchParams().GetProductId().GetFilteringkey(),
		IsProductID:          req.GetSearchParams().GetProductId().GetFilteringkey() != "",
		ApplicationName:      req.GetSearchParams().GetName().GetFilteringkey(),
		IsApplicationName:    req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		LkApplicationName:    !req.GetSearchParams().GetName().GetFilterType() && req.GetSearchParams().GetName().GetFilteringkey() != "",
		ApplicationOwner:     req.GetSearchParams().GetOwner().GetFilteringkey(),
		IsApplicationOwner:   req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		LkApplicationOwner:   !req.GetSearchParams().GetOwner().GetFilterType() && req.GetSearchParams().GetOwner().GetFilteringkey() != "",
		ApplicationNameAsc:   strings.Contains(req.GetSortBy().String(), "application_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationNameDesc:  strings.Contains(req.GetSortBy().String(), "application_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ApplicationOwnerAsc:  strings.Contains(req.GetSortBy().String(), "application_owner") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ApplicationOwnerDesc: strings.Contains(req.GetSortBy().String(), "application_owner") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfInstancesAsc:    strings.Contains(req.GetSortBy().String(), "num_of_instances") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfInstancesDesc:   strings.Contains(req.GetSortBy().String(), "num_of_instances") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfProductsAsc:     strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfProductsDesc:    strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:   strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:  strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		CostAsc:              strings.Contains(req.GetSortBy().String(), "cost") && strings.Contains(req.GetSortOrder().String(), "asc"),
		CostDesc:             strings.Contains(req.GetSortBy().String(), "cost") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		logger.Log.Error("service/v1 - ListApplications - GetApplicationsView", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
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
		ListAppResponse.Applications[i].Owner = resp[i].ApplicationOwner
		ListAppResponse.Applications[i].NumOfInstances = resp[i].NumOfInstances
		ListAppResponse.Applications[i].NumOfProducts = resp[i].NumOfProducts
	}
	return &ListAppResponse, nil
}

func (s *applicationServiceServer) ListInstances(ctx context.Context, req *v1.ListInstancesRequest) (*v1.ListInstancesResponse, error) {

	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}

	resp, err := s.applicationRepo.GetInstancesView(ctx, db.GetInstancesViewParams{
		Scope:                   userClaims.Socpes,
		ApplicationID:           req.GetSearchParams().GetApplicationId().GetFilteringkey(),
		IsApplicationID:         req.GetSearchParams().GetApplicationId().GetFilteringkey() != "",
		ProductID:               req.GetSearchParams().GetProductId().GetFilteringkey(),
		IsProductID:             req.GetSearchParams().GetProductId().GetFilteringkey() != "",
		InstanceIDAsc:           strings.Contains(req.GetSortBy().String(), "instance_id") && strings.Contains(req.GetSortOrder().String(), "asc"),
		InstanceIDDesc:          strings.Contains(req.GetSortBy().String(), "instance_id") && strings.Contains(req.GetSortOrder().String(), "asc"),
		InstanceEnvironmentAsc:  strings.Contains(req.GetSortBy().String(), "instance_environment") && strings.Contains(req.GetSortOrder().String(), "asc"),
		InstanceEnvironmentDesc: strings.Contains(req.GetSortBy().String(), "instance_environment") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfProductsAsc:        strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfProductsDesc:       strings.Contains(req.GetSortBy().String(), "num_of_products") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumOfEquipmentsAsc:      strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumOfEquipmentsDesc:     strings.Contains(req.GetSortBy().String(), "num_of_equipments") && strings.Contains(req.GetSortOrder().String(), "desc"),
		//API expect pagenum from 1 but the offset in DB starts with 0
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	})
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to get Instances-> "+err.Error())
	}

	ListAppResponse := v1.ListInstancesResponse{}

	ListAppResponse.Instances = make([]*v1.Instance, len(resp))

	if len(resp) > 0 {
		ListAppResponse.TotalRecords = int32(resp[0].Totalrecords)
	}

	for i := range resp {
		ListAppResponse.Instances[i] = &v1.Instance{}
		ListAppResponse.Instances[i].Id = resp[i].InstanceID
		ListAppResponse.Instances[i].Environment = resp[i].InstanceEnvironment
		ListAppResponse.Instances[i].NumOfProducts = resp[i].NumOfProducts
		ListAppResponse.Instances[i].NumOfEquipments = resp[i].NumOfEquipments

	}
	return &ListAppResponse, nil
}

// func (s *applicationServiceServer) ListProductsForApplication(ctx context.Context, req *v1.ListProductsForApplicationRequest) (*v1.ListProductsForApplicationResponse, error) {
// 	userClaims, ok := ctxmanage.RetrieveClaims(ctx)
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
