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
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_productServiceServer_ListAggregationProducts(t *testing.T) {
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
		req *v1.ListAggregationProductsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregationProductsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationProductsRequest{
					ID:     1,
					Editor: "editor",
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
			want: &v1.ListAggregationProductsResponse{
				AggrightsProducts: []*v1.AggregationProducts{
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
				SelectedProducts: []*v1.AggregationProducts{
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
				req: &v1.ListAggregationProductsRequest{
					Editor: "editor",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationProductsRequest{
					Editor: "editor",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.ListAggregationProductsResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/ListProductsForAggregation-no data",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationProductsRequest{
					Editor: "editor",
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
					Scope:  "scope1",
				}).Times(1).Return([]db.ListProductsForAggregationRow{}, sql.ErrNoRows)
			},
			want:    &v1.ListAggregationProductsResponse{},
			wantErr: false,
		},
		{name: "FAILURE-db/ListProductsForAggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationProductsRequest{
					Editor: "editor",
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
					Scope:  "scope1",
				}).Times(1).Return([]db.ListProductsForAggregationRow{}, errors.New("internal"))
			},
			want:    &v1.ListAggregationProductsResponse{},
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
			got, err := tt.s.ListAggregationProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ListAggregationProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.ListAggregationProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_ListAggregationEditors(t *testing.T) {
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
		req *v1.ListAggregationEditorsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregationEditorsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationEditorsRequest{
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
			want: &v1.ListAggregationEditorsResponse{
				Editor: []string{"e1", "e2", "e3"},
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAggregationEditorsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationEditorsRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.ListAggregationEditorsResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/ListEditorsForAggregation-no data",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationEditorsRequest{
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
			want:    &v1.ListAggregationEditorsResponse{},
			wantErr: false,
		},
		{name: "FAILURE-db/ListEditorsForAggregation",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationEditorsRequest{
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
			want:    &v1.ListAggregationEditorsResponse{},
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
			got, err := tt.s.ListAggregationEditors(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.ListAggregationEditors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.ListAggregationEditors() = %v, want %v", got, tt.want)
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
				mockRepo.EXPECT().ListAggregations(ctx, db.ListAggregationsParams{
					Scope: "scope1",
				}).Times(1).Return([]db.ListAggregationsRow{
					{

						ID:              1,
						AggregationName: "agg1",
						ProductEditor:   "aggedit1",
						Products:        []string{"prod1", "prod2"},
						Swidtags:        []string{"swid1", "swid2", "swid3"},
						Scope:           "scope1",
					},
					{

						ID:              2,
						AggregationName: "agg2",
						ProductEditor:   "aggedit1",
						Products:        []string{"prod1", "prod2"},
						Swidtags:        []string{"swid4", "swid5", "swid6"},
						Scope:           "scope1",
					},
				}, nil)
			},
			want: &v1.ListAggregationsResponse{
				TotalRecords: 2,
				Aggregations: []*v1.Aggregation{
					{
						ID:              1,
						AggregationName: "agg1",
						ProductEditor:   "aggedit1",
						ProductNames:    []string{"prod1", "prod2"},
						Swidtags:        []string{"swid1", "swid2", "swid3"},
						Scope:           "scope1",
					},
					{
						ID:              2,
						AggregationName: "agg2",
						ProductEditor:   "aggedit1",
						ProductNames:    []string{"prod1", "prod2"},
						Swidtags:        []string{"swid4", "swid5", "swid6"},
						Scope:           "scope1",
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
				mockRepo.EXPECT().ListAggregations(ctx, db.ListAggregationsParams{
					Scope: "scope1",
				}).Times(1).Return([]db.ListAggregationsRow{}, sql.ErrNoRows)
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
				mockRepo.EXPECT().ListAggregations(ctx, db.ListAggregationsParams{
					Scope: "scope1",
				}).Times(1).Return([]db.ListAggregationsRow{}, errors.New("internal"))
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
		req *v1.Aggregation
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AggregationResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					ProductNames:    []string{"prod1", "prod2"},
					Swidtags:        []string{"swid1", "swid2", "swid3"},
					Scope:           "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: "aggname",
					Scope:           "scope1",
				}).Times(1).Return(db.Aggregation{}, sql.ErrNoRows)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
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
				mockRepo.EXPECT().InsertAggregation(ctx, db.InsertAggregationParams{
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					Products:        []string{"prod1", "prod2"},
					Swidtags:        []string{"swid1", "swid2", "swid3"},
					Scope:           "scope1",
					CreatedBy:       "admin@superuser.com",
				}).Times(1).Return(int32(1), nil)
				jsonData, err := json.Marshal(&dgworker.UpsertAggregationRequest{
					ID:            1,
					Name:          "aggname",
					Swidtags:      []string{"swid1", "swid2", "swid3"},
					Products:      []string{"prod1", "prod2"},
					ProductEditor: "aggeditor",
					Scope:         "scope1",
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
			want: &v1.AggregationResponse{
				Success: true,
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.Aggregation{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					Scope: "scope4",
				},
			},
			setup: func() {},
			want: &v1.AggregationResponse{
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
		req *v1.Aggregation
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AggregationResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:              1,
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					ProductNames:    []string{"prod1", "prod2", "prod3", "prod4"},
					Swidtags:        []string{"swid1", "swid2", "swid3", "swid4"},
					Scope:           "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
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
				mockRepo.EXPECT().UpdateAggregation(ctx, db.UpdateAggregationParams{
					ID:              1,
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					ProductNames:    []string{"prod1", "prod2", "prod3", "prod4"},
					Swidtags:        []string{"swid1", "swid2", "swid3", "swid4"},
					Scope:           "scope1",
					UpdatedBy:       sql.NullString{String: "admin@superuser.com", Valid: true},
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(&dgworker.UpsertAggregationRequest{
					ID:            1,
					Name:          "aggname",
					Swidtags:      []string{"swid1", "swid2", "swid3", "swid4"},
					Products:      []string{"prod1", "prod2", "prod3", "prod4"},
					ProductEditor: "aggeditor",
					Scope:         "scope1",
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
			want: &v1.AggregationResponse{
				Success: true,
			},
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.Aggregation{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					Scope: "scope4",
				},
			},
			setup: func() {},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-UpdateAggregation-DBError",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:              1,
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					ProductNames:    []string{"prod1", "prod2", "prod3"},
					Swidtags:        []string{"swid1", "swid2", "swid3"},
					Scope:           "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:              1,
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
					Scope:  "scope1",
					Editor: "aggeditor",
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
				mockRepo.EXPECT().UpdateAggregation(ctx, db.UpdateAggregationParams{
					ID:              1,
					AggregationName: "aggname",
					ProductEditor:   "aggeditor",
					ProductNames:    []string{"prod1", "prod2", "prod3"},
					Swidtags:        []string{"swid1", "swid2", "swid3"},
					Scope:           "scope1",
					UpdatedBy:       sql.NullString{String: "admin@superuser.com", Valid: true},
				}).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.AggregationResponse{
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

func Test_productServiceServer_DeleteAggregation(t *testing.T) {
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
		req *v1.DeleteAggregationRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AggregationResponse
		wantErr bool
	}{
		{name: "Success-Delete Aggregations",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregationRequest{
					ID:    1,
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:              1,
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().DeleteAggregation(ctx, db.DeleteAggregationParams{
					ID:    1,
					Scope: "scope1",
				}).Return(nil).Times(1)
				jsonData, err := json.Marshal(&v1.DeleteAggregationRequest{
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
			want: &v1.AggregationResponse{
				Success: true,
			},
		},
		{name: "FAILURE-Claims Not Found",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteAggregationRequest{
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregationRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			want:    &v1.AggregationResponse{},
			wantErr: true,
		},
		{name: "FAILURE-aggregation does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregationRequest{
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
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{}, sql.ErrNoRows)
			},
			want:    &v1.AggregationResponse{},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAggregationByID",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregationRequest{
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
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{}, errors.New("internal"))
			},
			want:    &v1.AggregationResponse{},
			wantErr: true,
		},
		{name: "FAILURE-Delete Aggregation",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregationRequest{
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
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:              1,
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().DeleteAggregation(ctx, db.DeleteAggregationParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(errors.New("internal"))
			},
			want:    &v1.AggregationResponse{},
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
			got, err := tt.s.DeleteAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DeleteAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DeleteAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}
