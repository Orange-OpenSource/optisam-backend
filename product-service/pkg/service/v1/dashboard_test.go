package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	accv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/account-service/pkg/api/v1"
	accmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/account-service/pkg/api/v1/mock"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

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

func TestGetTotalSharedAmount(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	mockCtrl := gomock.NewController(t)
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.GetTotalSharedAmountRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetTotalSharedAmountResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetTotalSharedAmountRequest{
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				met = mockMetric
				mockRepo.EXPECT().GetSharedData(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{SharedLicences: int32(1), RecievedLicences: int32(1)}}, nil)
				mockRepo.EXPECT().GetUnitPriceBySku(ctx, gomock.Any()).Times(1).Return(db.GetUnitPriceBySkuRow{}, nil)
			},
			want:    &v1.GetTotalSharedAmountResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			_, err := tt.s.GetTotalSharedAmount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetEditorExpensesByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_OverviewProductQuality(t *testing.T) {

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	//dgObj := dgmock.NewMockProduct(mockCtrl)

	tests := []struct {
		name      string
		ctx       context.Context
		s         *ProductServiceServer
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
				mockRepository.EXPECT().ProductsNotDeployed(ctx, "s4").AnyTimes().Return([]db.ProductsNotDeployedRow{}, nil)
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

				mockRepository.EXPECT().ProductsNotDeployed(ctx, "s1").AnyTimes().Return([]db.ProductsNotDeployedRow{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.OverviewProductQuality(tt.ctx, tt.input)
			if (err != nil) != tt.outputErr {
				t.Errorf("ProductServiceServer.OverviewProductQuality() error = %v, wantErr %v", err, tt.outputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ProductServiceServer.OverviewProductQuality() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_ProductMaintenancePerc(t *testing.T) {

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	//dgObj := dgmock.NewMockProduct(mockCtrl)

	tests := []struct {
		name      string
		ctx       context.Context
		s         *ProductServiceServer
		input     *v1.ProductMaintenancePercRequest
		setup     func()
		output    *v1.ProductMaintenancePercResponse
		outputErr bool
	}{
		{
			name:  "Success: data_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductMaintenancePercRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductMaintenanceCount(ctx, "s1").Times(1).Return([]db.ProductMaintenanceCountRow{
					{
						NumberOfSwidtag: 1,
						Total:           10,
					},
				}, nil)
			},
			output: &v1.ProductMaintenancePercResponse{
				ProductWithMaintenancePercentage:    float64(10),
				ProductWithoutMaintenancePercentage: float64(90),
			},
		},
		{
			name:  "Success: No data in Scope",
			ctx:   customctx,
			input: &v1.ProductMaintenancePercRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductMaintenanceCount(ctx, "s1").Times(1).Return([]db.ProductMaintenanceCountRow{
					{
						NumberOfSwidtag: 0,
						Total:           0,
					},
				}, nil)
			},
			output: &v1.ProductMaintenancePercResponse{
				ProductWithMaintenancePercentage:    float64(0),
				ProductWithoutMaintenancePercentage: float64(0),
			},
		},
		{
			name:  "Failed: data_doesnot_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductMaintenancePercRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductMaintenanceCount(ctx, "s1").Times(1).Return(nil, errors.New("No data"))
			},
			outputErr: true,
		},
		{
			name:  "Failed : Scope not exist",
			ctx:   customctx,
			input: &v1.ProductMaintenancePercRequest{Scope: "s4"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().ProductMaintenanceCount(ctx, "s4").AnyTimes().Return([]db.ProductMaintenanceCountRow{}, nil)
			},
			outputErr: true,
		},
		{
			name:  "Failed : context not exist",
			input: &v1.ProductMaintenancePercRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				//mockDgraph := dgmock.NewMockProduct(mockCtrl)

				rep = mockRepository
				queue = mockQueue

				// mockRepository.EXPECT().ProductMaintenanceCount(ctx, "s1").Times(1).Return([]db.ProductMaintenanceCountRow{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.ProductMaintenancePerc(tt.ctx, tt.input)
			if (err != nil) != tt.outputErr {
				t.Errorf("ProductServiceServer.ProductMaintenancePerc() error = %v, wantErr %v", err, tt.outputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ProductServiceServer.ProductMaintenancePerc() = %v, want %v", got, tt.output)
			}
		})
	}
}

func Test_ProductNoMaintenanceDetails(t *testing.T) {
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	//dgObj := dgmock.NewMockProduct(mockCtrl)

	tests := []struct {
		name      string
		ctx       context.Context
		s         *ProductServiceServer
		input     *v1.ProductNoMaintenanceDetailsRequest
		setup     func()
		output    *v1.ProductNoMaintenanceDetailsResponse
		outputErr bool
	}{
		{
			name:  "Success: data_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s1").Times(1).Return([]string{"Acrobat_Capture_A_1", "Acrobat_Capture_A_2", "Acrobat_A_1"}, nil)
				mockRepository.EXPECT().ProductNoMaintenance(ctx, db.ProductNoMaintenanceParams{
					Swidtag: []string{"Acrobat_Capture_A_1", "Acrobat_Capture_A_2", "Acrobat_A_1"},
					Scope:   "s1",
				}).Times(1).Return([]db.ProductNoMaintenanceRow{
					{
						Swidtag:     "Acrobat_Capture_A_1",
						ProductName: "Acrobat Capture",
					},
					{
						Swidtag:     "Acrobat_Capture_A_2",
						ProductName: "Acrobat",
					},
					{
						Swidtag:     "Acrobat_A_1",
						ProductName: "",
					},
				}, nil)
				mockRepository.EXPECT().ProductCatalogVersion(ctx, db.ProductCatalogVersionParams{
					Swidtag: []string{"Acrobat_Capture_A_1", "Acrobat_Capture_A_2", "Acrobat_A_1"},
					Scope:   "s1",
				}).Times(1).Return([]db.ProductCatalogVersionRow{
					{
						Swidtag: "Acrobat_Capture_A_2",
						Version: "v2",
					},
				}, nil)
			},
			output: &v1.ProductNoMaintenanceDetailsResponse{
				TotalProducts: 3,
				ProductNoMain: []*v1.ProductNoMain{
					{
						ProductName: "Acrobat Capture",
						Version:     "",
						Swidtag:     "Acrobat_Capture_A_1",
					},
					{
						ProductName: "Acrobat Capture",
						Version:     "v2",
						Swidtag:     "Acrobat_Capture_A_2",
					},
					{
						ProductName: "Acrobat",
						Version:     "",
						Swidtag:     "Acrobat_A_1",
					},
				},
			},
		},
		{
			name:  "Failed: data_doesnot_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s1").Times(1).Return([]string{"PNA1", "PNA2"}, nil)
				mockRepository.EXPECT().ProductNoMaintenance(ctx, db.ProductNoMaintenanceParams{
					Swidtag: []string{"PNA1", "PNA2"},
					Scope:   "s1",
				}).Times(1).Return(nil, errors.New("No data"))
			},
			outputErr: true,
		},
		{
			name:  "FAIL: DB-error",
			ctx:   customctx,
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s1").Times(1).Return([]string{"PNA1", "PNA2"}, nil)
				mockRepository.EXPECT().ProductNoMaintenance(ctx, db.ProductNoMaintenanceParams{
					Swidtag: []string{"PNA1", "PNA2"},
					Scope:   "s1",
				}).Times(1).Return([]db.ProductNoMaintenanceRow{
					{
						Swidtag:     "PNA1",
						ProductName: "P1",
					},
					{
						Swidtag:     "PNA2",
						ProductName: "P2",
					},
				}, nil)
				mockRepository.EXPECT().ProductCatalogVersion(ctx, db.ProductCatalogVersionParams{
					Swidtag: []string{"PNA1", "PNA2"},
					Scope:   "s1",
				}).AnyTimes().Return(nil, errors.New("No data"))
			},
			outputErr: true,
		},
		{
			name:  "Failed : Scope not exist",
			ctx:   customctx,
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s4"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				//	mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s4").Times(1).Return([]string{}, nil)
			},
			outputErr: true,
		},
		{
			name:  "Failed : Swidtag query fail",
			ctx:   customctx,
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s1").Times(1).Return(nil, errors.New("No swidtag data"))
			},
			outputErr: true,
		},
		{
			name:  "Failed : context not exist",
			input: &v1.ProductNoMaintenanceDetailsRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				//mockDgraph := dgmock.NewMockProduct(mockCtrl)

				rep = mockRepository
				queue = mockQueue

				//	mockRepository.EXPECT().AllNoMaintainenceProducts(ctx, "s1").Times(1).Return([]string{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.ProductNoMaintenanceDetails(tt.ctx, tt.input)
			if (err != nil) != tt.outputErr {
				t.Errorf("ProductServiceServer.ProductNoMaintenanceDetails() error = %v, wantErr %v", err, tt.outputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ProductServiceServer.ProductNoMaintenanceDetails() = %v, want %v", got, tt.output)
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
		lr      *ProductServiceServer
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
				MockAccount.EXPECT().ListScopes(ctx, &accv1.ListScopesRequest{}).AnyTimes().Return(nil, errors.New("service error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.lr = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				account:     acc,
			}
			_, err := tt.lr.SoftwareExpenditureByScope(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.SoftwareExpenditureByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ProductServiceServer.SoftwareExpenditureByScope() = %v, want %v", got, tt.want)
			// }
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
		s       *ProductServiceServer
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
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.GetBanner(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetBanner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.GetBanner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_DashboardOverview(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(10.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
				mockRepository.EXPECT().GetTotalDeltaCost(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
				mockRepository.EXPECT().GetComputedCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetComputedCostRow{
					TotalCost:    decimal.New(123, 2),
					PurchaseCost: decimal.New(12, 2),
				}, nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:                 int32(3),
				NumProducts:                int32(40),
				TotalLicenseCost:           float64(12300),
				TotalMaintenanceCost:       float64(1200),
				TotalCounterfeitingAmount:  float64(10.0),
				TotalUnderusageAmount:      float64(40.0),
				ComputedMaintenance:        float64(12300),
				ComputedWithoutMaintenance: float64(1200),
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(10.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
				mockRepository.EXPECT().GetTotalDeltaCost(ctx, "Scope1").Times(1).Return(float64(20.0), nil)
				mockRepository.EXPECT().GetComputedCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetComputedCostRow{
					TotalCost:    decimal.New(123, 2),
					PurchaseCost: decimal.New(12, 2),
				}, nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:                 int32(1),
				TotalLicenseCost:           float64(12300),
				TotalMaintenanceCost:       float64(1200),
				TotalCounterfeitingAmount:  float64(10.0),
				TotalUnderusageAmount:      float64(40.0),
				ComputedMaintenance:        float64(12300),
				ComputedWithoutMaintenance: float64(1200),
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return(nil, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{}, nil)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalDeltaCost(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetComputedCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetComputedCostRow{}, nil)
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetLicensesCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetLicensesCostRow{}, sql.ErrNoRows)
				mockRepository.EXPECT().GetTotalCounterfietAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalUnderusageAmount(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetTotalDeltaCost(ctx, "Scope1").Times(1).Return(float64(0.0), nil)
				mockRepository.EXPECT().GetComputedCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetComputedCostRow{}, nil)
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
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.DashboardOverview(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DashboardOverview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DashboardOverview() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_ProductsPerEditor(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, nil)
				gomock.InOrder(
					mockRepository.EXPECT().GetProductsByEditor(ctx, db.GetProductsByEditorParams{
						ProductEditor: "e1",
						Scopes:        []string{"Scope1"},
					}).AnyTimes().Return([]db.GetProductsByEditorRow{
						{
							Swidtag:     "s1",
							ProductName: "p1",
						},
					}, nil),
				)
				mockRepository.EXPECT().GetProductsByEditorScope(ctx, gomock.Any()).Times(1).Return([]db.GetProductsByEditorScopeRow{}, nil)
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1"}, errors.New("internal error"))
				gomock.InOrder(
					mockRepository.EXPECT().GetProductsByEditor(ctx, db.GetProductsByEditorParams{
						ProductEditor: "e1",
						Scopes:        []string{"Scope1"},
					}).AnyTimes().Return([]db.GetProductsByEditorRow{
						{
							Swidtag:     "s1",
							ProductName: "p1",
						},
					}, errors.New("Internal")),
				)
				mockRepository.EXPECT().GetProductsByEditorScope(ctx, gomock.Any()).AnyTimes().Return([]db.GetProductsByEditorScopeRow{}, nil)
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
				mockRepository.EXPECT().ListEditorsScope(ctx, gomock.Any()).Times(1).Return(nil, errors.New("Internal"))
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
				mockRepository.EXPECT().ListEditorsScope(ctx, []string{"Scope1"}).Times(1).Return(nil, nil)

			},
			want: &v1.ProductsPerEditorResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			//	dgObj := dgmock.NewMockProduct(mockCtrl)

			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			_, err := s.ProductsPerEditor(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.ProductsPerEditor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ProductServiceServer.ProductsPerEditor() = %v, want %v", got, tt.want)
			// }
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
		lr      *ProductServiceServer
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
			lr := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
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
		lr      *ProductServiceServer
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
			lr := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
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
		lr      *ProductServiceServer
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
			lr := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
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
		lr      *ProductServiceServer
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
			lr := &ProductServiceServer{
				ProductRepo: rep,
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

func Test_ProductServiceServer_DashboardQualityProducts(t *testing.T) {
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
		s       *ProductServiceServer
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
		{name: "FAILURE - ProductServiceServer/DashboardQuality - ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ProductServiceServer/DashboardQuality - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DashboardQualityProductsRequest{
					Scope: "Scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ProductServiceServer/DashboardQuality - db/ProductsNotDeployedCount - DBError",
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
		{name: "FAILURE - ProductServiceServer/DashboardQuality - db/ProductsNotAcquiredCount - DBError",
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
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.DashboardQualityProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DashboardQuality() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DashboardQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_CreateDashboardUpdateJob(t *testing.T) {
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
		s       *ProductServiceServer
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
			name:    "SucessfullyJobCreated ppid not blank",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope1", Ppid: "job1"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: true},
			wantErr: false,
			ctx:     ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				f := mockRepository.EXPECT().GetJobsInExecution(gomock.Any(), gomock.Any()).Return(int64(5), nil)
				mockRepository.EXPECT().GetJobsInExecution(gomock.Any(), gomock.Any()).Return(int64(0), nil).AnyTimes().After(f)
				mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "lcalw").Return(int32(0), nil).AnyTimes()
			},
		},
		{
			name:    "error ppid not blank",
			input:   &v1.CreateDashboardUpdateJobRequest{Scope: "Scope1", Ppid: "job1"},
			want:    &v1.CreateDashboardUpdateJobResponse{Success: false},
			wantErr: true,
			ctx:     ctx,
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				f := mockRepository.EXPECT().GetJobsInExecution(gomock.Any(), gomock.Any()).Return(int64(5), nil)
				mockRepository.EXPECT().GetJobsInExecution(gomock.Any(), gomock.Any()).Return(int64(0), errors.New("err")).AnyTimes().After(f)
				// mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "lcalw").Return(int32(0), nil).AnyTimes()
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
			ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
				UserID: "admin@superuser.com",
				Role:   "SuperAdmin",
				Socpes: []string{},
			}),
			setup: func() {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.CreateDashboardUpdateJob(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.CreateDashboardUpdateJob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.CreateDashboardUpdateJob() = %v, want %v", got, tt.want)
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
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
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
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
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

func Test_ProductServiceServer_GetUnderusageLicenceByEditorProduct(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepo.EXPECT().ListUnderusageByEditor(ctx, gomock.Any()).Times(1).Return([]db.ListUnderusageByEditorRow{
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
		{name: "FAILURE - ProductServiceServer/GetUnderusageLicenceByEditorProduct - ClaimsNotFoundError",
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
		{name: "FAILURE - ProductServiceServer/GetUnderusageLicenceByEditorProduct - RoleValidationError",
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
		// {name: "FAILURE - ProductServiceServer/DashboardQuality - ScopeValidationError",
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
		{name: "FAILURE - ProductServiceServer/DashboardQuality - db/ProductsNotDeployedCount - DBError",
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
				mockRepo.EXPECT().ListUnderusageByEditor(ctx, gomock.Any()).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			_, err := s.GetUnderusageLicenceByEditorProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DashboardQuality() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ProductServiceServer.DashboardQuality() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_Dashboard_ProductLocationType(t *testing.T) {
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
		req *v1.GetDeploymentTypeRequest
	}
	tests := []struct {
		name    string
		lr      *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetDeploymentTypeResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetDeploymentTypeRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().DeploymentPercent(ctx, db.DeploymentPercentParams{
					Scope:       "Scope1",
					ProductType: db.ProductTypeONPREMISE,
				}).Times(1).Return(float64(30), nil)
				mockRepository.EXPECT().DeploymentPercent(ctx, db.DeploymentPercentParams{
					Scope:       "Scope1",
					ProductType: db.ProductTypeSAAS,
				}).Times(1).Return(float64(10), nil)
			},
			want: &v1.GetDeploymentTypeResponse{
				SaasPercentage:      float64(25),
				OnPremisePercentage: float64(75),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := lr.ProductLocationType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dashboard.ProductLocationType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dashboard.ProductLocationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Dashboard_GetWasteUpLicences(t *testing.T) {
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
		req *v1.GetWasteUpLicencesRequest
	}
	tests := []struct {
		name    string
		lr      *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetWasteUpLicencesResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetWasteUpLicencesRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().WasteCost(ctx, "Scope1").Times(1).Return([]db.WasteCostRow{
					{
						Editor:       "Oracle",
						ProductNames: "Oracle Enterprice DB 9",
						Cost:         100,
					},
				}, nil)
			},
			want: &v1.GetWasteUpLicencesResponse{
				TotalWasteUpCost: 100,
				EditorsWasteUpCost: []*v1.EditorsWasteCost{
					{
						Editor:     "Oracle",
						EditorCost: 100,
						ProductsWasteUpCost: []*v1.ProductsCost{
							{
								Product:         "Oracle Enterprice DB 9",
								ProductCost:     100,
								AggregationName: "",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := lr.GetWasteUpLicences(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dashboard.ProductLocationType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dashboard.ProductLocationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Dashboard_GetTrueUpLicences(t *testing.T) {
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
		req *v1.GetTrueUpLicencesRequest
	}
	tests := []struct {
		name    string
		lr      *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetTrueUpLicencesResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetTrueUpLicencesRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().TrueCost(ctx, "Scope1").Times(1).Return([]db.TrueCostRow{
					{
						Editor:       "Oracle",
						ProductNames: "Oracle Enterprice DB 9",
						Cost:         -100,
					},
				}, nil)
			},
			want: &v1.GetTrueUpLicencesResponse{
				TotalTrueUpCost: -100,
				EditorsTrueUpCost: []*v1.EditorsCost{
					{
						Editor:     "Oracle",
						EditorCost: -100,
						ProductsTrueUpCost: []*v1.ProductsCost{
							{
								Product:         "Oracle Enterprice DB 9",
								ProductCost:     -100,
								AggregationName: "",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := lr.GetTrueUpLicences(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dashboard.ProductLocationType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dashboard.ProductLocationType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductsPercOpenClosedSource(t *testing.T) {

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	//dgObj := dgmock.NewMockProduct(mockCtrl)

	tests := []struct {
		name      string
		ctx       context.Context
		s         *ProductServiceServer
		input     *v1.ProductsPercOpenClosedSourceRequest
		setup     func()
		output    *v1.ProductsPercOpenClosedSourceResponse
		outputErr bool
	}{
		{
			name:  "Success: data_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductsPercOpenClosedSourceRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().TotalProductsOfScope(ctx, "s1").Times(1).Return([]int32{100}, nil)
				mockRepository.EXPECT().GetOpenSourceCloseSourceData(ctx, "s1").Times(1).Return([]db.GetOpenSourceCloseSourceDataRow{
					{
						Oscount: 10,
						Cscount: 10,
					},
				}, nil)
			},
			output: &v1.ProductsPercOpenClosedSourceResponse{
				OpenSource: []*v1.OpenSource{
					{
						AmountOs:     10,
						PrecentageOs: 10,
					},
				},
				ClosedSource: []*v1.CloseSource{
					{
						AmountCs:     10,
						PrecentageCs: 10,
					},
				},
				TotalAmount: []*v1.TotalProductData{
					{
						Precentage: 80,
						Amount:     80,
					},
				},
			},
		},
		{
			name:  "Failed: data_doesnot_exist_for_scope",
			ctx:   customctx,
			input: &v1.ProductsPercOpenClosedSourceRequest{Scope: "s1"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().TotalProductsOfScope(ctx, "s1").Times(1).Return([]int32{100}, nil)
				mockRepository.EXPECT().GetOpenSourceCloseSourceData(ctx, "s1").Times(1).Return(nil, errors.New("No data"))
			},
			outputErr: true,
		},
		{
			name:  "Failed : Scope not exist",
			ctx:   customctx,
			input: &v1.ProductsPercOpenClosedSourceRequest{Scope: "s4"},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().TotalProductsOfScope(ctx, "s4").AnyTimes().Return([]int32{}, nil)
			},
			outputErr: true,
		},
		{
			name:  "Failed : context not exist",
			input: &v1.ProductsPercOpenClosedSourceRequest{Scope: "s1"},
			ctx:   context.Background(),
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				//mockDgraph := dgmock.NewMockProduct(mockCtrl)

				rep = mockRepository
				queue = mockQueue

				mockRepository.EXPECT().TotalProductsOfScope(ctx, "s1").AnyTimes().Return([]int32{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := NewProductServiceServer(rep, queue, nil, "", nil, nil, &config.Config{})
			got, err := s.ProductsPercOpenClosedSource(tt.ctx, tt.input)
			if (err != nil) != tt.outputErr {
				t.Errorf("ProductServiceServer.ProductsPercOpenClosedSource() error = %v, wantErr %v", err, tt.outputErr)
				return
			}
			if !reflect.DeepEqual(got, tt.output) {
				t.Errorf("ProductServiceServer.ProductsPercOpenClosedSource() = %v, want %v", got, tt.output)
			}
		})
	}
}
