// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	v1 "optisam-backend/acqrights-service/pkg/api/v1"
	dbmock "optisam-backend/acqrights-service/pkg/repository/v1/dbmock"
	"optisam-backend/acqrights-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/acqrights-service/pkg/repository/v1/queuemock"
	"optisam-backend/acqrights-service/pkg/rpc"
	"optisam-backend/acqrights-service/pkg/worker"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue/job"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	ctx = ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1"},
	})
)

func getJob(input interface{}, jtype worker.MessageType) (json.RawMessage, error) {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	e := worker.Envelope{Type: jtype, JSON: jsonData}
	envolveData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return envolveData, nil
}

func TestUpsertAcqRights(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
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
				AvgUnitPrice:            float32(5.0),
				AvgMaintenanceUnitPrice: float32(2.0),
				TotalPurchaseCost:       float32(500.0),
				TotalMaintenanceCost:    float32(20.0),
				TotalCost:               float32(532.0),
				Entity:                  "f",
				Scope:                   "s1",
			},
			output: &v1.UpsertAcqRightsResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAcqRightsRequest) {
				fcall := dbObj.EXPECT().UpsertAcqRights(ctx, db.UpsertAcqRightsParams{
					Sku:                     input.Sku,
					Swidtag:                 input.Swidtag,
					ProductName:             input.ProductName,
					ProductEditor:           input.ProductEditor,
					Metric:                  input.MetricType,
					NumLicensesAcquired:     input.NumLicensesAcquired,
					NumLicencesMaintainance: input.NumLicencesMaintainance,
					AvgUnitPrice:            input.AvgUnitPrice,
					AvgMaintenanceUnitPrice: input.AvgMaintenanceUnitPrice,
					TotalPurchaseCost:       input.TotalPurchaseCost,
					TotalMaintenanceCost:    input.TotalMaintenanceCost,
					TotalCost:               input.TotalCost,
					Entity:                  input.Entity,
					Scope:                   input.Scope,
				}).Return(nil).Times(1)

				eData, err := getJob(input, worker.UpsertAcqRightsRequest)
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
			s := NewAcqRightsServiceServer(dbObj, qObj)
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
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsRequest
		output *v1.ListAcqRightsResponse
		mock   func(*v1.ListAcqRightsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "ListAcqRightsWithCorrectData",
			input: &v1.ListAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			output: &v1.ListAcqRightsResponse{
				TotalRecords: int32(2),
				AcquiredRights: []*v1.AcqRights{
					&v1.AcqRights{
						Entity:                         "a",
						SKU:                            "b",
						SwidTag:                        "c",
						Editor:                         "d",
						ProductName:                    "e",
						Metric:                         "f",
						AcquiredLicensesNumber:         int32(2),
						LicensesUnderMaintenanceNumber: int32(2),
						AvgLicenesUnitPrice:            float32(1),
						AvgMaintenanceUnitPrice:        float32(1),
						TotalPurchaseCost:              float32(2),
						TotalMaintenanceCost:           float32(2),
						TotalCost:                      float32(4),
					},
					&v1.AcqRights{
						Entity:                         "a2",
						SKU:                            "b2",
						SwidTag:                        "c2",
						Editor:                         "d2",
						ProductName:                    "e2",
						Metric:                         "f2",
						AcquiredLicensesNumber:         int32(3),
						LicensesUnderMaintenanceNumber: int32(3),
						AvgLicenesUnitPrice:            float32(1),
						AvgMaintenanceUnitPrice:        float32(1),
						TotalPurchaseCost:              float32(3),
						TotalMaintenanceCost:           float32(3),
						TotalCost:                      float32(6),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("claims is missing in testing")
				}
				dbObj.EXPECT().ListAcqRightsIndividual(ctx, db.ListAcqRightsIndividualParams{
					Scope:     userClaims.Socpes,
					PageNum:   input.PageSize * (input.PageNum - 1),
					PageSize:  input.PageSize,
					EntityAsc: true,
				}).Return([]db.ListAcqRightsIndividualRow{
					{
						Totalrecords:            int64(2),
						Entity:                  "a",
						Sku:                     "b",
						Swidtag:                 "c",
						ProductEditor:           "d",
						ProductName:             "e",
						Metric:                  "f",
						NumLicencesMaintainance: int32(2),
						NumLicensesAcquired:     int32(2),
						AvgMaintenanceUnitPrice: float32(1),
						AvgUnitPrice:            float32(1),
						TotalMaintenanceCost:    float32(2),
						TotalPurchaseCost:       float32(2),
						TotalCost:               float32(4),
					},
					{
						Totalrecords:            int64(2),
						Entity:                  "a2",
						Sku:                     "b2",
						Swidtag:                 "c2",
						ProductEditor:           "d2",
						ProductName:             "e2",
						Metric:                  "f2",
						NumLicencesMaintainance: int32(3),
						NumLicensesAcquired:     int32(3),
						AvgMaintenanceUnitPrice: float32(1),
						AvgUnitPrice:            float32(1),
						TotalMaintenanceCost:    float32(3),
						TotalPurchaseCost:       float32(3),
						TotalCost:               float32(6),
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsWithputContext",
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsRequest) {},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.ListAcqRights(test.ctx, test.input)
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

func TestListAcqRightsProducts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsProductsRequest
		output *v1.ListAcqRightsProductsResponse
		mock   func(*v1.ListAcqRightsProductsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name:  "ListAcqRightsProductsWithCorrectData",
			input: &v1.ListAcqRightsProductsRequest{Scope: "s1", Editor: "b", Metric: "c"},
			output: &v1.ListAcqRightsProductsResponse{
				AcqrightsProducts: []*v1.ListAcqRightsProductsResponse_AcqRightsProducts{
					&v1.ListAcqRightsProductsResponse_AcqRightsProducts{
						Swidtag:     "p1",
						ProductName: "p1name",
					},
					&v1.ListAcqRightsProductsResponse_AcqRightsProducts{
						Swidtag:     "p2",
						ProductName: "p2name",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsProductsRequest) {
				dbObj.EXPECT().ListAcqRightsProducts(ctx, db.ListAcqRightsProductsParams{
					Editor: input.Editor,
					Metric: input.Metric,
					Scope:  input.Scope}).Return([]db.ListAcqRightsProductsRow{
					{
						Swidtag:     "p1",
						ProductName: "p1name",
					},
					{
						Swidtag:     "p2",
						ProductName: "p2name",
					}}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsProductsWithScopeMissmatch",
			input:  &v1.ListAcqRightsProductsRequest{Scope: "s2", Editor: "b", Metric: "c"},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListAcqRightsProductsRequest) {},
		},
		{
			name:   "ListAcqRightsProductsWithoutContext",
			input:  &v1.ListAcqRightsProductsRequest{Scope: "s2", Editor: "b", Metric: "c"},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsProductsRequest) {},
		},
		{
			name:   "ListAcqRightsProductsWithEmptyProductList",
			input:  &v1.ListAcqRightsProductsRequest{Scope: "s1", Editor: "b", Metric: "c"},
			output: &v1.ListAcqRightsProductsResponse{AcqrightsProducts: []*v1.ListAcqRightsProductsResponse_AcqRightsProducts{}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsProductsRequest) {
				dbObj.EXPECT().ListAcqRightsProducts(ctx, db.ListAcqRightsProductsParams{
					Editor: input.Editor,
					Metric: input.Metric,
					Scope:  input.Scope}).Return([]db.ListAcqRightsProductsRow{}, nil).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.ListAcqRightsProducts(test.ctx, test.input)
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

func TestListAcqRightsEditors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsEditorsRequest
		output *v1.ListAcqRightsEditorsResponse
		mock   func(*v1.ListAcqRightsEditorsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name:   "ListAcqRightsEditorsWithCorrectData",
			input:  &v1.ListAcqRightsEditorsRequest{Scope: "s1"},
			output: &v1.ListAcqRightsEditorsResponse{Editor: []string{"e1", "e2", "e3"}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsEditorsRequest) {
				dbObj.EXPECT().ListAcqRightsEditors(ctx, input.Scope).Return([]string{"e1", "e2", "e3"}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsEditorsWithScopeMismatch",
			input:  &v1.ListAcqRightsEditorsRequest{Scope: "s2"},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListAcqRightsEditorsRequest) {},
		},
		{
			name:   "ListAcqRightsEditorsWithEpmtyResult",
			input:  &v1.ListAcqRightsEditorsRequest{Scope: "s1"},
			output: &v1.ListAcqRightsEditorsResponse{Editor: []string{}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsEditorsRequest) {
				dbObj.EXPECT().ListAcqRightsEditors(ctx, input.Scope).Return([]string{}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsEditorsWithoutContext",
			input:  &v1.ListAcqRightsEditorsRequest{Scope: "s1"},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsEditorsRequest) {},
		},
		{
			name:   "ListAcqRightsEditorsWithoutScope",
			input:  &v1.ListAcqRightsEditorsRequest{},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListAcqRightsEditorsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.ListAcqRightsEditors(test.ctx, test.input)
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

func TestListAcqRightsMetrics(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsMetricsRequest
		output *v1.ListAcqRightsMetricsResponse
		mock   func(*v1.ListAcqRightsMetricsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name:   "ListAcqRightsMetricsWithCorrectData",
			input:  &v1.ListAcqRightsMetricsRequest{Scope: "s1"},
			output: &v1.ListAcqRightsMetricsResponse{Metric: []string{"m1", "m2", "m3"}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsMetricsRequest) {
				dbObj.EXPECT().ListAcqRightsMetrics(ctx, input.Scope).Return([]string{"m1", "m2", "m3"}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsMetricsWithScopeMistmatch",
			input:  &v1.ListAcqRightsMetricsRequest{Scope: "s2"},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListAcqRightsMetricsRequest) {},
		},
		{
			name:   "ListAcqRightsMetricsWithEmptyScope",
			input:  &v1.ListAcqRightsMetricsRequest{},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListAcqRightsMetricsRequest) {},
		},
		{
			name:   "ListAcqRightsMetricsWithoutContext",
			input:  &v1.ListAcqRightsMetricsRequest{Scope: "s1"},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsMetricsRequest) {},
		},
		{
			name:   "ListAcqRightsMetricsWithEmptyResult",
			input:  &v1.ListAcqRightsMetricsRequest{Scope: "s1"},
			output: &v1.ListAcqRightsMetricsResponse{Metric: []string{}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsMetricsRequest) {
				dbObj.EXPECT().ListAcqRightsMetrics(ctx, input.Scope).Return([]string{}, nil).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.ListAcqRightsMetrics(test.ctx, test.input)
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

func TestCreateProductAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductAggregationMessage
		output *v1.ProductAggregationMessage
		mock   func(*v1.ProductAggregationMessage)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "CreateProductAggregationWithCorrectData",
			input: &v1.ProductAggregationMessage{
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Scope:    "s1",
				Products: []string{"p1", "p2"},
			},
			output: &v1.ProductAggregationMessage{
				ID:       int32(1),
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Products: []string{"p1", "p2"},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductAggregationMessage) {
				fcall := dbObj.EXPECT().InsertAggregation(ctx, db.InsertAggregationParams{
					AggregationName:   input.Name,
					AggregationScope:  input.Scope,
					AggregationMetric: input.Metric,
					Products:          input.Products,
				}).Return(db.Aggregation{
					AggregationID:     int32(1),
					AggregationName:   "agg",
					AggregationMetric: "m",
					Products:          []string{"p1", "p2"},
				}, nil).Times(1)

				envData := rpc.Envelope{Id: int32(1),
					Name:       input.Name,
					Swidtags:   input.Products,
					ActionType: "add"}

				edata, err := json.Marshal(envData)
				if err != nil {
					t.Error("Failed to do json marshalling , something has been changed in test cases")
				}
				sec := qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rpc"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "rpc").Return(int32(1), nil).After(fcall)

				edata, err = getJob(input, worker.UpsertAggregation)
				if err != nil {
					t.Errorf("Something has been changed in testcases")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "aw").Return(int32(2), nil).After(sec)
			},
		},
		{
			name: "CreateAggregationWithScopeMismatch",
			input: &v1.ProductAggregationMessage{
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Scope:    "s2",
				Products: []string{"p1", "p2"},
			},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ProductAggregationMessage) {},
		},
		{
			name: "CreateAggregationWithNoScope",
			input: &v1.ProductAggregationMessage{
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Products: []string{"p1", "p2"},
			},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ProductAggregationMessage) {},
		},
		{
			name: "CreateAggregationWithoutContext",
			input: &v1.ProductAggregationMessage{
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Scope:    "s1",
				Products: []string{"p1", "p2"},
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ProductAggregationMessage) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.CreateProductAggregation(test.ctx, test.input)
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

func TestListProductAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListProductAggregationRequest
		output *v1.ListProductAggregationResponse
		mock   func(*v1.ListProductAggregationRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name:  "ListProductAggregationWithCorrectData",
			input: &v1.ListProductAggregationRequest{},
			output: &v1.ListProductAggregationResponse{
				Aggregations: []*v1.ProductAggregation{
					&v1.ProductAggregation{
						ID:           int32(1),
						Name:         "a1",
						Editor:       "b1",
						Scope:        "c1",
						Metric:       "d1",
						ProductNames: []string{"p1", "p2"},
						Products:     []string{"pp1", "pp2"},
					},
					&v1.ProductAggregation{
						ID:           int32(2),
						Name:         "a2",
						Editor:       "b2",
						Scope:        "c2",
						Metric:       "d2",
						ProductNames: []string{"p3", "p4"},
						Products:     []string{"pp3", "pp4"},
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductAggregationRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("Claim issue in testcases")
				}
				dbObj.EXPECT().ListAggregation(ctx, userClaims.Socpes).Return([]db.ListAggregationRow{
					{
						AggregationID:     int32(1),
						AggregationMetric: "d1",
						AggregationName:   "a1",
						AggregationScope:  "c1",
						ProductEditor:     "b1",
						ProductNames:      []string{"p1", "p2"},
						ProductSwidtags:   []string{"pp1", "pp2"},
					},
					{
						AggregationID:     int32(2),
						AggregationMetric: "d2",
						AggregationName:   "a2",
						AggregationScope:  "c2",
						ProductEditor:     "b2",
						ProductNames:      []string{"p3", "p4"},
						ProductSwidtags:   []string{"pp3", "pp4"},
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListProductAggregationWithoutContext",
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductAggregationRequest) {},
		},
		{
			name:   "ListProductAggregationWithEmptyResult",
			outErr: false,
			output: &v1.ListProductAggregationResponse{
				Aggregations: []*v1.ProductAggregation{},
			},
			ctx: ctx,
			mock: func(input *v1.ListProductAggregationRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("Claim issue in testcases")
				}
				dbObj.EXPECT().ListAggregation(ctx, userClaims.Socpes).Return([]db.ListAggregationRow{}, nil).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.ListProductAggregation(test.ctx, test.input)
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

func TestUpdateProductAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductAggregationMessage
		output *v1.ProductAggregationMessage
		mock   func(*v1.ProductAggregationMessage)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "UpdateAggregationWithCorrectData",
			input: &v1.ProductAggregationMessage{
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Scope:    "s1",
				Products: []string{"p1", "p2"},
			},
			output: &v1.ProductAggregationMessage{
				ID:       int32(1),
				Name:     "agg",
				Editor:   "e",
				Metric:   "m",
				Products: []string{"p1", "p2"},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductAggregationMessage) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("Testcase has been failed ")
				}
				fcall := dbObj.EXPECT().UpdateAggregation(ctx, db.UpdateAggregationParams{
					Scope:           userClaims.Socpes,
					AggregationID:   input.ID,
					AggregationName: input.Name,
					Products:        input.Products,
				}).Return(db.Aggregation{
					AggregationID:     int32(1),
					AggregationName:   "agg",
					AggregationMetric: "m",
					Products:          []string{"p1", "p2"},
				}, nil).Times(1)
				envData := rpc.Envelope{Id: int32(1),
					Name:       input.Name,
					Swidtags:   input.Products,
					ActionType: "upsert"}

				edata, err := json.Marshal(envData)
				if err != nil {
					t.Error("Failed to do json marshalling , something has been changed in test cases")
				}
				sec := qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rpc"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "rpc").Return(int32(1), nil).After(fcall)

				edata, err = getJob(input, worker.UpsertAggregation)
				if err != nil {
					t.Errorf("Something has been changed in testcases")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "aw").Return(int32(2), nil).After(sec)
			},
		},
		{
			name: "UpdateAggregationWithoutContext",
			input: &v1.ProductAggregationMessage{
				Name:     "agg1",
				Editor:   "e1",
				Metric:   "m1",
				Scope:    "s1",
				Products: []string{"p1", "p2"},
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ProductAggregationMessage) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.UpdateProductAggregation(test.ctx, test.input)
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

func TestDeleteProductAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockAcqRights(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.DeleteProductAggregationRequest
		output *v1.DeleteProductAggregationResponse
		mock   func(*v1.DeleteProductAggregationRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "DeleteProductAggregationWithCorrectData",
			input: &v1.DeleteProductAggregationRequest{
				ID:    int32(1),
				Scope: "s1",
			},
			output: &v1.DeleteProductAggregationResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.DeleteProductAggregationRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("Failed in testcases")
				}
				fcall := dbObj.EXPECT().DeleteAggregation(ctx, db.DeleteAggregationParams{
					AggregationID: input.ID,
					Scope:         userClaims.Socpes,
				}).Return(nil).Times(1)
				envData := rpc.Envelope{
					Id:         int32(1),
					Swidtags:   []string{},
					ActionType: "delete"}

				edata, err := json.Marshal(envData)
				if err != nil {
					t.Error("Failed to do json marshalling , something has been changed in test cases")
				}
				sec := qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "rpc"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "rpc").Return(int32(1), nil).After(fcall)

				edata, err = getJob(input, worker.DeleteAggregation)
				if err != nil {
					t.Errorf("Something has been changed in testcases")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   edata,
				}, "aw").Return(int32(2), nil).After(sec)
			},
		},
		{
			name:   "DeleteAggregationWithoutContext",
			input:  &v1.DeleteProductAggregationRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.DeleteProductAggregationRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewAcqRightsServiceServer(dbObj, qObj)
			got, err := s.DeleteProductAggregation(test.ctx, test.input)

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
