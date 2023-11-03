package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

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
	"github.com/shopspring/decimal"
)

func Test_ProductServiceServer_CreateAggregatedRights(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
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
				// mockRepo.EXPECT().GetAggRightMetricsByAggregationId(ctx, db.GetAggRightMetricsByAggregationIdParams{
				// 	Scope: "scope1",
				// 	AggID: 1,
				// }).Times(1).Return([]db.GetAggRightMetricsByAggregationIdRow{
				// 	{
				// 		Sku:    "sku2",
				// 		Metric: "met1",
				// 	},
				// }, nil)
				// mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
				// 	Scope:  "scope1",
				// 	Editor: "aggeditor",
				// }).Times(1).Return([]db.ListProductsForAggregationRow{
				// 	{
				// 		ProductName: "prod1",
				// 		Swidtag:     "swid1",
				// 	},
				// 	{
				// 		ProductName: "prod1",
				// 		Swidtag:     "swid2",
				// 	},
				// 	{
				// 		ProductName: "prod3",
				// 		Swidtag:     "swid3",
				// 	},
				// }, nil)
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
					SupportNumbers:            []string{"123"},
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
		{name: "Fail support number greater than 16",
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
					SupportNumber:             "12345678901234656890",
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
				// starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				// endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				// orderingtime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, db.GetAggregatedRightBySKUParams{
					Sku:   "aggsku",
					Scope: "scope1",
				}).Times(1).Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
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
				// mockRepo.EXPECT().GetAggRightMetricsByAggregationId(ctx, db.GetAggRightMetricsByAggregationIdParams{
				// 	Scope: "scope1",
				// 	AggID: 1,
				// }).Times(1).Return([]db.GetAggRightMetricsByAggregationIdRow{
				// 	{
				// 		Sku:    "sku2",
				// 		Metric: "met1",
				// 	},
				// }, nil)
				// mockRepo.EXPECT().ListProductsForAggregation(ctx, db.ListProductsForAggregationParams{
				// 	Scope:  "scope1",
				// 	Editor: "aggeditor",
				// }).Times(1).Return([]db.ListProductsForAggregationRow{
				// 	{
				// 		ProductName: "prod1",
				// 		Swidtag:     "swid1",
				// 	},
				// 	{
				// 		ProductName: "prod1",
				// 		Swidtag:     "swid2",
				// 	},
				// 	{
				// 		ProductName: "prod3",
				// 		Swidtag:     "swid3",
				// 	},
				// }, nil)
				// mockRepo.EXPECT().UpsertAggregatedRights(ctx, db.UpsertAggregatedRightsParams{
				// 	AggregationID:             1,
				// 	Sku:                       "aggsku",
				// 	Metric:                    "met1",
				// 	Scope:                     "scope1",
				// 	NumLicensesAcquired:       10,
				// 	NumLicencesMaintenance:    2,
				// 	AvgUnitPrice:              decimal.NewFromFloat(2),
				// 	AvgMaintenanceUnitPrice:   decimal.NewFromFloat(2),
				// 	TotalPurchaseCost:         decimal.NewFromFloat(20),
				// 	TotalMaintenanceCost:      decimal.NewFromFloat(4),
				// 	TotalCost:                 decimal.NewFromFloat(24),
				// 	StartOfMaintenance:        sql.NullTime{Time: starttime, Valid: true},
				// 	EndOfMaintenance:          sql.NullTime{Time: endtime, Valid: true},
				// 	Comment:                   sql.NullString{String: "aggregation 1", Valid: true},
				// 	OrderingDate:              sql.NullTime{Time: orderingtime, Valid: true},
				// 	CorporateSourcingContract: "csc",
				// 	SoftwareProvider:          "oracle",
				// 	LastPurchasedOrder:        "odernum",
				// 	SupportNumbers:            []string{"123"},
				// 	MaintenanceProvider:       "oracle",
				// 	FileName:                  "f1",
				// 	FileData:                  []byte{},
				// 	CreatedBy:                 "admin@superuser.com",
				// // }).Times(1).Return(nil)
				// jsonData, err := json.Marshal(&dgworker.UpsertAggregatedRight{
				// 	AggregationID:             1,
				// 	Sku:                       "aggsku",
				// 	Metric:                    "met1",
				// 	StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
				// 	EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
				// 	NumLicensesAcquired:       10,
				// 	AvgUnitPrice:              2,
				// 	AvgMaintenanceUnitPrice:   2,
				// 	TotalPurchaseCost:         20,
				// 	TotalMaintenanceCost:      4,
				// 	TotalCost:                 24,
				// 	Scope:                     "scope1",
				// 	NumLicencesMaintenance:    2,
				// 	OrderingDate:              "2020-01-01T10:58:56.026008Z",
				// 	CorporateSourcingContract: "csc",
				// 	SoftwareProvider:          "oracle",
				// 	LastPurchasedOrder:        "odernum",
				// 	SupportNumber:             "123",
				// 	MaintenanceProvider:       "oracle",
				// })
				// if err != nil {
				// 	t.Errorf("Failed to do json marshalling in test %v", err)
				// }
				// e := dgworker.Envelope{Type: dgworker.UpsertAggregatedRights, JSON: jsonData}

				// envolveData, err := json.Marshal(e)
				// if err != nil {
				// 	t.Errorf("Failed to do json marshalling in test  %v", err)
				// }
				// mockQueue.EXPECT().PushJob(ctx, job.Job{
				// 	Type:   sql.NullString{String: "aw"},
				// 	Status: job.JobStatusPENDING,
				// 	Data:   envolveData,
				// }, "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.AggregatedRightsResponse{
				Success: false,
			},
			wantErr: true,
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.CreateAggregatedRights(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.CreateAggregatedRights() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.CreateAggregatedRights() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_DeleteAggregatedRights(t *testing.T) {
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
		s       *ProductServiceServer
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.DeleteAggregatedRights(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DeleteAggregatedRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DeleteAggregatedRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_DownloadAggregatedRightsFile(t *testing.T) {
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
		s       *ProductServiceServer
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.DownloadAggregatedRightsFile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_UpdateAggregatedRights(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := dbmock.NewMockProduct(mockCtrl)
	mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
	mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)

	req := &v1.AggregatedRightsRequest{
		Sku:        "sku1",
		Scope:      "scope1",
		MetricName: "met1",
	}

	want := &v1.AggregatedRightsResponse{
		Success: true,
	}
	wanterr := false

	mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).Return(db.GetAggregatedRightBySKURow{Sku: "sku1"}, nil)
	mockRepo.EXPECT().UpsertAggregatedRights(ctx, gomock.Any()).Return(nil)
	mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(1), nil)
	mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
	mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
	mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
	mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
	mockRepo.EXPECT().GetAggregationByID(ctx, gomock.Any()).Times(1).Return(db.Aggregation{
		ID:    1,
		Scope: "scope1",
	}, nil)
	mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
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

	mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Times(1).Return(int32(1000), nil)

	svc := &ProductServiceServer{
		ProductRepo: mockRepo,
		queue:       mockQueue,
		metric:      mockMetric,
	}

	got, err := svc.UpdateAggregatedRights(ctx, req)
	if (err != nil) != wanterr {
		t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, wanterr)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() = %v, want %v", got, want)
	}

	ctx1 := context.Background()
	wanterr = true
	got, err = svc.UpdateAggregatedRights(ctx1, req)
	if (err != nil) != wanterr {
		t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, wanterr)
		return
	}

	req.Scope = "na"
	wanterr = true
	got, err = svc.UpdateAggregatedRights(ctx, req)
	if (err != nil) != wanterr {
		t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, wanterr)
		return
	}

	// wanterr = true
	// mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).Return(db.GetAggregatedRightBySKURow{Sku: "sku1"}, sql.ErrNoRows)
	// got, err = svc.UpdateAggregatedRights(ctx, req)
	// if (err != nil) != wanterr {
	// 	t.Errorf("ProductServiceServer.DownloadAggregatedRightsFile() error = %v, wantErr %v", err, wanterr)
	// 	return
	// }

}

func TestProductServiceServer_CreateAggregationIfNotExists(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := dbmock.NewMockProduct(mockCtrl)
	q := queuemock.NewMockWorkerqueue(mockCtrl)
	m := metmock.NewMockMetricServiceClient(mockCtrl)

	s := &ProductServiceServer{
		ProductRepo: mockRepo,
		queue:       q,
		metric:      m,
	}

	tests := []struct {
		name  string
		setup func()
		args  struct {
			ctx           context.Context
			senderScope   string
			receiverScope string
			aggName       string
		}
		wantErr bool
	}{
		{
			name:    "AggregationNotExists",
			wantErr: true,
			setup: func() {
				// Set up the mock expectations for GetAggregationByName
				mockRepo.EXPECT().GetAggregationByName(ctx, gomock.Any()).Times(1).Return(db.Aggregation{}, sql.ErrNoRows).AnyTimes()

				mockRepo.EXPECT().GetAggregationByName(ctx, gomock.Any()).Times(1).Return(db.Aggregation{
					Swidtags:      []string{"swidtag1", "swidtag2"},
					ProductEditor: "Editor1",
					Products:      []string{"Product1", "Product2"},
				}, nil).AnyTimes()

				// Set up the mock expectations for CreateProduct
				f := mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any()).Times(1).Return(db.GetProductInformationRow{}, sql.ErrNoRows).AnyTimes()

				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any()).Times(1).Return(db.GetProductInformationRow{
					ProductName:    "Product1",
					ProductEditor:  "Editor1",
					ProductVersion: "1.0",
				}, nil).After(f).AnyTimes()
				s := mockRepo.EXPECT().GetAcqBySwidtag(ctx, gomock.Any()).Times(1).Return(db.Acqright{Metric: "m1"}, sql.ErrNoRows).AnyTimes()
				mockRepo.EXPECT().GetAcqBySwidtag(ctx, gomock.Any()).Times(1).Return(db.Acqright{Metric: "m1"}, nil).After(s).AnyTimes()

				m1 := m.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{{Name: "m1"}}}, nil).AnyTimes()
				m.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{{Name: "m1"}}}, nil).After(m1).AnyTimes()
				mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).AnyTimes()
				// fcall := mockRepo.EXPECT().UpsertProductTx(ctx, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				q.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).AnyTimes()

			},
			args: struct {
				ctx           context.Context
				senderScope   string
				receiverScope string
				aggName       string
			}{
				ctx:           ctx,
				senderScope:   "s1",
				receiverScope: "s2",
				aggName:       "Agg1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := s.CreateAggregationIfNotExists(tt.args.ctx, tt.args.senderScope, tt.args.receiverScope, tt.args.aggName)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAggregationIfNotExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// func TestCreateAcqrights(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockProductRepo := dbmock.NewMockProduct(ctrl)
// 	mockMetric := metmock.NewMockMetricServiceClient(ctrl)
// 	qObj := queuemock.NewMockWorkerqueue(ctrl)
// 	ctx := context.Background()
// 	swidtag := "SWIDTAG123"
// 	senderScope := "sender"
// 	receiverScope := "receiver"

// 	t.Run("NoError", func(t *testing.T) {
// 		// Set up the expected calls in the mock objects
// 		mockProductRepo.EXPECT().GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
// 			Swidtag: swidtag,
// 			Scope:   receiverScope,
// 		}).Return(db.Acqright{}, sql.ErrNoRows)

// 		mockProductRepo.EXPECT().GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
// 			Swidtag: swidtag,
// 			Scope:   senderScope,
// 		}).Return(db.Acqright{
// 			Swidtag:       swidtag,
// 			ProductName:   "Product123",
// 			ProductEditor: "Editor123",
// 		}, nil)

// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{senderScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{
// 			Metrices: []*metv1.Metric{{Name: "m1"}},
// 		}, nil)

// 		// Set up the mock expectations for ListMetrices (receiverScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{receiverScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{}, nil)

// 		// Set up the mock expectations for CreateMetric
// 		mockMetric.EXPECT().CreateMetric(ctx, &metv1.CreateMetricRequest{
// 			Metric:        &metv1.Metric{Name: "m1"},
// 			SenderScope:   senderScope,
// 			RecieverScope: receiverScope,
// 		}).Times(1).Return(nil, nil)

// 		mockMetric.EXPECT().CreateMetric(ctx, &metv1.CreateMetricRequest{
// 			Metric:        &metv1.Metric{Name: "m1"},
// 			SenderScope:   senderScope,
// 			RecieverScope: receiverScope,
// 		}).Times(1).Return(nil, nil)

// 		fcall := mockProductRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).Times(1)

// 		qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
// 		// Create the service with the mocked dependencies
// 		service := &ProductServiceServer{
// 			ProductRepo: mockProductRepo,
// 			metric:      mockMetric,
// 			queue:       qObj,
// 		}

// 		// Call the function being tested
// 		err := service.CreateAcqrights(ctx, swidtag, senderScope, receiverScope)
// 		assert.Equal(t, status.Error(codes.NotFound, "AcRightNotFound"), err)

// 		// Assert that there is no error
// 	})

// t.Run("AcRightNotFound", func(t *testing.T) {
// 	// Set up the expected calls in the mock objects
// 	mockProductRepo.EXPECT().GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
// 		Swidtag: swidtag,
// 		Scope:   receiverScope,
// 	}).Return(nil, sql.ErrNoRows)

// 	mockProductRepo.EXPECT().GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
// 		Swidtag: swidtag,
// 		Scope:   senderScope,
// 	}).Return(nil, sql.ErrNoRows)

// 	// Create the service with the mocked dependencies
// 	service := &ProductServiceServer{
// 		ProductRepo:    mockProductRepo,
// 		productService: mockService,
// 	}

// 	// Call the function being tested
// 	err := service.CreateAcqrights(ctx, swidtag, senderScope, receiverScope)

// 	// Assert that the expected error is returned
// 	assert.Equal(t, status.Error(codes.NotFound, "AcRightNotFound"), err)
// })

// t.Run("DBError", func(t *testing.T) {
// 	// Set up the expected calls in the mock objects
// 	mockProductRepo.EXPECT().GetAcqBySwidtag(ctx, db.GetAcqBySwidtagParams{
// 		Swidtag: swidtag,
// 		Scope:   receiverScope,
// 	}).Return(nil, errors.New("DB error"))

// 	// Create the service with the mocked dependencies
// 	service := &ProductServiceServer{
// 		ProductRepo:    mockProductRepo,
// 		productService: mockService,
// 	}

// 	// Call the function being tested
// 	err := service.CreateAcqrights(ctx, swidtag, senderScope, receiverScope)

// 	// Assert that the expected error is returned
// 	assert.Equal(t, status.Error(codes.Internal, "DBError"), err)
// })
// }
