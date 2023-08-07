package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	accv1 "optisam-backend/account-service/pkg/api/v1"
	accmock "optisam-backend/account-service/pkg/api/v1/mock"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	customctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1", "s2", "s3"},
	})
)

func Test_OverviewProductQuality(t *testing.T) {

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	//dgObj := dgmock.NewMockProduct(mockCtrl)

	tests := []struct {
		name      string
		ctx       context.Context
		s         *productServiceServer
		input     *v1.OverviewProductQualityRequest
		setup     func(*v1.OverviewProductQualityRequest)
		output    *v1.OverviewProductQualityResponse
		outputErr bool
	}{
		{
			name:  "Success: data_exist_for_scope",
			ctx:   customctx,
			input: &v1.OverviewProductQualityRequest{Scope: "s1"},
			setup: func(s *v1.OverviewProductQualityRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsNotDeployed(ctx, "s1").Times(1).Return([]db.ProductsNotDeployedRow{
					{
						Swidtag:       "PND1",
						ProductName:   "ProNotDep1",
						ProductEditor: "e1",
						Version:       "v1",
					},
					{
						Swidtag:       "PND2",
						ProductName:   "ProNotDep2",
						ProductEditor: "e2",
						Version:       "v2",
					},
				}, nil)
				mockRepository.EXPECT().ProductsNotAcquired(ctx, "s1").Times(1).Return([]db.ProductsNotAcquiredRow{
					{
						Swidtag:        "PNA1",
						ProductName:    "ProNotAcq1",
						ProductEditor:  "e1",
						ProductVersion: "v1",
					},
				}, nil)
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"s1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return([]db.ListProductsViewRow{
					{
						Totalrecords: int64(40),
					},
				}, nil)
			},
			output: &v1.OverviewProductQualityResponse{
				NotAcquiredProducts:           int32(1),
				NotDeployedProducts:           int32(2),
				NotAcquiredProductsPercentage: float64(2.5),
				NotDeployedProductsPercentage: float64(5.0),
			},
		},
		{
			name:  "Failed: data_doesnot_exist_for_scope",
			ctx:   customctx,
			input: &v1.OverviewProductQualityRequest{Scope: "s1"},
			setup: func(s *v1.OverviewProductQualityRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsNotDeployed(ctx, "s1").Times(1).Return([]db.ProductsNotDeployedRow{}, nil)
				mockRepository.EXPECT().ProductsNotAcquired(ctx, "s1").Times(1).Return([]db.ProductsNotAcquiredRow{}, nil)
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"s1"},
					PageNum:  0,
					PageSize: 1,
				}).Times(1).Return([]db.ListProductsViewRow{}, nil)
			},
			output: &v1.OverviewProductQualityResponse{},
		},
		{
			name:  "Failed : Scope not exist",
			ctx:   customctx,
			input: &v1.OverviewProductQualityRequest{Scope: "s4"},
			setup: func(s *v1.OverviewProductQualityRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsNotDeployed(ctx, s.Scope).Times(1).Return([]db.ProductsNotDeployedRow{}, nil)
			},
			outputErr: true,
		},
		{
			name:  "Failed : context not exist",
			input: &v1.OverviewProductQualityRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func(s *v1.OverviewProductQualityRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				//mockDgraph := dgmock.NewMockProduct(mockCtrl)

				rep = mockRepository
				queue = mockQueue

				mockRepository.EXPECT().ProductsNotDeployed(ctx, s.Scope).Times(1).Return([]db.ProductsNotDeployedRow{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.OverviewProductQuality(tt.ctx, tt.input)
			if (err != nil) != tt.outputErr {
				t.Errorf("productServiceServer.OverviewProductQuality() error = %v, wantErr %v", err, tt.outputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("productServiceServer.OverviewProductQuality() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_SoftwareExpenditureByScope(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3", "OSN", "OFR"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var acc accv1.AccountServiceClient

	type args struct {
		ctx context.Context
		req *v1.SoftwareExpenditureByScopeRequest
	}
	tests := []struct {
		name    string
		lr      *productServiceServer
		args    args
		setup   func()
		want    *v1.SoftwareExpenditureByScopeResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.SoftwareExpenditureByScopeRequest{
					Scope: []string{"OSN", "OFR"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				MockAccount := accmock.NewMockAccountServiceClient(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				acc = MockAccount
				MockAccount.EXPECT().ListScopes(ctx, &accv1.ListScopesRequest{}).Times(1).Return(&accv1.ListScopesResponse{
					Scopes: []*accv1.Scope{
						{
							ScopeCode:   "OSN",
							ScopeName:   "osn123",
							ScopeType:   "osn type",
							CreatedBy:   "abcd@gmail.com",
							CreatedOn:   &tspb.Timestamp{Seconds: 10},
							Expenditure: 10,
							GroupNames:  []string{"G1", "G2"},
						}, {
							ScopeCode:   "OFR",
							ScopeName:   "ofr123",
							ScopeType:   "ofr type",
							CreatedBy:   "zxcv@gmail.com",
							CreatedOn:   &tspb.Timestamp{Seconds: 14},
							Expenditure: 20,
							GroupNames:  []string{"G1", "G2"},
						}, {
							ScopeCode:   "INM",
							ScopeName:   "inm123",
							ScopeType:   "inm type",
							CreatedBy:   "asdf@gmail.com",
							CreatedOn:   &tspb.Timestamp{Seconds: 15},
							Expenditure: 8,
							GroupNames:  []string{"G1"},
						},
					}}, nil)
				mockRepository.EXPECT().TotalCostOfEachScope(ctx, []string{"OSN", "OFR"}).Times(1).Return([]db.TotalCostOfEachScopeRow{
					{
						Scope:     "OSN",
						TotalCost: decimal.New(50, 0),
					}, {
						Scope:     "OFR",
						TotalCost: decimal.New(60, 0),
					},
				}, nil)
			},
			want: &v1.SoftwareExpenditureByScopeResponse{
				ExpensePercent: []*v1.SoftwareExpensePercent{
					{
						Scope:              "OSN",
						Expenditure:        float64(10),
						TotalCost:          float64(50),
						ExpenditurePercent: float64(20),
					}, {
						Scope:              "OFR",
						Expenditure:        float64(20),
						TotalCost:          float64(60),
						ExpenditurePercent: float64(33.333335876464844),
					},
				},
				TotalExpenditure: float64(30),
				TotalCost:        110,
			},
		},
		{
			name: "SoftwareExpenditureByScope without context",
			args: args{
				ctx: context.Background(),
				req: &v1.SoftwareExpenditureByScopeRequest{
					Scope: []string{"OSN", "OFR"},
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "SoftwareExpenditureByScope ListScopes - ServiceError",
			args: args{
				ctx: context.Background(),
				req: &v1.SoftwareExpenditureByScopeRequest{
					Scope: []string{"OSN", "OFR"},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				MockAccount := accmock.NewMockAccountServiceClient(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				acc = MockAccount
				MockAccount.EXPECT().ListScopes(ctx, &accv1.ListScopesRequest{}).Times(1).Return(nil, errors.New("service error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.lr = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				account:     acc,
			}
			got, err := tt.lr.SoftwareExpenditureByScope(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.SoftwareExpenditureByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.SoftwareExpenditureByScope() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetBanner(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	ct := time.Now()
	nt := ct.Add(time.Hour)
	cout := ct.Format("2006-01-02 15:04")
	nout := nt.Format("2006-01-02 15:04")

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.GetBannerRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.GetBannerResponse
		wantErr bool
	}{
		{
			name: "Success",
			args: args{
				ctx: ctx,
				req: &v1.GetBannerRequest{
					Scope:    "Scope1",
					TimeZone: "CET",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().GetDashboardUpdates(ctx, db.GetDashboardUpdatesParams{
					Scope:   "Scope1",
					Column2: "CET",
				}).Return(db.GetDashboardUpdatesRow{
					UpdatedAt:    ct,
					NextUpdateAt: nt,
				}, nil).Times(1)
			},
			want: &v1.GetBannerResponse{
				UpdatedAt:    cout,
				NextUpdateAt: nout,
			},
			wantErr: false,
		},
		{
			name: "DataNotFound",
			args: args{
				ctx: ctx,
				req: &v1.GetBannerRequest{
					Scope:    "Scope2",
					TimeZone: "CET",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().GetDashboardUpdates(ctx, db.GetDashboardUpdatesParams{
					Scope:   "Scope2",
					Column2: "CET",
				}).Return(db.GetDashboardUpdatesRow{}, sql.ErrNoRows).Times(1)
			},
			wantErr: true,
		},
		{
			name: "ScopeNotFound",
			args: args{
				ctx: ctx,
				req: &v1.GetBannerRequest{
					Scope:    "Scope20",
					TimeZone: "CET",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.GetBannerRequest{
					Scope:    "Scope20",
					TimeZone: "CET",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "DBError",
			args: args{
				ctx: ctx,
				req: &v1.GetBannerRequest{
					Scope:    "Scope2",
					TimeZone: "CET",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().GetDashboardUpdates(ctx, db.GetDashboardUpdatesParams{
					Scope:   "Scope2",
					Column2: "CET",
				}).Return(db.GetDashboardUpdatesRow{}, errors.New("DBError")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.GetBanner(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.GetBanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.GetBanner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DashboardOverview(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue

	type args struct {
		ctx context.Context
		req *v1.DashboardOverviewRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DashboardOverviewResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return([]db.ListProductsViewRow{
					{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(10.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:                int32(3),
				NumProducts:               int32(40),
				TotalLicenseCost:          float64(12300),
				TotalMaintenanceCost:      float64(1200),
				TotalCounterfeitingAmount: float64(10.0),
				TotalUnderusageAmount:     float64(20.0),
			},
		},
		{
			name: "FAILURE: Error in db/GetAcqRightsCost",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return([]db.ListProductsViewRow{
					{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{}, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE: No claims Found",
			args: args{
				ctx: context.Background(),
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE: User do not have access to scopes",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE: Error in db/ListProductsView",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE: Error in db/ListEditors",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return([]db.ListProductsViewRow{
					{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS - No Products in the System",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return(nil, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(10.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:                int32(1),
				TotalLicenseCost:          float64(12300),
				TotalMaintenanceCost:      float64(1200),
				TotalCounterfeitingAmount: float64(10.0),
				TotalUnderusageAmount:     float64(20.0),
			},
		},
		{
			name: "SUCCESS - No Editors in the System",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return(nil, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return(nil, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
			},
			want: &v1.DashboardOverviewResponse{},
		},
		{
			name: "SUCCESS -  No costs in the system",
			args: args{
				ctx: ctx,
				req: &v1.DashboardOverviewRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    []string{"Scope1"},
					PageNum:  int32(0),
					PageSize: int32(1),
				}).Times(1).Return([]db.ListProductsViewRow{
					{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{}, sql.ErrNoRows)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:  int32(3),
				NumProducts: int32(40),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.DashboardOverview(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DashboardOverview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DashboardOverview() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_ProductsPerEditor(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.ProductsPerEditorRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ProductsPerEditorResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, nil)
				gomock.InOrder(
					mockRepository.EXPECT().GetProductsByEditor(ctx, db.GetProductsByEditorParams{
						ProductEditor: "e1",
						Scopes:        []string{"Scope1"},
					}).Times(1).Return([]db.GetProductsByEditorRow{
						{
							Swidtag:     "s1",
							ProductName: "p1",
						},
					}, nil),
				)
			},
			want: &v1.ProductsPerEditorResponse{
				EditorsProducts: []*v1.EditorProducts{
					{
						Editor:      "e1",
						NumProducts: int32(1),
					},
				},
			},
		},
		{
			name: "FAILURE: Error in db/GetProductsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, nil)
				gomock.InOrder(
					mockRepository.EXPECT().GetProductsByEditor(ctx, db.GetProductsByEditorParams{
						ProductEditor: "e1",
						Scopes:        []string{"Scope1"},
					}).Times(1).Return(nil, errors.New("Internal")),
				)
			},
			wantErr: true,
		},
		{
			name: "FAILURE: No Claims Found",
			args: args{
				ctx: context.Background(),
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE: Do not have access to scopes",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{
			name: "FAILURE - Error in db/ListEditors",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS - No Editors Found",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerEditorRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)

				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return(nil, nil)

			},
			want: &v1.ProductsPerEditorResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			//	dgObj := dgmock.NewMockProduct(mockCtrl)

			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.ProductsPerEditor(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ProductsPerEditor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.ProductsPerEditor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_acqRightsServiceServer_ProductsPerMetricType(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.ProductsPerMetricTypeRequest
	}
	tests := []struct {
		name    string
		lr      *productServiceServer
		args    args
		setup   func()
		want    *v1.ProductsPerMetricTypeResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerMetricTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsPerMetric(ctx, "Scope1").Times(1).Return([]db.ProductsPerMetricRow{
					{
						Metric:      "OPS",
						Composition: int64(100),
					},
				}, nil)
			},
			want: &v1.ProductsPerMetricTypeResponse{
				MetricsProducts: []*v1.MetricProducts{
					{
						MetricName:  "OPS",
						NumProducts: int32(100),
					},
				},
			},
		},
		{
			name: "SUCCESS : No Result",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerMetricTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsPerMetric(ctx, "Scope1").Times(1).Return(nil, nil)
			},
			want: &v1.ProductsPerMetricTypeResponse{},
		},
		{
			name: "FAILURE: Can not find claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ProductsPerMetricTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {

			},
			wantErr: true,
		},
		{
			name: "FAILURE: User does not have permission to access given scope",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerMetricTypeRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "FAILURE: Error in db/getAcqRightsCost",
			args: args{
				ctx: ctx,
				req: &v1.ProductsPerMetricTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductsPerMetric(ctx, "Scope1").Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := NewProductServiceServer(rep, queue, nil, "")
			got, err := lr.ProductsPerMetricType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.ProductsPerMetricType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.ProductsPerMetricType() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_acqRightsServiceServer_CounterfeitedProducts(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.CounterfeitedProductsRequest
	}
	tests := []struct {
		name    string
		lr      *productServiceServer
		args    args
		setup   func()
		want    *v1.CounterfeitedProductsResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - Both licenses and costs",
			args: args{
				ctx: ctx,
				req: &v1.CounterfeitedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterFeitedProductsLicences(ctx, db.CounterFeitedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsLicencesRow{
					{
						SwidTags:            "p1",
						ProductNames:        "p1n1",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(1000, 0),
						NumAcquiredLicences: decimal.New(100, 0),
						Delta:               decimal.New(-900, 0),
					},
					{
						SwidTags:            "p2",
						ProductNames:        "p2n2",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(1000, 0),
						NumAcquiredLicences: decimal.New(200, 0),
						Delta:               decimal.New(-800, 0),
					},
				}, nil)
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsCostsRow{
					{
						SwidTags:        "p1",
						ProductNames:    "p1n1",
						AggregationName: "agg1",
						PurchaseCost:    decimal.New(100, 0),
						ComputedCost:    decimal.New(1000, 0),
						DeltaCost:       decimal.New(-900, 0),
					},
					{
						SwidTags:        "p2",
						ProductNames:    "p2n2",
						AggregationName: "agg1",
						ComputedCost:    decimal.New(1000, 0),
						PurchaseCost:    decimal.New(200, 0),
						DeltaCost:       decimal.New(-800, 0),
					},
				}, nil)
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(200),
						Delta:               int64(-800),
					},
				},
				ProductsCosts: []*v1.ProductsCosts{
					{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						AggregationName:      "agg1",
						LicensesAcquiredCost: float64(100),
						LicensesComputedCost: float64(1000),
						DeltaCost:            float64(-900),
					},
					{
						SwidTag:              "p2",
						ProductName:          "p2n2",
						AggregationName:      "agg1",
						LicensesComputedCost: float64(1000),
						LicensesAcquiredCost: float64(200),
						DeltaCost:            float64(-800),
					},
				},
			},
		},
		{
			name: "SUCCESS - Only Licences, Error in costs",
			args: args{
				ctx: ctx,
				req: &v1.CounterfeitedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterFeitedProductsLicences(ctx, db.CounterFeitedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsLicencesRow{
					{
						SwidTags:            "p1",
						ProductNames:        "p1n1",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(1000, 0),
						NumAcquiredLicences: decimal.New(100, 0),
						Delta:               decimal.New(-900, 0),
					},
					{
						SwidTags:            "p2",
						ProductNames:        "p2n2",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(1000, 0),
						NumAcquiredLicences: decimal.New(200, 0),
						Delta:               decimal.New(-800, 0),
					},
				}, nil)
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(200),
						Delta:               int64(-800),
					},
				},
			},
		},
		{
			name: "SUCCESS - Only costs, licenses error",
			args: args{
				ctx: ctx,
				req: &v1.CounterfeitedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterFeitedProductsLicences(ctx, db.CounterFeitedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsCostsRow{
					{
						SwidTags:        "p1",
						ProductNames:    "p1n1",
						AggregationName: "agg1",
						PurchaseCost:    decimal.New(100, 0),
						ComputedCost:    decimal.New(1000, 0),
						DeltaCost:       decimal.New(-900, 0),
					},
					{
						SwidTags:        "p2",
						ProductNames:    "p2n2",
						AggregationName: "agg1",
						ComputedCost:    decimal.New(1000, 0),
						PurchaseCost:    decimal.New(200, 0),
						DeltaCost:       decimal.New(-800, 0),
					},
				}, nil)
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsCosts: []*v1.ProductsCosts{
					{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						AggregationName:      "agg1",
						LicensesAcquiredCost: float64(100),
						LicensesComputedCost: float64(1000),
						DeltaCost:            float64(-900),
					},
					{
						SwidTag:              "p2",
						ProductName:          "p2n2",
						AggregationName:      "agg1",
						LicensesComputedCost: float64(1000),
						LicensesAcquiredCost: float64(200),
						DeltaCost:            float64(-800),
					},
				},
			},
		},
		{
			name: "FAILURE: Can not find claims",
			args: args{
				ctx: context.Background(),
				req: &v1.CounterfeitedProductsRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {

			},
			wantErr: true,
		},
		{
			name: "FAILURE: User does not have permission to access given scope",
			args: args{
				ctx: ctx,
				req: &v1.CounterfeitedProductsRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := NewProductServiceServer(rep, queue, nil, "")
			got, err := lr.CounterfeitedProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.CounterfeitedProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.CounterfeitedProducts() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_acqRightsServiceServer_OverdeployedProducts(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue

	type args struct {
		ctx context.Context
		req *v1.OverdeployedProductsRequest
	}
	tests := []struct {
		name    string
		lr      *productServiceServer
		args    args
		setup   func()
		want    *v1.OverdeployedProductsResponse
		wantErr bool
	}{
		{
			name: "SUCCESS - Both licenses and costs",
			args: args{
				ctx: ctx,
				req: &v1.OverdeployedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().OverDeployedProductsLicences(ctx, db.OverDeployedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsLicencesRow{
					{
						SwidTags:            "p1",
						ProductNames:        "p1n1",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(100, 0),
						NumAcquiredLicences: decimal.New(1000, 0),
						Delta:               decimal.New(900, 0),
					},
					{
						SwidTags:            "p2",
						ProductNames:        "p2n2",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(200, 0),
						NumAcquiredLicences: decimal.New(1000, 0),
						Delta:               decimal.New(800, 0),
					},
				}, nil)
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsCostsRow{
					{
						SwidTags:        "p1",
						ProductNames:    "p1n1",
						AggregationName: "agg1",
						PurchaseCost:    decimal.New(1000, 0),
						ComputedCost:    decimal.New(100, 0),
						DeltaCost:       decimal.New(900, 0),
					},
					{
						SwidTags:        "p2",
						ProductNames:    "p2n2",
						AggregationName: "agg1",
						ComputedCost:    decimal.New(200, 0),
						PurchaseCost:    decimal.New(1000, 0),
						DeltaCost:       decimal.New(800, 0),
					},
				}, nil)
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(200),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(800),
					},
				},
				ProductsCosts: []*v1.ProductsCosts{
					{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						AggregationName:      "agg1",
						LicensesAcquiredCost: float64(1000),
						LicensesComputedCost: float64(100),
						DeltaCost:            float64(900),
					},
					{
						SwidTag:              "p2",
						ProductName:          "p2n2",
						AggregationName:      "agg1",
						LicensesComputedCost: float64(200),
						LicensesAcquiredCost: float64(1000),
						DeltaCost:            float64(800),
					},
				},
			},
		},
		{
			name: "SUCCESS - Only Licences, Error in costs",
			args: args{
				ctx: ctx,
				req: &v1.OverdeployedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().OverDeployedProductsLicences(ctx, db.OverDeployedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsLicencesRow{
					{
						SwidTags:            "p1",
						ProductNames:        "p1n1",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(100, 0),
						NumAcquiredLicences: decimal.New(1000, 0),
						Delta:               decimal.New(900, 0),
					},
					{
						SwidTags:            "p2",
						ProductNames:        "p2n2",
						AggregationName:     "agg1",
						NumComputedLicences: decimal.New(200, 0),
						NumAcquiredLicences: decimal.New(1000, 0),
						Delta:               decimal.New(800, 0),
					},
				}, nil)
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						AggregationName:     "agg1",
						NumLicensesComputed: int64(200),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(800),
					},
				},
			},
		},
		{
			name: "SUCCESS - Only costs, licenses error",
			args: args{
				ctx: ctx,
				req: &v1.OverdeployedProductsRequest{
					Scope:  "Scope1",
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().OverDeployedProductsLicences(ctx, db.OverDeployedProductsLicencesParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:  "Scope1",
					Editor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsCostsRow{
					{
						SwidTags:        "p1",
						ProductNames:    "p1n1",
						AggregationName: "agg1",
						PurchaseCost:    decimal.New(1000, 0),
						ComputedCost:    decimal.New(100, 0),
						DeltaCost:       decimal.New(900, 0),
					},
					{
						SwidTags:        "p2",
						ProductNames:    "p2n2",
						AggregationName: "agg1",
						ComputedCost:    decimal.New(200, 0),
						PurchaseCost:    decimal.New(1000, 0),
						DeltaCost:       decimal.New(800, 0),
					},
				}, nil)
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsCosts: []*v1.ProductsCosts{
					{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						AggregationName:      "agg1",
						LicensesAcquiredCost: float64(1000),
						LicensesComputedCost: float64(100),
						DeltaCost:            float64(900),
					},
					{
						SwidTag:              "p2",
						ProductName:          "p2n2",
						AggregationName:      "agg1",
						LicensesComputedCost: float64(200),
						LicensesAcquiredCost: float64(1000),
						DeltaCost:            float64(800),
					},
				},
			},
		},
		{
			name: "FAILURE: Can not find claims",
			args: args{
				ctx: context.Background(),
				req: &v1.OverdeployedProductsRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {

			},
			wantErr: true,
		},
		{
			name: "FAILURE: User does not have permission to access given scope",
			args: args{
				ctx: ctx,
				req: &v1.OverdeployedProductsRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := NewProductServiceServer(rep, queue, nil, "")
			got, err := lr.OverdeployedProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.OverdeployedProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.OverdeployedProducts() got= %v, want= %v", got, tt.want)
			}
		})
	}
}

func Test_acqRightsServiceServer_ComplianceAlert(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.ComplianceAlertRequest
	}
	tests := []struct {
		name    string
		lr      *productServiceServer
		args    args
		setup   func()
		want    *v1.ComplianceAlertResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{
					Acq:         decimal.New(50000, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{
					Acq:         decimal.New(50000, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
			},
			want: &v1.ComplianceAlertResponse{
				CounterfeitingPercentage: float64(0.5),
				OverdeploymentPercentage: float64(0.5),
			},
		},
		{
			name: "FAILURE: overdeployment - Acq is zero",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{
					Acq:         decimal.New(50000, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{
					Acq:         decimal.New(0, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
			},
			want:    &v1.ComplianceAlertResponse{},
			wantErr: false,
		},
		{
			name: "FAILURE- acq is zero - Counterfeit",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{
					Acq:         decimal.New(0, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
			},
			want:    &v1.ComplianceAlertResponse{},
			wantErr: false,
		},
		{
			name: "FAILURE: error in db/CounterfeitPercent",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{}, errors.New("internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE: error in db/CounterfeitPercent - Not Found",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{}, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "FAILURE: error in db/OverdeployPercent",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{
					Acq:         decimal.New(50000, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{}, errors.New("internal"))
			},
			wantErr: true,
		},
		{
			name: "FAILURE: error in db/OverdeployPercent - Not Found",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{
					Acq:         decimal.New(50000, 0),
					DeltaRights: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{}, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "FAILURE: Can not find claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
		{
			name: "FAILURE: Permission Error",
			args: args{
				ctx: ctx,
				req: &v1.ComplianceAlertRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := &productServiceServer{
				productRepo: rep,
				queue:       queue,
			}
			got, err := lr.ComplianceAlert(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.ComplianceAlert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.ComplianceAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DashboardQualityProducts(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DashboardQualityProductsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DashboardQualityProductsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().ProductsNotDeployed(ctx, "Scope1").Times(1).Return([]db.ProductsNotDeployedRow{
					{
						Swidtag:       "PND1",
						ProductName:   "ProNotDep1",
						ProductEditor: "e1",
						Version:       "v1",
					},
					{
						Swidtag:       "PND2",
						ProductName:   "ProNotDep2",
						ProductEditor: "e2",
						Version:       "v2",
					},
				}, nil)
				mockRepo.EXPECT().ProductsNotAcquired(ctx, "Scope1").Times(1).Return([]db.ProductsNotAcquiredRow{
					{
						Swidtag:        "PNA1",
						ProductName:    "ProNotAcq1",
						ProductEditor:  "e1",
						ProductVersion: "v1",
					},
				}, nil)
			},
			want: &v1.DashboardQualityProductsResponse{
				ProductsNotDeployed: []*v1.DashboardQualityProducts{
					{
						SwidTag:     "PND1",
						ProductName: "ProNotDep1",
						Editor:      "e1",
						Version:     "v1",
					},
					{
						SwidTag:     "PND2",
						ProductName: "ProNotDep2",
						Editor:      "e2",
						Version:     "v2",
					},
				},
				ProductsNotAcquired: []*v1.DashboardQualityProducts{
					{
						SwidTag:     "PNA1",
						ProductName: "ProNotAcq1",
						Editor:      "e1",
						Version:     "v1",
					},
				},
			},
		},
		{name: "FAILURE - productServiceServer/DashboardQuality - ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - productServiceServer/DashboardQuality - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - productServiceServer/DashboardQuality - db/ProductsNotDeployedCount - DBError",
			args: args{
				ctx: ctx,
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().ProductsNotDeployed(ctx, "Scope1").Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE - productServiceServer/DashboardQuality - db/ProductsNotAcquiredCount - DBError",
			args: args{
				ctx: ctx,
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().ProductsNotDeployed(ctx, "Scope1").Times(1).Return([]db.ProductsNotDeployedRow{
					{
						Swidtag:     "PND1",
						ProductName: "ProNotDep1",
					},
					{
						Swidtag:     "PND2",
						ProductName: "ProNotDep2",
					},
				}, nil)
				mockRepo.EXPECT().ProductsNotAcquired(ctx, "Scope1").Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.DashboardQualityProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DashboardQuality() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DashboardQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_CreateDashboardUpdateJob(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	tests := []struct {
		name    string
		s       *productServiceServer
		input   *v1.CreateDashboardUpdateJobRequest
		want    *v1.CreateDashboardUpdateJobResponse
		wantErr bool
		ctx     context.Context
		setup   func()
	}{
		{
			name:    "SucessfullyJobCreated",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope1"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: true},
			wantErr: false,
			ctx:     ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(fmt.Sprintf(`{"updatedBy":"data_update" , "scope" :"%s"}`, "Scope1")),
				}, "lcalw").Return(int32(0), nil).Times(1)
			},
		},
		{
			name:    "FailedInJobCreation",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope1"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: false},
			wantErr: true,
			ctx:     ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "lcalw"},
					Status: job.JobStatusPENDING,
					Data:   json.RawMessage(fmt.Sprintf(`{"updatedBy":"data_update" , "scope" :"%s"}`, "Scope1")),
				}, "lcalw").Return(int32(0), errors.New("JobFailed")).Times(1)
			},
		},
		{
			name:    "ContextNotFound",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope1"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: false},
			wantErr: true,
			ctx:     context.Background(),
			setup:   func() {},
		},
		{
			name:    "ScopeNotFound",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope11"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: false},
			wantErr: true,
			ctx:     context.Background(),
			setup:   func() {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.CreateDashboardUpdateJob(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.CreateDashboardUpdateJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.CreateDashboardUpdateJob() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProductListByEditorRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GetProductListByEditorRequest
		output *v1.GetProductListByEditorResponse
		mock   func(*v1.GetProductListByEditorRequest)
		ctx    context.Context
		outErr bool
	}{

		{
			name:   "GetProductListByEditorRequest without context",
			input:  &v1.GetProductListByEditorRequest{Scopes: []string{"AAK", "MON"}, Editor: "Oracle"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GetProductListByEditorRequest) {},
		},
		{
			name:  "GetProductListByEditor with data",
			input: &v1.GetProductListByEditorRequest{Scopes: []string{"AAK", "MON"}, Editor: "Oracle"},
			ctx:   ctx,
			mock: func(data *v1.GetProductListByEditorRequest) {
				dbObj.EXPECT().GetProductListByEditor(ctx, db.GetProductListByEditorParams{
					Scope:  []string{"AAK", "MON"},
					Editor: "Oracle",
				}).Return([]string{
					"Oracle 1",
					"Oracle 2",
					"Oracle 3",
				}, nil).Times(1)
			},
			output: &v1.GetProductListByEditorResponse{
				Products: []string{
					"Oracle 1",
					"Oracle 2",
					"Oracle 3",
				},
			},
			outErr: false,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.GetProductListByEditor(test.ctx, test.input)
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

func TestGroupComplianceProductRequest(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GroupComplianceProductRequest
		output *v1.GroupComplianceProductResponse
		mock   func(*v1.GroupComplianceProductRequest)
		ctx    context.Context
		outErr bool
	}{

		{
			name:   "GroupComplianceProductRequest without context",
			input:  &v1.GroupComplianceProductRequest{Scopes: []string{"AAK", "MON"}, Editor: "Oracle", ProductName: "Oracle Enterprise Database 10"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GroupComplianceProductRequest) {},
		},
		{
			name:  "GroupComplianceProduct with data",
			input: &v1.GroupComplianceProductRequest{Scopes: []string{"AAK", "MON"}, Editor: "Oracle", ProductName: "Oracle Enterprise Database 10"},
			ctx:   ctx,
			mock: func(data *v1.GroupComplianceProductRequest) {
				dbObj.EXPECT().GetOverallLicencesByProduct(ctx, db.GetOverallLicencesByProductParams{
					Scope:       []string{"AAK", "MON"},
					Editor:      "Oracle",
					ProductName: "Oracle Enterprise Database 10",
				}).Return([]db.GetOverallLicencesByProductRow{
					{
						Scope:            "AAK",
						ComputedLicences: decimal.NewFromFloat(20),
						AcquiredLicences: decimal.NewFromFloat(50),
					},
					{
						Scope:            "MON",
						ComputedLicences: decimal.NewFromFloat(30),
						AcquiredLicences: decimal.NewFromFloat(40),
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetOverallCostByProduct(ctx, db.GetOverallCostByProductParams{
					Scope:       []string{"AAK", "MON"},
					Editor:      "Oracle",
					ProductName: "Oracle Enterprise Database 10",
				}).Return([]db.GetOverallCostByProductRow{
					{
						Scope:              "AAK",
						CounterfeitingCost: decimal.NewFromFloat(100),
						UnderusageCost:     decimal.NewFromFloat(0),
						TotalCost:          decimal.NewFromFloat(100),
					},
					{
						Scope:              "MON",
						CounterfeitingCost: decimal.NewFromFloat(0),
						UnderusageCost:     decimal.NewFromFloat(250),
						TotalCost:          decimal.NewFromFloat(250),
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetTotalCostByProduct(ctx, db.GetTotalCostByProductParams{
					Scope:       []string{"AAK", "MON"},
					Editor:      "Oracle",
					ProductName: "Oracle Enterprise Database 10",
				}).Return([]db.GetTotalCostByProductRow{
					{
						Scope:     "AAK",
						TotalCost: float64(100),
					},
					{
						Scope:     "MON",
						TotalCost: float64(250),
					},
				}, nil).Times(1)
			},
			output: &v1.GroupComplianceProductResponse{
				Licences: []*v1.LicencesData{
					{
						Scope:            "AAK",
						ComputedLicences: 20,
						AcquiredLicences: 50,
					},
					{
						Scope:            "MON",
						ComputedLicences: 30,
						AcquiredLicences: 40,
					},
				},
				Cost: []*v1.CostData{
					{
						Scope:              "AAK",
						CounterfeitingCost: 100,
						UnderusageCost:     0,
						TotalCost:          100,
					},
					{
						Scope:              "MON",
						CounterfeitingCost: 0,
						UnderusageCost:     250,
						TotalCost:          250,
					},
				},
			},
			outErr: false,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.GroupComplianceProduct(test.ctx, test.input)
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

func Test_productServiceServer_GetUnderusageLicenceByEditorProduct(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})

	ctxAdmin := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "user@optisam.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.GetUnderusageByEditorRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.GetUnderusageByEditorResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetUnderusageByEditorRequest{
					Scopes: []string{"Scope1", "Scope2", "Scope3"},
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListUnderusageByEditor(ctx, db.ListUnderusageByEditorParams{
					Scope:          []string{"Scope1", "Scope2", "Scope3"},
					Editor:         "Oracle",
					LkProductNames: false,
					ProductNames:   "",
				}).Times(1).Return([]db.ListUnderusageByEditorRow{
					{
						Metrics: "INS",
						Scope:   "Scope1",
						Delta:   decimal.New(200, 0),
					},
				}, nil)
			},
			want: &v1.GetUnderusageByEditorResponse{
				UnderusageByEditorData: []*v1.UnderusageByEditorData{
					{
						Metrics:     "INS",
						Scope:       "Scope1",
						DeltaNumber: 200,
					},
				},
			},
		},
		{name: "FAILURE - productServiceServer/GetUnderusageLicenceByEditorProduct - ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.GetUnderusageByEditorRequest{
					Scopes: []string{"Scope1", "Scope2", "Scope3"},
					Editor: "Oracle",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - productServiceServer/GetUnderusageLicenceByEditorProduct - RoleValidationError",
			args: args{
				ctx: ctxAdmin,
				req: &v1.GetUnderusageByEditorRequest{
					Scopes: []string{"Scope1", "Scope2", "Scope3"},
					Editor: "Oracle",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		// {name: "FAILURE - productServiceServer/DashboardQuality - ScopeValidationError",
		// 	args: args{
		// 		ctx: ctx,
		// 		req: &v1.GetUnderusageByEditorRequest{
		// 			Scopes: []string{"Scope4"},
		// 			Editor: "Oracle",
		// 		},
		// 	},
		// 	setup:   func() {},
		// 	wantErr: true,
		// },
		{name: "FAILURE - productServiceServer/DashboardQuality - db/ProductsNotDeployedCount - DBError",
			args: args{
				ctx: ctx,
				req: &v1.GetUnderusageByEditorRequest{
					Scopes: []string{"Scope1", "Scope2", "Scope3"},
					Editor: "Oracle",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().ListUnderusageByEditor(ctx, db.ListUnderusageByEditorParams{
					Scope:          []string{"Scope1", "Scope2", "Scope3"},
					Editor:         "Oracle",
					LkProductNames: false,
					ProductNames:   "",
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "")
			got, err := s.GetUnderusageLicenceByEditorProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DashboardQuality() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DashboardQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}
