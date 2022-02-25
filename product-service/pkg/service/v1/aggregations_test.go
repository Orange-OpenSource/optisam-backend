package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	metmock "optisam-backend/metric-service/pkg/api/v1/mock"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	mock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func Test_productServiceServer_ListAggregatedRightsProducts(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.ListAggregatedRightsProductsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregatedRightsProductsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsProductsRequest{
					ID:     1,
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
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
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
				}).Times(1).Return([]db.ListProductsForAggregationRow{
					{
						ProductName:   "pro1",
						Swidtag:       "swid1",
						ProductEditor: "abc",
					},
					{
						ProductName:   "pro1",
						Swidtag:       "swid2",
						ProductEditor: "abc",
					},
				}, nil)
				mockRepo.EXPECT().ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return([]db.ListSelectedProductsForAggregrationRow{
					{
						ProductName:   "pro2",
						Swidtag:       "swid1",
						ProductEditor: "abc",
					},
					{
						ProductName:   "pro2",
						Swidtag:       "swid2",
						ProductEditor: "abc",
					},
				}, nil)
			},
			want: &v1.ListAggregatedRightsProductsResponse{
				AggrightsProducts: []*v1.AggregatedRightsProducts{
					{
						ProductName: "pro1",
						Swidtag:     "swid1",
						Editor:      "abc",
					},
					{
						ProductName: "pro1",
						Swidtag:     "swid2",
						Editor:      "abc",
					},
				},
				SelectedProducts: []*v1.AggregatedRightsProducts{
					{
						ProductName: "pro2",
						Swidtag:     "swid1",
						Editor:      "abc",
					},
					{
						ProductName: "pro2",
						Swidtag:     "swid2",
						Editor:      "abc",
					},
				},
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAggregatedRightsProductsRequest{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsProductsRequest{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.ListAggregatedRightsProductsResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/ListProductsForAggregation-no data",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsProductsRequest{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
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
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
				}).Times(1).Return([]db.ListProductsForAggregationRow{}, sql.ErrNoRows)
			},
			want:    &v1.ListAggregatedRightsProductsResponse{},
			wantErr: false,
		},
		{name: "FAILURE-db/ListProductsForAggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsProductsRequest{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
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
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Editor: "editor",
					Metric: "metric",
					Scope:  "scope1",
				}).Times(1).Return([]db.ListProductsForAggregationRow{}, errors.New("internal"))
			},
			want:    &v1.ListAggregatedRightsProductsResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.ListAggregatedRightsProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ListAggregatedRightsProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.ListAggregatedRightsProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_ListAggregatedRightsEditors(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.ListAggregatedRightsEditorsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregatedRightsEditorsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsEditorsRequest{
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, "scope1").Times(1).Return([]string{"e1", "e2", "e3"}, nil)
			},
			want: &v1.ListAggregatedRightsEditorsResponse{
				Editor: []string{"e1", "e2", "e3"},
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAggregatedRightsEditorsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsEditorsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.ListAggregatedRightsEditorsResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/ListEditorsForAggregation-no data",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsEditorsRequest{
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, "scope1").Times(1).Return([]string{}, sql.ErrNoRows)
			},
			want:    &v1.ListAggregatedRightsEditorsResponse{},
			wantErr: false,
		},
		{name: "FAILURE-db/ListEditorsForAggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregatedRightsEditorsRequest{
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, "scope1").Times(1).Return([]string{}, errors.New("internal"))

			},
			want:    &v1.ListAggregatedRightsEditorsResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.ListAggregatedRightsEditors(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ListAggregatedRightsEditors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.ListAggregatedRightsEditors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_ListAggregations(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	timeStart := time.Now()
	timeEnd := timeStart.Add(10 * time.Hour)
	timestampStart, _ := ptypes.TimestampProto(timeStart)
	timestampEnd, _ := ptypes.TimestampProto(timeEnd)
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.ListAggregationsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregationsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationsRequest{
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
				mockRepo.EXPECT().ListAggregations(ctx, "scope1").Times(1).Return([]db.ListAggregationsRow{
					{
						ID:                      1,
						AggregationName:         "agg1",
						Sku:                     "aggsku1",
						ProductEditor:           "aggedit1",
						Metric:                  "aggmet1",
						Products:                []string{"prod1", "prod2"},
						Swidtags:                []string{"swid1", "swid2", "swid3"},
						Scope:                   "scope1",
						NumLicensesAcquired:     10,
						NumLicencesComputed:     0,
						NumLicencesMaintainance: 5,
						AvgUnitPrice:            decimal.NewFromFloat(10),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
						TotalPurchaseCost:       decimal.NewFromFloat(100),
						TotalComputedCost:       decimal.NewFromFloat(0),
						TotalMaintenanceCost:    decimal.NewFromFloat(10),
						TotalCost:               decimal.NewFromFloat(110),
						StartOfMaintenance:      sql.NullTime{Time: timeStart, Valid: true},
						EndOfMaintenance:        sql.NullTime{Time: timeEnd, Valid: true},
						Comment:                 sql.NullString{String: "aggregation 1"},
					},
					{
						ID:                  2,
						AggregationName:     "agg2",
						Sku:                 "aggsku2",
						ProductEditor:       "aggedit1",
						Metric:              "aggmet1",
						Products:            []string{"prod1", "prod2"},
						Swidtags:            []string{"swid4", "swid5", "swid6"},
						Scope:               "scope1",
						NumLicensesAcquired: 10,
						NumLicencesComputed: 0,
						AvgUnitPrice:        decimal.NewFromFloat(10),
						TotalPurchaseCost:   decimal.NewFromFloat(100),
						TotalComputedCost:   decimal.NewFromFloat(0),
						TotalCost:           decimal.NewFromFloat(100),
						Comment:             sql.NullString{String: "aggregation 2"},
					},
				}, nil)
			},
			want: &v1.ListAggregationsResponse{
				Aggregations: []*v1.ListAggregatedRights{
					{
						ID:                      1,
						AggregationName:         "agg1",
						Sku:                     "aggsku1",
						ProductEditor:           "aggedit1",
						MetricName:              "aggmet1",
						ProductNames:            []string{"prod1", "prod2"},
						Swidtags:                []string{"swid1", "swid2", "swid3"},
						Scope:                   "scope1",
						NumLicensesAcquired:     10,
						AvgUnitPrice:            10,
						StartOfMaintenance:      timestampStart,
						EndOfMaintenance:        timestampEnd,
						NumLicencesMaintainance: 5,
						AvgMaintenanceUnitPrice: 2,
						Comment:                 "aggregation 1",
					},
					{
						ID:                  2,
						AggregationName:     "agg2",
						Sku:                 "aggsku2",
						ProductEditor:       "aggedit1",
						MetricName:          "aggmet1",
						ProductNames:        []string{"prod1", "prod2"},
						Swidtags:            []string{"swid4", "swid5", "swid6"},
						Scope:               "scope1",
						NumLicensesAcquired: 10,
						AvgUnitPrice:        10,
						Comment:             "aggregation 2",
					},
				},
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAggregationsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.ListAggregationsResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/ListAggregations-no data",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationsRequest{
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
				mockRepo.EXPECT().ListAggregations(ctx, "scope1").Times(1).Return([]db.ListAggregationsRow{}, sql.ErrNoRows)
			},
			want:    &v1.ListAggregationsResponse{},
			wantErr: false,
		},
		{name: "FAILURE-db/ListAggregations",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationsRequest{
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
				mockRepo.EXPECT().ListAggregations(ctx, "scope1").Times(1).Return([]db.ListAggregationsRow{}, errors.New("internal"))
			},
			want:    &v1.ListAggregationsResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.ListAggregations(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ListAggregations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.Equal(t, tt.want, got, "productServiceServer.ListAggregations()") {
				t.Errorf("productServiceServer.ListAggregations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_CreateAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.AggregatedRights
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AggregatedRightsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRights{
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					MetricName:              "met1,met2",
					ProductNames:            []string{"prod1", "prod2"},
					Swidtags:                []string{"swid1", "swid2", "swid3"},
					NumLicensesAcquired:     10,
					AvgUnitPrice:            2,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 2,
					AvgMaintenanceUnitPrice: 2,
					Scope:                   "scope1",
					Comment:                 "aggregation 1",
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
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: "aggname",
					Scope:           "scope1",
				}).Times(1).Return(db.AggregatedRight{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregationBySKU(ctx, db.GetAggregationBySKUParams{
					Sku:   "aggsku",
					Scope: "scope1",
				}).Times(1).Return(db.AggregatedRight{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "met1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "met2",
							Description: "metricNup description",
						},
						{
							Type:        "NUP",
							Name:        "met3",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
					Metric: "met1,met2",
				}).Times(1).Return([]db.ListProductsForAggregationRow{
					{
						ProductName: "prod1",
						Swidtag:     "swid1",
					},
					{
						ProductName: "prod1",
						Swidtag:     "swid2",
					},
					{
						ProductName: "prod3",
						Swidtag:     "swid3",
					},
				}, nil)
				mockRepo.EXPECT().UpsertAggregation(ctx, db.UpsertAggregationParams{
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					Metric:                  "met1,met2",
					Products:                []string{"prod1", "prod2"},
					Swidtags:                []string{"swid1", "swid2", "swid3"},
					Scope:                   "scope1",
					NumLicensesAcquired:     10,
					NumLicencesMaintainance: 2,
					AvgUnitPrice:            decimal.NewFromFloat(2),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
					TotalPurchaseCost:       decimal.NewFromFloat(20),
					TotalMaintenanceCost:    decimal.NewFromFloat(4),
					TotalCost:               decimal.NewFromFloat(24),
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					Comment:                 sql.NullString{String: "aggregation 1", Valid: true},
					CreatedBy:               "admin@superuser.com",
				}).Times(1).Return(int32(1), nil)
				jsonData, err := json.Marshal(&dgworker.UpsertAggregatedRightsRequest{
					ID:                      1,
					Name:                    "aggname",
					Sku:                     "aggsku",
					Swidtags:                []string{"swid1", "swid2", "swid3"},
					Products:                []string{"prod1", "prod2"},
					ProductEditor:           "aggeditor",
					Metric:                  "met1,met2",
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicensesAcquired:     10,
					AvgUnitPrice:            2,
					AvgMaintenanceUnitPrice: 2,
					TotalPurchaseCost:       20,
					TotalMaintenanceCost:    4,
					TotalCost:               24,
					Scope:                   "scope1",
					NumLicencesMaintenance:  2,
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAggregation, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}, "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.AggregatedRightsResponse{
				Success: true,
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.AggregatedRights{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRights{
					Scope: "scope4",
				},
			},
			setup: func() {},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.CreateAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.CreateAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.CreateAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_UpdateAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.AggregatedRights
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AggregatedRightsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRights{
					ID:                      1,
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					MetricName:              "met1,met2",
					ProductNames:            []string{"prod1", "prod2", "prod3", "prod4"},
					Swidtags:                []string{"swid1", "swid2", "swid3", "swid4"},
					NumLicensesAcquired:     10,
					AvgUnitPrice:            2,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 2,
					AvgMaintenanceUnitPrice: 2,
					Scope:                   "scope1",
					Comment:                 "aggregation 1",
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
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.AggregatedRight{
					AggregationName: "aggname",
					Sku:             "aggsku",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "met1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "met2",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
					Metric: "met1,met2",
				}).Times(1).Return([]db.ListProductsForAggregationRow{
					{
						ProductName: "prod1",
						Swidtag:     "swid1",
					},
					{
						ProductName: "prod1",
						Swidtag:     "swid2",
					},
					{
						ProductName: "prod4",
						Swidtag:     "swid4",
					},
				}, nil)
				mockRepo.EXPECT().ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return([]db.ListSelectedProductsForAggregrationRow{
					{
						ProductName: "prod3",
						Swidtag:     "swid3",
					},
				}, nil)
				mockRepo.EXPECT().UpsertAggregation(ctx, db.UpsertAggregationParams{
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					Metric:                  "met1,met2",
					Products:                []string{"prod1", "prod2", "prod3", "prod4"},
					Swidtags:                []string{"swid1", "swid2", "swid3", "swid4"},
					Scope:                   "scope1",
					NumLicensesAcquired:     10,
					NumLicencesMaintainance: 2,
					AvgUnitPrice:            decimal.NewFromFloat(2),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
					TotalPurchaseCost:       decimal.NewFromFloat(20),
					TotalMaintenanceCost:    decimal.NewFromFloat(4),
					TotalCost:               decimal.NewFromFloat(24),
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					Comment:                 sql.NullString{String: "aggregation 1", Valid: true},
					CreatedBy:               "admin@superuser.com",
				}).Times(1).Return(int32(1), nil)
				jsonData, err := json.Marshal(&dgworker.UpsertAggregatedRightsRequest{
					ID:                      1,
					Name:                    "aggname",
					Sku:                     "aggsku",
					Swidtags:                []string{"swid1", "swid2", "swid3", "swid4"},
					Products:                []string{"prod1", "prod2", "prod3", "prod4"},
					ProductEditor:           "aggeditor",
					Metric:                  "met1,met2",
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicensesAcquired:     10,
					AvgUnitPrice:            2,
					AvgMaintenanceUnitPrice: 2,
					TotalPurchaseCost:       20,
					TotalMaintenanceCost:    4,
					TotalCost:               24,
					Scope:                   "scope1",
					NumLicencesMaintenance:  2,
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAggregation, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}, "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.AggregatedRightsResponse{
				Success: true,
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.AggregatedRights{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRights{
					Scope: "scope4",
				},
			},
			setup: func() {},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-InsertAcqRight-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRights{
					ID:                      1,
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					MetricName:              "met1,met2",
					ProductNames:            []string{"prod1", "prod2", "prod3"},
					Swidtags:                []string{"swid1", "swid2", "swid3"},
					NumLicensesAcquired:     10,
					AvgUnitPrice:            2,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 2,
					AvgMaintenanceUnitPrice: 2,
					Scope:                   "scope1",
					Comment:                 "aggregation 1",
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
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.AggregatedRight{
					AggregationName: "aggname",
					Sku:             "aggsku",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "met1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "met2",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
					Metric: "met1,met2",
				}).Times(1).Return([]db.ListProductsForAggregationRow{
					{
						ProductName: "prod1",
						Swidtag:     "swid1",
					},
					{
						ProductName: "prod1",
						Swidtag:     "swid2",
					},
				}, nil)
				mockRepo.EXPECT().ListSelectedProductsForAggregration(ctx, db.ListSelectedProductsForAggregrationParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return([]db.ListSelectedProductsForAggregrationRow{
					{
						ProductName: "prod3",
						Swidtag:     "swid3",
					},
				}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAggregation(ctx, db.UpsertAggregationParams{
					AggregationName:         "aggname",
					Sku:                     "aggsku",
					ProductEditor:           "aggeditor",
					Metric:                  "met1,met2",
					Products:                []string{"prod1", "prod2", "prod3"},
					Swidtags:                []string{"swid1", "swid2", "swid3"},
					Scope:                   "scope1",
					NumLicensesAcquired:     10,
					NumLicencesMaintainance: 2,
					AvgUnitPrice:            decimal.NewFromFloat(2),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
					TotalPurchaseCost:       decimal.NewFromFloat(20),
					TotalMaintenanceCost:    decimal.NewFromFloat(4),
					TotalCost:               decimal.NewFromFloat(24),
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					Comment:                 sql.NullString{String: "aggregation 1", Valid: true},
					CreatedBy:               "admin@superuser.com",
				}).Times(1).Return(int32(0), errors.New("Internal"))
			},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.UpdateAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.UpdateAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.UpdateAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DeleteProductAggregation(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	type args struct {
		ctx context.Context
		req *v1.DeleteProductAggregationRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DeleteProductAggregationResponse
		wantErr bool
	}{
		{name: "Success-Delete Product Aggregations",
			args: args{
				ctx: ctx,
				req: &v1.DeleteProductAggregationRequest{
					ID:    1,
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DeleteAggregation(ctx, db.DeleteAggregationParams{
					AggregationID: 1,
					Scope:         "scope1",
				}).Return(nil).Times(1)
				jsonData, err := json.Marshal(&v1.DeleteProductAggregationRequest{
					ID:    1,
					Scope: "scope1",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.DeleteAggregation, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}, "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.DeleteProductAggregationResponse{
				Success: true,
			},
		},
		{name: "FAILURE-Claims Not Found",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteProductAggregationRequest{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.DeleteProductAggregationResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.DeleteProductAggregationRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.DeleteProductAggregationResponse{},
			wantErr: true,
		},
		{name: "FAILURE-Delete Aggregation",
			args: args{
				ctx: ctx,
				req: &v1.DeleteProductAggregationRequest{
					ID:    1,
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
				mockRepo.EXPECT().DeleteAggregation(ctx, db.DeleteAggregationParams{
					AggregationID: 1,
					Scope:         "scope1",
				}).Times(1).Return(errors.New("internal"))
			},
			want:    &v1.DeleteProductAggregationResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.DeleteProductAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DeleteProductAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DeleteProductAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}
