package v1

import (
	"context"
	"errors"
	"fmt"
	"log"
	v1 "optisam-backend/account-service/pkg/api/v1"
	repv1 "optisam-backend/account-service/pkg/repository/v1"
	"optisam-backend/account-service/pkg/repository/v1/mock"
	"optisam-backend/account-service/pkg/repository/v1/postgres/db"
	application "optisam-backend/application-service/pkg/api/v1"
	appMock "optisam-backend/application-service/pkg/api/v1/mock"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	dpsv1 "optisam-backend/dps-service/pkg/api/v1"
	dpsMock "optisam-backend/dps-service/pkg/api/v1/mock"
	equipment "optisam-backend/equipment-service/pkg/api/v1"
	equipmentMock "optisam-backend/equipment-service/pkg/api/v1/mock"
	metricv1 "optisam-backend/metric-service/pkg/api/v1"
	metricMock "optisam-backend/metric-service/pkg/api/v1/mock"
	product "optisam-backend/product-service/pkg/api/v1"
	prodMock "optisam-backend/product-service/pkg/api/v1/mock"
	reportv1 "optisam-backend/report-service/pkg/api/v1"
	reportMock "optisam-backend/report-service/pkg/api/v1/mock"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_DeleteResourceByScope(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	var prod product.ProductServiceClient
	var app application.ApplicationServiceClient
	var report reportv1.ReportServiceClient
	var metric metricv1.MetricServiceClient
	var equip equipment.EquipmentServiceClient
	var dps dpsv1.DpsServiceClient

	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1", "s2"},
	})
	tests := []struct {
		name    string
		input   *v1.DropScopeDataRequest
		setup   func()
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "ScopeValidationFailure case",
			input:   &v1.DropScopeDataRequest{Scope: "s3"},
			ctx:     ctx,
			setup:   func() {},
			wantErr: true,
		},
		{
			name:    "claimsNotFound case",
			input:   &v1.DropScopeDataRequest{Scope: "s1"},
			ctx:     context.Background(),
			setup:   func() {},
			wantErr: true,
		},
		{
			name:  "DB Failure case",
			input: &v1.DropScopeDataRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropScopeTX(ctx, "s1").Return(errors.New("DBError")).Times(1)
				mockAppClient := appMock.NewMockApplicationServiceClient(mockCtrl)
				app = mockAppClient
				mockAppClient.EXPECT().DropApplicationData(ctx, &application.DropApplicationDataRequest{Scope: "s1"}).Return(&application.DropApplicationDataResponse{Success: true}, nil).Times(1)
				mockAppClient.EXPECT().DropObscolenscenceData(ctx, &application.DropObscolenscenceDataRequest{Scope: "s1"}).Return(&application.DropObscolenscenceDataResponse{Success: true}, nil).Times(1)

				mockProdClient := prodMock.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockProdClient.EXPECT().DropProductData(ctx, &product.DropProductDataRequest{Scope: "s1", DeletionType: product.DropProductDataRequest_FULL}).Return(&product.DropProductDataResponse{Success: true}, nil).Times(1)

				mockEquipClient := equipmentMock.NewMockEquipmentServiceClient(mockCtrl)
				equip = mockEquipClient
				ctx1, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
				defer cancel()
				mockEquipClient.EXPECT().DropEquipmentData(ctx1, &equipment.DropEquipmentDataRequest{Scope: "s1"}).Return(&equipment.DropEquipmentDataResponse{Success: true}, nil).Times(1)
				mockEquipClient.EXPECT().DropMetaData(ctx, &equipment.DropMetaDataRequest{Scope: "s1"}).Return(&equipment.DropMetaDataResponse{Success: true}, nil).Times(1)

				mockDpsClient := dpsMock.NewMockDpsServiceClient(mockCtrl)
				dps = mockDpsClient
				mockDpsClient.EXPECT().DropUploadedFileData(ctx, &dpsv1.DropUploadedFileDataRequest{Scope: "s1"}).Return(&dpsv1.DropUploadedFileDataResponse{Success: true}, nil).Times(1)

				mockMetricClient := metricMock.NewMockMetricServiceClient(mockCtrl)
				metric = mockMetricClient
				mockMetricClient.EXPECT().DropMetricData(ctx, &metricv1.DropMetricDataRequest{Scope: "s1"}).Return(&metricv1.DropMetricDataResponse{Success: true}, nil).Times(1)

				mockReportClient := reportMock.NewMockReportServiceClient(mockCtrl)
				report = mockReportClient
				mockReportClient.EXPECT().DropReportData(ctx, &reportv1.DropReportDataRequest{Scope: "s1"}).Return(&reportv1.DropReportDataResponse{Success: true}, nil).Times(1)

			},
			wantErr: true,
		},
		{
			name:  "Success Deletion",
			input: &v1.DropScopeDataRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().DropScopeTX(ctx, "s1").Return(nil).Times(1)

				mockAppClient := appMock.NewMockApplicationServiceClient(mockCtrl)
				app = mockAppClient
				mockAppClient.EXPECT().DropApplicationData(ctx, &application.DropApplicationDataRequest{Scope: "s1"}).Return(&application.DropApplicationDataResponse{Success: true}, nil).Times(1)
				mockAppClient.EXPECT().DropObscolenscenceData(ctx, &application.DropObscolenscenceDataRequest{Scope: "s1"}).Return(&application.DropObscolenscenceDataResponse{Success: true}, nil).Times(1)

				mockProdClient := prodMock.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockProdClient.EXPECT().DropProductData(ctx, &product.DropProductDataRequest{Scope: "s1"}).Return(&product.DropProductDataResponse{Success: true}, nil).Times(1)

				mockEquipClient := equipmentMock.NewMockEquipmentServiceClient(mockCtrl)
				equip = mockEquipClient
				ctx1, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*300))
				defer cancel()
				mockEquipClient.EXPECT().DropEquipmentData(ctx1, &equipment.DropEquipmentDataRequest{Scope: "s1"}).Return(&equipment.DropEquipmentDataResponse{Success: true}, nil).Times(1)
				mockEquipClient.EXPECT().DropMetaData(ctx, &equipment.DropMetaDataRequest{Scope: "s1"}).Return(&equipment.DropMetaDataResponse{Success: true}, nil).Times(1)

				mockDpsClient := dpsMock.NewMockDpsServiceClient(mockCtrl)
				dps = mockDpsClient
				mockDpsClient.EXPECT().DropUploadedFileData(ctx, &dpsv1.DropUploadedFileDataRequest{Scope: "s1"}).Return(&dpsv1.DropUploadedFileDataResponse{Success: true}, nil).Times(1)

				mockMetricClient := metricMock.NewMockMetricServiceClient(mockCtrl)
				metric = mockMetricClient
				mockMetricClient.EXPECT().DropMetricData(ctx, &metricv1.DropMetricDataRequest{Scope: "s1"}).Return(&metricv1.DropMetricDataResponse{Success: true}, nil).Times(1)

				mockReportClient := reportMock.NewMockReportServiceClient(mockCtrl)
				report = mockReportClient
				mockReportClient.EXPECT().DropReportData(ctx, &reportv1.DropReportDataRequest{Scope: "s1"}).Return(&reportv1.DropReportDataResponse{Success: true}, nil).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			obj := &accountServiceServer{
				accountRepo: rep,
				application: app,
				product:     prod,
				metric:      metric,
				equipment:   equip,
				report:      report,
				dps:         dps,
			}
			_, err := obj.DropScopeData(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.DropScopeData() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else {
				log.Println(tt.name, " passed")
			}
		})
	}
}

func Test_accountServiceServer_UpdateAccount(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account

	type args struct {
		ctx context.Context
		req *v1.UpdateAccountRequest
	}

	tests := []struct {
		name    string
		args    args
		s       *accountServiceServer
		want    *v1.UpdateAccountResponse
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - personal information",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:     "admin@test.com",
						FirstName:  "admin1",
						LastName:   "admin",
						Locale:     "en",
						ProfilePic: "profilepic1",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:     "admin@test.com",
					FirstName:  "admin2",
					LastName:   "user",
					Locale:     "fr",
					ProfilePic: []byte("profilepic"),
				}, nil)
				mockRepo.EXPECT().UpdateAccount(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", &repv1.UpdateAccount{
					FirstName:  "admin1",
					LastName:   "admin",
					Locale:     "en",
					ProfilePic: []byte("profilepic1"),
				}).Times(1).Return(nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: true,
			},
		},
		{name: "SUCCESS - role superadmin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
				mockRepo.EXPECT().UpdateUserAccount(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com", &repv1.UpdateUserAccount{
					Role: repv1.RoleAdmin,
				}).Times(1).Return(nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: true,
			},
		},
		{name: "SUCCESS - role admin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId: "admin1@test.com",
						Role:   v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin3@test.com",
					FirstName: "admin3",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", "admin1@test.com").Times(1).Return(true, nil)
				mockRepo.EXPECT().UpdateUserAccount(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin3@test.com", &repv1.UpdateUserAccount{
					Role: repv1.RoleAdmin,
				}).Times(1).Return(nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: true,
			},
		},
		{name: "FAILURE - UpdateAccount - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId: "admin1@test.com",
						Role:   v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - user does not exist",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com").Times(1).Return(nil, repv1.ErrNoData)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to get Account info",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com").Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - personal information|failed to update account",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:     "admin@test.com",
					FirstName:  "admin1",
					LastName:   "user",
					Locale:     "fr",
					ProfilePic: []byte("profilepic"),
				}, nil)
				mockRepo.EXPECT().UpdateAccount(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", &repv1.UpdateAccount{
					FirstName:  "admin1",
					LastName:   "admin",
					Locale:     "en",
					ProfilePic: []byte("profilepic"),
				}).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - user does not have the access to update other users",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "User",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "User",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to validate update account request|undefined role",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_UNDEFINED,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to validate update account request|can not update role of superadmin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleSuperAdmin,
				}, nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to validate update account request|can not update role to superadmin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_SUPER_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleAdmin,
				}, nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to update account",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "admin",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin1@test.com",
					FirstName: "admin",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleAdmin,
				}, nil)
				mockRepo.EXPECT().UpdateUserAccount(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com", &repv1.UpdateUserAccount{
					Role: repv1.RoleUser,
				}).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - failed to check if user belongs to the admin groups",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId: "admin1@test.com",
						Role:   v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin3@test.com",
					FirstName: "admin3",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", "admin1@test.com").Times(1).Return(false, errors.New("Internal"))
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - UpdateAccount - user does not belong to admin's group",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.UpdateAccountRequest{
					Account: &v1.UpdateAccount{
						UserId: "admin1@test.com",
						Role:   v1.ROLE_ADMIN,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:    "admin3@test.com",
					FirstName: "admin3",
					LastName:  "user",
					Locale:    "fr",
					Role:      repv1.RoleUser,
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", "admin1@test.com").Times(1).Return(false, nil)
			},
			want: &v1.UpdateAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.UpdateAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.UpdateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				return
			}
		})
	}
}

func Test_accountServiceServer_DeleteAccount(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "Admin",
	})
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	type args struct {
		ctx context.Context
		req *v1.DeleteAccountRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.DeleteAccountResponse
		wantErr bool
	}{
		{name: "SUCCESS - role superadmin",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleAdmin,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().InsertUserAudit(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), db.InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleAdmin.RoleToRoleString(),
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       db.AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				}).Times(1).Return(nil)
				mockRepo.EXPECT().DeleteUser(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(nil)
			},
			want: &v1.DeleteAccountResponse{
				Success: true,
			},
		},
		{name: "SUCCESS - role admin",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(ctx, "admin@test.com", "admin1@test.com").Times(1).Return(true, nil)
				mockRepo.EXPECT().InsertUserAudit(ctx, db.InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser.RoleToRoleString(),
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       db.AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				}).Times(1).Return(nil)
				mockRepo.EXPECT().DeleteUser(ctx, "admin1@test.com").Times(1).Return(nil)
			},
			want: &v1.DeleteAccountResponse{
				Success: true,
			},
		},
		{name: "FAILURE - DeleteAccount - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount - AccountInfo - user does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin1@test.com").Times(1).Return(nil, repv1.ErrNoData)
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount - AccountInfo - failed to get Account info",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin1@test.com").Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount - UserBelongsToAdminGroup - failed to check if user belongs to the admin groups",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(ctx, "admin@test.com", "admin1@test.com").Times(1).Return(false, errors.New("Internal"))
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount - UserBelongsToAdminGroup - user does not belong to admin's group",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().UserBelongsToAdminGroup(ctx, "admin@test.com", "admin1@test.com").Times(1).Return(false, nil)
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount -  InsertUserAudit - DBError",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().InsertUserAudit(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), db.InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser.RoleToRoleString(),
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       db.AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				}).Times(1).Return(errors.New("DBError"))
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAccount - DeleteAccount - DBError",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteAccountRequest{
					UserId: "admin1@test.com",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(&repv1.AccountInfo{
					UserID:          "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser,
					Locale:          "en",
					ContFailedLogin: int16(3),
				}, nil)
				mockRepo.EXPECT().InsertUserAudit(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), db.InsertUserAuditParams{
					Username:        "admin1@test.com",
					FirstName:       "admin1",
					LastName:        "test",
					Role:            repv1.RoleUser.RoleToRoleString(),
					Locale:          "en",
					ContFailedLogin: int16(3),
					Operation:       db.AuditStatusDELETED,
					UpdatedBy:       "admin@test.com",
				}).Times(1).Return(nil)
				mockRepo.EXPECT().DeleteUser(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "SuperAdmin",
				}), "admin1@test.com").Times(1).Return(errors.New("DBError"))
			},
			want: &v1.DeleteAccountResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.DeleteAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.DeleteAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_accountServiceServer_GetAccount(t *testing.T) {
	type args struct {
		ctx context.Context
		req *v1.GetAccountRequest
	}
	ctx := context.Background()
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	profilePic := "base64encoded"
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.GetAccountResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
				}),
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
				}), "admin@superuser.com").Times(1).Return(&repv1.AccountInfo{
					UserID:     "admin@superuser.com",
					FirstName:  "first",
					LastName:   "last",
					Role:       repv1.Role(1),
					Locale:     "fr",
					ProfilePic: []byte(profilePic),
					FirstLogin: true,
				}, nil)
			},
			want: &v1.GetAccountResponse{
				UserId:     "admin@superuser.com",
				FirstName:  "first",
				LastName:   "last",
				Role:       v1.ROLE(1),
				Locale:     "fr",
				ProfilePic: profilePic,
				FirstLogin: true,
			},
		},
		{name: "FAILURE - GetAccount - failed to get Account info",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
				}),
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
				}), "admin@superuser.com").Times(1).Return(nil, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure - can not retrieve claims",
			args: args{
				ctx: ctx,
			},
			setup:   func() {},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.GetAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.GetAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got) {
				return
			}
			if tt.setup == nil {
				mockCtrl.Finish()
			}
		})
	}
}

func Test_accountServiceServer_CreateAccount(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	type args struct {
		ctx context.Context
		req *v1.Account
	}
	mockCtrl = gomock.NewController(t)
	mockRepo := mock.NewMockAccount(mockCtrl)
	rep = mockRepo
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   claims.RoleAdmin,
	})
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.Account
		wantErr bool
	}{
		{name: "success",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
				gomock.InOrder(
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(6), gomock.Any()).Return(nil, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(3), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
					}, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(2), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 3,
						},
						{
							ID: 6,
						},
					}, nil),
					mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", gomock.Any()).Return(4, []*repv1.Group{
						{
							ID: 2,
						},
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 5,
						},
						{
							ID: 6,
						},
						{
							ID: 3,
						},
					}, nil),
					mockRepo.EXPECT().CreateAccount(ctx, &repv1.AccountInfo{
						UserID:    "user@test.com",
						FirstName: "abc",
						LastName:  "xyz",
						Password:  defaultPassHash,
						Locale:    "en",
						Role:      repv1.RoleAdmin,
						Group:     []int64{2},
					}),
				)
			},
			want: &v1.Account{
				UserId:    "user@test.com",
				FirstName: "abc",
				LastName:  "xyz",
				Locale:    "en",
				Role:      v1.ROLE_ADMIN,
				Groups:    []int64{2},
			},
		},
		{name: "failure - cannot find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure - only admin users can create users",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@user.com",
					Role:   claims.RoleUser,
				}),
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure - user already exists",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(true, nil)
			},
			wantErr: true,
		},
		{name: "failure - userExists db function failed",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(true, errors.New("test error"))
			},
			wantErr: true,
		},
		{name: "failure - first name should be non-empty",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
			},
			wantErr: true,
		},
		{name: "failure - last name should be non-empty",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
			},
			wantErr: true,
		},
		{name: "failure - locale name should be non-empty",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
			},
			wantErr: true,
		},
		{name: "failure - only admin and user roles are allowed",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_UNDEFINED,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
			},
			wantErr: true,
		},
		{name: "failure - CreateAccount - GetRootGroup - cannot get root group",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 4},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "failure - no user allowed to create in root group",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 1, 4},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "failure - highestAscendants - cannot create account",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
				gomock.InOrder(
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(6), gomock.Any()).Return(nil, errors.New("")),
				)
			},
			wantErr: true,
		},
		{name: "failure - UserOwnedGroups - cannot create user account",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
				gomock.InOrder(
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(6), gomock.Any()).Return(nil, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(3), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
					}, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(2), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 3,
						},
						{
							ID: 6,
						},
					}, nil),
					mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", gomock.Any()).Return(0, nil, errors.New("")),
				)
			},
			wantErr: true,
		},
		{name: "failure - cannot create user account group: groups not owned by user",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
				gomock.InOrder(
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(6), gomock.Any()).Return(nil, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(3), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
					}, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(2), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 3,
						},
						{
							ID: 6,
						},
					}, nil),
					mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", gomock.Any()).Return(4, []*repv1.Group{
						{
							ID: 1,
						},
						{
							ID: 7,
						},
					}, nil),
				)
			},
			wantErr: true,
		},
		{name: "failure - CreateAccount - cannot create user account",
			args: args{
				ctx: ctx,
				req: &v1.Account{
					UserId:    "user@test.com",
					FirstName: "abc",
					LastName:  "xyz",
					Locale:    "en",
					Role:      v1.ROLE_ADMIN,
					Groups:    []int64{6, 3, 2, 4, 5},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserExistsByID(ctx, "user@test.com").Return(false, nil)
				mockRepo.EXPECT().GetRootGroup(ctx).Return(&repv1.Group{ID: 1}, nil).Times(1)
				gomock.InOrder(
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(6), gomock.Any()).Return(nil, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(3), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
					}, nil),
					mockRepo.EXPECT().ChildGroupsAll(ctx, int64(2), gomock.Any()).Return([]*repv1.Group{
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 3,
						},
						{
							ID: 6,
						},
					}, nil),
					mockRepo.EXPECT().UserOwnedGroups(ctx, "admin@superuser.com", gomock.Any()).Return(4, []*repv1.Group{
						{
							ID: 2,
						},
						{
							ID: 4,
						},
						{
							ID: 5,
						},
						{
							ID: 5,
						},
						{
							ID: 6,
						},
						{
							ID: 3,
						},
					}, nil),
					mockRepo.EXPECT().CreateAccount(ctx, &repv1.AccountInfo{
						UserID:    "user@test.com",
						FirstName: "abc",
						LastName:  "xyz",
						Password:  defaultPassHash,
						Locale:    "en",
						Role:      repv1.RoleAdmin,
						Group:     []int64{2},
					}).Return(errors.New("")),
				)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.CreateAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.CreateAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_GetUsers(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@test.com",
		Role:   "SuperAdmin",
	})
	var mockCtrl *gomock.Controller
	var rep repv1.Account

	type args struct {
		ctx context.Context
		req *v1.GetUsersRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListUsersResponse
		wantErr bool
	}{
		{name: "SUCCESS - get all users",
			args: args{
				ctx: ctx,
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: true,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UsersAll(ctx, "admin@test.com").Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "SUCCESS - get list of users role superadmin",
			args: args{
				ctx: ctx,
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: false,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UsersAll(ctx, "admin@test.com").Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "SUCCESS - get list of users",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: false,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UsersWithUserSearchParams(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", &repv1.UserQueryParams{}).Return([]*repv1.AccountInfo{
					{
						UserID:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "user",
						Locale:    "en",
						GroupName: []string{"A"},
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "admin2@test.com",
						FirstName: "admin2",
						LastName:  "user",
						Locale:    "en",
						GroupName: []string{"A", "B"},
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "admin3@test.com",
						FirstName: "admin3",
						LastName:  "user",
						Locale:    "en",
						GroupName: []string{"B", "C"},
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "admin1@test.com",
						FirstName: "admin1",
						LastName:  "user",
						Locale:    "en",
						Groups:    []string{"A"},
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "admin2@test.com",
						FirstName: "admin2",
						LastName:  "user",
						Locale:    "en",
						Groups:    []string{"A", "B"},
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "admin3@test.com",
						FirstName: "admin3",
						LastName:  "user",
						Locale:    "en",
						Groups:    []string{"B", "C"},
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "FAILURE - GetUsers - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: true,
					},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetUsers - failed to get  all users",
			args: args{
				ctx: ctx,
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: true,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UsersAll(ctx, "admin@test.com").Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetUsers - failed to get list of users",
			args: args{
				ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}),
				req: &v1.GetUsersRequest{
					UserFilter: &v1.UserQueryParams{
						AllUsers: false,
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UsersWithUserSearchParams(grpc_middleware.AddClaims(context.Background(), &claims.Claims{
					UserID: "admin@test.com",
					Role:   "Admin",
				}), "admin@test.com", &repv1.UserQueryParams{}).Return(nil, errors.New("Internal")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.GetUsers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.GetUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUsers(t, "GetUsers", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_GetGroupUsers(t *testing.T) {
	ctx := context.Background()
	var mockCtrl *gomock.Controller
	var rep repv1.Account

	type args struct {
		ctx context.Context
		req *v1.GetGroupUsersRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListUsersResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.GetGroupUsersRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroups(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", nil).Return(2, []*repv1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(2)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "FAILURE - GetGroupUsers - can not retrieve claims",
			args: args{
				ctx: ctx,
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetGroupUsers - failed to get groups",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.GetGroupUsersRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroups(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", nil).Return(0, nil, errors.New("")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetGroupUsers - user does not have access to group",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.GetGroupUsersRequest{
					GroupId: 1,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroups(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", nil).Return(2, []*repv1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - GetGroupUsers - failed to get users",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.GetGroupUsersRequest{
					GroupId: 2,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnedGroups(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", nil).Return(2, []*repv1.Group{
					{
						ID:                 2,
						Name:               "OLS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OLS",
						Scopes:             []string{"A", "B"},
					},
					{
						ID:                 3,
						Name:               "OFS",
						ParentID:           1,
						FullyQualifiedName: "Orange.OBS.OFS",
						Scopes:             []string{"C", "D"},
					},
				}, nil).Times(1)
				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(2)).Return(nil, errors.New("")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.GetGroupUsers(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.GetGroupUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUsers(t, "GetUsers", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_AddGroupUser(t *testing.T) {
	ctx := context.Background()
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	type args struct {
		ctx context.Context
		req *v1.AddGroupUsersRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListUsersResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				gomock.InOrder(
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u1", int64(1)).Return(true, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u2", int64(1)).Return(false, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u3", int64(1)).Return(false, nil),
				)
				mockRepo.EXPECT().AddGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u2", "u3"}).Return(nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: ctx,
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetUsers - user doesnot have access to add users",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
				}),
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - user doesnt own group",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(false, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot access userOwnsGroupByID",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, errors.New("Test Error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot access userOwnsGroupByID of given users",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				gomock.InOrder(
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u1", int64(1)).Return(true, errors.New("Test error")),
				)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot add user",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				gomock.InOrder(
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u1", int64(1)).Return(true, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u2", int64(1)).Return(false, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u3", int64(1)).Return(false, nil),
				)
				mockRepo.EXPECT().AddGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u2", "u3"}).Return(errors.New("Test Error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch group users",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.AddGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				gomock.InOrder(
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u1", int64(1)).Return(true, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u2", int64(1)).Return(false, nil),
					mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
						UserID: "admin@superuser.com",
						Role:   "SuperAdmin",
					}), "u3", int64(1)).Return(false, nil),
				)
				mockRepo.EXPECT().AddGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u2", "u3"}).Return(nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(nil, errors.New("Test Error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.AddGroupUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.AddGroupUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUsers(t, "GetUsers", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_DeleteGroupUser(t *testing.T) {
	ctx := context.Background()
	var mockCtrl *gomock.Controller
	var rep repv1.Account
	type args struct {
		ctx context.Context
		req *v1.DeleteGroupUsersRequest
	}
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ListUsersResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u4", "u5"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u5",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)

				mockRepo.EXPECT().IsGroupRoot(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(true, nil)

				mockRepo.EXPECT().DeleteGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u4", "u5"}).Return(nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
			},
			want: &v1.ListUsersResponse{
				Users: []*v1.User{
					{
						UserId:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      v1.ROLE_ADMIN,
					},
					{
						UserId:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      v1.ROLE_USER,
					},
				},
			},
		},
		{name: "FAILURE - can not retrieve claims",
			args: args{
				ctx: ctx,
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetUsers - user doesnot have access to delete users",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "User",
				}),
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - user doesnt own group",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(false, nil)
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot access userOwnsGroupByID",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2", "u3"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, errors.New("Test Error"))
			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch groups of user",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u4", "u5"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(nil, errors.New("Test Error")).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - user doesnt exists",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u4", "u5"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot delete all admins of a root group",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u5",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)

				mockRepo.EXPECT().IsGroupRoot(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(true, nil)

			},
			wantErr: true,
		},
		{name: "FAILURE - IsRootGroup returns error",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u1", "u2"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u5",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)

				mockRepo.EXPECT().IsGroupRoot(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(false, errors.New("test error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot delete user",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u4", "u5"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u5",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().IsGroupRoot(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(true, nil)

				mockRepo.EXPECT().DeleteGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u4", "u5"}).Return(errors.New("Test Error"))

			},
			wantErr: true,
		},
		{name: "FAILURE - cannot fetch groups",
			args: args{
				ctx: grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}),
				req: &v1.DeleteGroupUsersRequest{
					GroupId: 1,
					UserId:  []string{"u4", "u5"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().UserOwnsGroupByID(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), "admin@superuser.com", int64(1)).Return(true, nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return([]*repv1.AccountInfo{
					{
						UserID:    "u1",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u2",
						FirstName: "first",
						LastName:  "last",
						Locale:    "fr",
						Role:      repv1.RoleAdmin,
					},
					{
						UserID:    "u3",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u4",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
					{
						UserID:    "u5",
						FirstName: "first",
						LastName:  "last",
						Locale:    "en",
						Role:      repv1.RoleUser,
					},
				}, nil).Times(1)
				mockRepo.EXPECT().IsGroupRoot(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(true, nil)

				mockRepo.EXPECT().DeleteGroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1), []string{"u4", "u5"}).Return(nil)

				mockRepo.EXPECT().GroupUsers(grpc_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "admin@superuser.com",
					Role:   "SuperAdmin",
				}), int64(1)).Return(nil, errors.New("Test Error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.DeleteGroupUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.DeleteGroupUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUsers(t, "GetUsers", got, tt.want)
			}
		})
	}
}

func Test_accountServiceServer_ChangePassword(t *testing.T) {
	ctx := context.Background()
	clms := &claims.Claims{
		UserID: "admin@superuser.com",
	}
	ctx = grpc_middleware.AddClaims(ctx, clms)
	var mockCtrl *gomock.Controller
	var rep repv1.Account

	type args struct {
		ctx context.Context
		req *v1.ChangePasswordRequest
	}
	abcHash := "$2a$11$m.t5BLK.8wmiPuQzesnaoeyk3EMisi9Q/MmyEbEcaMArNmvtxdi.6"
	tests := []struct {
		name    string
		s       *accountServiceServer
		args    args
		setup   func()
		want    *v1.ChangePasswordResponse
		wantErr bool
	}{
		{name: "SUCCESS - first login true",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					FirstLogin: true,
					Password:   abcHash,
				}, nil).Times(1)
				mockRepo.EXPECT().ChangePassword(ctx, "admin@superuser.com", gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().ChangeUserFirstLogin(ctx, "admin@superuser.com").Times(1).Return(nil)
			},
			want: &v1.ChangePasswordResponse{
				Success: true,
			},
		},
		{name: "SUCCESS - first login false",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					FirstLogin: false,
					Password:   abcHash,
				}, nil).Times(1)
				mockRepo.EXPECT().ChangePassword(ctx, "admin@superuser.com", gomock.Any()).Return(nil).Times(1)
			},
			want: &v1.ChangePasswordResponse{
				Success: true,
			},
		},
		{name: "FAILURE - ChangePassword - cannot retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - failed to get user info",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(nil, errors.New("test.error")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - check password  - fails does not exists in database",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "cde",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - old and new passwords are same",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "abc",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - password  did not validate - must contain a number",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - password  did not validate - must contain an upper case letter",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "xyz1@",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - password  did not validate - must contain a lower case letter",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "XYZ1@",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - password did not validate - invalid special character",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz1!",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
			},
			wantErr: true,
		},
		// {name: "FAILURE - generate hash - failed to generate hash of password",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.ChangePasswordRequest{
		// 			Old: "abc",
		// 			New: "Xyz@123",
		// 		},
		// 	},
		// 	setup: func() {
		// 		mockCtrl = gomock.NewController(t)
		// 		mockRepo := mock.NewMockAccount(mockCtrl)
		// 		rep = mockRepo
		// 		mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
		// 			Password: abcHash,
		// 		}, nil).Times(1)
		// 		mockRepo.EXPECT().ChangePassword(ctx, "admin@superuser.com", "Xyz@123").Return(errors.New("failed to change password")).Times(1)
		// 	},
		// 	wantErr: true,
		// },
		{name: "FAILURE - ChangePassword - failed to change password",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					Password: abcHash,
				}, nil).Times(1)
				mockRepo.EXPECT().ChangePassword(ctx, "admin@superuser.com", gomock.Any()).Return(errors.New("failed to change password")).Times(1)
			},
			wantErr: true,
		},
		{name: "FAILURE - ChangePassword - failed to change first login status",
			args: args{
				ctx: ctx,
				req: &v1.ChangePasswordRequest{
					Old: "abc",
					New: "Xyz@123",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockAccount(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AccountInfo(ctx, "admin@superuser.com").Return(&repv1.AccountInfo{
					FirstLogin: true,
					Password:   abcHash,
				}, nil).Times(1)
				mockRepo.EXPECT().ChangePassword(ctx, "admin@superuser.com", gomock.Any()).Return(nil).Times(1)
				mockRepo.EXPECT().ChangeUserFirstLogin(ctx, "admin@superuser.com").Times(1).Return(errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &accountServiceServer{
				accountRepo: rep,
			}
			got, err := tt.s.ChangePassword(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("accountServiceServer.ChangePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("accountServiceServer.ChangePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareUsers(t *testing.T, name string, exp *v1.ListUsersResponse, act *v1.ListUsersResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	for i := range exp.Users {
		compareUser(t, fmt.Sprintf("%s[%d]", name, i), exp.Users[i], act.Users[i])
	}
}

func compareUser(t *testing.T, name string, exp *v1.User, act *v1.User) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.UserId, act.UserId, "%s.UserId are not same", name)
	assert.Equalf(t, exp.FirstName, act.FirstName, "%s.FirstName are not same", name)
	assert.Equalf(t, exp.LastName, act.LastName, "%s.LastName are not same", name)
	assert.Equalf(t, exp.Locale, act.Locale, "%s.Locale are not same", name)
	assert.Equalf(t, exp.Groups, act.Groups, "%s.Groups are not same", name)
	assert.Equalf(t, exp.Role, act.Role, "%s.Role are not same", name)
}
