// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/product-service/pkg/api/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestListAcqRightsAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsAggregationRequest
		output *v1.ListAcqRightsAggregationResponse
		mock   func(*v1.ListAcqRightsAggregationRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "ListAcqRightsAggregationWithCorrectData",
			input: &v1.ListAcqRightsAggregationRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1"},
			},
			output: &v1.ListAcqRightsAggregationResponse{
				TotalRecords: int32(2),
				Aggregations: []*v1.AcqRightsAggregation{
					&v1.AcqRightsAggregation{
						ID:        int32(1),
						Name:      "x1",
						Editor:    "e1",
						Skus:      []string{"s1", "s2"},
						Swidtags:  []string{"p1", "p2"},
						Metric:    "m1",
						TotalCost: float64(100),
					},
					&v1.AcqRightsAggregation{
						ID:        int32(2),
						Name:      "x2",
						Editor:    "e2",
						Skus:      []string{"s3", "s4"},
						Swidtags:  []string{"p3", "p4"},
						Metric:    "m2",
						TotalCost: float64(200),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAcqRightsAggregationRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize,
					Scope:              input.Scopes,
					AggregationNameAsc: true,
				}).Return([]db.ListAcqRightsAggregationRow{
					{
						Totalrecords:    int64(2),
						AggregationID:   int32(1),
						AggregationName: "x1",
						ProductEditor:   "e1",
						Metric:          "m1",
						Skus:            []string{"s1", "s2"},
						Swidtags:        []string{"p1", "p2"},
						TotalCost:       decimal.NewFromFloat(100),
					},
					{
						Totalrecords:    int64(2),
						AggregationID:   int32(2),
						AggregationName: "x2",
						ProductEditor:   "e2",
						Metric:          "m2",
						Skus:            []string{"s3", "s4"},
						Swidtags:        []string{"p3", "p4"},
						TotalCost:       decimal.NewFromFloat(200),
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListAcqRightsAggregationWithoutContext",
			input: &v1.ListAcqRightsAggregationRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsAggregationRequest) {},
		},
		{
			name: "ListAcqRightsAggregationWithNoResultSet",
			input: &v1.ListAcqRightsAggregationRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			output: &v1.ListAcqRightsAggregationResponse{
				Aggregations: []*v1.AcqRightsAggregation{},
				TotalRecords: int32(0),
			},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListAcqRightsAggregationRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					PageNum:            input.PageSize * (input.PageNum - 1),
					PageSize:           input.PageSize,
					Scope:              []string{},
					AggregationNameAsc: true,
				}).Return([]db.ListAcqRightsAggregationRow{}, nil).Times(1)
			},
		},
		{
			name: "FAILURE: User does not have access to the scopes",
			mock: func(*v1.ListAcqRightsAggregationRequest) {},
			ctx:  ctx,
			input: &v1.ListAcqRightsAggregationRequest{
				Scopes: []string{"s4"},
			},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ListAcqRightsAggregation(test.ctx, test.input)

			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err [%s] ", test.name, err.Error())
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%+v], ex[ [%+v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestListAcqRightsAggregationRecords(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAcqRightsAggregationRecordsRequest
		output *v1.ListAcqRightsAggregationRecordsResponse
		mock   func(*v1.ListAcqRightsAggregationRecordsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name:   "ListAcqRightsAggregationRecordsWithCorrectData",
			input:  &v1.ListAcqRightsAggregationRecordsRequest{AggregationId: int32(1), Scopes: []string{"s1"}},
			outErr: false,
			output: &v1.ListAcqRightsAggregationRecordsResponse{
				AcquiredRights: []*v1.AcqRights{
					&v1.AcqRights{
						Entity:                         "ent",
						SKU:                            "sku",
						SwidTag:                        "prod",
						ProductName:                    "pname",
						Editor:                         "e",
						Metric:                         "m",
						AcquiredLicensesNumber:         int32(10),
						LicensesUnderMaintenanceNumber: int32(10),
						AvgLicenesUnitPrice:            float64(1),
						AvgMaintenanceUnitPrice:        float64(1),
						TotalPurchaseCost:              float64(20),
						TotalMaintenanceCost:           float64(20),
						TotalCost:                      float64(40),
					},
				},
			},
			ctx: ctx,
			mock: func(input *v1.ListAcqRightsAggregationRecordsRequest) {
				dbObj.EXPECT().ListAcqRightsAggregationIndividual(ctx, db.ListAcqRightsAggregationIndividualParams{
					Scope:         input.Scopes,
					AggregationID: input.AggregationId,
				}).Return([]db.ListAcqRightsAggregationIndividualRow{
					{
						Entity:                  "ent",
						Sku:                     "sku",
						Swidtag:                 "prod",
						ProductName:             "pname",
						ProductEditor:           "e",
						Metric:                  "m",
						NumLicensesAcquired:     int32(10),
						NumLicencesMaintainance: int32(10),
						AvgUnitPrice:            decimal.NewFromFloat(1),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(1),
						TotalPurchaseCost:       decimal.NewFromFloat(20),
						TotalMaintenanceCost:    decimal.NewFromFloat(20),
						TotalCost:               decimal.NewFromFloat(40),
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListAcqRightsAggregationRecordsWithoutContext",
			input:  &v1.ListAcqRightsAggregationRecordsRequest{AggregationId: int32(1)},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAcqRightsAggregationRecordsRequest) {},
		},
		{
			name: "FAILURE: User does not have access to the scopes",
			mock: func(*v1.ListAcqRightsAggregationRecordsRequest) {},
			ctx:  ctx,
			input: &v1.ListAcqRightsAggregationRecordsRequest{
				Scopes: []string{"s4"},
			},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ListAcqRightsAggregationRecords(test.ctx, test.input)
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
