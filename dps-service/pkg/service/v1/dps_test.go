// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"log"
	appv1 "optisam-backend/application-service/pkg/api/v1"
	mockapp "optisam-backend/application-service/pkg/api/v1/mock"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	repo "optisam-backend/dps-service/pkg/repository/v1"
	dbmock "optisam-backend/dps-service/pkg/repository/v1/dbmock"
	"optisam-backend/dps-service/pkg/repository/v1/postgres/db"
	equipv1 "optisam-backend/equipment-service/pkg/api/v1"
	mockequip "optisam-backend/equipment-service/pkg/api/v1/mock"
	prov1 "optisam-backend/product-service/pkg/api/v1"
	mockpro "optisam-backend/product-service/pkg/api/v1/mock"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
)

var (
	dps_ctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ut    = time.Now()
	gt, _ = ptypes.TimestampProto(ut)
)

func Test_ListUploadGlobalData(t *testing.T) {
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
				Scope:     "Scope1",
			},
			setup: func(req *v1.ListUploadRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository

				mockRepository.EXPECT().ListUploadedGlobalDataFiles(ctx, db.ListUploadedGlobalDataFilesParams{
					Scope:       []string{"Scope1"},
					PageNum:     int32(10) * (int32(1) - 1),
					PageSize:    int32(10),
					UploadIDAsc: true,
				}).Times(1).Return([]db.ListUploadedGlobalDataFilesRow{
					{
						Totalrecords: int64(2),
						UploadID:     int32(1),
						Scope:        "Scope1",
						DataType:     db.DataTypeGLOBALDATA,
						FileName:     "temp.xlsx",
						Status:       db.UploadStatusPENDING,
						UploadedBy:   "dummy",
						UploadedOn:   ut,
					},
					{
						Totalrecords: int64(2),
						UploadID:     int32(2),
						Scope:        "Scope1",
						DataType:     db.DataTypeGLOBALDATA,
						FileName:     "temp2.xlsx",
						Status:       db.UploadStatusCOMPLETED,
						UploadedBy:   "dummy2",
						UploadedOn:   ut,
					}}, nil)

			},
			output: &v1.ListUploadResponse{
				TotalRecords: int32(2),
				Uploads: []*v1.Upload{
					&v1.Upload{
						UploadId:   int32(1),
						Scope:      "Scope1",
						FileName:   "temp.xlsx",
						Status:     "PENDING",
						UploadedBy: "dummy",
						UploadedOn: gt,
					},
					&v1.Upload{
						UploadId:   int32(2),
						Scope:      "Scope1",
						FileName:   "temp2.xlsx",
						Status:     "COMPLETED",
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
				Scope: "Scope1",
			},
			setup:   func(req *v1.ListUploadRequest) {},
			wantErr: true,
		},
		{
			name: "scope out of context ",
			ctx:  context.Background(),
			input: &v1.ListUploadRequest{
				Scope: "Scope5",
			},
			setup:   func(req *v1.ListUploadRequest) {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			obj := NewDpsServiceServer(rep, queue, nil)
			got, err := obj.ListUploadGlobalData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.ListUploadGlobalData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("dpsServiceServer.ListUploadGlobalData() = %v, want %v", got, tt.output)
			}
			log.Println("Test Passed : ", tt.name)
		})
	}
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
					Scope: "Scope1",
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
				mockAppClient.EXPECT().DropApplicationData(ctx, &appv1.DropApplicationDataRequest{
					Scope: "Scope1",
				}).Times(1).Return(&appv1.DropApplicationDataResponse{
					Success: true,
				}, nil)
				mockEquipClient.EXPECT().DropEquipmentData(ctx, &equipv1.DropEquipmentDataRequest{
					Scope: "Scope1",
				}).Times(1).Return(&equipv1.DropEquipmentDataResponse{
					Success: true,
				}, nil)
				mockProdClient.EXPECT().DropProductData(ctx, &prov1.DropProductDataRequest{
					Scope: "Scope1",
				}).Times(1).Return(&prov1.DropProductDataResponse{
					Success: true,
				}, nil)
			},
			want: &v1.DeleteInventoryResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteInventoryRequest{
					Scope: "Scope1",
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
					Scope: "Scope4",
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
				queue:       queue,
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
