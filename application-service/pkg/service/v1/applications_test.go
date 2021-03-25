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
	"errors"
	"log"
	v1 "optisam-backend/application-service/pkg/api/v1"
	repo "optisam-backend/application-service/pkg/repository/v1"
	dbmock "optisam-backend/application-service/pkg/repository/v1/dbmock"
	"optisam-backend/application-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/application-service/pkg/repository/v1/queuemock"
	dgWorker "optisam-backend/application-service/pkg/worker/dgraph"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	ctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2"},
	})
)

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	os.Exit(m.Run())
}

func TestUpsertApplication(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.UpsertApplicationRequest
		output *v1.UpsertApplicationResponse
		mock   func(*v1.UpsertApplicationRequest)
		outErr bool
	}{
		{
			name: "UpsertApplicationWithCorrectData",
			input: &v1.UpsertApplicationRequest{
				ApplicationId: "a1",
				Name:          "a1name",
				Version:       "a1version",
				Owner:         "a1owner",
				Scope:         "Scope1",
				Domain:        "Payments",
			},
			output: &v1.UpsertApplicationResponse{Success: true},
			mock: func(input *v1.UpsertApplicationRequest) {
				firstCall := dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{
					ApplicationID:      "a1",
					ApplicationName:    "a1name",
					ApplicationOwner:   "a1owner",
					ApplicationVersion: "a1version",
					Scope:              "Scope1",
					ApplicationDomain:  "Payments",
				}).Return(nil).Times(1)
				jsonData, err := json.Marshal(input)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgWorker.Envelope{Type: dgWorker.UpsertApplicationRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				job := job.Job{
					Type:   sql.NullString{String: "lw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}
				qObj.EXPECT().PushJob(ctx, job, "lw").Return(int32(1000), nil).After(firstCall)
			},
			outErr: false,
		},
		{
			name:   "UpsertApplicationWithMissingData",
			input:  &v1.UpsertApplicationRequest{Scope: "Scope1"},
			output: &v1.UpsertApplicationResponse{Success: false},
			mock: func(input *v1.UpsertApplicationRequest) {
				dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{Scope: "Scope1", ApplicationDomain: ""}).Return(errors.New("rpc error: code = Internal desc = DBError")).Times(1)
			},
			outErr: true,
		},
		{
			name: "UpsertApplicationWithMissingapplicationID",
			input: &v1.UpsertApplicationRequest{
				Name:    "a1name",
				Owner:   "a1owner",
				Version: "a1version",
				Scope:   "Scope1",
			},
			output: &v1.UpsertApplicationResponse{Success: false},
			mock: func(input *v1.UpsertApplicationRequest) {
				dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{
					ApplicationName:    "a1name",
					ApplicationVersion: "a1version",
					ApplicationOwner:   "a1owner",
					ApplicationDomain:  "",
					Scope:              "Scope1",
				}).Return(errors.New("rpc error: code = Internal desc = DBError")).Times(1)
			},
			outErr: true,
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewApplicationServiceServer(dbObj, qObj)
			got, err := s.UpsertApplication(ctx, test.input)
			log.Println(" log to be removed RESP[", got, "][", err, "]")
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if got.Success != test.output.Success {
				t.Errorf("Failed case [%s] because  exepcted output [%v] is not same as actual output [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListApplications(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)

	testSet := []struct {
		name   string
		input  *v1.ListApplicationsRequest
		output *v1.ListApplicationsResponse
		mock   func(*v1.ListApplicationsRequest)
		isErr  bool
		errVal string
		ctx    context.Context
	}{
		{
			name: "ListApplicationWithCorrectData",
			input: &v1.ListApplicationsRequest{
				PageNum:  int32(1),
				PageSize: int32(2),
				Scopes:   []string{"Scope1"},
			},
			output: &v1.ListApplicationsResponse{
				TotalRecords: 2,
				Applications: []*v1.Application{
					&v1.Application{
						ApplicationId:    "a1",
						Name:             "a1name",
						Owner:            "a1owner",
						NumOfInstances:   int32(5),
						NumOfProducts:    int32(5),
						NumOfEquipments:  int32(5),
						Domain:           "Payments",
						ObsolescenceRisk: "Risk1",
					},
					&v1.Application{
						ApplicationId:   "a2",
						Name:            "a2name",
						Owner:           "a2owner",
						NumOfInstances:  int32(3),
						NumOfProducts:   int32(3),
						NumOfEquipments: int32(3),
					},
				},
			},
			mock: func(input *v1.ListApplicationsRequest) {
				dbObj.EXPECT().GetApplicationsView(ctx, db.GetApplicationsViewParams{
					Scope:              []string{"Scope1"},
					ApplicationNameAsc: true,
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize}).Return([]db.GetApplicationsViewRow{
					{
						Totalrecords:      int64(2),
						ApplicationID:     "a1",
						ApplicationName:   "a1name",
						ApplicationOwner:  "a1owner",
						NumOfInstances:    int32(5),
						NumOfProducts:     int32(5),
						NumOfEquipments:   int32(5),
						ApplicationDomain: "Payments",
						ObsolescenceRisk:  sql.NullString{String: "Risk1", Valid: true},
					},
					{
						Totalrecords:     int64(2),
						ApplicationID:    "a2",
						ApplicationName:  "a2name",
						ApplicationOwner: "a2owner",
						NumOfInstances:   int32(3),
						NumOfProducts:    int32(3),
						NumOfEquipments:  int32(3),
					}}, nil).Times(1)
			},
			isErr: false,
			ctx:   ctx,
		},
		{
			name:  "ListApplicationWithClaimNotfound",
			input: &v1.ListApplicationsRequest{},
			mock:  func(input *v1.ListApplicationsRequest) {},
			isErr: true,
			ctx:   context.Background(),
		},
		{
			name: "ListApplicationWithNoRecords",
			input: &v1.ListApplicationsRequest{
				PageNum:  int32(1),
				PageSize: int32(2),
				Scopes:   []string{"Scope1"},
			},
			mock: func(input *v1.ListApplicationsRequest) {
				dbObj.EXPECT().GetApplicationsView(ctx, db.GetApplicationsViewParams{
					Scope:              []string{"Scope1"},
					ApplicationNameAsc: true,
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize}).Return([]db.GetApplicationsViewRow{}, nil).Times(1)
			},
			output: &v1.ListApplicationsResponse{
				TotalRecords: 0,
				Applications: []*v1.Application{},
			},
			isErr: false,
			ctx:   ctx,
		},
		{
			name: "ListApplicationWithInvalidArguments",
			input: &v1.ListApplicationsRequest{
				PageNum:  int32(-1),
				PageSize: int32(-1),
				Scopes:   []string{"Scope1"},
			},
			mock: func(input *v1.ListApplicationsRequest) {
				dbObj.EXPECT().GetApplicationsView(ctx, db.GetApplicationsViewParams{
					Scope:              []string{"Scope1"},
					ApplicationNameAsc: true,
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize,
				}).Return(nil, errors.New("rpc error: code = Unknown desc = DBError")).Times(1)
			},
			isErr: true,
			ctx:   ctx,
		},
	}

	for _, test := range testSet {
		t.Run(test.name, func(t *testing.T) {
			test.mock(test.input)
			s := NewApplicationServiceServer(dbObj, qObj)
			got, err := s.ListApplications(test.ctx, test.input)
			log.Println(" log to be removed RESP[", got, "][", err, "]")
			if (err != nil) != test.isErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)

	testSet := []struct {
		name   string
		input  *v1.ListInstancesRequest
		output *v1.ListInstancesResponse
		mock   func(*v1.ListInstancesRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "ListInstanceWithCorrectData",
			input: &v1.ListInstancesRequest{
				PageNum:  int32(1),
				PageSize: int32(1),
				Scopes:   []string{"Scope1"},
			},
			output: &v1.ListInstancesResponse{
				TotalRecords: int32(2),
				Instances: []*v1.Instance{
					&v1.Instance{
						Id:              "a1",
						Environment:     "env1",
						NumOfProducts:   int32(5),
						NumOfEquipments: int32(5),
					},
					&v1.Instance{
						Id:              "a2",
						Environment:     "env2",
						NumOfProducts:   int32(10),
						NumOfEquipments: int32(10),
					},
				},
			},
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:          []string{"Scope1"},
					PageNum:        input.PageSize*input.PageNum - 1,
					InstanceIDAsc:  true,
					InstanceIDDesc: true,
					PageSize:       input.PageSize}).Return([]db.GetInstancesViewRow{
					{
						Totalrecords:        2,
						InstanceID:          "a1",
						InstanceEnvironment: "env1",
						NumOfProducts:       int32(5),
						NumOfEquipments:     int32(5),
					},
					{
						Totalrecords:        2,
						InstanceID:          "a2",
						InstanceEnvironment: "env2",
						NumOfProducts:       int32(10),
						NumOfEquipments:     int32(10),
					},
				}, nil).Times(1)
			},
			outErr: false,
			ctx:    ctx,
		},
		{
			name:   "ListInstanceWithClaimNotFound",
			input:  &v1.ListInstancesRequest{},
			outErr: true,
			mock:   func(*v1.ListInstancesRequest) {},
			ctx:    context.Background(),
		},
		{
			name: "ListInstanceWithInvalidData",
			input: &v1.ListInstancesRequest{
				PageNum:  -1,
				PageSize: -1,
				Scopes:   []string{"Scope1"},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:          []string{"Scope1"},
					PageNum:        input.PageSize * (input.PageNum - 1),
					InstanceIDAsc:  true,
					InstanceIDDesc: true,
					PageSize:       input.PageSize}).Return(nil, errors.New("rpc error: code = Unknown desc = DBError")).Times(1)
			},
		},
		{
			name: "ListInstanceWithNoRecordFound",
			input: &v1.ListInstancesRequest{
				Scopes: []string{"Scope1"},
			},
			outErr: false,
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:          []string{"Scope1"},
					PageNum:        input.PageSize * (input.PageNum - 1),
					InstanceIDAsc:  true,
					InstanceIDDesc: true,
					PageSize:       input.PageSize}).Return([]db.GetInstancesViewRow{}, nil).Times(1)
			},
			ctx: ctx,
			output: &v1.ListInstancesResponse{
				TotalRecords: 0,
				Instances:    []*v1.Instance{},
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewApplicationServiceServer(dbObj, qObj)
			got, err := s.ListInstances(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestUpsertInstance(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.UpsertInstanceRequest
		output *v1.UpsertInstanceResponse
		mock   func(*v1.UpsertInstanceRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "UpsertInstanceWithCorrectData",
			input: &v1.UpsertInstanceRequest{
				ApplicationId: "a1",
				InstanceId:    "i1",
				InstanceName:  "iname",
				Products: &v1.UpsertInstanceRequestProduct{
					Operation: "add",
					ProductId: []string{"p1", "p2"},
				},
				Equipments: &v1.UpsertInstanceRequestEquipment{
					Operation:   "add",
					EquipmentId: []string{"e1", "e2"},
				},
			},
			output: &v1.UpsertInstanceResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertInstanceRequest) {
				firstCall := dbObj.EXPECT().UpsertInstanceTX(ctx, input).Return(nil).Times(1)
				jsonData, err := json.Marshal(input)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgWorker.Envelope{Type: dgWorker.UpsertInstanceRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				job := job.Job{
					Type:   sql.NullString{String: "lw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}
				qObj.EXPECT().PushJob(ctx, job, "lw").Return(int32(1000), nil).After(firstCall)
			},
		},
		{
			name: "UpsertInstanceWithMissingInstanceId",
			input: &v1.UpsertInstanceRequest{
				ApplicationId: "a1",
				InstanceName:  "iname",
				Products: &v1.UpsertInstanceRequestProduct{
					Operation: "add",
					ProductId: []string{"p1", "p2"},
				},
				Equipments: &v1.UpsertInstanceRequestEquipment{
					Operation:   "add",
					EquipmentId: []string{"e1", "e2"},
				},
			},
			output: &v1.UpsertInstanceResponse{Success: false},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.UpsertInstanceRequest) {
				dbObj.EXPECT().UpsertInstanceTX(ctx, input).Return(errors.New("DB Error")).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewApplicationServiceServer(dbObj, qObj)
			got, err := s.UpsertInstance(test.ctx, test.input)
			log.Println(" log to be removed RESP[", got, "][", err, "]")
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func Test_applicationServiceServer_DropApplicationData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Application
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DropApplicationDataRequest
	}
	tests := []struct {
		name    string
		s       *applicationServiceServer
		args    args
		setup   func()
		want    *v1.DropApplicationDataResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DropApplicationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DropApplicationDataTX(ctx, "Scope1").Times(1).Return(nil)
				jsonData, err := json.Marshal(&v1.DropApplicationDataRequest{
					Scope: "Scope1",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgWorker.Envelope{Type: dgWorker.DropApplicationDataRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				job := job.Job{
					Type:   sql.NullString{String: "lw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}
				mockQueue.EXPECT().PushJob(ctx, job, "lw").Return(int32(1000), nil)
			},
			want: &v1.DropApplicationDataResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.DropApplicationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {},
			want: &v1.DropApplicationDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DropApplicationDataRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {},
			want: &v1.DropApplicationDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DropApplicationDataTX - DBError",
			args: args{
				ctx: ctx,
				req: &v1.DropApplicationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DropApplicationDataTX(ctx, "Scope1").Times(1).Return(errors.New("Internal"))
			},
			want: &v1.DropApplicationDataResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &applicationServiceServer{
				applicationRepo: rep,
				queue:           queue,
			}
			got, err := tt.s.DropApplicationData(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("applicationServiceServer.DropApplicationData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applicationServiceServer.DropApplicationData() = %v, want %v", got, tt.want)
			}
		})
	}
}
