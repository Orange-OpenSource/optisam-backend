// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

var (
	customctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1", "s2", "s3"},
	})
)

func Test_OverviewProeuctQuality(t *testing.T) {

	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
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
				mockRepository.EXPECT().GetProductQualityOverview(ctx, s.Scope).Times(1).Return(db.GetProductQualityOverviewRow{
					TotalRecords:          int64(4),
					NotDeployed:           int64(1),
					NotAcquired:           int64(1),
					NotDeployedPercentage: decimal.New(25, 0),
					NotAcquiredPercentage: decimal.New(25, 0),
				}, nil)
			},
			output: &v1.OverviewProductQualityResponse{
				NotAcquiredProducts:           int32(1),
				NotDeployedProducts:           int32(1),
				NotAcquiredProductsPercentage: float64(25),
				NotDeployedProductsPercentage: float64(25),
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
				mockRepository.EXPECT().GetProductQualityOverview(ctx, s.Scope).Times(1).Return(db.GetProductQualityOverviewRow{}, nil)
			},
			outputErr: true,
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
				mockRepository.EXPECT().GetProductQualityOverview(ctx, s.Scope).Times(1).Return(db.GetProductQualityOverviewRow{}, nil)
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
				rep = mockRepository
				queue = mockQueue
				mockRepository.EXPECT().GetProductQualityOverview(ctx, s.Scope).Times(1).Return(db.GetProductQualityOverviewRow{}, nil)
			},
			outputErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			s := NewProductServiceServer(rep, queue)
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
					db.ListProductsViewRow{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetAcqRightsCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetAcqRightsCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:           int32(3),
				NumProducts:          int32(40),
				TotalLicenseCost:     float64(12300),
				TotalMaintenanceCost: float64(1200),
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
					db.ListProductsViewRow{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetAcqRightsCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetAcqRightsCostRow{}, errors.New("Internal"))
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
					db.ListProductsViewRow{
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
				mockRepository.EXPECT().GetAcqRightsCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetAcqRightsCostRow{
					TotalCost:            decimal.New(123, 2),
					TotalMaintenanceCost: decimal.New(12, 2),
				}, nil)
			},
			want: &v1.DashboardOverviewResponse{
				NumEditors:           int32(1),
				TotalLicenseCost:     float64(12300),
				TotalMaintenanceCost: float64(1200),
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
				mockRepository.EXPECT().GetAcqRightsCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetAcqRightsCostRow{}, nil)
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
					db.ListProductsViewRow{
						Totalrecords: int64(40),
					},
				}, nil)
				mockRepository.EXPECT().ListEditors(ctx, []string{"Scope1"}).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
				mockRepository.EXPECT().GetAcqRightsCost(ctx, []string{"Scope1"}).Times(1).Return(db.GetAcqRightsCostRow{}, sql.ErrNoRows)
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
			s := NewProductServiceServer(rep, queue)
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
						db.GetProductsByEditorRow{
							Swidtag:     "s1",
							ProductName: "p1",
						},
					}, nil),
				)
			},
			want: &v1.ProductsPerEditorResponse{
				EditorsProducts: []*v1.EditorProducts{
					&v1.EditorProducts{
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
			s := NewProductServiceServer(rep, queue)
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
				mockRepository.EXPECT().ProductsPerMetric(ctx, []string{"Scope1"}).Times(1).Return([]db.ProductsPerMetricRow{
					db.ProductsPerMetricRow{
						Metric:      "OPS",
						NumProducts: int64(100),
					},
				}, nil)
			},
			want: &v1.ProductsPerMetricTypeResponse{
				MetricsProducts: []*v1.MetricProducts{
					&v1.MetricProducts{
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
				mockRepository.EXPECT().ProductsPerMetric(ctx, []string{"Scope1"}).Times(1).Return(nil, nil)
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
				mockRepository.EXPECT().ProductsPerMetric(ctx, []string{"Scope1"}).Times(1).Return(nil, errors.New("Internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			lr := NewProductServiceServer(rep, queue)
			got, err := lr.ProductsPerMetricType(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.ProductsPerMetricType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.ProductsPerMetricType() = %v, want %v", got, tt.want)
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsLicencesRow{
					db.CounterFeitedProductsLicencesRow{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicencesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					db.CounterFeitedProductsLicencesRow{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicencesComputed: int64(1000),
						NumLicensesAcquired: int64(200),
						Delta:               int64(-800),
					},
				}, nil)
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsCostsRow{
					db.CounterFeitedProductsCostsRow{
						SwidTag:           "p1",
						ProductName:       "p1n1",
						TotalPurchaseCost: decimal.New(100, 0),
						TotalComputedCost: decimal.New(1000, 0),
						DeltaCost:         decimal.New(-900, 0),
					},
					db.CounterFeitedProductsCostsRow{
						SwidTag:           "p2",
						ProductName:       "p2n2",
						TotalComputedCost: decimal.New(1000, 0),
						TotalPurchaseCost: decimal.New(200, 0),
						DeltaCost:         decimal.New(-800, 0),
					},
				}, nil)
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					&v1.ProductsLicenses{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					&v1.ProductsLicenses{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(200),
						Delta:               int64(-800),
					},
				},
				ProductsCosts: []*v1.ProductsCosts{
					&v1.ProductsCosts{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						LicensesAcquiredCost: float64(100),
						LicensesComputedCost: float64(1000),
						DeltaCost:            float64(-900),
					},
					&v1.ProductsCosts{
						SwidTag:              "p2",
						ProductName:          "p2n2",
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsLicencesRow{
					db.CounterFeitedProductsLicencesRow{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicencesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					db.CounterFeitedProductsLicencesRow{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicencesComputed: int64(1000),
						NumLicensesAcquired: int64(200),
						Delta:               int64(-800),
					},
				}, nil)
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					&v1.ProductsLicenses{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicensesComputed: int64(1000),
						NumLicensesAcquired: int64(100),
						Delta:               int64(-900),
					},
					&v1.ProductsLicenses{
						SwidTag:             "p2",
						ProductName:         "p2n2",
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
				mockRepository.EXPECT().CounterFeitedProductsCosts(ctx, db.CounterFeitedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.CounterFeitedProductsCostsRow{
					db.CounterFeitedProductsCostsRow{
						SwidTag:           "p1",
						ProductName:       "p1n1",
						TotalPurchaseCost: decimal.New(100, 0),
						TotalComputedCost: decimal.New(1000, 0),
						DeltaCost:         decimal.New(-900, 0),
					},
					db.CounterFeitedProductsCostsRow{
						SwidTag:           "p2",
						ProductName:       "p2n2",
						TotalComputedCost: decimal.New(1000, 0),
						TotalPurchaseCost: decimal.New(200, 0),
						DeltaCost:         decimal.New(-800, 0),
					},
				}, nil)
			},
			want: &v1.CounterfeitedProductsResponse{
				ProductsCosts: []*v1.ProductsCosts{
					&v1.ProductsCosts{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						LicensesAcquiredCost: float64(100),
						LicensesComputedCost: float64(1000),
						DeltaCost:            float64(-900),
					},
					&v1.ProductsCosts{
						SwidTag:              "p2",
						ProductName:          "p2n2",
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
			lr := NewProductServiceServer(rep, queue)
			got, err := lr.CounterfeitedProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.CounterfeitedProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.CounterfeitedProducts() = %v, want %v", got, tt.want)
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsLicencesRow{
					db.OverDeployedProductsLicencesRow{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicencesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					db.OverDeployedProductsLicencesRow{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicencesComputed: int64(200),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(800),
					},
				}, nil)
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsCostsRow{
					db.OverDeployedProductsCostsRow{
						SwidTag:           "p1",
						ProductName:       "p1n1",
						TotalPurchaseCost: decimal.New(1000, 0),
						TotalComputedCost: decimal.New(100, 0),
						DeltaCost:         decimal.New(900, 0),
					},
					db.OverDeployedProductsCostsRow{
						SwidTag:           "p2",
						ProductName:       "p2n2",
						TotalComputedCost: decimal.New(200, 0),
						TotalPurchaseCost: decimal.New(1000, 0),
						DeltaCost:         decimal.New(800, 0),
					},
				}, nil)
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					&v1.ProductsLicenses{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicensesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					&v1.ProductsLicenses{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicensesComputed: int64(200),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(800),
					},
				},
				ProductsCosts: []*v1.ProductsCosts{
					&v1.ProductsCosts{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						LicensesAcquiredCost: float64(1000),
						LicensesComputedCost: float64(100),
						DeltaCost:            float64(900),
					},
					&v1.ProductsCosts{
						SwidTag:              "p2",
						ProductName:          "p2n2",
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsLicencesRow{
					db.OverDeployedProductsLicencesRow{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicencesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					db.OverDeployedProductsLicencesRow{
						SwidTag:             "p2",
						ProductName:         "p2n2",
						NumLicencesComputed: int64(200),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(800),
					},
				}, nil)
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsLicenses: []*v1.ProductsLicenses{
					&v1.ProductsLicenses{
						SwidTag:             "p1",
						ProductName:         "p1n1",
						NumLicensesComputed: int64(100),
						NumLicensesAcquired: int64(1000),
						Delta:               int64(900),
					},
					&v1.ProductsLicenses{
						SwidTag:             "p2",
						ProductName:         "p2n2",
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
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return(nil, errors.New("Internal"))
				mockRepository.EXPECT().OverDeployedProductsCosts(ctx, db.OverDeployedProductsCostsParams{
					Scope:         "Scope1",
					ProductEditor: "Oracle",
				}).Times(1).Return([]db.OverDeployedProductsCostsRow{
					db.OverDeployedProductsCostsRow{
						SwidTag:           "p1",
						ProductName:       "p1n1",
						TotalPurchaseCost: decimal.New(1000, 0),
						TotalComputedCost: decimal.New(100, 0),
						DeltaCost:         decimal.New(900, 0),
					},
					db.OverDeployedProductsCostsRow{
						SwidTag:           "p2",
						ProductName:       "p2n2",
						TotalComputedCost: decimal.New(200, 0),
						TotalPurchaseCost: decimal.New(1000, 0),
						DeltaCost:         decimal.New(800, 0),
					},
				}, nil)
			},
			want: &v1.OverdeployedProductsResponse{
				ProductsCosts: []*v1.ProductsCosts{
					&v1.ProductsCosts{
						SwidTag:              "p1",
						ProductName:          "p1n1",
						LicensesAcquiredCost: float64(1000),
						LicensesComputedCost: float64(100),
						DeltaCost:            float64(900),
					},
					&v1.ProductsCosts{
						SwidTag:              "p2",
						ProductName:          "p2n2",
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
			lr := NewProductServiceServer(rep, queue)
			got, err := lr.OverdeployedProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("acqRightsServiceServer.OverdeployedProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("acqRightsServiceServer.OverdeployedProducts() = %v, want %v", got, tt.want)
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
					Tpc:       decimal.New(50000, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{
					Tpc:       decimal.New(50000, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)
			},
			want: &v1.ComplianceAlertResponse{
				CounterfeitingPercentage: float64(1),
				OverdeploymentPercentage: float64(1),
			},
		},
		{
			name: "FAILURE: overdeployment - tpc is zero",
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
					Tpc:       decimal.New(50000, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{
					Tpc:       decimal.New(0, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)
			},
			wantErr: true,
		},
		{
			name: "FAILURE- tpc is zero - Counterfeit",
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
					Tpc:       decimal.New(0, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)

			},
			wantErr: true,
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
				mockRepository.EXPECT().CounterfeitPercent(ctx, "Scope1").Times(1).Return(db.CounterfeitPercentRow{}, errors.New("Internal"))

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
					Tpc:       decimal.New(50000, 0),
					DeltaCost: decimal.New(500, 0),
				}, nil)
				mockRepository.EXPECT().OverdeployPercent(ctx, "Scope1").Times(1).Return(db.OverdeployPercentRow{}, errors.New("Internal"))
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
					Tpc:       decimal.New(50000, 0),
					DeltaCost: decimal.New(500, 0),
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
			lr := NewProductServiceServer(rep, queue)
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
						Swidtag:     "PND1",
						ProductName: "ProNotDep1",
					},
					{
						Swidtag:     "PND2",
						ProductName: "ProNotDep2",
					},
				}, nil)
				mockRepo.EXPECT().ProductsNotAcquired(ctx, "Scope1").Times(1).Return([]db.ProductsNotAcquiredRow{
					{
						Swidtag:     "PNA1",
						ProductName: "ProNotAcq1",
					},
				}, nil)
			},
			want: &v1.DashboardQualityProductsResponse{
				ProductsNotDeployed: []*v1.DashboardQualityProducts{
					{
						SwidTag:     "PND1",
						ProductName: "ProNotDep1",
					},
					{
						SwidTag:     "PND2",
						ProductName: "ProNotDep2",
					},
				},
				ProductsNotAcquired: []*v1.DashboardQualityProducts{
					{
						SwidTag:     "PNA1",
						ProductName: "ProNotAcq1",
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
			s := NewProductServiceServer(rep, queue)
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
