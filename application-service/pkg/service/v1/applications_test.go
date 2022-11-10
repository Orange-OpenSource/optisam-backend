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
	prov1 "optisam-backend/product-service/pkg/api/v1"
	promock "optisam-backend/product-service/pkg/api/v1/mock"
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

func Test_DropObscolenscence(t *testing.T) {
	ctx1 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2"},
	})
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name    string
		input   *v1.DropObscolenscenceDataRequest
		setup   func()
		wantErr bool
		ctx     context.Context
	}{
		{
			name:    "ScopeValidationFailure",
			wantErr: true,
			setup:   func() {},
			ctx:     ctx1,
			input:   &v1.DropObscolenscenceDataRequest{Scope: "Scope11"},
		},
		{
			name:    "ClaimsNotFound",
			wantErr: true,
			setup:   func() {},
			ctx:     ctx,
			input:   &v1.DropObscolenscenceDataRequest{Scope: "Scope1"},
		},
		{
			name:    "DBError",
			wantErr: true,
			setup: func() {
				dbObj.EXPECT().DropObscolenscenceDataTX(ctx1, "Scope1").Return(errors.New("DBError")).Times(1)
			},
			ctx:   ctx1,
			input: &v1.DropObscolenscenceDataRequest{Scope: "Scope1"},
		},
		{
			name:    "SuccessFullyApplicationResourceDeleted",
			wantErr: false,
			setup: func() {
				dbObj.EXPECT().DropObscolenscenceDataTX(ctx1, "Scope1").Return(nil).Times(1)
			},
			ctx:   ctx1,
			input: &v1.DropObscolenscenceDataRequest{Scope: "Scope1"},
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.setup()
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			_, err := s.DropObscolenscenceData(test.ctx, test.input)
			if (err != nil) != test.wantErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.wantErr, err)
				return
			}
		})
	}
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
				// Version:       "a1version",
				// Owner:         "a1owner",
				Scope:  "Scope1",
				Domain: "Payments",
			},
			output: &v1.UpsertApplicationResponse{Success: true},
			mock: func(input *v1.UpsertApplicationRequest) {
				firstCall := dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{
					ApplicationID:   "a1",
					ApplicationName: "a1name",
					// ApplicationOwner:   "a1owner",
					// ApplicationVersion: "a1version",
					Scope:             "Scope1",
					ApplicationDomain: "Payments",
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
				dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{Scope: "Scope1", ApplicationDomain: "Not specified"}).Return(errors.New("rpc error: code = Internal desc = DBError")).Times(1)
			},
			outErr: true,
		},
		{
			name: "UpsertApplicationWithMissingapplicationID",
			input: &v1.UpsertApplicationRequest{
				Name: "a1name",
				// Owner:   "a1owner",
				// Version: "a1version",
				Scope: "Scope1",
			},
			output: &v1.UpsertApplicationResponse{Success: false},
			mock: func(input *v1.UpsertApplicationRequest) {
				dbObj.EXPECT().UpsertApplication(ctx, db.UpsertApplicationParams{
					ApplicationName: "a1name",
					// ApplicationVersion: "a1version",
					// ApplicationOwner:   "a1owner",
					ApplicationDomain: "Not specified",
					Scope:             "Scope1",
				}).Return(errors.New("rpc error: code = Internal desc = DBError")).Times(1)
			},
			outErr: true,
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewApplicationServiceServer(dbObj, qObj, nil)
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
					{
						ApplicationId:    "a1",
						Name:             "a1name",
						Owner:            "a1owner",
						NumOfInstances:   int32(5),
						NumOfProducts:    int32(5),
						NumOfEquipments:  int32(5),
						Domain:           "Payments",
						ObsolescenceRisk: "Risk1",
					},
					{
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
						NumOfEquipments:   int32(5),
						ApplicationDomain: "Payments",
						ObsolescenceRisk:  sql.NullString{String: "Risk1", Valid: true},
					},
					{
						Totalrecords:     int64(2),
						ApplicationID:    "a2",
						ApplicationName:  "a2name",
						ApplicationOwner: "a2owner",
						NumOfEquipments:  int32(3),
					}}, nil).Times(1)
			},
			isErr: false,
			ctx:   ctx,
		},
		{
			name: "SUCCESS - ListApplications With Multiple filtering key",
			input: &v1.ListApplicationsRequest{
				PageNum:  int32(1),
				PageSize: int32(2),
				Scopes:   []string{"Scope1"},
				SearchParams: &v1.ApplicationSearchParams{
					ProductId: &v1.StringFilter{
						FilteringkeyMultiple: []string{"swid1", "swid2"},
					},
				},
			},
			output: &v1.ListApplicationsResponse{
				TotalRecords: 2,
				Applications: []*v1.Application{
					{
						ApplicationId:    "a1",
						Name:             "a1name",
						Owner:            "a1owner",
						NumOfInstances:   int32(5),
						NumOfEquipments:  int32(5),
						Domain:           "Payments",
						ObsolescenceRisk: "Risk1",
					},
					{
						ApplicationId:   "a2",
						Name:            "a2name",
						Owner:           "a2owner",
						NumOfInstances:  int32(3),
						NumOfEquipments: int32(3),
					},
				},
			},
			mock: func(input *v1.ListApplicationsRequest) {
				dbObj.EXPECT().GetApplicationsByProduct(ctx, db.GetApplicationsByProductParams{
					Scope:              []string{"Scope1"},
					ApplicationID:      []string{"a1", "a2"},
					ApplicationNameAsc: true,
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize}).Return([]db.GetApplicationsByProductRow{
					{
						Totalrecords:      int64(2),
						ApplicationID:     "a1",
						ApplicationName:   "a1name",
						ApplicationOwner:  "a1owner",
						NumOfEquipments:   int32(5),
						ApplicationDomain: "Payments",
						ObsolescenceRisk:  sql.NullString{String: "Risk1", Valid: true},
					},
					{
						Totalrecords:     int64(2),
						ApplicationID:    "a2",
						ApplicationName:  "a2name",
						ApplicationOwner: "a2owner",
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
			name: "ListApplicationWithScopeError",
			input: &v1.ListApplicationsRequest{
				Scopes: []string{"Scope4"},
			},
			mock:  func(input *v1.ListApplicationsRequest) {},
			isErr: true,
			ctx:   ctx,
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
		{
			name: "FAILURE - GetApplicationsByProduct - DBError",
			input: &v1.ListApplicationsRequest{
				PageNum:  int32(-1),
				PageSize: int32(-1),
				Scopes:   []string{"Scope1"},
				SearchParams: &v1.ApplicationSearchParams{
					ProductId: &v1.StringFilter{
						FilteringkeyMultiple: []string{"swid1", "swid2"},
					},
				},
			},
			mock: func(input *v1.ListApplicationsRequest) {
				dbObj.EXPECT().GetApplicationsByProduct(ctx, db.GetApplicationsByProductParams{
					Scope:              []string{"Scope1"},
					ApplicationID:      []string{"a1", "a2"},
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
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.ListApplications(test.ctx, test.input)
			// log.Println(" log to be removed RESP[", got, "][", err, "]")
			if (err != nil) != test.isErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, got, (test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func Test_applicationServiceServer_ListInstances(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockApplication(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	proObj := promock.NewMockProductServiceClient(mockCtrl)
	var pro prov1.ProductServiceClient
	pro = proObj
	testSet := []struct {
		name   string
		input  *v1.ListInstancesRequest
		output *v1.ListInstancesResponse
		mock   func(*v1.ListInstancesRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "Success:No Filter",
			input: &v1.ListInstancesRequest{
				PageNum:   int32(1),
				PageSize:  int32(10),
				SortOrder: v1.SortOrder_asc,
				Scopes:    []string{"Scope1"},
			},
			output: &v1.ListInstancesResponse{
				TotalRecords: int32(2),
				Instances: []*v1.Instance{
					{
						Id:              "a1",
						Environment:     "env1",
						NumOfProducts:   int32(5),
						NumOfEquipments: int32(2),
					},
					{
						Id:              "a2",
						Environment:     "env2",
						NumOfProducts:   int32(10),
						NumOfEquipments: int32(3),
					},
				},
			},
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:                   []string{"Scope1"},
					ApplicationID:           "",
					IsApplicationID:         false,
					ProductID:               "",
					IsProductID:             false,
					InstanceIDAsc:           true,
					InstanceIDDesc:          false,
					InstanceEnvironmentAsc:  false,
					InstanceEnvironmentDesc: false,
					NumOfProductsAsc:        false,
					NumOfProductsDesc:       false,
					PageNum:                 input.PageSize * (input.PageNum - 1),
					PageSize:                input.PageSize,
				}).Return([]db.GetInstancesViewRow{
					{
						Totalrecords:        int64(2),
						InstanceID:          "a1",
						InstanceEnvironment: "env1",
						NumOfProducts:       int32(5),
					},
					{
						Totalrecords:        int64(2),
						InstanceID:          "a2",
						InstanceEnvironment: "env2",
						NumOfProducts:       int32(10),
					},
				}, nil).Times(1)
				gomock.InOrder(
					dbObj.EXPECT().GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
						Scope:           "Scope1",
						InstanceID:      "a1",
						EquipmentIds:    []string{},
						ProductID:       "",
						ApplicationID:   "",
						IsProductID:     false,
						IsApplicationID: false,
					}).Times(1).Return([]int64{2}, nil),
					dbObj.EXPECT().GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
						Scope:           "Scope1",
						InstanceID:      "a2",
						EquipmentIds:    []string{},
						ProductID:       "",
						ApplicationID:   "",
						IsProductID:     false,
						IsApplicationID: false,
					}).Times(1).Return([]int64{3}, nil),
				)
			},
			outErr: false,
			ctx:    ctx,
		},
		{
			name: "Success:ProductId Filter",
			input: &v1.ListInstancesRequest{
				PageNum:   int32(1),
				PageSize:  int32(10),
				SortOrder: v1.SortOrder_asc,
				Scopes:    []string{"Scope1"},
				SearchParams: &v1.InstanceSearchParams{
					ProductId: &v1.StringFilter{
						Filteringkey: "Oracle",
					},
				},
			},
			output: &v1.ListInstancesResponse{
				TotalRecords: int32(2),
				Instances: []*v1.Instance{
					{
						Id:              "a1",
						Environment:     "env1",
						NumOfProducts:   int32(5),
						NumOfEquipments: int32(2),
					},
					{
						Id:              "a2",
						Environment:     "env2",
						NumOfProducts:   int32(10),
						NumOfEquipments: int32(3),
					},
				},
			},
			mock: func(input *v1.ListInstancesRequest) {
				proObj.EXPECT().GetEquipmentsByProduct(ctx, &prov1.GetEquipmentsByProductRequest{
					Scope:   "Scope1",
					SwidTag: "Oracle",
				}).Times(1).Return(&prov1.GetEquipmentsByProductResponse{
					EquipmentId: []string{"eq1", "eq2", "eq3"},
				}, nil)
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:                   []string{"Scope1"},
					ApplicationID:           "",
					IsApplicationID:         false,
					ProductID:               "Oracle",
					IsProductID:             true,
					InstanceIDAsc:           true,
					InstanceIDDesc:          false,
					InstanceEnvironmentAsc:  false,
					InstanceEnvironmentDesc: false,
					NumOfProductsAsc:        false,
					NumOfProductsDesc:       false,
					PageNum:                 input.PageSize * (input.PageNum - 1),
					PageSize:                input.PageSize,
				}).Return([]db.GetInstancesViewRow{
					{
						Totalrecords:        2,
						InstanceID:          "a1",
						InstanceEnvironment: "env1",
						NumOfProducts:       int32(5),
					},
					{
						Totalrecords:        2,
						InstanceID:          "a2",
						InstanceEnvironment: "env2",
						NumOfProducts:       int32(10),
					},
				}, nil).Times(1)
				gomock.InOrder(
					dbObj.EXPECT().GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
						Scope:           "Scope1",
						InstanceID:      "a1",
						EquipmentIds:    []string{"eq1", "eq2", "eq3"},
						ProductID:       "Oracle",
						ApplicationID:   "",
						IsProductID:     true,
						IsApplicationID: false,
					}).Times(1).Return([]int64{2}, nil),
					dbObj.EXPECT().GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
						Scope:           "Scope1",
						InstanceID:      "a2",
						EquipmentIds:    []string{"eq1", "eq2", "eq3"},
						ProductID:       "Oracle",
						ApplicationID:   "",
						IsProductID:     true,
						IsApplicationID: false,
					}).Times(1).Return([]int64{3}, nil),
				)
			},
			outErr: false,
			ctx:    ctx,
		},
		{
			name:   "Failure:ClaimNotFound",
			input:  &v1.ListInstancesRequest{},
			outErr: true,
			mock:   func(*v1.ListInstancesRequest) {},
			ctx:    context.Background(),
		},
		{
			name: "Failure:ScopeError",
			input: &v1.ListInstancesRequest{
				Scopes: []string{"Scope4"},
			},
			mock:   func(*v1.ListInstancesRequest) {},
			outErr: true,
			ctx:    ctx,
		},
		{
			name: "Failure:InvalidData",
			input: &v1.ListInstancesRequest{
				PageNum:  -1,
				PageSize: -1,
				Scopes:   []string{"Scope1"},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:                   []string{"Scope1"},
					ApplicationID:           "",
					IsApplicationID:         false,
					ProductID:               "",
					IsProductID:             false,
					InstanceIDAsc:           true,
					InstanceIDDesc:          false,
					InstanceEnvironmentAsc:  false,
					InstanceEnvironmentDesc: false,
					NumOfProductsAsc:        false,
					NumOfProductsDesc:       false,
					PageNum:                 input.PageSize * (input.PageNum - 1),
					PageSize:                input.PageSize}).Return(nil, errors.New("rpc error: code = Unknown desc = DBError")).Times(1)
			},
		},
		{
			name: "Failure:NoRecordFound",
			input: &v1.ListInstancesRequest{
				Scopes: []string{"Scope1"},
			},
			outErr: false,
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:                   []string{"Scope1"},
					ApplicationID:           "",
					IsApplicationID:         false,
					ProductID:               "",
					IsProductID:             false,
					InstanceIDAsc:           true,
					InstanceIDDesc:          false,
					InstanceEnvironmentAsc:  false,
					InstanceEnvironmentDesc: false,
					NumOfProductsAsc:        false,
					NumOfProductsDesc:       false,
					PageNum:                 input.PageSize * (input.PageNum - 1),
					PageSize:                input.PageSize}).Return([]db.GetInstancesViewRow{}, nil).Times(1)
			},
			ctx: ctx,
			output: &v1.ListInstancesResponse{
				TotalRecords: 0,
				Instances:    []*v1.Instance{},
			},
		},
		{
			name: "Failure:NoEquipmentFound",
			input: &v1.ListInstancesRequest{
				Scopes: []string{"Scope1"},
			},
			ctx: ctx,
			mock: func(input *v1.ListInstancesRequest) {
				dbObj.EXPECT().GetInstancesView(ctx, db.GetInstancesViewParams{
					Scope:                   []string{"Scope1"},
					ApplicationID:           "",
					IsApplicationID:         false,
					ProductID:               "",
					IsProductID:             false,
					InstanceIDAsc:           true,
					InstanceIDDesc:          false,
					InstanceEnvironmentAsc:  false,
					InstanceEnvironmentDesc: false,
					NumOfProductsAsc:        false,
					NumOfProductsDesc:       false,
					PageNum:                 input.PageSize * (input.PageNum - 1),
					PageSize:                input.PageSize,
				}).Return([]db.GetInstancesViewRow{
					{
						InstanceID:          "",
						InstanceEnvironment: "",
						NumOfProducts:       int32(5),
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetInstanceViewEquipments(ctx, db.GetInstanceViewEquipmentsParams{
					Scope:           "Scope1",
					InstanceID:      "",
					IsProductID:     false,
					ProductID:       "",
					IsApplicationID: false,
					ApplicationID:   "",
					EquipmentIds:    []string{},
				}).Times(1).Return([]int64{}, errors.New("Internal"))
			},
			outErr: true,
		},
		{
			name: "Failure:GetEquipmentByProduct",
			input: &v1.ListInstancesRequest{
				Scopes: []string{"Scope1"},
				SearchParams: &v1.InstanceSearchParams{
					ProductId: &v1.StringFilter{
						Filteringkey: "Oracle",
					},
				},
			},
			ctx: ctx,
			mock: func(input *v1.ListInstancesRequest) {
				proObj.EXPECT().GetEquipmentsByProduct(ctx, &prov1.GetEquipmentsByProductRequest{
					Scope:   "Scope1",
					SwidTag: "Oracle",
				}).Times(1).Return(&prov1.GetEquipmentsByProductResponse{
					EquipmentId: []string{},
				}, errors.New("Internal"))
			},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := &applicationServiceServer{
				applicationRepo: dbObj,
				product:         pro,
				queue:           qObj,
			}
			got, err := s.ListInstances(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			}
			if !reflect.DeepEqual(got, test.output) {
				t.Errorf("applicationServiceServer.ListInstances() = %v, want %v", got, test.output)
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
			s := NewApplicationServiceServer(dbObj, qObj, nil)
			got, err := s.UpsertInstance(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, got, (test.output)) {
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

func Test_applicationServiceServer_GetEquipmentsByApplication(t *testing.T) {
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
		req *v1.GetEquipmentsByApplicationRequest
	}
	tests := []struct {
		name    string
		s       *applicationServiceServer
		args    args
		setup   func()
		want    *v1.GetEquipmentsByApplicationResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByApplicationRequest{
					Scope:         "Scope1",
					ApplicationId: "App_3",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetEquipmentsByApplicationID(ctx, db.GetEquipmentsByApplicationIDParams{
					Scope:         "Scope1",
					ApplicationID: "App_3",
				}).Times(1).Return([]string{"Eq1", "Eq2", "Eq3"}, nil)
			},
			want: &v1.GetEquipmentsByApplicationResponse{
				EquipmentId: []string{"Eq1", "Eq2", "Eq3"},
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.GetEquipmentsByApplicationRequest{
					Scope:         "Scope1",
					ApplicationId: "App_3",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByApplicationRequest{
					Scope:         "Scope4",
					ApplicationId: "App_3",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetEquipmentsByApplicationID - DBError",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByApplicationRequest{
					Scope:         "Scope1",
					ApplicationId: "App_3",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockApplication(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetEquipmentsByApplicationID(ctx, db.GetEquipmentsByApplicationIDParams{
					Scope:         "Scope1",
					ApplicationID: "App_3",
				}).Times(1).Return([]string{}, errors.New("Internal"))
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
			got, err := tt.s.GetEquipmentsByApplication(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("applicationServiceServer.GetEquipmentsByApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applicationServiceServer.GetEquipmentsByApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func Test_applicationServiceServer_GetProductsByApplication(t *testing.T) {
// 	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
// 		UserID: "admin@superuser.com",
// 		Role:   "Admin",
// 		Socpes: []string{"Scope1", "Scope2", "Scope3"},
// 	})
// 	var mockCtrl *gomock.Controller
// 	var rep repo.Application
// 	var queue workerqueue.Workerqueue
// 	type args struct {
// 		ctx context.Context
// 		req *v1.GetProductsByApplicationRequest
// 	}
// 	tests := []struct {
// 		name    string
// 		s       *applicationServiceServer
// 		args    args
// 		setup   func()
// 		want    *v1.GetProductsByApplicationResponse
// 		wantErr bool
// 	}{
// 		{name: "SUCCESS",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.GetProductsByApplicationInstanceRequest{
// 					Scope:         "Scope1",
// 					ApplicationId: "App_3",
// 					InstanceId:    "Ins_1",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := dbmock.NewMockApplication(mockCtrl)
// 				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
// 				rep = mockRepo
// 				queue = mockQueue
// 				mockRepo.EXPECT().GetProductsByApplicationInstanceID(ctx, db.GetProductsByApplicationInstanceIDParams{
// 					Scope:         "Scope1",
// 					ApplicationID: "App_3",
// 					InstanceID:    "Ins_1",
// 				}).Times(1).Return([]string{"a1", "a2", "a3"}, nil)
// 			},
// 			want: &v1.GetProductsByApplicationInstanceResponse{
// 				ProductId: []string{"a1", "a2", "a3"},
// 			},
// 			wantErr: false,
// 		},
// 		{name: "FAILURE - ClaimsNotFound",
// 			args: args{
// 				ctx: context.Background(),
// 				req: &v1.GetProductsByApplicationInstanceRequest{
// 					Scope:         "Scope1",
// 					ApplicationId: "App_3",
// 					InstanceId:    "Ins_1",
// 				},
// 			},
// 			setup:   func() {},
// 			wantErr: true,
// 		},
// 		{name: "FAILURE - ScopeValidationError",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.GetProductsByApplicationInstanceRequest{
// 					Scope:         "Scope4",
// 					ApplicationId: "App_3",
// 					InstanceId:    "Ins_1",
// 				},
// 			},
// 			setup:   func() {},
// 			wantErr: true,
// 		},
// 		{name: "FAILURE - GetProductsByApplicationInstanceID - DBError",
// 			args: args{
// 				ctx: ctx,
// 				req: &v1.GetProductsByApplicationInstanceRequest{
// 					Scope:         "Scope1",
// 					ApplicationId: "App_3",
// 					InstanceId:    "Ins_1",
// 				},
// 			},
// 			setup: func() {
// 				mockCtrl = gomock.NewController(t)
// 				mockRepo := dbmock.NewMockApplication(mockCtrl)
// 				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
// 				rep = mockRepo
// 				queue = mockQueue
// 				mockRepo.EXPECT().GetProductsByApplicationInstanceID(ctx, db.GetProductsByApplicationInstanceIDParams{
// 					Scope:         "Scope1",
// 					ApplicationID: "App_3",
// 					InstanceID:    "Ins_1",
// 				}).Times(1).Return([]string{}, errors.New("Internal"))
// 			},
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.setup()
// 			tt.s = &applicationServiceServer{
// 				applicationRepo: rep,
// 				queue:           queue,
// 			}
// 			got, err := tt.s.GetProductsByApplicationInstance(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("applicationServiceServer.GetProductsByApplicationInstance() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("applicationServiceServer.GetProductsByApplicationInstance() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
