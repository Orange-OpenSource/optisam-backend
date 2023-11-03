package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

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
func TestDeleteSharedLicenses(t *testing.T) {
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
		req *v1.DeleteSharedLicensesRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.DeleteSharedLicensesResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DeleteSharedLicensesRequest{
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.DeleteSharedLicensesResponse{},
			wantErr: false,
		},
		{name: "db err 1",
			args: args{
				ctx: ctx,
				req: &v1.DeleteSharedLicensesRequest{
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("error"))
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.DeleteSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 2",
			args: args{
				ctx: ctx,
				req: &v1.DeleteSharedLicensesRequest{
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("error"))
			},
			want:    &v1.DeleteSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "ctx not found",
			args: args{
				ctx: context.Background(),
				req: &v1.DeleteSharedLicensesRequest{
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.DeleteSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "ctx not found",
			args: args{
				ctx: ctx,
				req: &v1.DeleteSharedLicensesRequest{
					Scope: "scope1 not found",
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.DeleteSharedLicensesResponse{},
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
			_, err := tt.s.DeleteSharedLicenses(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetEditorExpensesByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpsertAcqRights(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	//dgObj := dgmock.NewMockProduct(mockCtrl)
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
				SupportNumber:           "abc",
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "Context error",
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
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: true,
			ctx:    context.Background(),
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start of maintainace and end are rfc",
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
				SupportNumber:           "abc",
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start of maintainace and end are rfc error",
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
				StartOfMaintenance:      "20a19-08-2s7T09:s58:56sss.0260078afdsZ",
				EndOfMaintenance:        "20d21-01-2s9dT0d9:58d:56.0260ds078Z",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start of maintainace and end are blank",
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
				StartOfMaintenance:      "",
				EndOfMaintenance:        "",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "NumLicencesMaintainance less than 0",
			input: &v1.UpsertAcqRightsRequest{
				Sku:                     "a",
				Swidtag:                 "b",
				ProductName:             "c",
				ProductEditor:           "d",
				MetricType:              "e",
				NumLicensesAcquired:     int32(100),
				NumLicencesMaintainance: int32(-1),
				AvgUnitPrice:            float64(5.0),
				AvgMaintenanceUnitPrice: float64(2.0),
				TotalPurchaseCost:       float64(500.0),
				TotalMaintenanceCost:    float64(20.0),
				TotalCost:               float64(532.0),
				Scope:                   "s1",
				StartOfMaintenance:      "",
				EndOfMaintenance:        "",
				Version:                 "vv",
				SupportNumber:           "abc",
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: " start and end maintanace empty",
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
				StartOfMaintenance:      "",
				EndOfMaintenance:        "",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: true,
			ctx: grpc_middleware.AddClaims(context.Background(), &claims.Claims{
				UserID: "admin@superuser.com",
				Role:   "User",
				Socpes: []string{},
			}),
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start time format1",
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
				StartOfMaintenance:      "1/2/06",
				EndOfMaintenance:        "1/2/06",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start time format2",
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
				StartOfMaintenance:      "02/01/2006",
				EndOfMaintenance:        "02/01/2006",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start time format 02-01-2006",
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
				StartOfMaintenance:      "02-01-2006",
				EndOfMaintenance:        "02-01-2006",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "start time format wrong",
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
				StartOfMaintenance:      "wrong",
				EndOfMaintenance:        "wrong",
				Version:                 "vv",
				SupportNumber:           "abc",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "UpsertAcqRightsWithCompleteData1",
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
				SupportNumber:           "abchghdhasdksahjlafhadfjaldkfhadhfdafkhlkhasfa",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
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
					SupportNumbers:          []string{"abc"},
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
					SupportNumber:           "abc",
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
		{
			name: "ordering date",
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
				SupportNumber:           "abc",
				OrderingDate:            "1/2/06",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).Times(1)

				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "ordering date 02-01-2006",
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
				SupportNumber:           "abc",
				OrderingDate:            "02-01-2006",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).Times(1)

				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "ordering date 2006-01-02T15:04:05",
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
				SupportNumber:           "abc",
				OrderingDate:            "2006-01-02T15:04:05",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).Times(1)

				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
		{
			name: "ordering date err",
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
				SupportNumber:           "abc",
				OrderingDate:            "err",
			},
			output: &v1.UpsertAcqRightsResponse{Success: false},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Return(nil).Times(1)

				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
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
	timestampOrderDate, _ := ptypes.TimestampProto(timeStart)
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	// pmock := prodmock.NewMockProductServiceClient(mockCtrl)
	//dgObj := dgmock.NewMockProduct(mockCtrl)

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
						OrderingDate:                   timestampOrderDate,
						SoftwareProvider:               "abc",
						MaintenanceProvider:            "xyz",
						CorporateSourcingContract:      "pqr",
						SupportNumber:                  "123",
						LastPurchasedOrder:             "def",
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
						SupportNumber:                  "123",
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
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "db error",
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
						OrderingDate:                   timestampOrderDate,
						SoftwareProvider:               "abc",
						MaintenanceProvider:            "xyz",
						CorporateSourcingContract:      "pqr",
						SupportNumber:                  "123",
						LastPurchasedOrder:             "def",
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
						SupportNumber:                  "123",
					},
				},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					SkuAsc:   true,
				}).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				}, errors.New("error")).Times(1)
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "context not found",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1"},
			},
			output: &v1.ListAcqRightsResponse{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					SkuAsc:   true,
				}).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "ListAcqRightsWithCorrectData",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"notfound"},
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
						OrderingDate:                   timestampOrderDate,
						SoftwareProvider:               "abc",
						MaintenanceProvider:            "xyz",
						CorporateSourcingContract:      "pqr",
						SupportNumber:                  "123",
						LastPurchasedOrder:             "def",
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
						SupportNumber:                  "123",
					},
				},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					SkuAsc:   true,
				}).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "ListAcqRightsWithCorrectData filtering key",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.AcqRightsSearchParams{
					OrderingDate: &v1.StringFilter{Filteringkey: "1/2/06"},
				},
				Scopes: []string{"s1"},
			},
			output: &v1.ListAcqRightsResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, gomock.Any()).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "ListAcqRightsWithCorrectData filtering key 1",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.AcqRightsSearchParams{
					OrderingDate: &v1.StringFilter{Filteringkey: "02-01-2006"},
				},
				Scopes: []string{"s1"},
			},
			output: &v1.ListAcqRightsResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, gomock.Any()).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
		{
			name: "ListAcqRightsWithCorrectData filtering key 2",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.AcqRightsSearchParams{
					OrderingDate: &v1.StringFilter{Filteringkey: "2006-01-02T15:04:05.999999999Z"},
				},
				Scopes: []string{"s1"},
			},
			output: &v1.ListAcqRightsResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, gomock.Any()).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:              int64(2),
						Sku:                       "b",
						Swidtag:                   "c",
						ProductEditor:             "d",
						ProductName:               "e",
						Metric:                    "f",
						NumLicensesAcquired:       int32(2),
						NumLicencesMaintainance:   int32(2),
						AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
						AvgUnitPrice:              decimal.NewFromFloat(1),
						TotalMaintenanceCost:      decimal.NewFromFloat(2),
						TotalPurchaseCost:         decimal.NewFromFloat(2),
						TotalCost:                 decimal.NewFromFloat(4),
						StartOfMaintenance:        sql.NullTime{Time: *s, Valid: true},
						EndOfMaintenance:          sql.NullTime{Time: *e, Valid: true},
						Version:                   "vv",
						OrderingDate:              sql.NullTime{Time: *s, Valid: true},
						SoftwareProvider:          "abc",
						MaintenanceProvider:       "xyz",
						CorporateSourcingContract: "pqr",
						SupportNumbers:            []string{"123"},
						LastPurchasedOrder:        "def",
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
				dbObj.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(0), nil).AnyTimes()
				dbObj.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				dbObj.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Return([]db.SharedLicense{}, nil).AnyTimes()
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input, &timeStart, &timeEnd)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.ListAcqRights(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err [%s] ", test.name, err.Error())
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
				// 	return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}
func Test_ProductServiceServer_CreateAcqRight(t *testing.T) {
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
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.AcqRightResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                       "sku1",
					ProductName:               "product name",
					Version:                   "prodversion",
					ProductEditor:             "producteditor",
					MetricName:                "metricNup",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              10,
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance:   5,
					AvgMaintenanceUnitPrice:   5,
					Scope:                     "scope1",
					Comment:                   "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
							Type:        "oracle.nup.standard",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				// mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, db.GetAcqRightMetricsBySwidtagParams{
				// 	Scope:   "scope1",
				// 	Swidtag: "product_name_producteditor_prodversion",
				// }).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				orderingtime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                       "sku1",
					Swidtag:                   "product_name_producteditor_prodversion",
					ProductName:               "product name",
					ProductEditor:             "producteditor",
					Scope:                     "scope1",
					Metric:                    "metricNup",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice:   decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:         decimal.NewFromFloat(float64(200)),
					TotalMaintenanceCost:      decimal.NewFromFloat(float64(25)),
					TotalCost:                 decimal.NewFromFloat(float64(225)),
					CreatedBy:                 "admin@superuser.com",
					StartOfMaintenance:        sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:          sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance:   5,
					Version:                   "prodversion",
					Comment:                   sql.NullString{String: "acqright created from UI", Valid: true},
					OrderingDate:              sql.NullTime{Time: orderingtime, Valid: true},
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumbers:            []string{"123"},
					MaintenanceProvider:       "oracle",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.UpsertAcqRightsRequest{
					Sku:                       "sku1",
					Swidtag:                   "product_name_producteditor_prodversion",
					ProductName:               "product name",
					ProductEditor:             "producteditor",
					MetricType:                "metricNup",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              10,
					AvgMaintenanceUnitPrice:   5,
					TotalPurchaseCost:         200,
					TotalMaintenanceCost:      25,
					TotalCost:                 225,
					Scope:                     "scope1",
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintenance:    5,
					Version:                   "prodversion",
					OrderingDate:              "2020-01-01T10:58:56.026008Z",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             strings.Join([]string{"123"}, ","),
					MaintenanceProvider:       "oracle",
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
		{
			name: "SUCCESS-no maintenance",
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
					Scope:                   "scope1", CorporateSourcingContract: "csc",
					SoftwareProvider:    "oracle",
					LastPurchasedOrder:  "odernum",
					SupportNumber:       "123",
					MaintenanceProvider: "oracle",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
							Type:        "oracle.nup.standard",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				// mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, db.GetAcqRightMetricsBySwidtagParams{
				// 	Scope:   "scope1",
				// 	Swidtag: "product_name_producteditor_prodversion",
				// }).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{
				// 	{
				// 		Sku:    "sku2",
				// 		Metric: "metricNup",
				// 	},
				// }, nil)
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                       "sku1",
					Swidtag:                   "product_name_producteditor_prodversion",
					ProductName:               "product name",
					ProductEditor:             "producteditor",
					Scope:                     "scope1",
					Metric:                    "ops",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              decimal.NewFromFloat(float64(10)),
					TotalPurchaseCost:         decimal.NewFromFloat(float64(200)),
					TotalCost:                 decimal.NewFromFloat(float64(200)),
					AvgMaintenanceUnitPrice:   decimal.NewFromFloat(2),
					TotalMaintenanceCost:      decimal.NewFromFloat(0),
					CreatedBy:                 "admin@superuser.com",
					Version:                   "prodversion",
					Comment:                   sql.NullString{String: "", Valid: true},
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumbers:            []string{"123"},
					MaintenanceProvider:       "oracle",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.UpsertAcqRightsRequest{
					Sku:                       "sku1",
					Swidtag:                   "product_name_producteditor_prodversion",
					ProductName:               "product name",
					ProductEditor:             "producteditor",
					MetricType:                "ops",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              10,
					TotalPurchaseCost:         200,
					TotalCost:                 200,
					AvgMaintenanceUnitPrice:   2,
					Scope:                     "scope1",
					Version:                   "prodversion",
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumber:             strings.Join([]string{"123"}, ","),
					MaintenanceProvider:       "oracle",
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
		{
			name: "FAILURE-ClaimsNotFoundError",
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
		{
			name: "FAILURE-ScopeValidationError",
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
		{
			name: "FAILURE-GetAcqRightBySKU-DBError",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, errors.New("sql: no rows in result set"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE-sku already exists",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE-ListMetrices-ServiceError",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)

				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(nil, errors.New("service error"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE-ListMetrices-MetricNotExists-no metrics exists in scope",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE-ListMetrices-MetricNotExists-metrics exists in scope",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
		{
			name: "FAILURE-unable to parse ordering date",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                 "sku1",
					ProductName:         "product name",
					Version:             "prodversion",
					ProductEditor:       "producteditor",
					MetricName:          "acs",
					NumLicensesAcquired: 20,
					AvgUnitPrice:        10,
					OrderingDate:        "notparsable",
					Scope:               "scope1",
					Comment:             "acqright created from UI",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"scope1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "attribute.counter.standard",
							Name:        "acs",
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
		{
			name: "FAILURE-unable to parse start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "acs",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
							Type:        "attribute.counter.standard",
							Name:        "acs",
							Description: "metric description",
						},
					},
				}, nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name: "FAILURE-unable to parse end time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "metricNup",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
		{
			name: "FAILURE-end time is less than start time",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "metricNup",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
		{
			name: "FAILURE-all or none maintenance fields should be present",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "metricNup",
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
		{
			name: "FAILURE-UpsertAcqRights-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				met = mockMetric
				// mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
				// 	AcqrightSku: "sku1",
				// 	Scope:       "scope1",
				// }).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				// mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
				// mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
				// 	Metrices: []*metv1.Metric{
				// 		{
				// 			Type:        "oracle.processor.standard",
				// 			Name:        "ops",
				// 			Description: "metric description",
				// 		},
				// 		{
				// 			Type:        "NUP",
				// 			Name:        "metricNup",
				// 			Description: "metricNup description",
				// 		},
				// 	},
				// }, nil)
				// mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, db.GetAcqRightMetricsBySwidtagParams{
				// 	Scope:   "scope1",
				// 	Swidtag: "product_name_producteditor_prodversion",
				// }).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{}, nil)
				// mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
			wantErr: true,
		},
		{
			name:    "fail db",
			wantErr: true,
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                       "sku1",
					ProductName:               "product name",
					Version:                   "prodversion",
					ProductEditor:             "producteditor",
					MetricName:                "metricNup",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              10,
					StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
					EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
					NumLicencesMaintainance:   5,
					AvgMaintenanceUnitPrice:   5,
					Scope:                     "scope1",
					Comment:                   "acqright created from UI",
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
				mockRepo.EXPECT().GetAggregatedRightBySKU(ctx, gomock.Any()).AnyTimes().Return(db.GetAggregatedRightBySKURow{}, sql.ErrNoRows)
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
							Type:        "oracle.nup.standard",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil)
				// mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, db.GetAcqRightMetricsBySwidtagParams{
				// 	Scope:   "scope1",
				// 	Swidtag: "product_name_producteditor_prodversion",
				// }).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{}, nil)
				starttime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				endtime, _ := time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				orderingtime, _ := time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                       "sku1",
					Swidtag:                   "product_name_producteditor_prodversion",
					ProductName:               "product name",
					ProductEditor:             "producteditor",
					Scope:                     "scope1",
					Metric:                    "metricNup",
					NumLicensesAcquired:       20,
					AvgUnitPrice:              decimal.NewFromFloat(float64(10)),
					AvgMaintenanceUnitPrice:   decimal.NewFromFloat(float64(5)),
					TotalPurchaseCost:         decimal.NewFromFloat(float64(200)),
					TotalMaintenanceCost:      decimal.NewFromFloat(float64(25)),
					TotalCost:                 decimal.NewFromFloat(float64(225)),
					CreatedBy:                 "admin@superuser.com",
					StartOfMaintenance:        sql.NullTime{Time: starttime, Valid: true},
					EndOfMaintenance:          sql.NullTime{Time: endtime, Valid: true},
					NumLicencesMaintainance:   5,
					Version:                   "prodversion",
					Comment:                   sql.NullString{String: "acqright created from UI", Valid: true},
					OrderingDate:              sql.NullTime{Time: orderingtime, Valid: true},
					CorporateSourcingContract: "csc",
					SoftwareProvider:          "oracle",
					LastPurchasedOrder:        "odernum",
					SupportNumbers:            []string{"123"},
					MaintenanceProvider:       "oracle",
				}).Times(1).Return(errors.New("db error"))
				// // jsonData, err := json.Marshal(dgworker.UpsertAcqRightsRequest{
				// // 	Sku:                       "sku1",
				// // 	Swidtag:                   "product_name_producteditor_prodversion",
				// // 	ProductName:               "product name",
				// // 	ProductEditor:             "producteditor",
				// // 	MetricType:                "metricNup",
				// // 	NumLicensesAcquired:       20,
				// // 	AvgUnitPrice:              10,
				// // 	AvgMaintenanceUnitPrice:   5,
				// // 	TotalPurchaseCost:         200,
				// // 	TotalMaintenanceCost:      25,
				// // 	TotalCost:                 225,
				// // 	Scope:                     "scope1",
				// // 	StartOfMaintenance:        "2020-01-01T10:58:56.026008Z",
				// // 	EndOfMaintenance:          "2023-01-01T05:40:56.026008Z",
				// // 	NumLicencesMaintenance:    5,
				// // 	Version:                   "prodversion",
				// // 	OrderingDate:              "2020-01-01T10:58:56.026008Z",
				// // 	CorporateSourcingContract: "csc",
				// // 	SoftwareProvider:          "oracle",
				// // 	LastPurchasedOrder:        "odernum",
				// // 	SupportNumber:             strings.Join([]string{"123"}, ","),
				// // 	MaintenanceProvider:       "oracle",
				// // })
				// if err != nil {
				// 	t.Errorf("Failed to do json marshalling in test %v", err)
				// }
				// e := dgworker.Envelope{Type: dgworker.UpsertAcqRights, JSON: jsonData}

				// envelopeData, err := json.Marshal(e)
				// if err != nil {
				// 	t.Errorf("Failed to do json marshalling in test  %v", err)
				// }
				// mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Times(1).Return(int32(1000), nil)
			},
			want: &v1.AcqRightResponse{
				Success: false,
			},
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
			got, err := tt.s.CreateAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.CreateAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.CreateAcqRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_UpdateAcqRight(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2", "scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	var met metv1.MetricServiceClient
	//var prodS *ProductServiceServer
	type args struct {
		ctx context.Context
		req *v1.AcqRightRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
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
				}).Times(1).Return(db.GetAcqRightBySKURow{
					Sku:    "sku1",
					Metric: "ops,metricNup",
				}, nil)
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
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
				mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, gomock.Any()).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{}, nil)
				_, _ = time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				_, _ = time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.UpsertAcqRightsRequest{
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

				_, err = json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Times(1).Return(int32(1000), nil)
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
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
				mockRepo.EXPECT().GetAcqRightMetricsBySwidtag(ctx, gomock.Any()).Times(1).Return([]db.GetAcqRightMetricsBySwidtagRow{}, nil)
				mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.UpsertAcqRightsRequest{
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

				_, err = json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Times(1).Return(int32(1000), nil)
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, errors.New("Internal"))
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
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
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
					MetricName:              "metricNup",
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
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
					MetricName:              "metricNup",
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
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
					MetricName:              "metricNup",
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
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
					MetricName:              "metricNup",
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{Sku: "sku1"}, nil)
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
		{name: "FAILURE-UpsertAcqRights-DBError",
			args: args{
				ctx: ctx,
				req: &v1.AcqRightRequest{
					Sku:                     "sku1",
					ProductName:             "product name",
					Version:                 "prodversion",
					ProductEditor:           "producteditor",
					MetricName:              "metricNup",
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(0), nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil)
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil)

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
				_, _ = time.Parse(time.RFC3339Nano, "2020-01-01T10:58:56.026008Z")
				_, _ = time.Parse(time.RFC3339Nano, "2023-01-01T05:40:56.026008Z")
				mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any()).Times(1).Return(errors.New("Internal"))
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			_, err := tt.s.UpdateAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.UpdateAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("ProductServiceServer.UpdateAcqRight() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func Test_ProductServiceServer_DeleteAcqRight(t *testing.T) {
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
		s       *ProductServiceServer
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
				mockRepo.EXPECT().DeleteSharedLicences(ctx, db.DeleteSharedLicencesParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(nil)
				mockRepo.EXPECT().DeleteAcqrightBySKU(ctx, db.DeleteAcqrightBySKUParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(nil)
				jsonData, err := json.Marshal(dgworker.DeleteAcqRightRequest{
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
		{
			name: "FAILURE-DBError",
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
				mockRepo.EXPECT().DeleteSharedLicences(ctx, db.DeleteSharedLicencesParams{
					Sku:   "sku1",
					Scope: "scope1",
				}).Times(1).Return(nil)
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
		{
			name: "FAILURE-DBError",
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
				mockRepo.EXPECT().DeleteSharedLicences(ctx, db.DeleteSharedLicencesParams{
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
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.DeleteAcqRight(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DeleteAcqRight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DeleteAcqRight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_DownloadAcqRightFile(t *testing.T) {
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
		req *v1.DownloadAcqRightFileRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.DownloadAcqRightFileResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{
					Sku:      "sku1",
					Metric:   "ops,metricNup",
					FileName: "sku1_file.pdf",
				}, nil)
				mockRepo.EXPECT().GetAcqRightFileDataBySKU(ctx, db.GetAcqRightFileDataBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return([]byte("filedata"), nil)
			},
			want: &v1.DownloadAcqRightFileResponse{
				FileData: []byte("filedata"),
			},
		},
		{name: "FAILURE-ClaimsNotFoundError",
			args: args{
				ctx: context.Background(),
				req: &v1.DownloadAcqRightFileRequest{
					Sku:   "sku1",
					Scope: "scope1",
				},
			},
			setup:   func() {},
			want:    &v1.DownloadAcqRightFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
					Sku:   "sku1",
					Scope: "scope5",
				},
			},
			setup:   func() {},
			want:    &v1.DownloadAcqRightFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-SKU does not exist",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{}, sql.ErrNoRows)
			},
			want:    &v1.DownloadAcqRightFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{}, errors.New("internal"))
			},
			want:    &v1.DownloadAcqRightFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-Acquired Right does not contain file",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{
					Sku:    "sku1",
					Metric: "ops,metricNup",
				}, nil)
			},
			want:    &v1.DownloadAcqRightFileResponse{},
			wantErr: true,
		},
		{name: "FAILURE-GetAcqRightFileDataBySKU-DBError",
			args: args{
				ctx: ctx,
				req: &v1.DownloadAcqRightFileRequest{
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
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return(db.GetAcqRightBySKURow{
					Sku:      "sku1",
					Metric:   "ops,metricNup",
					FileName: "sku1_file.pdf",
				}, nil)
				mockRepo.EXPECT().GetAcqRightFileDataBySKU(ctx, db.GetAcqRightFileDataBySKUParams{
					AcqrightSku: "sku1",
					Scope:       "scope1",
				}).Times(1).Return([]byte(""), errors.New("internal"))
			},
			want:    &v1.DownloadAcqRightFileResponse{},
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
			got, err := tt.s.DownloadAcqRightFile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DownloadAcqRightFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DownloadAcqRightFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProductServiceServer_GetMaintenanceBySwidtag(t *testing.T) {
	timeStart := time.Now()
	timeEnd := timeStart.Add(10 * time.Hour)
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GetMaintenanceBySwidtagRequest
		output *v1.GetMaintenanceBySwidtagResponse
		mock   func(*v1.GetMaintenanceBySwidtagRequest)
		ctx    context.Context
		outErr bool
	}{

		{
			name:   "GetMaintenanceBySwidtagRequest without context",
			input:  &v1.GetMaintenanceBySwidtagRequest{Scope: "MON", Acqsku: "sku1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GetMaintenanceBySwidtagRequest) {},
		},
		{
			name:  "GetMaintenanceBySwidtag with data",
			input: &v1.GetMaintenanceBySwidtagRequest{Scope: "s1", Acqsku: "sku1"},
			ctx:   ctx,
			mock: func(data *v1.GetMaintenanceBySwidtagRequest) {
				dbObj.EXPECT().GetAcqRightBySKU(ctx, db.GetAcqRightBySKUParams{
					Scope:       "s1",
					AcqrightSku: "sku1",
				}).Return(db.GetAcqRightBySKURow{
					Sku:                       "sku1",
					ProductEditor:             "d",
					ProductName:               "e",
					Metric:                    "f",
					NumLicensesAcquired:       int32(2),
					NumLicencesMaintainance:   int32(2),
					AvgMaintenanceUnitPrice:   decimal.NewFromFloat(1),
					AvgUnitPrice:              decimal.NewFromFloat(1),
					TotalMaintenanceCost:      decimal.NewFromFloat(2),
					TotalPurchaseCost:         decimal.NewFromFloat(2),
					TotalCost:                 decimal.NewFromFloat(4),
					StartOfMaintenance:        sql.NullTime{Time: timeStart, Valid: true},
					EndOfMaintenance:          sql.NullTime{Time: timeEnd, Valid: true},
					Version:                   "vv",
					OrderingDate:              sql.NullTime{Time: timeStart, Valid: true},
					SoftwareProvider:          "abc",
					MaintenanceProvider:       "xyz",
					CorporateSourcingContract: "pqr",
					SupportNumbers:            []string{"123"},
					LastPurchasedOrder:        "def",
				}, nil).Times(1)
			},
			output: &v1.GetMaintenanceBySwidtagResponse{
				Success: true,
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.GetMaintenanceBySwidtag(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}
func TestUpdateAcqrightsSharedLicenses(t *testing.T) {
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
		req *v1.UpdateSharedLicensesRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		setup2  func()
		want    *v1.UpdateSharedLicensesResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
					Scope:       "scope1",
					LicenseData: []*v1.UpdateSharedLicensesRequest_SharedLicenses{{SharedLicenses: int32(0)}},
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()

				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: false,
		},
		{name: "ctx not found",
			args: args{
				ctx: context.Background(),
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "ctx scope not found",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
					Scope: "scope1 not found",
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
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 1",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), errors.New("db error")).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 2",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 3",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("some error")).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 4",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("some error")).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "db err 5",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, errors.New("some error")).AnyTimes()
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(errors.New("some error")).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
		{name: "create metric",
			args: args{
				ctx: ctx,
				req: &v1.UpdateSharedLicensesRequest{
					Scope:       "scope1",
					LicenseData: []*v1.UpdateSharedLicensesRequest_SharedLicenses{{SharedLicenses: int32(0)}},
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
				mockRepo.EXPECT().GetAvailableAggLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				mockRepo.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Times(1).Return(int32(1), nil).AnyTimes()
				first := mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{}, nil)
				mockRepo.EXPECT().GetTotalSharedLicenses(ctx, gomock.Any()).Times(1).Return(db.GetTotalSharedLicensesRow{}, nil).AnyTimes()
				mockRepo.EXPECT().GetSharedLicenses(ctx, gomock.Any()).Times(1).Return([]db.SharedLicense{{}}, nil).AnyTimes()
				mockRepo.EXPECT().GetAcqRightBySKU(ctx, gomock.Any()).Times(1).Return(db.GetAcqRightBySKURow{Metric: "ops"}, sql.ErrNoRows).AnyTimes().After(first)
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "oracle.processor.standard",
							Name:        "ops",
							Description: "metric description",
						},
						{
							Type:        "oracle.nup.standard",
							Name:        "metricNup",
							Description: "metricNup description",
						},
					}}, nil).AnyTimes()

				mockRepo.EXPECT().UpsertAcqRights(ctx, gomock.Any())
				mockQueue.EXPECT().PushJob(ctx, gomock.Any(), "aw").Times(1).Return(int32(1000), nil)
				mockRepo.EXPECT().UpsertSharedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
				mockRepo.EXPECT().UpsertRecievedLicenses(ctx, gomock.Any()).Times(1).Return(nil).AnyTimes()
			},
			want:    &v1.UpdateSharedLicensesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			// tt.s = prod
			tt.s = &ProductServiceServer{
				ProductRepo: rep,
				queue:       queue,
				metric:      met,
			}
			_, err := tt.s.UpdateAcqrightsSharedLicenses(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetEditorExpensesByScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_ProductServiceServer_GetMetric(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GetMetricRequest
		output *v1.GetMetricResponse
		mock   func(*v1.GetMetricRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:   "GetMetricRequest without context",
			input:  &v1.GetMetricRequest{Scope: "MON", Sku: "sku1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GetMetricRequest) {},
		},
		{
			name:  "GetMetric with data",
			input: &v1.GetMetricRequest{Scope: "s1", Sku: "sku1"},
			ctx:   ctx,
			mock: func(data *v1.GetMetricRequest) {
				dbObj.EXPECT().GetMetricsBySku(ctx, db.GetMetricsBySkuParams{
					Scope: "s1",
					Sku:   "sku1",
				}).Return(db.GetMetricsBySkuRow{}, nil).Times(1)
			},
			output: &v1.GetMetricResponse{},
		},
		{
			name:  "db err",
			input: &v1.GetMetricRequest{Scope: "s1", Sku: "sku1"},
			ctx:   ctx,
			mock: func(data *v1.GetMetricRequest) {
				dbObj.EXPECT().GetMetricsBySku(ctx, db.GetMetricsBySkuParams{
					Scope: "s1",
					Sku:   "sku1",
				}).Return(db.GetMetricsBySkuRow{}, sql.ErrNoRows).Times(1)
			},
			output: &v1.GetMetricResponse{},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, nil)
			_, err := s.GetMetric(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

// func TestCreateMetricIfNotExists(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockMetric := metmock.NewMockMetricServiceClient(ctrl)

// 	server := &ProductServiceServer{
// 		metric: mockMetric,
// 	}

// 	senderScope := "SenderScope"
// 	receiverScope := "ReceiverScope"
// 	metric := "Metric1,Metric2"

// 	t.Run("Metrics exist in sender scope", func(t *testing.T) {
// 		// Set up the mock expectations for ListMetrices (senderScope)
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

// 		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, metric)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Metrics do not exist in sender scope", func(t *testing.T) {
// 		// Set up the mock expectations for ListMetrices (senderScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{senderScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{}, errors.New("some error"))

// 		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, metric)
// 		assert.EqualError(t, err, "ServiceError")
// 	})

// 	t.Run("Metrics exist in both sender and receiver scope", func(t *testing.T) {
// 		// Set up the mock expectations for ListMetrices (senderScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{senderScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{
// 			Metrices: []*metv1.Metric{{Name: "m1"}},
// 		},
// 			nil)

// 		// Set up the mock expectations for ListMetrices (receiverScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{receiverScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{
// 			Metrices: []*metv1.Metric{{Name: "m1"}},
// 		},
// 			nil)
// 		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, metric)
// 		assert.NoError(t, err)
// 	})

// 	t.Run("Metrics do not exist in receiver scope", func(t *testing.T) {
// 		// Set up the mock expectations for ListMetrices (senderScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{senderScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{
// 			Metrices: []*metv1.Metric{{Name: "m1"}},
// 		}, nil)

// 		// Set up the mock expectations for ListMetrices (receiverScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{receiverScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{}, errors.New("some error"))

// 		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, metric)
// 		assert.EqualError(t, err, "ServiceError")
// 	})

// 	t.Run("Metric does not exist in sender scope", func(t *testing.T) {
// 		// Set up the mock expectations for ListMetrices (senderScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{senderScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{
// 			Metrices: []*metv1.Metric{
// 				&metv1.Metric{Name: "m1"},
// 			},
// 		}, nil)

// 		// Set up the mock expectations for ListMetrices (receiverScope)
// 		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
// 			Scopes: []string{receiverScope},
// 		}).Times(1).Return(&metv1.ListMetricResponse{}, nil)

// 		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "Metric3")
// 		assert.EqualError(t, err, "MetricNotExists")
// 	})
// }

func TestCreateMetricIfNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetric := metmock.NewMockMetricServiceClient(ctrl)

	server := &ProductServiceServer{
		metric: mockMetric,
	}

	ctx := context.Background()
	senderScope := "SenderScope"
	receiverScope := "ReceiverScope"

	t.Run("MetricsExistInSenderScope_OnlyCreateInReceiverScope", func(t *testing.T) {
		senderMetrics := []*metv1.Metric{
			{Name: "m1"},
			{Name: "m2"},
		}
		receiverMetrics := []*metv1.Metric{
			{Name: "m1"},
		}

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{senderScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: senderMetrics,
		}, nil)

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{receiverScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: receiverMetrics,
		}, nil)

		mockMetric.EXPECT().CreateMetric(ctx, &metv1.CreateMetricRequest{
			Metric:        senderMetrics[1],
			SenderScope:   senderScope,
			RecieverScope: receiverScope,
		}).Return(nil, nil)

		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "m2")
		assert.NoError(t, err)
	})

	t.Run("MetricsNotExistInSenderScope", func(t *testing.T) {
		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{senderScope},
		}).Return(&metv1.ListMetricResponse{}, errors.New("some error"))

		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "m1,m2")
		assert.EqualError(t, err, "rpc error: code = Internal desc = ServiceError")
	})

	t.Run("MetricsExistInBothSenderAndReceiverScope", func(t *testing.T) {
		senderMetrics := []*metv1.Metric{
			{Name: "m1"},
		}
		receiverMetrics := []*metv1.Metric{
			{Name: "m1"},
		}

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{senderScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: senderMetrics,
		}, nil)

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{receiverScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: receiverMetrics,
		}, nil)

		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "m1")
		assert.NoError(t, err)
	})

	t.Run("MetricsNotExistInReceiverScope", func(t *testing.T) {
		senderMetrics := []*metv1.Metric{
			{Name: "m1"},
		}

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{senderScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: senderMetrics,
		}, nil)

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{receiverScope},
		}).Return(&metv1.ListMetricResponse{}, errors.New("some error"))

		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "m1")
		assert.EqualError(t, err, "rpc error: code = Internal desc = ServiceError")
	})

	t.Run("MetricDoesNotExistInSenderScope", func(t *testing.T) {
		senderMetrics := []*metv1.Metric{
			{Name: "m1"},
		}

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{senderScope},
		}).Return(&metv1.ListMetricResponse{
			Metrices: senderMetrics,
		}, nil)

		mockMetric.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
			Scopes: []string{receiverScope},
		}).Return(&metv1.ListMetricResponse{}, nil)

		err := server.CreateMetricIfNotExists(ctx, senderScope, receiverScope, "m3")
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = MetricNotExists")
	})
}
