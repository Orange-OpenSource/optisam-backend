package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	appv1 "optisam-backend/application-service/pkg/api/v1"
	mockapp "optisam-backend/application-service/pkg/api/v1/mock"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	job "optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/config"
	repo "optisam-backend/dps-service/pkg/repository/v1"
	dbmock "optisam-backend/dps-service/pkg/repository/v1/dbmock"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/dps-service/pkg/repository/v1/queuemock"
	"optisam-backend/dps-service/pkg/worker/constants"
	equipv1 "optisam-backend/equipment-service/pkg/api/v1"
	mockequip "optisam-backend/equipment-service/pkg/api/v1/mock"
	prov1 "optisam-backend/product-service/pkg/api/v1"
	mockpro "optisam-backend/product-service/pkg/api/v1/mock"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func Test_NotifyUpload(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	tm := time.Now()
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.NotifyUploadRequest
		setup   func(*v1.NotifyUploadRequest)
		output  *v1.NotifyUploadResponse
		wantErr bool
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.NotifyUploadRequest{
				Scope:      "Scope1",
				Type:       "data",
				UploadedBy: "admin@test.com",
				Files:      []string{"Scope1_applications.csv"},
			},
			setup:   func(*v1.NotifyUploadRequest) {},
			output:  &v1.NotifyUploadResponse{Success: false},
			wantErr: true,
		},
		{
			name: "Scope Not found",
			ctx:  ctx,
			input: &v1.NotifyUploadRequest{
				Scope:      "Scope10",
				Type:       "data",
				UploadedBy: "admin@test.com",
				Files:      []string{"Scope0_applications.csv"},
			},
			setup:   func(*v1.NotifyUploadRequest) {},
			output:  &v1.NotifyUploadResponse{Success: false},
			wantErr: true,
		},
		{
			name: "Deletion is already running",
			ctx:  context.Background(),
			input: &v1.NotifyUploadRequest{
				Scope:      "Scope1",
				Type:       "data",
				UploadedBy: "admin@test.com",
				Files:      []string{"Scope1_applications.csv"},
			},
			setup: func(*v1.NotifyUploadRequest) {

				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetDeletionStatus(ctx, "Scope1").Return(int64(1), nil).Times(1)
			},
			output:  &v1.NotifyUploadResponse{Success: false},
			wantErr: true,
		},
		{
			name: "SUCCESS",
			ctx:  ctx,
			input: &v1.NotifyUploadRequest{
				Scope:      "Scope1",
				Type:       "data",
				UploadedBy: "admin@test.com",
				Files:      []string{"Scope1_applications.csv"},
			},
			setup: func(*v1.NotifyUploadRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)

				rep = mockRepo
				mockRepo.EXPECT().GetDeletionStatus(ctx, "Scope1").Return(int64(0), nil).Times(1)
				mockRepo.EXPECT().GetInjectionStatus(ctx, "Scope1").Return(int64(0), nil).Times(1)
				mockRepo.EXPECT().InsertUploadedData(ctx, db.InsertUploadedDataParams{
					FileName:   "Scope1_applications.csv",
					DataType:   db.DataTypeDATA,
					Scope:      "Scope1",
					UploadedBy: "admin@test.com",
					Gid:        int32(0),
					Status:     db.UploadStatusPENDING,
					ScopeType:  db.ScopeTypesGENERIC,
					AnalysisID: sql.NullString{String: "", Valid: true},
				}).Return(db.UploadedDataFile{
					UploadID:       int32(1),
					Scope:          "Scope1",
					DataType:       db.DataTypeDATA,
					FileName:       "Scope1_applications.csv",
					Status:         db.UploadStatusPENDING,
					UploadedBy:     "admin@test.com",
					UploadedOn:     tm,
					TotalRecords:   int32(0),
					SuccessRecords: int32(0),
					FailedRecords:  int32(0),
					Gid:            int32(0),
					ScopeType:      db.ScopeTypesGENERIC,
					AnalysisID:     sql.NullString{String: "", Valid: true},
				}, nil).Times(1)

				dataForJob, _ := json.Marshal(db.UploadedDataFile{
					UploadID:       int32(1),
					Scope:          "Scope1",
					DataType:       db.DataTypeDATA,
					FileName:       "Scope1_applications.csv",
					Status:         db.UploadStatusPENDING,
					UploadedBy:     "admin@test.com",
					UploadedOn:     tm,
					TotalRecords:   int32(0),
					SuccessRecords: int32(0),
					FailedRecords:  int32(0),
					Gid:            int32(0),
					ScopeType:      db.ScopeTypesGENERIC,
					AnalysisID:     sql.NullString{String: "", Valid: true},
				})
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   constants.FILETYPE,
					Status: job.JobStatusPENDING,
					Data:   dataForJob,
				}, constants.FILEWORKER).Return(int32(0), nil).Times(1)
			},
			output:  &v1.NotifyUploadResponse{Success: true},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, qObj, nil)
			_, err := obj.NotifyUpload(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.NotifyUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("Test passed ", tt.name)
		})
	}
}

func Test_dpsServiceServer_ListUploadGlobalData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	ut := time.Now()
	gt, _ := ptypes.TimestampProto(ut)
	var mockCtrl *gomock.Controller
	var rep repo.Dps
	var queue workerqueue.Queue
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.ListUploadRequest
		setup   func(*v1.ListUploadRequest)
		output  *v1.ListUploadResponse
		wantErr bool
	}{
		{
			name: "SuccessCaseNoError",
			ctx:  ctx,
			input: &v1.ListUploadRequest{
				PageNum:   int32(1),
				PageSize:  int32(10),
				SortBy:    v1.ListUploadRequest_SortBy(0),
				SortOrder: v1.ListUploadRequest_SortOrder(0),
				Scope:     "scope1",
			},
			setup: func(req *v1.ListUploadRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				config.SetConfig(config.Config{RawdataLocation: "testErrFileLocation"})
				mockRepository.EXPECT().ListUploadedGlobalDataFiles(ctx, db.ListUploadedGlobalDataFilesParams{
					Scope:       []string{"scope1"},
					PageNum:     int32(10) * (int32(1) - 1),
					PageSize:    int32(10),
					UploadIDAsc: true,
				}).Times(1).Return([]db.ListUploadedGlobalDataFilesRow{
					{
						Totalrecords: int64(2),
						UploadID:     int32(1),
						Scope:        "scope1",
						DataType:     db.DataTypeGLOBALDATA,
						FileName:     "temp.xlsx",
						Status:       db.UploadStatusUPLOADED,
						UploadedBy:   "dummy",
						UploadedOn:   ut,
						AnalysisID:   sql.NullString{String: "121", Valid: true},
					},
					{
						Totalrecords: int64(2),
						UploadID:     int32(2),
						Scope:        "scope1",
						DataType:     db.DataTypeGLOBALDATA,
						FileName:     "temp2.xlsx",
						Status:       db.UploadStatusUPLOADED,
						UploadedBy:   "dummy2",
						UploadedOn:   ut,
						AnalysisID:   sql.NullString{String: "121", Valid: true},
					}}, nil)

			},
			output: &v1.ListUploadResponse{
				TotalRecords: int32(2),
				Uploads: []*v1.Upload{
					{
						UploadId:     int32(1),
						Scope:        "scope1",
						FileName:     "temp.xlsx",
						Status:       "UPLOADED",
						UploadedBy:   "dummy",
						UploadedOn:   gt,
						ErrorFileApi: "",
					},
					{
						UploadId:   int32(2),
						Scope:      "scope1",
						FileName:   "temp2.xlsx",
						Status:     "UPLOADED",
						UploadedBy: "dummy2",
						UploadedOn: gt,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Context not found ",
			ctx:  context.Background(),
			input: &v1.ListUploadRequest{
				Scope: "scope1",
			},
			setup:   func(req *v1.ListUploadRequest) {},
			wantErr: true,
		},
		{
			name: "scope out of context ",
			ctx:  context.Background(),
			input: &v1.ListUploadRequest{
				Scope: "scope5",
			},
			setup:   func(req *v1.ListUploadRequest) {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, &queue, nil)
			got, err := obj.ListUploadGlobalData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.ListUploadGlobalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListUploadResponse(t, "dpsServiceServer.ListUploadGlobalData()", tt.output, got)
			}
		})
	}
}

func Test_dpsServiceServer_GetGlobalFileInfo(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	ctx2 := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope10"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Dps
	var queue workerqueue.Queue
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.GetAnalysisFileInfoRequest
		setup   func(*v1.GetAnalysisFileInfoRequest)
		output  *v1.GetAnalysisFileInfoResponse
		wantErr bool
	}{
		{
			name: "SuccessCaseNoError",
			ctx:  ctx,
			input: &v1.GetAnalysisFileInfoRequest{
				Scope:    "scope1",
				UploadId: int32(10),
				FileType: "error",
			},
			setup: func(req *v1.GetAnalysisFileInfoRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().GetGlobalFileInfo(ctx, db.GetGlobalFileInfoParams{UploadID: int32(10), Scope: "scope1"}).Return(db.GetGlobalFileInfoRow{AnalysisID: sql.NullString{String: "123", Valid: true}, FileName: "temp.xlsx", UploadID: int32(1), ScopeType: db.ScopeTypesGENERIC}, nil).Times(1)
			},
			output: &v1.GetAnalysisFileInfoResponse{
				FileName: "bad_123_temp.xlsx",
			},
			wantErr: false,
		},
		{
			name: "Context not found ",
			ctx:  context.Background(),
			input: &v1.GetAnalysisFileInfoRequest{
				Scope: "scope1",
			},
			setup:   func(req *v1.GetAnalysisFileInfoRequest) {},
			wantErr: true,
		},
		{
			name: "scope out of context ",
			ctx:  context.Background(),
			input: &v1.GetAnalysisFileInfoRequest{
				Scope: "scope5",
			},
			setup:   func(req *v1.GetAnalysisFileInfoRequest) {},
			wantErr: true,
		},
		{
			name: "Unauthorised Role",
			ctx:  ctx2,
			input: &v1.GetAnalysisFileInfoRequest{
				Scope: "scope10",
			},
			setup:   func(req *v1.GetAnalysisFileInfoRequest) {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, &queue, nil)
			got, err := obj.GetAnalysisFileInfo(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.GetAnalysisFileInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.FileName != tt.output.FileName {
				t.Errorf("dpsServiceServer.GetAnalysisFileInfo() error got = %v, want %v", got, tt.output)
				return
			}
		})
	}
}

type contextMatcher struct {
	q context.Context
	t *testing.T
}

func (c *contextMatcher) Matches(x interface{}) bool {
	expCtx, ok := x.(context.Context)
	if !ok {
		return ok
	}
	return compareContext(c, expCtx)
}
func compareContext(c *contextMatcher, exp context.Context) bool {
	if exp == nil {
		return false
	}
	matcherClaims, ok := grpc_middleware.RetrieveClaims(c.q)
	if !ok {
		return false
	}
	expClaims, ok := grpc_middleware.RetrieveClaims(exp)
	if !ok {
		return false
	}
	if !assert.Equalf(c.t, matcherClaims.UserID, expClaims.UserID, "UserID are not same") {
		return false
	}
	if !assert.Equalf(c.t, matcherClaims.Role, expClaims.Role, "Role are not same") {
		return false
	}
	if !assert.ElementsMatchf(c.t, matcherClaims.Socpes, expClaims.Socpes, "Socpes are not same") {
		return false
	}
	return true
}

func (p *contextMatcher) String() string {
	return "compareContext"
}

func Test_dpsServiceServer_DeleteInventory(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var appClient appv1.ApplicationServiceClient
	var equipClient equipv1.EquipmentServiceClient
	var prodClient prov1.ProductServiceClient
	var rep repo.Dps
	var queue workerqueue.Queue
	type args struct {
		ctx context.Context
		req *v1.DeleteInventoryRequest
	}
	tests := []struct {
		name    string
		d       *dpsServiceServer
		args    args
		setup   func()
		want    *v1.DeleteInventoryResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteInventoryRequest{
					Scope:        "Scope1",
					DeletionType: v1.DeleteInventoryRequest_ACQRIGHTS,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockAppClient := mockapp.NewMockApplicationServiceClient(mockCtrl)
				appClient = mockAppClient
				mockEquipClient := mockequip.NewMockEquipmentServiceClient(mockCtrl)
				equipClient = mockEquipClient
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prodClient = mockProdClient
				mockRepo.EXPECT().GetInjectionStatus(ctx, "Scope1").Return(int64(0), nil).Times(1)
				mockRepo.EXPECT().GetDeletionStatus(ctx, "Scope1").Return(int64(0), nil).Times(1)
				mockRepo.EXPECT().SetDeletionActive(ctx, db.SetDeletionActiveParams{
					Scope:        "Scope1",
					DeletionType: "ACQRIGHTS",
					CreatedBy:    "admin@superuser.com",
				}).Return(int32(1), nil).Times(1)
				mockAppClient.EXPECT().DropApplicationData(ctx, &appv1.DropApplicationDataRequest{
					Scope: "Scope1",
				}).Times(1).Return(&appv1.DropApplicationDataResponse{
					Success: true,
				}, nil)
				ctx1, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
				defer cancel()
				mockEquipClient.EXPECT().DropEquipmentData(&contextMatcher{q: ctx1, t: t}, &equipv1.DropEquipmentDataRequest{
					Scope: "Scope1",
				}).Times(1).Return(&equipv1.DropEquipmentDataResponse{
					Success: true,
				}, nil)
				mockProdClient.EXPECT().DropProductData(ctx, &prov1.DropProductDataRequest{
					Scope:        "Scope1",
					DeletionType: prov1.DropProductDataRequest_ACQRIGHTS,
				}).Times(1).Return(&prov1.DropProductDataResponse{
					Success: true,
				}, nil)
				mockRepo.EXPECT().UpdateDeletionStatus(ctx, db.UpdateDeletionStatusParams{
					Status: db.UploadStatusSUCCESS,
					Reason: sql.NullString{String: "", Valid: true},
					ID:     int32(1)}).Return(nil).Times(1)
			},
			want: &v1.DeleteInventoryResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "Already Deletion running",
			args: args{
				ctx: ctx,
				req: &v1.DeleteInventoryRequest{
					Scope:        "Scope1",
					DeletionType: v1.DeleteInventoryRequest_FULL,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetInjectionStatus(ctx, "Scope1").Return(int64(0), nil).Times(1)
				mockRepo.EXPECT().GetDeletionStatus(ctx, "Scope1").Return(int64(1), nil).Times(1)
				mockAppClient := mockapp.NewMockApplicationServiceClient(mockCtrl)
				appClient = mockAppClient
				mockEquipClient := mockequip.NewMockEquipmentServiceClient(mockCtrl)
				equipClient = mockEquipClient
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prodClient = mockProdClient

			},
			want: &v1.DeleteInventoryResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "Already Injection running",
			args: args{
				ctx: ctx,
				req: &v1.DeleteInventoryRequest{
					Scope:        "Scope1",
					DeletionType: v1.DeleteInventoryRequest_FULL,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockAppClient := mockapp.NewMockApplicationServiceClient(mockCtrl)
				appClient = mockAppClient
				mockEquipClient := mockequip.NewMockEquipmentServiceClient(mockCtrl)
				equipClient = mockEquipClient
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prodClient = mockProdClient
				mockRepo.EXPECT().GetInjectionStatus(ctx, "Scope1").Return(int64(1), nil).Times(1)
			},
			want: &v1.DeleteInventoryResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteInventoryRequest{
					Scope:        "Scope1",
					DeletionType: v1.DeleteInventoryRequest_FULL,
				},
			},
			setup: func() {},
			want: &v1.DeleteInventoryResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteInventoryRequest{
					Scope:        "Scope4",
					DeletionType: v1.DeleteInventoryRequest_FULL,
				},
			},
			setup: func() {},
			want: &v1.DeleteInventoryResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.d = &dpsServiceServer{
				dpsRepo:     rep,
				queue:       &queue,
				application: appClient,
				equipment:   equipClient,
				product:     prodClient,
			}
			got, err := tt.d.DeleteInventory(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.DeleteInventory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dpsServiceServer.DeleteInventory() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_DropUploadedFileData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})

	var rep repo.Dps
	tests := []struct {
		name    string
		d       *dpsServiceServer
		input   *v1.DropUploadedFileDataRequest
		setup   func()
		wantErr bool
		ctx     context.Context
		cleanup func()
	}{
		{
			name:    "ScopeNotFound",
			input:   &v1.DropUploadedFileDataRequest{Scope: "s1"},
			setup:   func() {},
			wantErr: true,
			cleanup: func() {},
			ctx:     ctx,
		},
		{
			name:    "ClaimsNotFound",
			input:   &v1.DropUploadedFileDataRequest{Scope: "Scope1"},
			setup:   func() {},
			wantErr: true,
			cleanup: func() {},
			ctx:     context.Background(),
		},
		{
			name:  "ScopeDataNotPresent",
			input: &v1.DropUploadedFileDataRequest{Scope: "Scope2"},
			setup: func() {
				mockCtrl := gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropFileRecords(ctx, "Scope2").Return(nil).Times(1)
			},
			wantErr: false,
			ctx:     ctx,
		},
		{
			name:  "SuccessfullyDeletedFilesRecords",
			input: &v1.DropUploadedFileDataRequest{Scope: "Scope1"},
			setup: func() {
				mockCtrl := gomock.NewController(t)
				mockRepo := dbmock.NewMockDps(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropFileRecords(ctx, "Scope1").Return(nil).Times(1)
			},
			wantErr: false,
			ctx:     ctx,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.d = &dpsServiceServer{
				dpsRepo: rep,
			}
			_, err := tt.d.DropUploadedFileData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.DropUploadedFileData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func compareListUploadResponse(t *testing.T, name string, exp, act *v1.ListUploadResponse) {
	assert.Equalf(t, exp.TotalRecords, act.TotalRecords, "%s. TotalRecords are not same", name)
	if !assert.Equal(t, len(exp.Uploads), len(act.Uploads), "number of elements are not same") {
		return
	}
	for i := range exp.Uploads {
		compareUpload(t, fmt.Sprintf("%s[%d]", name, i), exp.Uploads[i], act.Uploads[i])
	}
}

func compareUpload(t *testing.T, name string, exp, act *v1.Upload) {
	assert.Equalf(t, exp.UploadId, act.UploadId, "%s. UploadId are not same", name)
	assert.Equalf(t, exp.FileName, act.FileName, "%s. FileName are not same", name)
	assert.Equalf(t, exp.Status, act.Status, "%s. Status are not same", name)
	assert.Equalf(t, exp.UploadedBy, act.UploadedBy, "%s. UploadedBy are not same", name)
	assert.Equalf(t, exp.Scope, act.Scope, "%s. Scope are not same", name)
	assert.Equalf(t, exp.UploadedOn, act.UploadedOn, "%s. UploadedOn are not same", name)
	assert.Equalf(t, exp.SuccessRecords, act.SuccessRecords, "%s. SuccessRecords are not same", name)
	assert.Equalf(t, exp.FailedRecords, act.FailedRecords, "%s. FailedRecords are not same", name)
	assert.Equalf(t, exp.ErrorFileApi, act.ErrorFileApi, "%s. ErrorFileApi are not same", name)
}

func Test_dpsServiceServer_ListDeletionRecords(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	ut := time.Now()
	gt, _ := ptypes.TimestampProto(ut)
	var mockCtrl *gomock.Controller
	var rep repo.Dps
	var queue workerqueue.Queue
	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.ListDeletionRequest
		setup   func(*v1.ListDeletionRequest)
		output  *v1.ListDeletionResponse
		wantErr bool
	}{
		{
			name: "SuccessCaseNoError",
			ctx:  ctx,
			input: &v1.ListDeletionRequest{
				PageNum:   int32(1),
				PageSize:  int32(10),
				SortBy:    v1.ListDeletionRequest_SortBy(0),
				SortOrder: v1.ListDeletionRequest_SortOrder(0),
				Scope:     "scope1",
			},
			setup: func(req *v1.ListDeletionRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockRepository.EXPECT().ListDeletionRecrods(ctx, db.ListDeletionRecrodsParams{
					Scope:           "scope1",
					PageNum:         int32(10) * (int32(1) - 1),
					PageSize:        int32(10),
					DeletionTypeAsc: true,
				}).Times(1).Return([]db.ListDeletionRecrodsRow{
					{
						Totalrecords: int64(2),
						Scope:        "scope1",
						Status:       db.UploadStatusFAILED,
						CreatedBy:    "admin@superuser.com",
						CreatedOn:    ut,
						DeletionType: db.DeletionTypeACQRIGHTS,
					},
					{
						Totalrecords: int64(2),
						Scope:        "scope1",
						Status:       db.UploadStatusSUCCESS,
						CreatedBy:    "admin@superuser.com",
						CreatedOn:    ut,
						DeletionType: db.DeletionTypeWHOLEINVENTORY,
					}}, nil)

			},
			output: &v1.ListDeletionResponse{
				TotalRecords: int32(2),
				Deletions: []*v1.Deletion{
					{
						DeletionType: "Acquired Rights",
						CreatedBy:    "admin@superuser.com",
						Status:       "FAILED",
						CreatedOn:    gt,
					},
					{
						DeletionType: "Whole Inventory",
						CreatedBy:    "admin@superuser.com",
						Status:       "SUCCESS",
						CreatedOn:    gt,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Context not found ",
			ctx:  context.Background(),
			input: &v1.ListDeletionRequest{
				Scope: "scope1",
			},
			setup:   func(req *v1.ListDeletionRequest) {},
			wantErr: true,
		},
		{
			name: "scope out of context ",
			ctx:  context.Background(),
			input: &v1.ListDeletionRequest{
				Scope: "scope5",
			},
			setup:   func(req *v1.ListDeletionRequest) {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, &queue, nil)
			got, err := obj.ListDeletionRecords(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.ListDeletionRecords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("dpsServiceServer.ListDeletionRecords() = %v, want %v", got, tt.output)
			}
		})
	}
}
