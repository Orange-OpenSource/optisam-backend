package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	//"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	YYYYMMDD string = "2006-01-02"
	DDMMYYYY string = "02-01-2006"
)

var (
	NoOfRetries = "no_of_retries"
)
var dateFormats = []string{YYYYMMDD, DDMMYYYY}

// UpsertProductConcurrentUser will add or update product concurrent users
func (s *ProductServiceServer) UpsertProductConcurrentUser(ctx context.Context, req *v1.ProductConcurrentUserRequest) (*v1.ProductConcurrentUserResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	var swittag string
	// Check if aggregation is true
	if !req.IsAggregations {

		// listEditors, err := s.ProductRepo.ListEditors(ctx, []string{req.GetScope()})
		// if err != nil {
		// 	logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		// 	return nil, status.Error(codes.Internal, "DBError")
		// }
		// if !helper.Contains(listEditors, req.GetProductEditor()) {
		// 	return nil, status.Error(codes.PermissionDenied, "editor doesn't exists")
		// }

		pName := removeSpecialChars(req.GetProductName())
		pEditor := removeSpecialChars(req.GetProductEditor())

		if req.GetProductVersion() != "" {
			swittag = strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{pName, pEditor, req.GetProductVersion()}, "_"), " ", "_"), "-", "_")
		} else {
			swittag = strings.ReplaceAll(strings.ReplaceAll(strings.Join([]string{pName, pEditor}, "_"), " ", "_"), "-", "_")
		}
		req.Swidtag = swittag
		_, errProduct := s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{Scope: req.GetScope(), Swidtag: swittag})
		if errProduct != nil {
			if errProduct == sql.ErrNoRows {
				productUpsertReq := &v1.UpsertProductRequest{
					Name:        req.GetProductName(),
					Editor:      req.GetProductEditor(),
					Version:     req.GetProductVersion(),
					Scope:       req.GetScope(),
					SwidTag:     swittag,
					ProductType: v1.Producttype_saas,
				}
				_, err := s.UpsertProduct(ctx, productUpsertReq)
				if err != nil {
					logger.Log.Error("UpsertProduct Failed", zap.Error(err))
					return nil, status.Error(codes.Internal, "DBError")
				}
			} else {
				logger.Log.Error("service/v1 - GetProductInformation - GetProductInformation", zap.Error(errProduct))
				return nil, status.Error(codes.Internal, "DBError unable to get product info")
			}
		}

	} else {
		_, err := s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
			ID:    req.GetAggregationId(),
			Scope: req.GetScope(),
		})
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.PermissionDenied, "Aggregation doesn't exists")
		}
		if err != nil {
			logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
	}

	err := s.ProductRepo.UpsertConcurrentUserTx(ctx, req, userClaims.UserID)
	if err != nil {
		logger.Log.Error("can ", zap.Error(err))
		return nil, status.Error(codes.Internal, "can not add product concurrent users")
	}

	upsertConcurrentReqDgraph := UpsertConcurrentUserDgraphRequest(req, userClaims.UserID)
	currentDateTime := time.Now()
	theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), 1, 00, 00, 00, 000, time.Local)

	if req.GetId() > 0 {
		pConUser, err := s.ProductRepo.GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
		if err != nil {
			logger.Log.Error("failed to update product concurrent user, unable to get data", zap.Error(err))
		}
		purchaseDate := pConUser.PurchaseDate
		if purchaseDate.Month() == currentDateTime.Month() {
			theDate = time.Date(purchaseDate.Year(), purchaseDate.Month(), 1, 00, 00, 00, 000, time.Local)
		}
	}

	upsertConcurrentReqDgraph.PurchaseDate = theDate.String()

	// For Worker Queue
	jsonData, err := json.Marshal(upsertConcurrentReqDgraph)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertConcurrentUserRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	logger.Log.Sugar().Infow("ARg", "swidtag", swittag)
	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	return &v1.ProductConcurrentUserResponse{Success: true}, nil
}

// ListConcurrentUsers will return list of concurrent users
func (s *ProductServiceServer) ListConcurrentUsers(ctx context.Context, req *v1.ListConcurrentUsersRequest) (*v1.ListConcurrentUsersResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	listCouncurrentReq := db.ListConcurrentUsersParams{
		Scope:               []string{req.GetScopes()},
		IsAggregations:      req.IsAggregation,
		LkProductName:       !req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		ProductName:         req.GetSearchParams().GetProductName().GetFilteringkey(),
		IsProductName:       req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		LkEditorName:        !req.GetSearchParams().GetProductEditor().GetFilterType() && req.GetSearchParams().GetProductEditor().GetFilteringkey() != "",
		ProductEditor:       req.GetSearchParams().GetProductEditor().GetFilteringkey(),
		IsEditorName:        req.GetSearchParams().GetProductEditor().GetFilterType() && req.GetSearchParams().GetProductEditor().GetFilteringkey() != "",
		LkAggregationName:   !req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		AggregationName:     req.GetSearchParams().GetAggregationName().GetFilteringkey(),
		IsAggregationName:   req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		LkProductVersion:    !req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
		ProductVersion:      req.GetSearchParams().GetProductVersion().GetFilteringkey(),
		IsProductVersion:    req.GetSearchParams().GetProductVersion().GetFilterType(),
		LkTeam:              !req.GetSearchParams().GetTeam().GetFilterType() && req.GetSearchParams().GetTeam().GetFilteringkey() != "",
		Team:                req.GetSearchParams().GetTeam().GetFilteringkey(),
		IsTeam:              req.GetSearchParams().GetTeam().GetFilterType() && req.GetSearchParams().GetTeam().GetFilteringkey() != "",
		LkProfileUser:       !req.GetSearchParams().GetProfileUser().GetFilterType() && req.GetSearchParams().GetProfileUser().GetFilteringkey() != "",
		ProfileUser:         req.GetSearchParams().GetProfileUser().GetFilteringkey(),
		IsProfileUser:       req.GetSearchParams().GetProfileUser().GetFilterType() && req.GetSearchParams().GetProfileUser().GetFilteringkey() != "",
		LkNumberOfUsers:     !req.GetSearchParams().GetNumberOfUsers().GetFilterType() && req.GetSearchParams().GetNumberOfUsers().GetFilteringkey() != "",
		NumberOfUsers:       req.GetSearchParams().GetNumberOfUsers().GetFilteringkey(),
		IsNumberOfUsers:     req.GetSearchParams().GetNumberOfUsers().GetFilterType() && req.GetSearchParams().GetNumberOfUsers().GetFilteringkey() != "",
		ProductNameAsc:      strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:     strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AggregationNameAsc:  strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggregationNameDesc: strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:   strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:  strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProfileUserAsc:      strings.Contains(req.GetSortBy(), "profile_user") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProfileUserDesc:     strings.Contains(req.GetSortBy(), "profile_user") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TeamAsc:             strings.Contains(req.GetSortBy(), "team") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TeamDesc:            strings.Contains(req.GetSortBy(), "team") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumberOfUsersAsc:    strings.Contains(req.GetSortBy(), "number_of_users") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumberOfUsersDesc:   strings.Contains(req.GetSortBy(), "number_of_users") && strings.Contains(req.GetSortOrder().String(), "desc"),
		PurchaseDateAsc:     strings.Contains(req.GetSortBy(), "purchase_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
		PurchaseDateDesc:    strings.Contains(req.GetSortBy(), "purchase_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductEditorAsc:    strings.Contains(req.GetSortBy(), "product_editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductEditorDesc:   strings.Contains(req.GetSortBy(), "product_editor") && strings.Contains(req.GetSortOrder().String(), "desc"),

		// API expect pagenum from 1 but the offset in DB starts
		PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
		PageSize: req.GetPageSize(),
	}

	if req.GetSearchParams().GetPurchaseDate().IsValid() {
		listCouncurrentReq.IsPurchaseDate = true
		listCouncurrentReq.PurchaseDate = req.GetSearchParams().PurchaseDate.AsTime()
	}
	dbresp, err := s.ProductRepo.ListConcurrentUsers(ctx, listCouncurrentReq)
	if err != nil {
		logger.Log.Error("service/v1 - listConcurrentUsers - db/ListConcurrentUsers", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ListConcurrentUsersResponse{}
	apiresp.ConcurrentUser = make([]*v1.ConcurrentUser, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.ConcurrentUser[i] = &v1.ConcurrentUser{}
		apiresp.ConcurrentUser[i].IsAggregation = dbresp[i].IsAggregations.Bool
		apiresp.ConcurrentUser[i].AggregationId = dbresp[i].AggregationID.Int32
		apiresp.ConcurrentUser[i].AggregationName = dbresp[i].AggregationName.String
		apiresp.ConcurrentUser[i].ProductName = dbresp[i].ProductName.String
		apiresp.ConcurrentUser[i].ProductVersion = dbresp[i].ProductVersion.String
		apiresp.ConcurrentUser[i].Team = dbresp[i].Team.String
		apiresp.ConcurrentUser[i].ProfileUser = dbresp[i].ProfileUser.String
		apiresp.ConcurrentUser[i].NumberOfUsers = dbresp[i].NumberOfUsers.Int32
		apiresp.ConcurrentUser[i].PurchaseDate = timestamppb.New(dbresp[i].UpdatedOn)
		apiresp.ConcurrentUser[i].ProductEditor = dbresp[i].ProductEditor.String
		apiresp.ConcurrentUser[i].Id = dbresp[i].ID

	}
	return &apiresp, nil
}

// DeleteConcurrentUsers will be delete a record from storage
func (s *ProductServiceServer) DeleteConcurrentUsers(ctx context.Context, req *v1.DeleteConcurrentUsersRequest) (*v1.DeleteConcurrentUsersResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	pConUser, err := s.ProductRepo.GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
	if err != nil {
		logger.Log.Error("failed to delete product concurrent user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	err = s.ProductRepo.DeletConcurrentUserByID(ctx, db.DeletConcurrentUserByIDParams{
		Scope: req.GetScope(), ID: req.GetId(),
	})

	if err != nil {
		logger.Log.Error("failed to delete product concurrent user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	deleteConcurrentReqDgraph := DeleteConcurrentUserRequest(pConUser)
	deleteConcurrentReqDgraph.Scope = req.GetScope()
	// For Worker Queue
	jsonData, err := json.Marshal(deleteConcurrentReqDgraph)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DeleteConcurrentUserRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	//logger.Log.Sugar().Infow("ARg", "swidtag", swittag)
	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	if !pConUser.IsAggregations.Bool {
		productDeleted := s.DeleteSaasProductUsers(ctx, pConUser.Swidtag.String, pConUser.Scope)
		logger.Log.Sugar().Debugw("Product is deleted ",
			"productDeleted", productDeleted,
			"swidtag", pConUser.Swidtag.String,
			"scope", pConUser.Scope,
		)
	}

	return &v1.DeleteConcurrentUsersResponse{Success: true}, nil
}

// UpsertConcurrentUserDgraphRequest will map v1.ProductConcurrentUserReq to dgworker.UpserConcurrentUserRequest
func UpsertConcurrentUserDgraphRequest(req *v1.ProductConcurrentUserRequest, createdBy string) (resp dgworker.UpserConcurrentUserRequest) {
	resp.IsAggregations = req.GetIsAggregations()
	resp.AggregationID = 0
	resp.SwidTag = req.Swidtag
	if req.GetIsAggregations() && req.GetAggregationId() > 0 {
		resp.AggregationID = req.GetAggregationId()
		resp.SwidTag = ""
	}
	resp.Editor = req.GetProductEditor()
	resp.ProductName = req.GetProductVersion()
	resp.ProductVersion = req.GetProductVersion()
	resp.Scope = req.GetScope()
	resp.NumberOfUsers = req.GetNumberOfUsers()
	resp.ProfileUser = req.GetProfileUser()
	resp.Team = req.GetTeam()
	resp.CreatedBy = createdBy
	return
}

// DeleteConcurrentUserRequest will map db.ProductConcurrentUser to dgworker.UpserConcurrentUserRequest
func DeleteConcurrentUserRequest(dbConUser db.ProductConcurrentUser) (resp dgworker.UpserConcurrentUserRequest) {
	resp.IsAggregations = dbConUser.IsAggregations.Bool
	resp.AggregationID = 0
	resp.SwidTag = dbConUser.Swidtag.String
	if dbConUser.IsAggregations.Bool && dbConUser.AggregationID.Int32 > 0 {
		resp.AggregationID = dbConUser.AggregationID.Int32
		resp.SwidTag = ""
	}

	purchaseDate := dbConUser.PurchaseDate
	theDate := time.Date(purchaseDate.Year(), purchaseDate.Month(), purchaseDate.Day(), 00, 00, 00, 000, time.Local)
	resp.PurchaseDate = theDate.String()
	return
}

// func (s *ProductServiceServer) UpsertNominativeUser(ctx context.Context, req *v1.UpserNominativeUserRequest) (resp *v1.UpserNominativeUserResponse, err error) {
// 	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 	if !ok {
// 		return nil, status.Error(codes.Internal, "cannot find claims in context")
// 	}
// 	if req.GetScope() == "" {
// 		return nil, status.Error(codes.Internal, "cannot find scope")
// 	}
// 	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
// 		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
// 	}
// 	if !(req.AggregationId > 0) && (req.Editor == "" || req.ProductName == "") {
// 		return nil, status.Error(codes.InvalidArgument, "Either aggrigation or product details are required")
// 	}
// 	var swid = ""
// 	aggr := db.Aggregation{}
// 	if req.GetAggregationId() > 0 {
// 		aggr, err = s.ProductRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
// 			ID:    req.GetAggregationId(),
// 			Scope: req.GetScope(),
// 		})
// 		if err == sql.ErrNoRows {
// 			return nil, status.Error(codes.InvalidArgument, "Aggregation doesn't exists")
// 		}
// 		if err != nil {
// 			logger.Log.Error("service/v1 - UpsertNominativeUser - GetAggregationByID", zap.Error(err))
// 			return nil, status.Error(codes.Internal, "DBError")
// 		}
// 	} else {
// 		// listEditors, err := s.ProductRepo.ListEditors(ctx, []string{req.GetScope()})
// 		// if err != nil {
// 		// 	logger.Log.Error("service/v1 - UpsertNominativeUser - ListEditors", zap.Error(err))
// 		// 	return nil, status.Error(codes.Internal, "DBError")
// 		// }
// 		// if !helper.Contains(listEditors, req.GetEditor()) {
// 		// 	return nil, status.Error(codes.InvalidArgument, "editor doesn't exists")
// 		// }

// 		pName := removeSpecialChars(req.ProductName)
// 		pEditor := removeSpecialChars(req.GetEditor())

// 		if req.GetProductVersion() != "" {
// 			swid = strings.ReplaceAll(strings.Join([]string{pName, pEditor, req.GetProductVersion()}, "_"), " ", "_")
// 		} else {
// 			swid = strings.ReplaceAll(strings.Join([]string{pName, pEditor}, "_"), " ", "_")
// 		}
// 		_, err = s.ProductRepo.GetProductInformation(ctx, db.GetProductInformationParams{
// 			Scope:   req.GetScope(),
// 			Swidtag: swid,
// 		})
// 		if err != nil && err != sql.ErrNoRows {
// 			logger.Log.Error("service/v1 - UpsertNominativeUser - GetProductInformation", zap.Error(err))
// 			return nil, status.Error(codes.Internal, "DBError")
// 		}
// 		if err == sql.ErrNoRows {
// 			_, err := s.UpsertProduct(ctx, &v1.UpsertProductRequest{
// 				SwidTag:     swid,
// 				Name:        req.GetProductName(),
// 				Editor:      req.GetEditor(),
// 				Scope:       req.GetScope(),
// 				Version:     req.GetProductVersion(),
// 				ProductType: v1.Producttype_saas,
// 			})
// 			if err != nil {
// 				logger.Log.Error("service/v1 - UpsertNominativeUser - UpsertProductRequest :can not add product", zap.Error(err))
// 				return nil, status.Error(codes.Internal, "can not add product")
// 			}
// 		}
// 	}
// 	ppId := uuid.New().String()
// 	//validUsers,inValidUsers,err:=filterNominativeUsers(req)
// 	err, users := s.ProductRepo.UpsertNominativeUsersTx(ctx, req, userClaims.UserID, userClaims.UserID, swid, ppId)
// 	if err != nil {
// 		logger.Log.Error("service/v1 - UpsertNominativeUser - UpsertNominativeUsersTx : can not upsert users", zap.Error(err))
// 		return nil, status.Error(codes.Internal, "can not upsert users")
// 	}
// 	upsertNominativeReqDgraph := prepairUpsertNominativeUserDgraphRequest(req, swid, userClaims.UserID, aggr.AggregationName, users)
// 	for _, v := range upsertNominativeReqDgraph {

// 		// For Worker Queue
// 		jsonData, err := json.Marshal(v)
// 		if err != nil {
// 			logger.Log.Error("Failed to do json marshalling", zap.Error(err))
// 		}
// 		e := dgworker.Envelope{Type: dgworker.UpsertNominativeUserRequest, JSON: jsonData}

// 		envolveData, err := json.Marshal(e)
// 		if err != nil {
// 			logger.Log.Error("Failed to do json marshalling", zap.Error(err))
// 		}

// 		_, err = s.queue.PushJob(ctx, job.Job{
// 			Type:   sql.NullString{String: "aw"},
// 			Status: job.JobStatusPENDING,
// 			Data:   envolveData,
// 			PPID:   ppId,
// 		}, "aw")
// 		if err != nil {
// 			logger.Log.Error("Failed to push job to the queue", zap.Error(err))
// 		}
// 	}
// 	return &v1.UpserNominativeUserResponse{Status: true}, nil
// }

func (s *ProductServiceServer) ListNominativeUser(ctx context.Context, req *v1.ListNominativeUsersRequest) (*v1.ListNominativeUsersResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	apiresp := v1.ListNominativeUsersResponse{}
	if req.IsProduct {
		listNomiDbReq := db.ListNominativeUsersProductsParams{
			Scope:            []string{req.GetScopes()},
			LkProductName:    !req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
			ProductName:      req.GetSearchParams().GetProductName().GetFilteringkey(),
			IsProductName:    req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
			LkProductVersion: !req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
			ProductVersion:   req.GetSearchParams().GetProductVersion().GetFilteringkey(),
			IsProductVersion: req.GetSearchParams().GetProductVersion().GetFilterType(),
			LkUserName:       !req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			UserName:         req.GetSearchParams().GetUserName().GetFilteringkey(),
			IsUserName:       req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			LkFirstName:      !req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			FirstName:        req.GetSearchParams().GetFirstName().GetFilteringkey(),
			IsFirstName:      req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			LkUserEmail:      !req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			UserEmail:        req.GetSearchParams().GetUserEmail().GetFilteringkey(),
			IsUserEmail:      req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			LkProfile:        !req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			Profile:          req.GetSearchParams().GetProfile().GetFilteringkey(),
			IsProfile:        req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			//LkActivationDate:    !req.GetSearchParams().GetActivationDate().GetFilterType() && req.GetSearchParams().GetActivationDate().GetFilteringkey() != "",
			// ActivationDate:     sql.NullTime{Time: req.GetSearchParams().ActivationDate.AsTime()},
			// IsActivationDate:   req.GetSearchParams().GetActivationDate().IsValid(),
			ProductNameAsc:     strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductNameDesc:    strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductVersionAsc:  strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductVersionDesc: strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserNameAsc:        strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserNameDesc:       strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FirstNameAsc:       strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FirstNameDesc:      strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserEmailAsc:       strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserEmailDesc:      strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProfileAsc:         strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProfileDesc:        strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ActivationDateAsc:  strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ActivationDateDesc: strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductEditorAsc:   strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductEditorDesc:  strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
			LkProductEditor:    !req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
			ProductEditor:      req.GetSearchParams().GetEditor().GetFilteringkey(),
			IsProductEditor:    req.GetSearchParams().GetEditor().GetFilterType() && req.GetSearchParams().GetEditor().GetFilteringkey() != "",
			// API expect pagenum from 1 but the offset in DB starts
			PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
			PageSize: req.GetPageSize(),
		}

		if req.GetSearchParams().GetActivationDate().IsValid() {
			t := req.GetSearchParams().ActivationDate.AsTime()
			year, month, day := t.Date()
			monthInString := strconv.Itoa(int(month))
			if len(monthInString) == 1 {
				monthInString = "0" + monthInString
			}
			dayInString := strconv.Itoa(int(day))
			if len(dayInString) == 1 {
				dayInString = "0" + dayInString
			}
			listNomiDbReq.IsActivationDate = true
			listNomiDbReq.ActivationDate = strconv.Itoa(year) + "-" + monthInString + "-" + dayInString
		}
		dbresp, err := s.ProductRepo.ListNominativeUsersProducts(ctx, listNomiDbReq)
		if err != nil {
			logger.Log.Error("service/v1 - listNominativeUsers - db/ListNominativeUsers", zap.Error(err))
			return nil, status.Error(codes.Unknown, "DBError")
		}
		apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		//apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		if len(dbresp) > 0 {
			apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
		}

		for i := range dbresp {
			apiresp.NominativeUser[i] = &v1.NominativeUser{}
			apiresp.NominativeUser[i].Id = dbresp[i].UserID
			apiresp.NominativeUser[i].FirstName = dbresp[i].FirstName.String
			apiresp.NominativeUser[i].ProductName = dbresp[i].ProductName
			apiresp.NominativeUser[i].ProductVersion = dbresp[i].ProductVersion
			apiresp.NominativeUser[i].Profile = dbresp[i].Profile.String
			apiresp.NominativeUser[i].UserEmail = dbresp[i].UserEmail
			apiresp.NominativeUser[i].UserName = dbresp[i].UserName.String
			apiresp.NominativeUser[i].Editor = dbresp[i].ProductEditor.String
			if dbresp[i].ActivationDate.Valid && dbresp[i].ActivationDate.Time.String() != "0001-01-01 00:00:00 +0000 +0000" {
				apiresp.NominativeUser[i].ActivationDate = timestamppb.New(dbresp[i].ActivationDate.Time)
			} else {
				apiresp.NominativeUser[i].ActivationDate = nil
			}
		}
	} else {
		listNomiDbReq := db.ListNominativeUsersAggregationParams{
			Scope:             []string{req.GetScopes()},
			LkAggregationName: !req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
			AggregationName:   req.GetSearchParams().GetAggregationName().GetFilteringkey(),
			IsAggregationName: req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
			LkUserName:        !req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			UserName:          req.GetSearchParams().GetUserName().GetFilteringkey(),
			IsUserName:        req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			LkFirstName:       !req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			FirstName:         req.GetSearchParams().GetFirstName().GetFilteringkey(),
			IsFirstName:       req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			LkUserEmail:       !req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			UserEmail:         req.GetSearchParams().GetUserEmail().GetFilteringkey(),
			IsUserEmail:       req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			LkProfile:         !req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			Profile:           req.GetSearchParams().GetProfile().GetFilteringkey(),
			IsProfile:         req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			//LkActivationDate:    !req.GetSearchParams().GetActivationDate().GetFilterType() && req.GetSearchParams().GetActivationDate().GetFilteringkey() != "",
			// ActivationDate:      sql.NullTime{Time: req.GetSearchParams().ActivationDate.AsTime()},
			// IsActivationDate:    req.GetSearchParams().GetActivationDate().IsValid(),
			AggregationNameAsc:  strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			AggregationNameDesc: strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserNameAsc:         strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserNameDesc:        strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FirstNameAsc:        strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FirstNameDesc:       strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserEmailAsc:        strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserEmailDesc:       strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProfileAsc:          strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProfileDesc:         strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ActivationDateAsc:   strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ActivationDateDesc:  strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
			// API expect pagenum from 1 but the offset in DB starts
			PageNum:  req.GetPageSize() * (req.GetPageNum() - 1),
			PageSize: req.GetPageSize(),
		}

		if req.GetSearchParams().GetActivationDate().IsValid() {
			t := req.GetSearchParams().ActivationDate.AsTime()
			year, month, day := t.Date()
			monthInString := strconv.Itoa(int(month))
			if len(monthInString) == 1 {
				monthInString = "0" + monthInString
			}
			dayInString := strconv.Itoa(int(day))
			if len(dayInString) == 1 {
				dayInString = "0" + dayInString
			}
			listNomiDbReq.IsActivationDate = true
			listNomiDbReq.ActivationDate = strconv.Itoa(year) + "-" + monthInString + "-" + dayInString
		}
		dbresp, err := s.ProductRepo.ListNominativeUsersAggregation(ctx, listNomiDbReq)
		if err != nil {
			logger.Log.Error("service/v1 - listNominativeUsers - db/ListNominativeUsers", zap.Error(err))
			return nil, status.Error(codes.Unknown, "DBError")
		}
		apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		//apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		if len(dbresp) > 0 {
			apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
		}

		for i := range dbresp {
			apiresp.NominativeUser[i] = &v1.NominativeUser{}
			apiresp.NominativeUser[i].Id = dbresp[i].UserID
			apiresp.NominativeUser[i].ActivationDate = timestamppb.New(dbresp[i].ActivationDate.Time)
			apiresp.NominativeUser[i].AggregationId = dbresp[i].AggregationsID.Int32
			apiresp.NominativeUser[i].AggregationName = dbresp[i].AggregationName.String
			apiresp.NominativeUser[i].FirstName = dbresp[i].FirstName.String
			apiresp.NominativeUser[i].Profile = dbresp[i].Profile.String
			apiresp.NominativeUser[i].UserEmail = dbresp[i].UserEmail
			apiresp.NominativeUser[i].UserName = dbresp[i].UserName.String
			apiresp.NominativeUser[i].Editor = dbresp[i].ProductEditor.String
			if dbresp[i].ActivationDate.Valid && dbresp[i].ActivationDate.Time.String() != "0001-01-01 00:00:00 +0000 +0000" {
				apiresp.NominativeUser[i].ActivationDate = timestamppb.New(dbresp[i].ActivationDate.Time)
			} else {
				apiresp.NominativeUser[i].ActivationDate = nil
			}
		}
	}
	return &apiresp, nil
}

func (s *ProductServiceServer) NominativeUserExport(ctx context.Context, req *v1.NominativeUsersExportRequest) (*v1.ListNominativeUsersExportResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	apiresp := v1.ListNominativeUsersExportResponse{}
	if req.IsProduct {
		listNomiDbReq := db.ExportNominativeUsersProductsParams{
			Scope:            []string{req.GetScopes()},
			LkProductName:    !req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
			ProductName:      req.GetSearchParams().GetProductName().GetFilteringkey(),
			IsProductName:    req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
			LkProductVersion: !req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
			ProductVersion:   req.GetSearchParams().GetProductVersion().GetFilteringkey(),
			IsProductVersion: req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
			LkUserName:       !req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			UserName:         req.GetSearchParams().GetUserName().GetFilteringkey(),
			IsUserName:       req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			LkFirstName:      !req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			FirstName:        req.GetSearchParams().GetFirstName().GetFilteringkey(),
			IsFirstName:      req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			LkUserEmail:      !req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			UserEmail:        req.GetSearchParams().GetUserEmail().GetFilteringkey(),
			IsUserEmail:      req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			LkProfile:        !req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			Profile:          req.GetSearchParams().GetProfile().GetFilteringkey(),
			IsProfile:        req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			//LkActivationDate:    !req.GetSearchParams().GetActivationDate().GetFilterType() && req.GetSearchParams().GetActivationDate().GetFilteringkey() != "",
			//ActivationDate:      sql.NullTime{Time: req.GetSearchParams().ActivationDate.AsTime()},
			//IsActivationDate:    req.GetSearchParams().GetActivationDate().IsValid(),
			ProductNameAsc:     strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductNameDesc:    strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductVersionAsc:  strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductVersionDesc: strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserNameAsc:        strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserNameDesc:       strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FirstNameAsc:       strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FirstNameDesc:      strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserEmailAsc:       strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserEmailDesc:      strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProfileAsc:         strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProfileDesc:        strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ActivationDateAsc:  strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ActivationDateDesc: strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
		}
		if req.GetSearchParams().GetActivationDate().IsValid() {
			t := req.GetSearchParams().ActivationDate.AsTime()
			year, month, day := t.Date()
			monthInString := strconv.Itoa(int(month))
			if len(monthInString) == 1 {
				monthInString = "0" + monthInString
			}
			dayInString := strconv.Itoa(int(day))
			if len(dayInString) == 1 {
				dayInString = "0" + dayInString
			}
			listNomiDbReq.IsActivationDate = true
			listNomiDbReq.ActivationDate = strconv.Itoa(year) + "-" + monthInString + "-" + dayInString
		}
		dbresp, err := s.ProductRepo.ExportNominativeUsersProducts(ctx, listNomiDbReq)
		if err != nil {
			logger.Log.Error("service/v1 - listNominativeUsers - db/ListNominativeUsers", zap.Error(err))
			return nil, status.Error(codes.Unknown, "DBError")
		}
		apiresp.NominativeUser = make([]*v1.NominativeUserExport, len(dbresp))
		//apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		if len(dbresp) > 0 {
			apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
		}

		for i := range dbresp {
			apiresp.NominativeUser[i] = &v1.NominativeUserExport{}
			apiresp.NominativeUser[i].FirstName = dbresp[i].FirstName.String
			apiresp.NominativeUser[i].ProductName = dbresp[i].ProductName
			apiresp.NominativeUser[i].ProductVersion = dbresp[i].ProductVersion
			apiresp.NominativeUser[i].Profile = dbresp[i].Profile.String
			apiresp.NominativeUser[i].UserEmail = dbresp[i].UserEmail
			apiresp.NominativeUser[i].UserName = dbresp[i].UserName.String
			apiresp.NominativeUser[i].Editor = dbresp[i].ProductEditor.String
			if dbresp[i].ActivationDate.Valid && dbresp[i].ActivationDate.Time.String() != "0001-01-01 00:00:00 +0000 +0000" {
				apiresp.NominativeUser[i].ActivationDate = dbresp[i].ActivationDate.Time.Format("2006-01-02")
			}
		}
	} else {
		listNomiDbReq := db.ExportNominativeUsersAggregationParams{
			Scope:             []string{req.GetScopes()},
			LkAggregationName: !req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
			AggregationName:   req.GetSearchParams().GetAggregationName().GetFilteringkey(),
			IsAggregationName: req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
			LkUserName:        !req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			UserName:          req.GetSearchParams().GetUserName().GetFilteringkey(),
			IsUserName:        req.GetSearchParams().GetUserName().GetFilterType() && req.GetSearchParams().GetUserName().GetFilteringkey() != "",
			LkFirstName:       !req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			FirstName:         req.GetSearchParams().GetFirstName().GetFilteringkey(),
			IsFirstName:       req.GetSearchParams().GetFirstName().GetFilterType() && req.GetSearchParams().GetFirstName().GetFilteringkey() != "",
			LkUserEmail:       !req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			UserEmail:         req.GetSearchParams().GetUserEmail().GetFilteringkey(),
			IsUserEmail:       req.GetSearchParams().GetUserEmail().GetFilterType() && req.GetSearchParams().GetUserEmail().GetFilteringkey() != "",
			LkProfile:         !req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			Profile:           req.GetSearchParams().GetProfile().GetFilteringkey(),
			IsProfile:         req.GetSearchParams().GetProfile().GetFilterType() && req.GetSearchParams().GetProfile().GetFilteringkey() != "",
			//LkActivationDate:    !req.GetSearchParams().GetActivationDate().GetFilterType() && req.GetSearchParams().GetActivationDate().GetFilteringkey() != "",
			//ActivationDate:      sql.NullTime{Time: req.GetSearchParams().ActivationDate.AsTime()},
			//IsActivationDate:    req.GetSearchParams().GetActivationDate().IsValid(),
			AggregationNameAsc:  strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			AggregationNameDesc: strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserNameAsc:         strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserNameDesc:        strings.Contains(req.GetSortBy(), "user_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FirstNameAsc:        strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FirstNameDesc:       strings.Contains(req.GetSortBy(), "first_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UserEmailAsc:        strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UserEmailDesc:       strings.Contains(req.GetSortBy(), "user_email") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProfileAsc:          strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProfileDesc:         strings.Contains(req.GetSortBy(), "profile") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ActivationDateAsc:   strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ActivationDateDesc:  strings.Contains(req.GetSortBy(), "activation_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
		}

		if req.GetSearchParams().GetActivationDate().IsValid() {
			t := req.GetSearchParams().ActivationDate.AsTime()
			year, month, day := t.Date()
			monthInString := strconv.Itoa(int(month))
			if len(monthInString) == 1 {
				monthInString = "0" + monthInString
			}
			dayInString := strconv.Itoa(int(day))
			if len(dayInString) == 1 {
				dayInString = "0" + dayInString
			}
			listNomiDbReq.IsActivationDate = true
			listNomiDbReq.ActivationDate = strconv.Itoa(year) + "-" + monthInString + "-" + dayInString
		}
		dbresp, err := s.ProductRepo.ExportNominativeUsersAggregation(ctx, listNomiDbReq)
		if err != nil {
			logger.Log.Error("service/v1 - listNominativeUsers - db/ListNominativeUsers", zap.Error(err))
			return nil, status.Error(codes.Unknown, "DBError")
		}
		apiresp.NominativeUser = make([]*v1.NominativeUserExport, len(dbresp))
		//apiresp.NominativeUser = make([]*v1.NominativeUser, len(dbresp))
		if len(dbresp) > 0 {
			apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
		}

		for i := range dbresp {
			apiresp.NominativeUser[i] = &v1.NominativeUserExport{}
			// apiresp.NominativeUser[i].ActivationDate = timestamppb.New(dbresp[i].ActivationDate.Time)
			apiresp.NominativeUser[i].AggregationId = dbresp[i].AggregationsID.Int32
			apiresp.NominativeUser[i].AggregationName = dbresp[i].AggregationName.String
			apiresp.NominativeUser[i].FirstName = dbresp[i].FirstName.String
			apiresp.NominativeUser[i].Profile = dbresp[i].Profile.String
			apiresp.NominativeUser[i].UserEmail = dbresp[i].UserEmail
			apiresp.NominativeUser[i].UserName = dbresp[i].UserName.String
			apiresp.NominativeUser[i].Editor = dbresp[i].ProductEditor.String
			if dbresp[i].ActivationDate.Valid && dbresp[i].ActivationDate.Time.String() != "0001-01-01 00:00:00 +0000 +0000" {
				apiresp.NominativeUser[i].ActivationDate = dbresp[i].ActivationDate.Time.Format("2006-01-02")
			}
		}
	}
	return &apiresp, nil
}

// DeleteNominativeUsers will be delete a record from storage
func (s *ProductServiceServer) DeleteNominativeUsers(ctx context.Context, req *v1.DeleteNominativeUserRequest) (*v1.DeleteNominativeUserResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	pNomUser, err := s.ProductRepo.GetNominativeUserByID(ctx, db.GetNominativeUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
	if err != nil {
		logger.Log.Error("failed to delete product nominative user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	err = s.ProductRepo.DeleteNominativeUserByID(ctx, db.DeleteNominativeUserByIDParams{
		Scope: req.GetScope(), ID: req.GetId(),
	})

	if err != nil {
		logger.Log.Error("failed to delete product nominative user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	deleteNominativeReqDgraph := DeleteNominativeUserRequest(pNomUser)
	deleteNominativeReqDgraph.Scope = req.GetScope()
	// For Worker Queue
	jsonData, err := json.Marshal(deleteNominativeReqDgraph)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.DeleteNominativeUserRequest, JSON: jsonData}

	envolveData, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}

	_, err = s.queue.PushJob(ctx, job.Job{
		Type:   sql.NullString{String: "aw"},
		Status: job.JobStatusPENDING,
		Data:   envolveData,
	}, "aw")
	if err != nil {
		logger.Log.Error("Failed to push job to the queue", zap.Error(err))
	}
	if pNomUser.AggregationsID.Int32 == 0 {
		productDeleted := s.DeleteSaasProductUsers(ctx, pNomUser.Swidtag.String, pNomUser.Scope)
		logger.Log.Sugar().Debugw("Product is deleted ",
			"productDeleted", productDeleted,
			"swidtag", pNomUser.Swidtag.String,
			"scope", pNomUser.Scope,
		)
	}

	return &v1.DeleteNominativeUserResponse{Success: true}, nil
}

// DeleteNominativeUserRequest will map db.NominativeUser to dgworker.UpserNominativeUserRequest
func DeleteNominativeUserRequest(dbNomUser db.NominativeUser) (resp dgworker.UpserNominativeUserRequest) {
	resp.AggregationId = dbNomUser.AggregationsID.Int32
	resp.SwidTag = dbNomUser.Swidtag.String
	var userDetails dgworker.NominativeUserDetails
	userDetails.Email = dbNomUser.UserEmail
	userDetails.Profile = dbNomUser.Profile.String
	resp.UserDetails = append(resp.UserDetails, &userDetails)
	return
}

// GetConcurrentUsersHistroy will get all concurrent users data by month or day from storage
func (s *ProductServiceServer) GetConcurrentUsersHistroy(ctx context.Context, req *v1.GetConcurrentUsersHistroyRequest) (*v1.GetConcurrentUsersHistroyResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	if req.GetAggID() == 0 && req.GetSwidtag() == "" {
		return nil, status.Error(codes.Internal, "Swidtag or aggregation cannot be blank")
	}

	if !req.GetStartDate().IsValid() || !req.GetEndDate().IsValid() {
		return nil, status.Error(codes.Internal, "please provide valid start and end date")
	}

	startDate := req.GetStartDate().AsTime()
	endDate := req.GetEndDate().AsTime()
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 00, 00, 00, 000, time.Local)
	endDate = time.Date(endDate.Year(), endDate.Month(), 1, 00, 00, 00, 000, time.Local)
	//daysDifferance := endDate.Sub(startDate).Hours() / 24
	var response = &v1.GetConcurrentUsersHistroyResponse{}
	//	if daysDifferance > 60 {
	concurrentUsersbyMonth, err := s.ProductRepo.GetConcurrentUsersByMonth(ctx, db.GetConcurrentUsersByMonthParams{
		Scope:               req.GetScope(),
		IsPurchaseStartDate: req.GetStartDate().IsValid(),
		StartDate:           startDate,
		IsPurchaseEndDate:   req.GetEndDate().IsValid(),
		EndDate:             endDate,
		IsSwidtag:           req.GetSwidtag() != "",
		Swidtag:             sql.NullString{String: req.Swidtag, Valid: true},
		IsAggregationID:     req.GetAggID() > 0,
		AggregationID:       sql.NullInt32{Int32: req.GetAggID(), Valid: true},
	})
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
		return nil, status.Error(codes.Internal, "DBError")
	}
	response.ConcurrentUsersByMonths = make([]*v1.ConcurrentUsersByMonth, len(concurrentUsersbyMonth))

	for i := range concurrentUsersbyMonth {
		response.ConcurrentUsersByMonths[i] = &v1.ConcurrentUsersByMonth{}
		response.ConcurrentUsersByMonths[i].PurchaseMonth = concurrentUsersbyMonth[i].Purchasemonthyear.(string)
		response.ConcurrentUsersByMonths[i].ConcurrentUsers = int32(concurrentUsersbyMonth[i].Totalconusers)
	}
	//	} else {
	// 	concurrentUsersbyDay, err := s.ProductRepo.GetConcurrentUsersByDay(ctx, db.GetConcurrentUsersByDayParams{
	// 		Scope:               req.GetScope(),
	// 		IsPurchaseStartDate: req.GetStartDate().IsValid(),
	// 		StartDate:           startDate,
	// 		IsPurchaseEndDate:   req.GetEndDate().IsValid(),
	// 		EndDate:             endDate,
	// 		IsSwidtag:           req.GetSwidtag() != "",
	// 		Swidtag:             sql.NullString{String: req.Swidtag, Valid: true},
	// 		IsAggregationID:     req.GetAggID() > 0,
	// 		AggregationID:       sql.NullInt32{Int32: req.GetAggID(), Valid: true},
	// 	})
	// 	if err != nil {
	// 		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	// 		return nil, status.Error(codes.Internal, "DBError")
	// 	}
	// 	response.ConcurrentUsersByDays = make([]*v1.ConcurrentUsersByDay, len(concurrentUsersbyDay))

	// 	for i := range concurrentUsersbyDay {
	// 		response.ConcurrentUsersByDays[i] = &v1.ConcurrentUsersByDay{}
	// 		response.ConcurrentUsersByDays[i].PurchaseDate = timestamppb.New(concurrentUsersbyDay[i].PurchaseDate)
	// 		response.ConcurrentUsersByDays[i].CouncurrentUsers = int32(concurrentUsersbyDay[i].Totalconusers)
	// 	}
	// }

	return response, nil
}

func (s *ProductServiceServer) ConcurrentUserExport(ctx context.Context, req *v1.ListConcurrentUsersExportRequest) (*v1.ListConcurrentUsersResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScopes()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	listCouncurrentReq := db.ExportConcurrentUsersParams{
		Scope:               []string{req.GetScopes()},
		IsAggregations:      req.IsAggregation,
		LkProductName:       !req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		ProductName:         req.GetSearchParams().GetProductName().GetFilteringkey(),
		IsProductName:       req.GetSearchParams().GetProductName().GetFilterType() && req.GetSearchParams().GetProductName().GetFilteringkey() != "",
		LkAggregationName:   !req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		AggregationName:     req.GetSearchParams().GetAggregationName().GetFilteringkey(),
		IsAggregationName:   req.GetSearchParams().GetAggregationName().GetFilterType() && req.GetSearchParams().GetAggregationName().GetFilteringkey() != "",
		LkProductVersion:    !req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
		ProductVersion:      req.GetSearchParams().GetProductVersion().GetFilteringkey(),
		IsProductVersion:    req.GetSearchParams().GetProductVersion().GetFilterType() && req.GetSearchParams().GetProductVersion().GetFilteringkey() != "",
		LkTeam:              !req.GetSearchParams().GetTeam().GetFilterType() && req.GetSearchParams().GetTeam().GetFilteringkey() != "",
		Team:                req.GetSearchParams().GetTeam().GetFilteringkey(),
		IsTeam:              req.GetSearchParams().GetTeam().GetFilterType() && req.GetSearchParams().GetTeam().GetFilteringkey() != "",
		LkProfileUser:       !req.GetSearchParams().GetProfileUser().GetFilterType() && req.GetSearchParams().GetProfileUser().GetFilteringkey() != "",
		ProfileUser:         req.GetSearchParams().GetProfileUser().GetFilteringkey(),
		IsProfileUser:       req.GetSearchParams().GetProfileUser().GetFilterType() && req.GetSearchParams().GetProfileUser().GetFilteringkey() != "",
		LkNumberOfUsers:     !req.GetSearchParams().GetNumberOfUsers().GetFilterType() && req.GetSearchParams().GetNumberOfUsers().GetFilteringkey() != "",
		NumberOfUsers:       req.GetSearchParams().GetNumberOfUsers().GetFilteringkey(),
		IsNumberOfUsers:     req.GetSearchParams().GetNumberOfUsers().GetFilterType() && req.GetSearchParams().GetNumberOfUsers().GetFilteringkey() != "",
		ProductNameAsc:      strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductNameDesc:     strings.Contains(req.GetSortBy(), "product_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		AggregationNameAsc:  strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "asc"),
		AggregationNameDesc: strings.Contains(req.GetSortBy(), "aggregation_name") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProductVersionAsc:   strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProductVersionDesc:  strings.Contains(req.GetSortBy(), "product_version") && strings.Contains(req.GetSortOrder().String(), "desc"),
		ProfileUserAsc:      strings.Contains(req.GetSortBy(), "profile_user") && strings.Contains(req.GetSortOrder().String(), "asc"),
		ProfileUserDesc:     strings.Contains(req.GetSortBy(), "profile_user") && strings.Contains(req.GetSortOrder().String(), "desc"),
		TeamAsc:             strings.Contains(req.GetSortBy(), "team") && strings.Contains(req.GetSortOrder().String(), "asc"),
		TeamDesc:            strings.Contains(req.GetSortBy(), "team") && strings.Contains(req.GetSortOrder().String(), "desc"),
		NumberOfUsersAsc:    strings.Contains(req.GetSortBy(), "number_of_users") && strings.Contains(req.GetSortOrder().String(), "asc"),
		NumberOfUsersDesc:   strings.Contains(req.GetSortBy(), "number_of_users") && strings.Contains(req.GetSortOrder().String(), "desc"),
		PurchaseDateAsc:     strings.Contains(req.GetSortBy(), "purchase_date") && strings.Contains(req.GetSortOrder().String(), "asc"),
		PurchaseDateDesc:    strings.Contains(req.GetSortBy(), "purchase_date") && strings.Contains(req.GetSortOrder().String(), "desc"),
		// API expect pagenum from 1 but the offset in DB starts
	}

	if req.GetSearchParams().GetPurchaseDate().IsValid() {
		listCouncurrentReq.PurchaseDate = req.GetSearchParams().PurchaseDate.AsTime()
	}
	dbresp, err := s.ProductRepo.ExportConcurrentUsers(ctx, listCouncurrentReq)
	if err != nil {
		logger.Log.Error("service/v1 - ConcurrentUserExport - db/ListConcurrentUsers", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	apiresp := v1.ListConcurrentUsersResponse{}
	apiresp.ConcurrentUser = make([]*v1.ConcurrentUser, len(dbresp))
	if len(dbresp) > 0 {
		apiresp.TotalRecords = int32(dbresp[0].Totalrecords)
	}

	for i := range dbresp {
		apiresp.ConcurrentUser[i] = &v1.ConcurrentUser{}
		apiresp.ConcurrentUser[i].IsAggregation = dbresp[i].IsAggregations.Bool
		apiresp.ConcurrentUser[i].AggregationId = dbresp[i].AggregationID.Int32
		apiresp.ConcurrentUser[i].AggregationName = dbresp[i].AggregationName.String
		apiresp.ConcurrentUser[i].ProductName = dbresp[i].ProductName.String
		apiresp.ConcurrentUser[i].ProductVersion = dbresp[i].ProductVersion.String
		apiresp.ConcurrentUser[i].Team = dbresp[i].Team.String
		apiresp.ConcurrentUser[i].ProfileUser = dbresp[i].ProfileUser.String
		apiresp.ConcurrentUser[i].NumberOfUsers = dbresp[i].NumberOfUsers.Int32
		apiresp.ConcurrentUser[i].PurchaseDate = timestamppb.New(dbresp[i].PurchaseDate)
		apiresp.ConcurrentUser[i].Id = dbresp[i].ID

	}
	return &apiresp, nil
}

func (s *ProductServiceServer) ListNominativeUserFileUpload(ctx context.Context, req *v1.ListNominativeUsersFileUploadRequest) (*v1.ListNominativeUsersFileUploadResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}
	apiresp := v1.ListNominativeUsersFileUploadResponse{}
	fDetails := []*v1.ListNominativeUsersFileUpload{}
	var err error
	var fileDetails []db.ListNominativeUsersUploadedFilesRow
	if req.GetId() > 0 {
		fileDetails, err = s.ProductRepo.ListNominativeUsersUploadedFiles(ctx, db.ListNominativeUsersUploadedFilesParams{
			Scope:              []string{req.GetScope()},
			FileUploadID:       true,
			ID:                 req.GetId(),
			PageNum:            req.GetPageSize() * (req.GetPageNum() - 1),
			PageSize:           req.GetPageSize(),
			FileNameAsc:        strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FileNameDesc:       strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FileStatusAsc:      strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FileStatusDesc:     strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductEditorAsc:   strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductEditorDesc:  strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
			NameAsc:            strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			NameDesc:           strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductVersionAsc:  strings.Contains(req.GetSortBy(), "productVersion") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductVersionDesc: strings.Contains(req.GetSortBy(), "productVersion") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UploadedByAsc:      strings.Contains(req.GetSortBy(), "uploadedBy") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UploadedByDesc:     strings.Contains(req.GetSortBy(), "uploadedBy") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UploadedOnAsc:      strings.Contains(req.GetSortBy(), "UploadedOn") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UploadedOnDesc:     strings.Contains(req.GetSortBy(), "UploadedOn") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProducttypeAsc:     strings.Contains(req.GetSortBy(), "productType") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProducttypeDesc:    strings.Contains(req.GetSortBy(), "productType") && strings.Contains(req.GetSortOrder().String(), "desc"),
		})
	} else {
		fileDetails, err = s.ProductRepo.ListNominativeUsersUploadedFiles(ctx, db.ListNominativeUsersUploadedFilesParams{
			Scope:              []string{req.GetScope()},
			FileUploadID:       false,
			PageNum:            req.GetPageSize() * (req.GetPageNum() - 1),
			PageSize:           req.GetPageSize(),
			FileNameAsc:        strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FileNameDesc:       strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "desc"),
			FileStatusAsc:      strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "asc"),
			FileStatusDesc:     strings.Contains(req.GetSortBy(), "fileName") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductEditorAsc:   strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductEditorDesc:  strings.Contains(req.GetSortBy(), "editor") && strings.Contains(req.GetSortOrder().String(), "desc"),
			NameAsc:            strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "asc"),
			NameDesc:           strings.Contains(req.GetSortBy(), "name") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProductVersionAsc:  strings.Contains(req.GetSortBy(), "productVersion") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProductVersionDesc: strings.Contains(req.GetSortBy(), "productVersion") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UploadedByAsc:      strings.Contains(req.GetSortBy(), "uploadedBy") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UploadedByDesc:     strings.Contains(req.GetSortBy(), "uploadedBy") && strings.Contains(req.GetSortOrder().String(), "desc"),
			UploadedOnAsc:      strings.Contains(req.GetSortBy(), "UploadedOn") && strings.Contains(req.GetSortOrder().String(), "asc"),
			UploadedOnDesc:     strings.Contains(req.GetSortBy(), "UploadedOn") && strings.Contains(req.GetSortOrder().String(), "desc"),
			ProducttypeAsc:     strings.Contains(req.GetSortBy(), "productType") && strings.Contains(req.GetSortOrder().String(), "asc"),
			ProducttypeDesc:    strings.Contains(req.GetSortBy(), "productType") && strings.Contains(req.GetSortOrder().String(), "desc"),
		})
	}
	if err != nil && err != sql.ErrNoRows {
		logger.Log.Error("service/v1 - ListNominativeUserFileUpload - db/ListNominativeUsersUploadedFiles", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}
	for _, fD := range fileDetails {
		usrs := []*v1.NominativeUser{}
		u := []*v1.NominativeUserDetails{}
		if len(fD.NominativeUsersDetails.RawMessage) > 0 {
			err := json.Unmarshal(fD.NominativeUsersDetails.RawMessage, &usrs)
			if err != nil {
				logger.Log.Error("service/v1 - ListNominativeUserFileUpload", zap.Error(err))
				return nil, status.Error(codes.Unknown, "error Unmarshal")
			}
			for _, v := range usrs {
				u = append(u, &v1.NominativeUserDetails{
					UserName:       v.GetUserName(),
					FirstName:      v.GetFirstName(),
					Email:          v.GetUserEmail(),
					Profile:        v.GetProfile(),
					ActivationDate: v.GetActivationDateString(),
					Comments:       v.GetComment(),
				})
			}
		}
		fDetail := &v1.ListNominativeUsersFileUpload{
			Id: fD.ID,
			//	Scope:                  fD.Scope,
			Swidtag:                fD.Swidtag.String,
			AggregationsId:         fD.AggregationsID.Int32,
			ProductEditor:          fD.ProductEditor.String,
			UploadedBy:             fD.UploadedBy,
			NominativeUsersDetails: u,
			RecordSucceed:          fD.RecordSucceed.Int32,
			RecordFailed:           fD.RecordFailed.Int32,
			FileName:               fD.FileName.String,
			SheetName:              fD.SheetName.String,
			//FileStatus:             string(fD.FileStatus),
			UploadedAt:      timestamppb.New(fD.UploadedAt),
			UploadId:        fD.UploadID,
			ProductName:     fD.ProductName.String,
			ProductVersion:  fD.ProductVersion.String,
			AggregationName: fD.AggregationName.String,
			Type:            fD.Nametype.(string),
			Name:            fD.Pname.(string),
		}
		if fD.Jobnotcompleted > 0 {
			fDetail.FileStatus = "PENDING"
		} else if fD.Jobnotcompleted == 0 {
			fDetail.FileStatus = "COMPLETED"
		} else {
			fDetail.FileStatus = string(fD.FileStatus)
		}
		fDetails = append(fDetails, fDetail)
	}
	apiresp.FileDetails = fDetails
	if len(fileDetails) > 0 {
		apiresp.Total = int32(fileDetails[0].Totalrecords)
	}
	return &apiresp, nil
}

// DeleteSaasProductUsers will check & delete SAAS product when concurrent & nominative users last user deleted
func (s *ProductServiceServer) DeleteSaasProductUsers(ctx context.Context, switag string, scope string) bool {

	// Check if product have concurrent & nominative users or not
	productUsers, err := s.ProductRepo.GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
		Swidtag: []string{switag},
		Scope:   []string{scope},
	})

	if err != nil {
		logger.Log.Sugar().Errorw("Service/DeleteSaasProductUsers - Error while getting nominative & concurrent users of product ",
			"scope", scope,
			"swidtag", switag,
			"error", err.Error(),
		)
		return false
	}

	if len(productUsers) == 0 {
		err = s.ProductRepo.DeleteProductsBySwidTagScope(ctx, db.DeleteProductsBySwidTagScopeParams{
			Scope:   scope,
			Swidtag: switag,
		})
		if err != nil {
			logger.Log.Sugar().Errorw("Service/DeleteSaasProductUsers Repo/DeleteProductsBySwidTagScope - Error while deleting product from DB ",
				"scope", scope,
				"swidtag", switag,
				"error", err.Error(),
			)
			return false
		}

		deleteProductReqDgraph := dgraph.DeleteProductRequest{
			Scope:   scope,
			SwidTag: switag,
		}

		// For Worker Queue
		jsonData, err := json.Marshal(deleteProductReqDgraph)
		if err != nil {
			logger.Log.Sugar().Errorw("Service/DeleteSaasProductUsers RepodGraph/DeleteProductsBySwidTagScope - Failed to do json marshalling ",
				"scope", scope,
				"swidtag", switag,
				"error", err.Error(),
				"deleteProductReqDgraph", deleteProductReqDgraph,
			)
			return false
		}
		e := dgworker.Envelope{Type: dgworker.DeletSaaSProductRequest, JSON: jsonData}

		envolveData, err := json.Marshal(e)
		if err != nil {
			logger.Log.Sugar().Errorw("Service/DeleteSaasProductUsers RepodGraph/DeleteProductsBySwidTagScope - Failed to do json marshalling ",
				"scope", scope,
				"swidtag", switag,
				"error", err.Error(),
				"envolveData", e,
			)
			return false
		}
		_, err = s.queue.PushJob(ctx, job.Job{
			Type:   sql.NullString{String: "aw"},
			Status: job.JobStatusPENDING,
			Data:   envolveData,
		}, "aw")
		if err != nil {
			logger.Log.Sugar().Errorw("Service/DeleteSaasProductUsers RepodGraph/DeleteProductsBySwidTagScope - Failed to push job to the queue ",
				"scope", scope,
				"swidtag", switag,
				"error", err.Error(),
			)
			return false
		}
		return true
	}

	return true
}
