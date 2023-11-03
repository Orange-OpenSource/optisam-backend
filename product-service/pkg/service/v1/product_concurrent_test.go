package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/tabbed/pqtype"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const NOCLAIMFOUND = "cannot find claims in context"
const FAILEDJSONUNMARSHAL = "Failed to do json marshalling"

func TestUpsertProductConcurrentUser(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "s1"},
	})
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductConcurrentUserRequest
		output *v1.ProductConcurrentUserResponse
		mock   func(*v1.ProductConcurrentUserRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "UpsertProductConcurrentUserData--Success",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: false,
				AggregationId:  0,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "s1",
				Id:             0,
			},
			output: &v1.ProductConcurrentUserResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				//dbObj.EXPECT().ListEditors(ctx, []string{input.Scope}).Return([]string{"abc", "abc2"}, nil)
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.Swidtag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{
					Swidtag:           "ABC_abc_v1",
					ProductName:       "ABC",
					ProductEditor:     "abc",
					ProductVersion:    "v1",
					NumOfApplications: 0,
					NumOfEquipments:   0,
					Metrics:           []string{"m1", "m2", "m3"},
				}, nil)

				fcall := dbObj.EXPECT().UpsertConcurrentUserTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				currentDateTime := time.Now()
				theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), 00, 00, 00, 000, time.Local)
				resp := UpsertConcurrentUserDgraphRequest(input, userClaims.UserID)
				resp.PurchaseDate = theDate.String()
				jsonData, err := json.Marshal(resp)
				if err != nil {
					t.Errorf(FAILEDJSONUNMARSHAL)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertConcurrentUserRequest, JSON: jsonData}

				_, err = json.Marshal(e)
				if err != nil {
					t.Error(FAILEDJSONUNMARSHAL)
				}
				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "UpsertProductConcurrentUserData--Success with aggregation",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: true,
				AggregationId:  408,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "s1",
				Id:             0,
			},
			output: &v1.ProductConcurrentUserResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				//dbObj.EXPECT().ListEditors(ctx, []string{input.Scope}).Return([]string{"abc", "abc2"}, nil)
				dbObj.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    input.AggregationId,
					Scope: input.Scope}).Return(db.Aggregation{
					ID:              input.AggregationId,
					AggregationName: "Test",
					Scope:           "s1",
				}, nil)

				fcall := dbObj.EXPECT().UpsertConcurrentUserTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				currentDateTime := time.Now()
				theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), 00, 00, 00, 000, time.Local)
				resp := UpsertConcurrentUserDgraphRequest(input, userClaims.UserID)
				resp.PurchaseDate = theDate.String()
				jsonData, err := json.Marshal(resp)
				if err != nil {
					t.Errorf(FAILEDJSONUNMARSHAL)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertConcurrentUserRequest, JSON: jsonData}

				_, err = json.Marshal(e)
				if err != nil {
					t.Error(FAILEDJSONUNMARSHAL)
				}
				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "UpsertProductConcurrentUserData--Success while updating data.",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: false,
				AggregationId:  0,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "s1",
				Id:             10,
			},
			output: &v1.ProductConcurrentUserResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				//dbObj.EXPECT().ListEditors(ctx, []string{input.Scope}).Return([]string{"abc", "abc2"}, nil)
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.Swidtag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{
					Swidtag:           "ABC_abc_v1",
					ProductName:       "ABC",
					ProductEditor:     "abc",
					ProductVersion:    "v1",
					NumOfApplications: 0,
					NumOfEquipments:   0,
					Metrics:           []string{"m1", "m2", "m3"},
				}, nil)

				fcall := dbObj.EXPECT().UpsertConcurrentUserTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				currentDateTime := time.Now()
				theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), 00, 00, 00, 000, time.Local)
				resp := UpsertConcurrentUserDgraphRequest(input, userClaims.UserID)
				if input.Id > 0 {
					dbObj.EXPECT().GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(db.ProductConcurrentUser{
						ID:           input.Id,
						PurchaseDate: theDate,
					}, nil)

				}
				resp.PurchaseDate = theDate.String()
				jsonData, err := json.Marshal(resp)
				if err != nil {
					t.Errorf(FAILEDJSONUNMARSHAL)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertConcurrentUserRequest, JSON: jsonData}

				_, err = json.Marshal(e)
				if err != nil {
					t.Error(FAILEDJSONUNMARSHAL)
				}
				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "UpsertProductConcurrentUserData--Success with aggregation while updating data",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: true,
				AggregationId:  408,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "s1",
				Id:             10,
			},
			output: &v1.ProductConcurrentUserResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				//dbObj.EXPECT().ListEditors(ctx, []string{input.Scope}).Return([]string{"abc", "abc2"}, nil)
				dbObj.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    input.AggregationId,
					Scope: input.Scope}).Return(db.Aggregation{
					ID:              input.AggregationId,
					AggregationName: "Test",
					Scope:           "s1",
				}, nil)

				fcall := dbObj.EXPECT().UpsertConcurrentUserTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				currentDateTime := time.Now()
				theDate := time.Date(currentDateTime.Year(), currentDateTime.Month(), currentDateTime.Day(), 00, 00, 00, 000, time.Local)
				resp := UpsertConcurrentUserDgraphRequest(input, userClaims.UserID)
				if input.Id > 0 {
					dbObj.EXPECT().GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(db.ProductConcurrentUser{
						ID:           input.Id,
						PurchaseDate: theDate,
					}, nil)

				}
				resp.PurchaseDate = theDate.String()
				jsonData, err := json.Marshal(resp)
				if err != nil {
					t.Errorf(FAILEDJSONUNMARSHAL)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertConcurrentUserRequest, JSON: jsonData}

				_, err = json.Marshal(e)
				if err != nil {
					t.Error(FAILEDJSONUNMARSHAL)
				}
				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name:   "UpsertProductConcurrentUserWithoutContext",
			input:  &v1.ProductConcurrentUserRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ProductConcurrentUserRequest) {
				// For adding coverage when no context is passed
			},
		},
		{
			name: "UpsertProductConcurrentUser FAILURE - No access to scopes",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: false,
				AggregationId:  0,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "s1",
				Id:             0,
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "UpsertProductConcurrentUser FAILURE - ScopeValidationError",
			input: &v1.ProductConcurrentUserRequest{
				IsAggregations: false,
				AggregationId:  0,
				Swidtag:        "ABC_abc_v1",
				ProductName:    "ABC",
				ProductEditor:  "abc",
				ProductVersion: "v1",
				NumberOfUsers:  100,
				ProfileUser:    "Sr",
				Team:           "IT",
				Scope:          "S33",
				Id:             0,
			},
			ctx:    ctx,
			outErr: true,
			mock: func(input *v1.ProductConcurrentUserRequest) {
				// // When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.UpsertProductConcurrentUser(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListConcurrentUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.ListConcurrentUsersRequest
		output *v1.ListConcurrentUsersResponse
		mock   func(*v1.ListConcurrentUsersRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListConcurrentUsersSuccess",
			input: &v1.ListConcurrentUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			output: &v1.ListConcurrentUsersResponse{
				TotalRecords: 1,
				ConcurrentUser: []*v1.ConcurrentUser{
					{
						ProductName:     "p1",
						ProductVersion:  "14.1.1",
						AggregationName: "a1",
						AggregationId:   0,
						NumberOfUsers:   32,
						Team:            "f1",
						ProfileUser:     "p1",
						Id:              32,
						PurchaseDate:    timestamppb.New(timeNow),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListConcurrentUsersRequest) {
				_, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				dbObj.EXPECT().ListConcurrentUsers(ctx, db.ListConcurrentUsersParams{Scope: []string{"s1"},
					PurchaseDateAsc: true, PageNum: 0, PageSize: 20}).Return([]db.ListConcurrentUsersRow{
					{
						Totalrecords:    1,
						ProductName:     sql.NullString{String: "p1", Valid: true},
						ProductVersion:  sql.NullString{String: "14.1.1", Valid: true},
						AggregationName: sql.NullString{String: "a1", Valid: true},
						AggregationID:   sql.NullInt32{Int32: 0, Valid: true},
						NumberOfUsers:   sql.NullInt32{Int32: 32, Valid: true},
						Team:            sql.NullString{String: "f1", Valid: true},
						ProfileUser:     sql.NullString{String: "p1", Valid: true},
						ID:              32,
						PurchaseDate:    timeNow,
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListConcurrentUserWithOutContext",
			input: &v1.ListConcurrentUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.ListConcurrentUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "SSS!",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name:   "ListConcurrentUserWithoutContext",
			input:  &v1.ListConcurrentUsersRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.ListConcurrentUsers(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestDeleteConcurrentUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.DeleteConcurrentUsersRequest
		output *v1.DeleteConcurrentUsersResponse
		mock   func(*v1.DeleteConcurrentUsersRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "DeleteConcurrentUsersSuccess",
			input: &v1.DeleteConcurrentUsersRequest{
				Id:    1,
				Scope: "s1",
			},
			output: &v1.DeleteConcurrentUsersResponse{
				Success: true,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.DeleteConcurrentUsersRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				if !helper.Contains(userClaims.Socpes, input.Scope) {
					t.Errorf("ScopeValidationError")
				}
				dbConnUser := db.ProductConcurrentUser{
					ID:             32,
					AggregationID:  sql.NullInt32{Int32: 20, Valid: true},
					IsAggregations: sql.NullBool{Bool: true, Valid: true},
					Swidtag:        sql.NullString{String: "ABC_abc_14", Valid: true},
					NumberOfUsers:  sql.NullInt32{Int32: 23, Valid: true},
					ProfileUser:    sql.NullString{String: "p1", Valid: true},
					Team:           sql.NullString{String: "t1", Valid: true},
					Scope:          input.Scope,
					PurchaseDate:   timeNow,
				}
				dbObj.EXPECT().GetConcurrentUserByID(ctx, db.GetConcurrentUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(dbConnUser, nil).Times(1)
				fcall := dbObj.EXPECT().DeletConcurrentUserByID(ctx, db.DeletConcurrentUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(nil).Times(1)
				deleteConcurrentReqDgraph := DeleteConcurrentUserRequest(dbConnUser)
				deleteConcurrentReqDgraph.Scope = input.Scope

				jsonData, err := json.Marshal(deleteConcurrentReqDgraph)
				if err != nil {
					t.Errorf(FAILEDJSONUNMARSHAL)
				}
				e := dgworker.Envelope{Type: dgworker.DeleteConcurrentUserRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Error(FAILEDJSONUNMARSHAL)
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData}, "aw").Return(int32(1), nil).After(fcall)

			},
		},
		{
			name: "DeleteConcurrentUserWithOutContext",
			input: &v1.DeleteConcurrentUsersRequest{
				Id:    1,
				Scope: "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.DeleteConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.DeleteConcurrentUsersRequest{
				Id:    1,
				Scope: "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.DeleteConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name:   "DeleteConcurrentUserWithoutContext",
			input:  &v1.DeleteConcurrentUsersRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.DeleteConcurrentUsersRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.DeleteConcurrentUsers(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestGetConcurrentUsersHistroy(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	timeStartDate := time.Date(timeNow.Year(), timeNow.Month(), 01, 00, 00, 00, 000, time.Local)
	timeParamStartDate := time.Date(timeNow.Year(), timeNow.Month(), 02, 00, 00, 00, 000, time.Local)
	timeEndDate := time.Date(timeNow.Year(), timeNow.Month(), 30, 00, 00, 00, 000, time.Local)
	//timeParamEndDate := time.Date(timeNow.Year(), timeNow.Month(), 31, 00, 00, 00, 000, time.Local)
	secondEndDate := time.Date(timeNow.Year(), timeNow.Month()+3, 01, 00, 00, 00, 000, time.Local)
	testSet := []struct {
		name   string
		input  *v1.GetConcurrentUsersHistroyRequest
		output *v1.GetConcurrentUsersHistroyResponse
		mock   func(*v1.GetConcurrentUsersHistroyRequest)
		ctx    context.Context
		outErr bool
	}{
		// {
		// 	name: "GetConcurrentUsersHistroyByDaySuccess",
		// 	input: &v1.GetConcurrentUsersHistroyRequest{
		// 		Swidtag:   "ABC_abc_2",
		// 		Scope:     "s1",
		// 		AggID:     0,
		// 		StartDate: timestamppb.New(timeParamStartDate),
		// 		EndDate:   timestamppb.New(timeParamEndDate),
		// 	},
		// 	output: &v1.GetConcurrentUsersHistroyResponse{
		// 		ConcurrentUsersByDays: []*v1.ConcurrentUsersByDay{
		// 			{
		// 				PurchaseDate:    timestamppb.New(timeParamStartDate),
		// 				ConcurrentUsers: 32,
		// 			},
		// 			{
		// 				PurchaseDate:    timestamppb.New(timeParamEndDate),
		// 				ConcurrentUsers: 322,
		// 			},
		// 		},
		// 	},
		// 	outErr: false,
		// 	ctx:    ctx,
		// 	mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
		// 		userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
		// 		if !ok {
		// 			t.Errorf(NOCLAIMFOUND)
		// 		}
		// 		if !helper.Contains(userClaims.Socpes, input.Scope) {
		// 			t.Errorf("ScopeValidationError")
		// 		}
		// 		startDate := input.GetStartDate().AsTime()
		// 		endDate := input.GetEndDate().AsTime()
		// 		startDate = time.Date(startDate.Year(), startDate.Month(), 01, 00, 00, 00, 000, time.Local)
		// 		endDate = time.Date(endDate.Year(), endDate.Month(), 30, 00, 00, 00, 000, time.Local)
		// 		daysDifferance := endDate.Sub(startDate).Hours() / 24
		// 		if daysDifferance > 60 {
		// 		} else {
		// 			dbConnUser := []db.GetConcurrentUsersByDayRow{
		// 				{
		// 					PurchaseDate:  input.StartDate.AsTime(),
		// 					Totalconusers: 32,
		// 				},
		// 				{
		// 					PurchaseDate:  input.EndDate.AsTime(),
		// 					Totalconusers: 322,
		// 				},
		// 			}
		// 			dbObj.EXPECT().GetConcurrentUsersByDay(ctx, db.GetConcurrentUsersByDayParams{
		// 				Scope:               input.Scope,
		// 				IsPurchaseStartDate: input.StartDate.IsValid(),
		// 				StartDate:           timeStartDate,
		// 				IsPurchaseEndDate:   input.EndDate.IsValid(),
		// 				EndDate:             timeEndDate,
		// 				IsSwidtag:           input.GetSwidtag() != "",
		// 				Swidtag:             sql.NullString{String: input.Swidtag, Valid: true},
		// 				IsAggregationID:     input.GetAggID() > 0,
		// 				AggregationID:       sql.NullInt32{Int32: input.AggID, Valid: true},
		// 			}).Return(dbConnUser, nil).Times(1)
		// 			dbObj.EXPECT().GetConcurrentUsersByMonth(ctx, gomock.Any()).Return([]db.GetConcurrentUsersByMonthRow{{Purchasemonthyear: "string"}}, nil).Times(1)
		// 		}

		// 	},
		// },
		{
			name: "GetConcurrentUsersHistroyByMonthSuccess",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "ABC_abc_2",
				Scope:     "s1",
				AggID:     0,
				StartDate: timestamppb.New(timeParamStartDate),
				EndDate:   timestamppb.New(secondEndDate),
			},
			output: &v1.GetConcurrentUsersHistroyResponse{
				ConcurrentUsersByMonths: []*v1.ConcurrentUsersByMonth{
					{
						PurchaseMonth:   "November 2022",
						ConcurrentUsers: 32,
					},
					{
						PurchaseMonth:   "December 2022",
						ConcurrentUsers: 322,
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				if !helper.Contains(userClaims.Socpes, input.Scope) {
					t.Errorf("ScopeValidationError")
				}
				startDate := input.GetStartDate().AsTime()
				endDate := input.GetEndDate().AsTime()
				startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 00, 00, 00, 000, time.Local)
				endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 00, 00, 00, 000, time.Local)
				daysDifferance := endDate.Sub(startDate).Hours() / 24
				if daysDifferance > 60 {
					var tmp interface{} = "November 2022"
					var tmp2 interface{} = "December 2022"
					dbConnUser := []db.GetConcurrentUsersByMonthRow{
						{
							Purchasemonthyear: tmp,
							Totalconusers:     32,
						},
						{
							Purchasemonthyear: tmp2,
							Totalconusers:     322,
						},
					}
					dbObj.EXPECT().GetConcurrentUsersByMonth(ctx, gomock.Any()).Return(dbConnUser, nil).Times(1)
				}

			},
		},
		{
			name: "Failer - GetConcurrentUsersHistroyByMonth",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "ABC_abc_2",
				Scope:     "s1",
				AggID:     0,
				StartDate: timestamppb.New(timeParamStartDate),
				EndDate:   timestamppb.New(secondEndDate),
			},
			output: &v1.GetConcurrentUsersHistroyResponse{
				ConcurrentUsersByMonths: []*v1.ConcurrentUsersByMonth{
					{
						PurchaseMonth:   "November 2022",
						ConcurrentUsers: 32,
					},
					{
						PurchaseMonth:   "December 2022",
						ConcurrentUsers: 322,
					},
				},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				if !helper.Contains(userClaims.Socpes, input.Scope) {
					t.Errorf("ScopeValidationError")
				}
				startDate := input.GetStartDate().AsTime()
				endDate := input.GetEndDate().AsTime()
				startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 00, 00, 00, 000, time.Local)
				endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 00, 00, 00, 000, time.Local)
				dbObj.EXPECT().GetConcurrentUsersByMonth(ctx, gomock.Any()).Return(nil, errors.New("Test")).Times(1)

			},
		},
		{
			name: "GetConcurrentUsersHistroyWithOutContext",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "ABC_abc_2",
				Scope:     "s1",
				AggID:     30,
				StartDate: timestamppb.New(timeStartDate),
				EndDate:   timestamppb.New(timeEndDate),
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "GetConcurrentUsersHistroyWithOutSwidtagORAggID",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "",
				Scope:     "s1",
				AggID:     0,
				StartDate: timestamppb.New(timeStartDate),
				EndDate:   timestamppb.New(timeEndDate),
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "ABC_abc_2",
				Scope:     "no access",
				AggID:     30,
				StartDate: timestamppb.New(timeStartDate),
				EndDate:   timestamppb.New(timeEndDate),
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name:   "GetConcurrentUsersHistroyerWithoutContext",
			input:  &v1.GetConcurrentUsersHistroyRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: agg id and swidtag empty",
			input: &v1.GetConcurrentUsersHistroyRequest{
				Swidtag:   "",
				Scope:     "s1",
				AggID:     0,
				StartDate: timestamppb.New(timeStartDate),
				EndDate:   timestamppb.New(timeEndDate),
			},
			ctx:    ctx,
			outErr: true,
			mock: func(input *v1.GetConcurrentUsersHistroyRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.GetConcurrentUsersHistroy(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestConcurrentUserExport(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.ListConcurrentUsersExportRequest
		output *v1.ListConcurrentUsersResponse
		mock   func(*v1.ListConcurrentUsersExportRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListConcurrentUsersExportSuccess",
			input: &v1.ListConcurrentUsersExportRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			output: &v1.ListConcurrentUsersResponse{
				TotalRecords: 1,
				ConcurrentUser: []*v1.ConcurrentUser{
					{
						ProductName:     "p1",
						ProductVersion:  "14.1.1",
						AggregationName: "a1",
						AggregationId:   0,
						NumberOfUsers:   32,
						Team:            "f1",
						ProfileUser:     "p1",
						Id:              32,
						PurchaseDate:    timestamppb.New(timeNow),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListConcurrentUsersExportRequest) {
				_, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				dbObj.EXPECT().ExportConcurrentUsers(ctx, db.ExportConcurrentUsersParams{Scope: []string{"s1"},
					PurchaseDateAsc: true}).Return([]db.ExportConcurrentUsersRow{
					{
						Totalrecords:    1,
						ProductName:     sql.NullString{String: "p1", Valid: true},
						ProductVersion:  sql.NullString{String: "14.1.1", Valid: true},
						AggregationName: sql.NullString{String: "a1", Valid: true},
						AggregationID:   sql.NullInt32{Int32: 0, Valid: true},
						NumberOfUsers:   sql.NullInt32{Int32: 32, Valid: true},
						Team:            sql.NullString{String: "f1", Valid: true},
						ProfileUser:     sql.NullString{String: "p1", Valid: true},
						ID:              32,
						PurchaseDate:    timeNow,
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListConcurrentUserExportWithOutContext",
			input: &v1.ListConcurrentUsersExportRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListConcurrentUsersExportRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.ListConcurrentUsersExportRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "SSS!",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListConcurrentUsersExportRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name:   "ListConcurrentUserExportWithoutContext",
			input:  &v1.ListConcurrentUsersExportRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListConcurrentUsersExportRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.ConcurrentUserExport(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListNominativeUserFileUpload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	_ = time.Now()
	testSet := []struct {
		name   string
		input  *v1.ListNominativeUsersFileUploadRequest
		output *v1.ListNominativeUsersFileUploadResponse
		mock   func(*v1.ListNominativeUsersFileUploadRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListConcurrentUsersExportSuccess",
			input: &v1.ListNominativeUsersFileUploadRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scope:     "s1",
			},
			output: &v1.ListNominativeUsersFileUploadResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListNominativeUsersFileUploadRequest) {
				_, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				dbObj.EXPECT().ListNominativeUsersUploadedFiles(ctx, gomock.Any()).Return(
					[]db.ListNominativeUsersUploadedFilesRow{
						{
							NominativeUsersDetails: pqtype.NullRawMessage{
								RawMessage: json.RawMessage(`[{"user": "test"}]`),
							},
							Pname:    "",
							Nametype: "",
						},
					},
					nil,
				).Times(1)

			},
		},
		{
			name: "ListConcurrentUsersExportSuccess1",
			input: &v1.ListNominativeUsersFileUploadRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scope:     "s1",
				Id:        int32(10),
			},
			output: &v1.ListNominativeUsersFileUploadResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListNominativeUsersFileUploadRequest) {
				_, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf(NOCLAIMFOUND)
				}
				dbObj.EXPECT().ListNominativeUsersUploadedFiles(ctx, gomock.Any()).Return(
					[]db.ListNominativeUsersUploadedFilesRow{
						{
							NominativeUsersDetails: pqtype.NullRawMessage{
								RawMessage: json.RawMessage(`[{"user": "test"}]`),
							},
							Pname:    "",
							Nametype: "",
						},
					},
					nil,
				).Times(1)

			},
		},
		{
			name: "ListListNominativeUserFileUploadWithOutContext",
			input: &v1.ListNominativeUsersFileUploadRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scope:     "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListNominativeUsersFileUploadRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.ListNominativeUsersFileUploadRequest{
				SortBy:    "purchase_date",
				SortOrder: v1.SortOrder_asc,
				Scope:     "SSS!",
			},
			ctx:    context.Background(),
			outErr: true,
			mock: func(input *v1.ListNominativeUsersFileUploadRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
		{
			name:   "ListListNominativeUserFileUploadWithoutContext",
			input:  &v1.ListNominativeUsersFileUploadRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListNominativeUsersFileUploadRequest) {
				// When user doesn't have access to scope error condition coverage
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.ListNominativeUserFileUpload(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestDeleteSaasProductUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := dbmock.NewMockProduct(ctrl)
	mockQueue := queuemock.NewMockWorkerqueue(ctrl)

	server := &ProductServiceServer{
		ProductRepo: mockProductRepo,
		queue:       mockQueue,
	}

	ctx := context.Background()

	t.Run("Delete product with no users", func(t *testing.T) {
		swidTag := "exampleSwidTag"
		scope := "exampleScope"

		mockProductRepo.EXPECT().GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
			Swidtag: []string{swidTag},
			Scope:   []string{scope},
		}).Return(nil, nil)

		mockProductRepo.EXPECT().DeleteProductsBySwidTagScope(ctx, db.DeleteProductsBySwidTagScopeParams{
			Swidtag: swidTag,
			Scope:   scope,
		}).Return(nil).AnyTimes()

		mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(32), nil).AnyTimes()

		// Make the API call
		result := server.DeleteSaasProductUsers(ctx, swidTag, scope)

		// Assertions
		assert.True(t, result)
	})

	t.Run("Delete product with users", func(t *testing.T) {
		swidTag := "exampleSwidTag"
		scope := "exampleScope"

		mockProductRepo.EXPECT().GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
			Swidtag: []string{swidTag},
			Scope:   []string{scope},
		}).Return([]db.GetConcurrentNominativeUsersBySwidTagRow{
			// Mock users here
		}, nil)

		// Make the API call
		result := server.DeleteSaasProductUsers(ctx, swidTag, scope)

		// Assertions
		assert.True(t, result)
	})

	t.Run("Error retrieving product users", func(t *testing.T) {
		swidTag := "exampleSwidTag"
		scope := "exampleScope"

		mockProductRepo.EXPECT().GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
			Swidtag: []string{swidTag},
			Scope:   []string{scope},
		}).Return(nil, errors.New("DBError"))

		// Make the API call
		result := server.DeleteSaasProductUsers(ctx, swidTag, scope)

		// Assertions
		assert.False(t, result)
	})

	// t.Run("Error deleting product from DB", func(t *testing.T) {
	// 	swidTag := "exampleSwidTag"
	// 	scope := "exampleScope"

	// 	mockProductRepo.EXPECT().GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
	// 		Swidtag: []string{swidTag},
	// 		Scope:   []string{scope},
	// 	}).Return(nil, nil)

	// 	mockProductRepo.EXPECT().DeleteProductsBySwidTagScope(ctx, db.DeleteProductsBySwidTagScopeParams{
	// 		Swidtag: swidTag,
	// 		Scope:   scope,
	// 	}).Return(errors.New("DBError")).AnyTimes()

	// 	// Make the API call
	// 	result := server.DeleteSaasProductUsers(ctx, swidTag, scope)

	// 	// Assertions
	// 	assert.False(t, result)
	// })

	// t.Run("Error pushing job to the queue", func(t *testing.T) {
	// 	swidTag := "exampleSwidTag"
	// 	scope := "exampleScope"

	// 	mockProductRepo.EXPECT().GetConcurrentNominativeUsersBySwidTag(ctx, db.GetConcurrentNominativeUsersBySwidTagParams{
	// 		Swidtag: []string{swidTag},
	// 		Scope:   []string{scope},
	// 	}).Return(nil, nil)

	// 	mockProductRepo.EXPECT().DeleteProductsBySwidTagScope(ctx, db.DeleteProductsBySwidTagScopeParams{
	// 		Swidtag: swidTag,
	// 		Scope:   scope,
	// 	}).Return(nil).AnyTimes()

	// 	mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), errors.New("QueueError")).AnyTimes()

	// 	// Make the API call
	// 	result := server.DeleteSaasProductUsers(ctx, swidTag, scope)

	// 	// Assertions
	// 	assert.False(t, result)
	// })
}
