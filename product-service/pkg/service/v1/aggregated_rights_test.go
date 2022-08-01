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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
)

func Test_productServiceServer_CreateAggregatedRights(t *testing.T) {
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
		req *v1.AggregatedRightsRequest
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
				req: &v1.AggregatedRightsRequest{
					AggregationID:             1,
					Sku:                       "aggsku",
					MetricName:                "met1",
					NumLicensesAcquired:       10,
					AvgUnitPrice:              2,
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:    2,
					AvgMaintenanceUnitPrice:   2,
					Scope:                     "scope1",
					Comment:                   "aggregation 1",
					OrderingDate:              "2020-01-01T10:58:56.026008Z",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             "123",
					MaintenanceProvider:       "oracle",
					FileName:                  "f1",
					FileData:                  []byte{},
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
				orderingtime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "aggsku",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:    1,
					Scope: "scope1",
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
						{
							Type:        "NUP",
							Name:        "met3",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().GetAggRightMetricsByAggregationId(ctx, db.GetAggRightMetricsByAggregationIdParams{
					Scope: "scope1",
					AggID: 1,
				}).Times(1).Return([]db.GetAggRightMetricsByAggregationIdRow{
					{
						Sku:    "sku2",
						Metric: "met1",
					},
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
						ProductName: "prod3",
						Swidtag:     "swid3",
					},
				}, nil)
				mockRepo.EXPECT().UpsertAggregatedRights(ctx, db.UpsertAggregatedRightsParams{
					AggregationID:             1,
					Sku:                       "aggsku",
					Metric:                    "met1",
					Scope:                     "scope1",
					NumLicensesAcquired:       10,
					NumLicencesMaintenance:    2,
					AvgUnitPrice:              decimal.NewFromFloat(2),
					AvgMaintenanceUnitPrice:   decimal.NewFromFloat(2),
					TotalPurchaseCost:         decimal.NewFromFloat(20),
					TotalMaintenanceCost:      decimal.NewFromFloat(4),
					TotalCost:                 decimal.NewFromFloat(24),
					StartOfMaintenance:        sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:          sql.NullTime{Time: endtime, Valid: true},
					Comment:                   sql.NullString{String: "aggregation 1", Valid: true},
					OrderingDate:              sql.NullTime{Time: orderingtime, Valid: true},
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             "123",
					MaintenanceProvider:       "oracle",
					FileName:                  "f1",
					FileData:                  []byte{},
					CreatedBy:                 "admin@superuser.com",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(&dgworker.UpsertAggregatedRight{
					AggregationID:             1,
					Sku:                       "aggsku",
					Metric:                    "met1",
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicensesAcquired:       10,
					AvgUnitPrice:              2,
					AvgMaintenanceUnitPrice:   2,
					TotalPurchaseCost:         20,
					TotalMaintenanceCost:      4,
					TotalCost:                 24,
					Scope:                     "scope1",
					NumLicencesMaintenance:    2,
					OrderingDate:              "2020-01-01T10:58:56.026008Z",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             "123",
					MaintenanceProvider:       "oracle",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAggregatedRights, JSON: jsonData}

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
				req: &v1.AggregatedRightsRequest{
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
				req: &v1.AggregatedRightsRequest{
					Scope: "scope4",
				},
			},
			setup: func() {},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-ServiceError",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRightsRequest{
					Sku:                       "aggsku",
					AggregationID:             1,
					MetricName:                "met1,met2",
					NumLicensesAcquired:       10,
					AvgUnitPrice:              2,
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:    2,
					AvgMaintenanceUnitPrice:   2,
					Scope:                     "scope1",
					Comment:                   "aggregation 1",
					OrderingDate:              "2020-01-01T10:58:56.026008Z",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             "123",
					MaintenanceProvider:       "oracle",
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
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "aggsku",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
					ID:    1,
					Scope: "scope1",
				}).Times(1).Return(db.Aggregation{
					ID:    1,
					Scope: "scope1",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(nil, errors.New("service error"))
			},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-sku already exists",
			args: args{
				ctx: ctx,
				req: &v1.AggregatedRightsRequest{
					Sku:                       "aggsku",
					MetricName:                "met1,met2",
					NumLicensesAcquired:       10,
					AvgUnitPrice:              2,
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:    2,
					AvgMaintenanceUnitPrice:   2,
					Scope:                     "scope1",
					Comment:                   "aggregation 1",
					OrderingDate:              "2020-01-01T10:58:56.026008Z",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             "123",
					MaintenanceProvider:       "oracle",
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
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "aggsku",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{Sku: "sku1"}, nil)
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
			got, err := tt.s.CreateAggregatedRights(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.CreateAggregatedRights() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.CreateAggregatedRights() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DeleteAggregatedRights(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DeleteAggregatedRightsRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DeleteAggregatedRightsResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregatedRightsRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DeleteAggregatedRightBySKU(ctx, db.DeleteAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.DeleteAggregatedRightRequest{
					Sku:   "sku1",
					Scope: "scope1",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.DeleteAggregatedRights, JSON: jsonData}

				envelopeData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envelopeData,
				}, "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.DeleteAggregatedRightsResponse{
				Success: true,
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteAggregatedRightsRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.DeleteAggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregatedRightsRequest{
					Sku:   "sku1",
					Scope: "scope5",
				},
			},
			setup: func() {},
			want: &v1.DeleteAggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAggregatedRightsRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DeleteAggregatedRightBySKU(ctx, db.DeleteAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(errors.New("internal"))
			},
			want: &v1.DeleteAggregatedRightsResponse{
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
			}
			got, err := tt.s.DeleteAggregatedRights(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DeleteAggregatedRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DeleteAggregatedRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DownloadAggregatedRightsFile(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DownloadAggregatedRightsFileRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DownloadAggregatedRightsFileResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{
					Sku:      "sku1",
					Metric:   "ops,metricNup",
					FileName: "sku1_file.pdf",
				}, nil)
				mockRepo.EXPECT().GetAggregatedRightsFileDataBySKU(ctx, db.GetAggregatedRightsFileDataBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return([]byte("filedata"), nil)
			},
			want: &v1.DownloadAggregatedRightsFileResponse{
				FileData: []byte("filedata"),
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup:   func() {},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope5",
				},
			},
			setup:   func() {},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-SKU does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
			},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{}, errors.New("internal"))
			},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-Acquired Right does not contain file",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{
					Sku:    "sku1",
					Metric: "ops,metricNup",
				}, nil)
			},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightFileDataBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAggregatedRightsFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{
					Sku:      "sku1",
					Metric:   "ops,metricNup",
					FileName: "sku1_file.pdf",
				}, nil)
				mockRepo.EXPECT().GetAggregatedRightsFileDataBySKU(ctx, db.GetAggregatedRightsFileDataBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return([]byte(""), errors.New("internal"))
			},
			want:    &v1.DownloadAggregatedRightsFileResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.DownloadAggregatedRightsFile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DownloadAggregatedRightsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
