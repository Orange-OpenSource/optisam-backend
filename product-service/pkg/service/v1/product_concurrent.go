package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/mail"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	"optisam-backend/product-service/pkg/worker/dgraph"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UpsertProductConcurrentUser will add or update product concurrent users
func (s *productServiceServer) UpsertProductConcurrentUser(ctx context.Context, req *v1.ProductConcurrentUserRequest) (*v1.ProductConcurrentUserResponse, error) {

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

		// listEditors, err := s.productRepo.ListEditors(ctx, []string{req.GetScope()})
		// if err != nil {
		// 	logger.Log.Error("service/v1 - ListEditors - ListEditors", zap.Error(err))
		// 	return nil, status.Error(codes.Internal, "DBError")
		// }
		// if !helper.Contains(listEditors, req.GetProductEditor()) {
		// 	return nil, status.Error(codes.PermissionDenied, "editor doesn't exists")
		// }

		if req.GetProductVersion() != "" {
			swittag = strings.ReplaceAll(strings.Join([]string{req.GetProductName(), req.GetProductEditor(), req.GetProductVersion()}, "_"), " ", "_")
		} else {
			swittag = strings.ReplaceAll(strings.Join([]string{req.GetProductName(), req.GetProductEditor()}, "_"), " ", "_")
		}
		req.Swidtag = swittag
		_, errProduct := s.productRepo.GetProductInformation(ctx, db.GetProductInformationParams{Scope: req.GetScope(), Swidtag: swittag})
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
		_, err := s.productRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
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

	err := s.productRepo.UpsertConcurrentUserTx(ctx, req, userClaims.UserID)
	if err != nil {
		logger.Log.Error("can ", zap.Error(err))
		return nil, status.Error(codes.Internal, "can not add product concurrent users")
	}

	upsertConcurrentReqDgraph := UpsertConcurrentUserDgraphRequest(req, userClaims.UserID)
	currentDateTime := time.Now()
	theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), 1, 00, 00, 00, 000, time.Local)

	if req.GetId() > 0 {
		pConUser, err := s.productRepo.GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
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
func (s *productServiceServer) ListConcurrentUsers(ctx context.Context, req *v1.ListConcurrentUsersRequest) (*v1.ListConcurrentUsersResponse, error) {
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
	dbresp, err := s.productRepo.ListConcurrentUsers(ctx, listCouncurrentReq)
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
func (s *productServiceServer) DeleteConcurrentUsers(ctx context.Context, req *v1.DeleteConcurrentUsersRequest) (*v1.DeleteConcurrentUsersResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	pConUser, err := s.productRepo.GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
	if err != nil {
		logger.Log.Error("failed to delete product concurrent user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	err = s.productRepo.DeletConcurrentUserByID(ctx, db.DeletConcurrentUserByIDParams{
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

func (s *productServiceServer) UpsertNominativeUser(ctx context.Context, req *v1.UpserNominativeUserRequest) (resp *v1.UpserNominativeUserResponse, err error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if req.GetScope() == "" {
		return nil, status.Error(codes.Internal, "cannot find scope")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if !(req.AggregationId > 0) && (req.Editor == "" || req.ProductName == "") {
		return nil, status.Error(codes.InvalidArgument, "Either aggrigation or product details are required")
	}
	var swid = ""
	aggr := db.Aggregation{}
	if req.GetAggregationId() > 0 {
		aggr, err = s.productRepo.GetAggregationByID(ctx, db.GetAggregationByIDParams{
			ID:    req.GetAggregationId(),
			Scope: req.GetScope(),
		})
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.InvalidArgument, "Aggregation doesn't exists")
		}
		if err != nil {
			logger.Log.Error("service/v1 - UpsertNominativeUser - GetAggregationByID", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
	} else {
		// listEditors, err := s.productRepo.ListEditors(ctx, []string{req.GetScope()})
		// if err != nil {
		// 	logger.Log.Error("service/v1 - UpsertNominativeUser - ListEditors", zap.Error(err))
		// 	return nil, status.Error(codes.Internal, "DBError")
		// }
		// if !helper.Contains(listEditors, req.GetEditor()) {
		// 	return nil, status.Error(codes.InvalidArgument, "editor doesn't exists")
		// }
		if req.GetProductVersion() != "" {
			swid = strings.ReplaceAll(strings.Join([]string{req.ProductName, req.GetEditor(), req.GetProductVersion()}, "_"), " ", "_")
		} else {
			swid = strings.ReplaceAll(strings.Join([]string{req.ProductName, req.GetEditor()}, "_"), " ", "_")
		}
		_, err = s.productRepo.GetProductInformation(ctx, db.GetProductInformationParams{
			Scope:   req.GetScope(),
			Swidtag: swid,
		})
		if err != nil && err != sql.ErrNoRows {
			logger.Log.Error("service/v1 - UpsertNominativeUser - GetProductInformation", zap.Error(err))
			return nil, status.Error(codes.Internal, "DBError")
		}
		if err == sql.ErrNoRows {
			_, err := s.UpsertProduct(ctx, &v1.UpsertProductRequest{
				SwidTag:     swid,
				Name:        req.GetProductName(),
				Editor:      req.GetEditor(),
				Scope:       req.GetScope(),
				Version:     req.GetProductVersion(),
				ProductType: v1.Producttype_saas,
			})
			if err != nil {
				logger.Log.Error("service/v1 - UpsertNominativeUser - UpsertProductRequest :can not add product", zap.Error(err))
				return nil, status.Error(codes.Internal, "can not add product")
			}
		}
	}

	err = s.productRepo.UpsertNominativeUsersTx(ctx, req, userClaims.UserID, userClaims.UserID, swid)
	if err != nil {
		logger.Log.Error("service/v1 - UpsertNominativeUser - UpsertNominativeUsersTx : can not upsert users", zap.Error(err))
		return nil, status.Error(codes.Internal, "can not upsert users")
	}
	upsertNominativeReqDgraph := PrepairUpsertNominativeUserDgraphRequest(req, swid, userClaims.UserID, aggr.AggregationName)

	// For Worker Queue
	jsonData, err := json.Marshal(upsertNominativeReqDgraph)
	if err != nil {
		logger.Log.Error("Failed to do json marshalling", zap.Error(err))
	}
	e := dgworker.Envelope{Type: dgworker.UpsertNominativeUserRequest, JSON: jsonData}

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
	return &v1.UpserNominativeUserResponse{Status: true}, nil
}

func PrepairUpsertNominativeUserDgraphRequest(req *v1.UpserNominativeUserRequest, swidTag, createdBy string, aggrName string) (resp dgworker.UpserNominativeUserRequest) {
	resp.AggregationId = req.GetAggregationId()
	resp.Editor = req.GetEditor()
	resp.ProductName = req.GetProductVersion()
	resp.ProductVersion = req.GetProductVersion()
	resp.Scope = req.GetScope()
	resp.SwidTag = swidTag
	resp.CreatedBy = createdBy
	respUsers := []*dgworker.NominativeUserDetails{}
	users := make(map[string]bool)
	for _, v := range req.UserDetails {
		var userDetails dgworker.NominativeUserDetails
		var startTime time.Time
		var err error
		if v.ActivationDate != "" {
			if len(v.ActivationDate) <= 10 {
				if strings.Contains(v.ActivationDate, "/") && len(v.ActivationDate) <= 8 {
					startTime, err = time.Parse("06/2/1", v.ActivationDate)
				} else if strings.Contains(v.ActivationDate, "/") {
					startTime, err = time.Parse("2006/01/02", v.ActivationDate)
				} else {
					startTime, err = time.Parse("2006-01-02", v.ActivationDate)
				}
				if err == nil {
					userDetails.ActivationDate = startTime
				}
			} else if len(v.ActivationDate) > 10 && len(v.ActivationDate) <= 24 {
				if strings.Contains(v.ActivationDate, "/") && len(v.ActivationDate) <= 8 {
					startTime, err = time.Parse("06/2/1T15:04:05.000Z", v.ActivationDate)
				} else if strings.Contains(v.ActivationDate, "/") {
					startTime, err = time.Parse("2006/01/02T15:04:05.000Z", v.ActivationDate)
				} else {
					startTime, err = time.Parse("2006-01-02T15:04:05.000Z", v.ActivationDate)
				}
				if err == nil {
					userDetails.ActivationDate = startTime
				}
			}
			// } else {
			// 	startTime, err = time.Parse(time.RFC3339Nano, v.ActivationDate)
			// 	if err != nil {
			// 		logger.Log.Error("service/v1 - UpsertAcqRights - unable to parse start time", zap.String("reason", err.Error()))
			// 	}
			// }
		}
		err = nil
		userDetails.Email = v.GetEmail()
		userDetails.FirstName = v.GetFirstName()
		userDetails.Profile = v.GetProfile()
		userDetails.UserName = v.GetUserName()
		_, err = mail.ParseAddress(v.GetEmail())
		if err != nil {
			err = errors.New("Invalid email format")
		}
		if _, ok := users[v.GetEmail()+v.GetProfile()]; ok {
			err = errors.New("duplicate entry")
		} else {
			users[v.Email+v.Profile] = true
		}
		if err == nil {
			respUsers = append(respUsers, &userDetails)
		}
	}
	resp.UserDetails = respUsers
	return
}

func (s *productServiceServer) ListNominativeUser(ctx context.Context, req *v1.ListNominativeUsersRequest) (*v1.ListNominativeUsersResponse, error) {
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
		dbresp, err := s.productRepo.ListNominativeUsersProducts(ctx, listNomiDbReq)
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
		dbresp, err := s.productRepo.ListNominativeUsersAggregation(ctx, listNomiDbReq)
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

func (s *productServiceServer) NominativeUserExport(ctx context.Context, req *v1.NominativeUsersExportRequest) (*v1.ListNominativeUsersExportResponse, error) {
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
		dbresp, err := s.productRepo.ExportNominativeUsersProducts(ctx, listNomiDbReq)
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
		dbresp, err := s.productRepo.ExportNominativeUsersAggregation(ctx, listNomiDbReq)
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
func (s *productServiceServer) DeleteNominativeUsers(ctx context.Context, req *v1.DeleteNominativeUserRequest) (*v1.DeleteNominativeUserResponse, error) {

	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Unknown, "ClaimsNotFoundError")
	}
	if !helper.Contains(userClaims.Socpes, req.Scope) {
		return nil, status.Error(codes.PermissionDenied, "ScopeValidationError")
	}

	pNomUser, err := s.productRepo.GetNominativeUserByID(ctx, db.GetNominativeUserByIDParams{Scope: req.GetScope(), ID: req.GetId()})
	if err != nil {
		logger.Log.Error("failed to delete product nominative user, unable to get data", zap.Error(err))
		return nil, status.Error(codes.Unknown, "DBError")
	}

	err = s.productRepo.DeleteNominativeUserByID(ctx, db.DeleteNominativeUserByIDParams{
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
func (s *productServiceServer) GetConcurrentUsersHistroy(ctx context.Context, req *v1.GetConcurrentUsersHistroyRequest) (*v1.GetConcurrentUsersHistroyResponse, error) {
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
	concurrentUsersbyMonth, err := s.productRepo.GetConcurrentUsersByMonth(ctx, db.GetConcurrentUsersByMonthParams{
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
		response.ConcurrentUsersByMonths[i].CouncurrentUsers = int32(concurrentUsersbyMonth[i].Totalconusers)
	}
	//	} else {
	// 	concurrentUsersbyDay, err := s.productRepo.GetConcurrentUsersByDay(ctx, db.GetConcurrentUsersByDayParams{
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

func (s *productServiceServer) ConcurrentUserExport(ctx context.Context, req *v1.ListConcurrentUsersExportRequest) (*v1.ListConcurrentUsersResponse, error) {
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
	dbresp, err := s.productRepo.ExportConcurrentUsers(ctx, listCouncurrentReq)
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

func (s *productServiceServer) ListNominativeUserFileUpload(ctx context.Context, req *v1.ListNominativeUsersFileUploadRequest) (*v1.ListNominativeUsersFileUploadResponse, error) {
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
		fileDetails, err = s.productRepo.ListNominativeUsersUploadedFiles(ctx, db.ListNominativeUsersUploadedFilesParams{
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
		fileDetails, err = s.productRepo.ListNominativeUsersUploadedFiles(ctx, db.ListNominativeUsersUploadedFilesParams{
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
		fDetails = append(fDetails, &v1.ListNominativeUsersFileUpload{
			Id:                     fD.ID,
			Scope:                  fD.Scope,
			Swidtag:                fD.Swidtag.String,
			AggregationsId:         fD.AggregationsID.Int32,
			ProductEditor:          fD.ProductEditor.String,
			UploadedBy:             fD.UploadedBy,
			NominativeUsersDetails: u,
			RecordSucceed:          fD.RecordSucceed.Int32,
			RecordFailed:           fD.RecordFailed.Int32,
			FileName:               fD.FileName.String,
			SheetName:              fD.SheetName.String,
			FileStatus:             string(fD.FileStatus),
			UploadedAt:             timestamppb.New(fD.UploadedAt),
			UploadId:               fD.UploadID,
			ProductName:            fD.ProductName.String,
			ProductVersion:         fD.ProductVersion.String,
			AggregationName:        fD.AggregationName.String,
			Type:                   fD.Nametype.(string),
			Name:                   fD.Pname.(string),
		})
	}
	apiresp.FileDetails = fDetails
	if len(fileDetails) > 0 {
		apiresp.Total = int32(fileDetails[0].Totalrecords)
	}
	return &apiresp, nil
}

// DeleteSaasProductUsers will check & delete SAAS product when concurrent & nominative users last user deleted
func (s *productServiceServer) DeleteSaasProductUsers(ctx context.Context, switag string, scope string) bool {

	// Check if product have concurrent & nominative users or not
	productUsers, err := s.productRepo.GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
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
		err = s.productRepo.DeleteProductsBySwidTagScope(ctx, db.DeleteProductsBySwidTagScopeParams{
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
