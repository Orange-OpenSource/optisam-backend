package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"github.com/golang/mock/gomock"
)

func Test_ProductServiceServer_ListAggregationProducts(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepo.EXPECT().ListSelectedProductsForAggregration(ctx, gomock.Any()).Times(1).Return([]db.ListSelectedProductsForAggregrationRow{
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.ListAggregationProducts(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.ListAggregationProducts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.ListAggregationProducts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_ListAggregationEditors(t *testing.T) {
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
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.ListAggregationEditorsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.ListAggregationEditorsRequest{
					Scope: "scope1,scope3",
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, gomock.Any()).Times(1).Return([]string{"e1", "e2", "e3"}, nil)
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, gomock.Any()).Times(1).Return([]string{}, sql.ErrNoRows)
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
				mockRepo.EXPECT().ListEditorsForAggregation(ctx, gomock.Any()).Times(1).Return([]string{}, errors.New("internal"))

			},
			want:    &v1.ListAggregationEditorsResponse{},
			wantErr: true,
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
			got, err := tt.s.ListAggregationEditors(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.ListAggregationEditors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.ListAggregationEditors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_ListAggregations(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepo.EXPECT().ListAggregations(ctx, gomock.Any()).Times(1).Return([]db.ListAggregationsRow{
					{
						ID:              1,
						AggregationName: "agg1",
						ProductEditor:   "aggedit1",
						Products:        []string{"prod1", "prod2"},
						Swidtags:        []string{"swid1", "swid2", "swid3"},
						Scope:           "scope1",
						Coalesce:        []byte{},
					},
					{
						ID:              2,
						AggregationName: "agg2",
						ProductEditor:   "aggedit1",
						Products:        []string{"prod1", "prod2"},
						Swidtags:        []string{"swid4", "swid5", "swid6"},
						Scope:           "scope1",
						Coalesce:        []byte{},
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
						EditorId:        "e1",
						Mapping:         []*v1.Mapping{&v1.Mapping{ProductName: "p1", ProductVersion: "v1"}},
					},
					{
						ID:              2,
						AggregationName: "agg2",
						ProductEditor:   "aggedit1",
						ProductNames:    []string{"prod1", "prod2"},
						Swidtags:        []string{"swid4", "swid5", "swid6"},
						Scope:           "scope1",
						EditorId:        "e1",
						Mapping:         []*v1.Mapping{&v1.Mapping{ProductName: "p1", ProductVersion: "v1"}},
					},
				},
			},
			wantErr: false,
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
				mockRepo.EXPECT().ListAggregations(ctx, gomock.Any()).Times(1).Return([]db.ListAggregationsRow{}, sql.ErrNoRows)
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
				mockRepo.EXPECT().ListAggregations(ctx, gomock.Any()).Times(1).Return([]db.ListAggregationsRow{}, errors.New("internal"))
			},
			want:    &v1.ListAggregationsResponse{},
			wantErr: true,
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
			_, err := tt.s.ListAggregations(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.ListAggregations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !assert.Equal(t, tt.want, got, "ProductServiceServer.ListAggregations()") {
			// 	t.Errorf("ProductServiceServer.ListAggregations() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_ProductServiceServer_CreateAggregation(t *testing.T) {
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
		s       *ProductServiceServer
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.CreateAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.CreateAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.CreateAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_UpdateAggregation(t *testing.T) {
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
		s       *ProductServiceServer
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
					ID:     1,
					Scope:  "scope1",
					Editor: "aggeditor",
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
					ID:     1,
					Scope:  "scope1",
					Editor: "aggeditor",
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
		// Existing test cases...

		// Test case: UpdateAggregation - Aggregation not found
		{
			name: "FAILURE-AggregationNotFound",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:    1,
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{}, sql.ErrNoRows).AnyTimes()
			},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},

		// Test case: UpdateAggregation - Validation error
		{
			name: "FAILURE-ValidationError",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:              1,
					AggregationName: "",
					Scope:           "scope1",
				},
			},
			setup:   func() {},
			want:    &v1.AggregationResponse{Success: false},
			wantErr: true,
		},

		// Test case: UpdateAggregation - DB error during product retrieval
		{
			name: "FAILURE-DBError-ProductRetrieval",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:    1,
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:              1,
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, gomock.Any()).Times(1).Return(nil, errors.New("DBError"))
			},
			want: &v1.AggregationResponse{
				Success: false,
			},
			wantErr: true,
		},

		// Test case: UpdateAggregation - DB error during aggregation update
		{
			name: "FAILURE-DBError-AggregationUpdate",
			args: args{
				ctx: ctx,
				req: &v1.Aggregation{
					ID:    1,
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:              1,
					AggregationName: "aggname",
				}, nil)
				mockRepo.EXPECT().ListProductsForAggregation(ctx, gomock.Any()).Times(1).Return([]db.ListProductsForAggregationRow{
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
				mockRepo.EXPECT().UpdateAggregation(ctx, gomock.Any()).Times(1).Return(errors.New("DBError"))
				mockRepo.EXPECT().ListSelectedProductsForAggregration(ctx, gomock.Any()).Times(1).Return([]db.ListSelectedProductsForAggregrationRow{
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.UpdateAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.UpdateAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.UpdateAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_DeleteAggregation(t *testing.T) {
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
		s       *ProductServiceServer
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.DeleteAggregation(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DeleteAggregation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DeleteAggregation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestUpdateAggregatedRights(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	ProductRepo := dbmock.NewMockProduct(mockCtrl)

// 	testCases := []struct {
// 		name          string
// 		ctx           context.Context
// 		request       *v1.AggregatedRightsRequest
// 		expectedError error
// 		mockSetup     func()
// 	}{
// 		{
// 			name: "ValidRequest",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(&db.AggregatedRight{
// 					Sku: "SKU1",
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.GetAvailableLicenses(gomock.Any(), gomock.Any()).Return(&v1.GetAvailableLicensesResponse{
// 					TotalSharedLicenses: 5,
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.EXPECT().UpsertAggregatedRights(gomock.Any(), gomock.Any()).Return(nil)
// 				// Add expectations for other method calls
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "ClaimsNotFoundError",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// No mock expectations needed
// 			},
// 			expectedError: status.Error(codes.Internal, "ClaimsNotFoundError"),
// 		},
// 		{
// 			name: "ScopeValidationError",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(&db.AggregatedRight{
// 					Sku: "SKU1",
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.EXPECT().GetAvailableLicenses(gomock.Any(), gomock.Any()).Return(&v1.GetAvailableLicensesResponse{
// 					TotalSharedLicenses: 5,
// 					// Set other necessary fields
// 				}, nil)
// 				// Add expectations for other method calls
// 			},
// 			expectedError: status.Error(codes.InvalidArgument, "ScopeValidationError"),
// 		},
// 		{
// 			name: "AggregationDoesNotExist",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(nil, sql.ErrNoRows)
// 				// Add expectations for other method calls
// 			},
// 			expectedError: status.Error(codes.InvalidArgument, "aggregation does not exist"),
// 		},
// 		{
// 			name: "SkuCannotBeUpdated",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(&db.AggregatedRight{
// 					Sku: "DifferentSKU",
// 					// Set other necessary fields
// 				}, nil)
// 				// Add expectations for other method calls
// 			},
// 			expectedError: status.Error(codes.InvalidArgument, "sku cannot be updated"),
// 		},
// 		{
// 			name: "AcquiredLicencesLessThanSharedLicences",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 5,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(&db.AggregatedRight{
// 					Sku: "SKU1",
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.EXPECT().GetAvailableLicenses(gomock.Any(), gomock.Any()).Return(&v1.GetAvailableLicensesResponse{
// 					TotalSharedLicenses: 10,
// 					// Set other necessary fields
// 				}, nil)
// 				// Add expectations for other method calls
// 			},
// 			expectedError: status.Error(codes.InvalidArgument, "AcquiredLicences less than sharedLicences"),
// 		},
// 		{
// 			name: "DBErrorOnUpsertAggregatedRights",
// 			ctx:  context.Background(),
// 			request: &v1.AggregatedRightsRequest{
// 				Sku:                 "SKU1",
// 				Scope:               "scope1",
// 				NumLicensesAcquired: 10,
// 				// Add other necessary fields to the request
// 			},
// 			mockSetup: func() {
// 				// Set up expectations for the mock
// 				ProductRepo.EXPECT().GetAggregatedRightBySKU(gomock.Any(), gomock.Any()).Return(&db.AggregatedRight{
// 					Sku: "SKU1",
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.EXPECT().GetAvailableLicenses(gomock.Any(), gomock.Any()).Return(&v1.GetAvailableLicensesResponse{
// 					TotalSharedLicenses: 10,
// 					// Set other necessary fields
// 				}, nil)
// 				ProductRepo.EXPECT().UpsertAggregatedRights(gomock.Any(), gomock.Any()).Return(errors.New("DB error"))
// 				// Add expectations for other method calls
// 			},
// 			expectedError: status.Error(codes.Unknown, "DBError"),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			tc.mockSetup()

// 			// Create an instance of the service with the mocked dependencies
// 			service := &ProductServiceServer{
// 				ProductRepo: ProductRepo,
// 			}

// 			// Call the method being tested
// 			response, err := service.UpdateAggregatedRights(tc.ctx, tc.request)

// 			// Check the error response
// 			if tc.expectedError != nil {
// 				if status.Code(err) != status.Code(tc.expectedError) {
// 					t.Errorf("Expected error code %s, but got %s", status.Code(tc.expectedError), status.Code(err))
// 				}
// 				return
// 			}

// 			// Check the success response
// 			if err != nil {
// 				t.Errorf("Unexpected error: %v", err)
// 			}
// 			if !response.Success {
// 				t.Error("Expected success to be true, but it was false")
// 			}
// 		})
// 	}

// 	mockCtrl.Finish()
// }

func Test_GetAggregationById(t *testing.T) {
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
		req *v1.GetAggregationByIdRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetAggregationByIdResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetAggregationByIdRequest{
					Scope:         "scope1",
					AggregationId: 56,
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

				mockRepo.EXPECT().GetAggregationByID(ctx, gomock.Any()).Times(1).Return(db.Aggregation{
					ID:              56,
					AggregationName: "agg1",
					Scope:           "scope1",
					ProductEditor:   "e1",
				}, nil)
			},
			want:    &v1.GetAggregationByIdResponse{Id: 56, AggregationName: "agg1", Scope: "scope1", ProductEditor: "e1"},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.GetAggregationByIdRequest{
					Scope:         "scope1",
					AggregationId: 56,
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
				mockRepo.EXPECT().GetAggregationByID(ctx, gomock.Any()).Times(1).Return(db.Aggregation{
					ID:              56,
					AggregationName: "agg1",
					Scope:           "scope1",
					ProductEditor:   "e1",
				}, errors.New("error"))

			},
			want:    &v1.GetAggregationByIdResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetAggregationByIdRequest{
					Scope:         "scope1",
					AggregationId: 56,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetAggregationByIdRequest{
					Scope:         "scope121",
					AggregationId: 56,
				},
			},
			setup:   func() {},
			wantErr: true,
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
			_, err := tt.s.GetAggregationById(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetEditorExpensesByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ProductServiceServer.GetEditorExpensesByScope() = %v, want %v", got, tt.want)
			// }
		})
	}
}
