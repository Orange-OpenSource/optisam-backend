package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"optisam-backend/common/optisam/logger"
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
	"optisam-backend/product-service/pkg/worker/dgraph"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func getJob(input interface{}, jtype dgworker.MessageType) (json.RawMessage, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	e := dgworker.Envelope{Type: jtype, JSON: jsonData}
	envolveData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return envolveData, nil
}

func TestUpsertAcqRights(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.UpsertAcqRightsRequest
		output *v1.UpsertAcqRightsResponse
		mock   func(*v1.UpsertAcqRightsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "UpsertAcqRightsWithCompleteData",
			input: &v1.UpsertAcqRightsRequest{
				Sku:                     "a",
				Swidtag:                 "b",
				ProductName:             "c",
				ProductEditor:           "d",
				MetricType:              "e",
				NumLicensesAcquired:     int32(100),
				NumLicencesMaintainance: int32(10),
				AvgUnitPrice:            float64(5.0),
				AvgMaintenanceUnitPrice: float64(2.0),
				TotalPurchaseCost:       float64(500.0),
				TotalMaintenanceCost:    float64(20.0),
				TotalCost:               float64(532.0),
				Scope:                   "s1",
				StartOfMaintenance:      "2019-08-27T09:58:56.0260078Z",
				EndOfMaintenance:        "2021-01-29T09:58:56.0260078Z",
				Version:                 "vv",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				startOfMaintenance := sql.NullTime{Valid: false}
				endOfMaintenance := sql.NullTime{Valid: false}
				startTime, err1 := time.Parse(time.RFC3339Nano, input.StartOfMaintenance)
				endTime, err2 := time.Parse(time.RFC3339Nano, input.EndOfMaintenance)
				if err1 == nil {
					startOfMaintenance = sql.NullTime{Time: startTime, Valid: true}
				}
				if err2 == nil {
					endOfMaintenance = sql.NullTime{Time: endTime, Valid: true}
				}
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     input.Sku,
					Swidtag:                 input.Swidtag,
					ProductName:             input.ProductName,
					ProductEditor:           input.ProductEditor,
					Metric:                  input.MetricType,
					NumLicensesAcquired:     input.NumLicensesAcquired,
					NumLicencesMaintainance: input.NumLicencesMaintainance,
					AvgUnitPrice:            decimal.NewFromFloat(input.AvgUnitPrice),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(input.AvgMaintenanceUnitPrice),
					TotalPurchaseCost:       decimal.NewFromFloat(input.TotalPurchaseCost),
					TotalMaintenanceCost:    decimal.NewFromFloat(input.TotalMaintenanceCost),
					TotalCost:               decimal.NewFromFloat(input.TotalCost),
					Scope:                   input.Scope,
					StartOfMaintenance:      startOfMaintenance,
					EndOfMaintenance:        endOfMaintenance,
					Version:                 input.Version,
					CreatedBy:               "admin@superuser.com",
				}).Return(nil).Times(1)

				eData, err := getJob(dgworker.UpsertAcqRightsRequest{
					Sku:                     input.Sku,
					Swidtag:                 input.Swidtag,
					ProductName:             input.ProductName,
					ProductEditor:           input.ProductEditor,
					MetricType:              input.MetricType,
					NumLicensesAcquired:     input.NumLicensesAcquired,
					NumLicencesMaintenance:  input.NumLicencesMaintainance,
					AvgMaintenanceUnitPrice: input.AvgMaintenanceUnitPrice,
					AvgUnitPrice:            input.AvgUnitPrice,
					TotalPurchaseCost:       input.TotalPurchaseCost,
					TotalMaintenanceCost:    input.TotalMaintenanceCost,
					TotalCost:               input.TotalCost,
					Scope:                   input.Scope,
					StartOfMaintenance:      input.StartOfMaintenance,
					EndOfMaintenance:        input.EndOfMaintenance,
					Version:                 input.Version,
				}, dgworker.UpsertAcqRights)
				if err != nil {
					t.Errorf("Test cases has beed modiefied or test data has been modified")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   eData,
				}, "aw").Return(int32(1), nil).After(fcall)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.UpsertAcqRights(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListAcqRights(t *testing.T) {
	timeStart := time.Now()
	timeEnd := timeStart.Add(10 * time.Hour)
	timestampStart, _ := ptypes.TimestampProto(timeStart)
	timestampEnd, _ := ptypes.TimestampProto(timeEnd)
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsRequest
		output *v1.ListAcqRightsResponse
		mock   func(*v1.ListAcqRightsRequest, *time.Time, *time.Time)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "ListAcqRightsWithCorrectData",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1"},
			},
			output: &v1.ListAcqRightsResponse{
				TotalRecords: int32(2),
				AcquiredRights: []*v1.AcqRights{
					{
						SKU:                            "b",
						SwidTag:                        "c",
						Editor:                         "d",
						ProductName:                    "e",
						Metric:                         "f",
						AcquiredLicensesNumber:         int32(2),
						LicensesUnderMaintenanceNumber: int32(2),
						AvgLicenesUnitPrice:            float64(1),
						AvgMaintenanceUnitPrice:        float64(1),
						TotalPurchaseCost:              float64(2),
						TotalMaintenanceCost:           float64(2),
						TotalCost:                      float64(4),
						StartOfMaintenance:             timestampStart,
						EndOfMaintenance:               timestampEnd,
						LicensesUnderMaintenance:       "yes",
						Version:                        "vv",
					},
					{
						SKU:                            "b2",
						SwidTag:                        "c2",
						Editor:                         "d2",
						ProductName:                    "e2",
						Metric:                         "f2",
						AcquiredLicensesNumber:         int32(3),
						LicensesUnderMaintenanceNumber: int32(3),
						AvgLicenesUnitPrice:            float64(1),
						AvgMaintenanceUnitPrice:        float64(1),
						TotalPurchaseCost:              float64(3),
						TotalMaintenanceCost:           float64(3),
						TotalCost:                      float64(6),
						LicensesUnderMaintenance:       "no",
						Version:                        "vv1",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					SkuAsc:   true,
				}).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:            int64(2),
						Sku:                     "b",
						Swidtag:                 "c",
						ProductEditor:           "d",
						ProductName:             "e",
						Metric:                  "f",
						NumLicensesAcquired:     int32(2),
						NumLicencesMaintainance: int32(2),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(1),
						AvgUnitPrice:            decimal.NewFromFloat(1),
						TotalMaintenanceCost:    decimal.NewFromFloat(2),
						TotalPurchaseCost:       decimal.NewFromFloat(2),
						TotalCost:               decimal.NewFromFloat(4),
						StartOfMaintenance:      sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:        sql.NullTime{Time: *e, Valid: true},
						Version:                 "vv",
					},
					{
						Totalrecords:            int64(2),
						Sku:                     "b2",
						Swidtag:                 "c2",
						ProductEditor:           "d2",
						ProductName:             "e2",
						Metric:                  "f2",
						NumLicensesAcquired:     int32(3),
						NumLicencesMaintainance: int32(3),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(1),
						AvgUnitPrice:            decimal.NewFromFloat(1),
						TotalMaintenanceCost:    decimal.NewFromFloat(3),
						TotalPurchaseCost:       decimal.NewFromFloat(3),
						TotalCost:               decimal.NewFromFloat(6),
						Version:                 "vv1",
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsWithputContext",
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsRequest, s *time.Time, es *time.Time) {},
			outErr: true,
		},
		// {
		// 	name: "FAILURE: User does not have access to the scope",
		// 	ctx:  ctx,
		// 	input: &v1.ListAcqRightsRequest{
		// 		Scopes: []string{"s4"},
		// 	},
		// 	mock:   func(*v1.ListAcqRightsRequest) {},
		// 	outErr: true,
		// },
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input, &timeStart, &timeEnd)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.ListAcqRights(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err [%s] ", test.name, err.Error())
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func Test_productServiceServer_CreateAcqRight(t *testing.T) {
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
		req *v1.AcqRightRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AcqRightResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops,metricNup",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops,metricNup",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(200)),
					TotalMaintenanceCost:    decimal.NewFromFloat(float64(25)),
					TotalCost:               decimal.NewFromFloat(float64(225)),
					CreatedBy:               "admin@superuser.com",
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance: 5,
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "acqright created from UI", Valid: true},
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgraph.UpsertAcqRightsRequest{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					MetricType:              "ops,metricNup",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					AvgMaintenanceUnitPrice: 5,
					TotalPurchaseCost:       200,
					TotalMaintenanceCost:    25,
					TotalCost:               225,
					Scope:                   "scope1",
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:  5,
					Version:                 "prodversion",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

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
			want: &v1.AcqRightResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-no maintenance",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgMaintenanceUnitPrice: 2,
					AvgUnitPrice:            10,
					Scope:                   "scope1",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(200)),
					TotalCost:               decimal.NewFromFloat(float64(200)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
					TotalMaintenanceCost:    decimal.NewFromFloat(0),
					CreatedBy:               "admin@superuser.com",
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "", Valid: true},
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgraph.UpsertAcqRightsRequest{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					MetricType:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					TotalPurchaseCost:       200,
					TotalCost:               200,
					AvgMaintenanceUnitPrice: 2,
					Scope:                   "scope1",
					Version:                 "prodversion",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

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
			want: &v1.AcqRightResponse{
				Success: true,
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
				},
			},
			setup: func() {},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope5",
					Comment:                 "acqright created from UI",
				},
			},
			setup: func() {},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, errors.New("Internal"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-sku already exists",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-ServiceError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(nil, errors.New("service error"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-MetricNotExists-no metrics exists in scope",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-MetricNotExists-metrics exists in scope",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "sag",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-unable to parse start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "notparsable",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-unable to parse end time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "notparsable",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-end time is less than start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2019-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-all or none maintenance fields should be present",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 0,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-InsertAcqRight-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(200)),
					TotalMaintenanceCost:    decimal.NewFromFloat(float64(25)),
					TotalCost:               decimal.NewFromFloat(float64(225)),
					CreatedBy:               "admin@superuser.com",
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance: 5,
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "acqright created from UI", Valid: true},
				}).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.AcqRightResponse{
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
			got, err := tt.s.CreateAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.CreateAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.CreateAcqRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_UpdateAcqRight(t *testing.T) {
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
		req *v1.AcqRightRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.AcqRightResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     10,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{
					Sku:    "sku1",
					Metric: "ops,metricNup",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops",
					NumLicensesAcquired:     10,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(100)),
					TotalMaintenanceCost:    decimal.NewFromFloat(float64(25)),
					TotalCost:               decimal.NewFromFloat(float64(125)),
					CreatedBy:               "admin@superuser.com",
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance: 5,
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "acqright created from UI", Valid: true},
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgraph.UpsertAcqRightsRequest{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					MetricType:              "ops",
					NumLicensesAcquired:     10,
					AvgUnitPrice:            10,
					AvgMaintenanceUnitPrice: 5,
					TotalPurchaseCost:       100,
					TotalMaintenanceCost:    25,
					TotalCost:               125,
					Scope:                   "scope1",
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:  5,
					Version:                 "prodversion",
					IsSwidtagModified:       true,
					IsMetricModifed:         true,
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

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
			want: &v1.AcqRightResponse{
				Success: true,
			},
		},
		{name: "SUCCESS-no maintenance",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                 "sku1",
					ProductName:         "product name",
					Version:             "prodversion",
					ProductEditor:       "producteditor",
					MetricName:          "ops",
					NumLicensesAcquired: 20,
					AvgUnitPrice:        10,
					Scope:               "scope1",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{
					Sku:    "sku1",
					Metric: "ops",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(200)),
					TotalCost:               decimal.NewFromFloat(float64(200)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(0),
					TotalMaintenanceCost:    decimal.NewFromFloat(0),
					CreatedBy:               "admin@superuser.com",
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "", Valid: true},
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgraph.UpsertAcqRightsRequest{
					Sku:                 "sku1",
					Swidtag:             "product_name_producteditor_prodversion",
					ProductName:         "product name",
					ProductEditor:       "producteditor",
					MetricType:          "ops",
					NumLicensesAcquired: 20,
					AvgUnitPrice:        10,
					TotalPurchaseCost:   200,
					TotalCost:           200,
					Scope:               "scope1",
					Version:             "prodversion",
					IsSwidtagModified:   true,
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

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
			want: &v1.AcqRightResponse{
				Success: true,
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
				},
			},
			setup: func() {},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope5",
					Comment:                 "acqright created from UI",
				},
			},
			setup: func() {},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, errors.New("Internal"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-sku does not exist",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{}, sql.ErrNoRows)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-ServiceError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{
					Sku: "sku1",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(nil, errors.New("service error"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-MetricNotExists-no metrics exists in scope",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{
					Sku: "sku1",
				}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ListMetrices-MetricNotExists-metrics exists in scope",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "sag",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-unable to parse start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "notparsable",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-unable to parse end time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "notparsable",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-end time is less than start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2019-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-all or none maintenance fields should be present",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 0,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-InsertAcqRight-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            10,
					StartOfMaintenance:      "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:        "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance: 5,
					AvgMaintenanceUnitPrice: 5,
					Scope:                   "scope1",
					Comment:                 "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.Acqright{Sku: "sku1"}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					},
				}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     "sku1",
					Swidtag:                 "product_name_producteditor_prodversion",
					ProductName:             "product name",
					ProductEditor:           "producteditor",
					Scope:                   "scope1",
					Metric:                  "ops",
					NumLicensesAcquired:     20,
					AvgUnitPrice:            decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice: decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:       decimal.NewFromFloat(float64(200)),
					TotalMaintenanceCost:    decimal.NewFromFloat(float64(25)),
					TotalCost:               decimal.NewFromFloat(float64(225)),
					CreatedBy:               "admin@superuser.com",
					StartOfMaintenance:      sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:        sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance: 5,
					Version:                 "prodversion",
					Comment:                 sql.NullString{String: "acqright created from UI", Valid: true},
				}).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.AcqRightResponse{
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
			got, err := tt.s.UpdateAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.UpdateAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.UpdateAcqRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_DeleteAcqRight(t *testing.T) {
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
		req *v1.DeleteAcqRightRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.DeleteAcqRightResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAcqRightRequest{
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
				mockRepo.EXPECT().DeleteAcqrightBySKU(ctx, db.DeleteAcqrightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgraph.DeleteAcqRightRequest{
					Sku:   "sku1",
					Scope: "scope1",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.DeleteAcqright, JSON: jsonData}

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
			want: &v1.DeleteAcqRightResponse{
				Success: true,
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteAcqRightRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup: func() {},
			want: &v1.DeleteAcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAcqRightRequest{
					Sku:   "sku1",
					Scope: "scope5",
				},
			},
			setup: func() {},
			want: &v1.DeleteAcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DeleteAcqRightRequest{
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
				mockRepo.EXPECT().DeleteAcqrightBySKU(ctx, db.DeleteAcqrightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(errors.New("internal"))
			},
			want: &v1.DeleteAcqRightResponse{
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
			got, err := tt.s.DeleteAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DeleteAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DeleteAcqRight() = %v, want %v", got, tt.want)
			}
		})
	}
}
