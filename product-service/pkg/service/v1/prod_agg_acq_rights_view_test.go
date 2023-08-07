package v1

import (
	"context"
	"errors"
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

func TestListAggregatedAcqRights(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAggregatedAcqRightsRequest
		output *v1.ListAggregatedAcqRightsResponse
		mock   func(*v1.ListAggregatedAcqRightsRequest)
		ctx    context.Context
		outErr bool
	}{

		{
			name:   "ListAggregatedAcqRights without context",
			input:  &v1.ListAggregatedAcqRightsRequest{Scope: "s1", PageNum: int32(1), PageSize: int32(10)},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ListAggregatedAcqRightsRequest) {},
		},
		{
			name:   "ListAggregatedAcqRights scope validation failure",
			input:  &v1.ListAggregatedAcqRightsRequest{Scope: "s4", PageNum: int32(1), PageSize: int32(10)},
			ctx:    ctx,
			outErr: true,
			mock:   func(input *v1.ListAggregatedAcqRightsRequest) {},
		},

		{
			name:  "ListAggregatedAcqRights DB Error",
			input: &v1.ListAggregatedAcqRightsRequest{Scope: "s1", PageNum: int32(1), PageSize: int32(10)},
			ctx:   ctx,
			mock: func(data *v1.ListAggregatedAcqRightsRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					Scope:    "s1",
					PageNum:  int32(data.PageNum-1) * data.PageSize,
					PageSize: int32(data.PageSize),
					SkuAsc:   true,
				}).Return(nil, errors.New("DBError")).Times(1)
			},
			outErr: true,
		},
		{
			name:  "ListAggregatedAcqRights with data",
			input: &v1.ListAggregatedAcqRightsRequest{Scope: "s1", PageNum: int32(1), PageSize: int32(10)},
			ctx:   ctx,
			mock: func(data *v1.ListAggregatedAcqRightsRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					Scope:    "s1",
					PageNum:  int32(data.PageNum-1) * data.PageSize,
					PageSize: int32(data.PageSize),
					SkuAsc:   true,
				}).Return([]db.ListAcqRightsAggregationRow{
					{
						Totalrecords:            int64(1),
						AggregationID:           int32(1),
						AggregationName:         "name",
						Sku:                     "sku",
						Swidtags:                []string{"a", "b"},
						Products:                []string{"x", "y"},
						ProductEditor:           "pe",
						AvgUnitPrice:            decimal.NewFromFloat(1.0),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(1.0),
						NumLicensesAcquired:     int32(1),
						NumLicencesMaintenance:  int32(1),
						TotalPurchaseCost:       decimal.NewFromFloat(1.0),
						TotalMaintenanceCost:    decimal.NewFromFloat(1.0),
						TotalComputedCost:       decimal.NewFromFloat(2.0),
						TotalCost:               decimal.NewFromFloat(2.0),
						Scope:                   "s1",
						SoftwareProvider:        "Software",
					},
				}, nil).Times(1)

				// dbObj.EXPECT().GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
				// 	Swidtag: []string{"a", "b"},
				// 	Scope:   "s1",
				// }).Return([]db.GetAcqBySwidtagsRow{}, nil).Times(1)
			},
			output: &v1.ListAggregatedAcqRightsResponse{
				TotalRecords: int32(1),
				Aggregations: []*v1.AggregatedRightsView{
					{
						ID:                      int32(1),
						AggregationName:         "name",
						Sku:                     "sku",
						Swidtags:                []string{"a", "b"},
						ProductNames:            []string{"x", "y"},
						ProductEditor:           "pe",
						AvgUnitPrice:            float64(1),
						AvgMaintenanceUnitPrice: float64(1),
						TotalPurchaseCost:       float64(1),
						TotalMaintenanceCost:    float64(1),
						TotalCost:               float64(2),
						Scope:                   "s1",
						NumLicensesAcquired:     int32(1),
						NumLicencesMaintenance:  int32(1),
						LicenceUnderMaintenance: "no",
						SoftwareProvider:        "Software",
					},
				},
			},
			outErr: false,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.ListAggregatedAcqRights(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

// func TestListAcqRightsAggregationRecords(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	dbObj := dbmock.NewMockProduct(mockCtrl)
// 	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
// 	testSet := []struct {
// 		name   string
// 		input  *v1.ListAcqRightsAggregationRecordsRequest
// 		output *v1.ListAcqRightsAggregationRecordsResponse
// 		mock   func(*v1.ListAcqRightsAggregationRecordsRequest)
// 		outErr bool
// 		ctx    context.Context
// 	}{
// 		{
// 			name:   "ListAcqRightsAggregationRecordsWithCorrectData",
// 			input:  &v1.ListAcqRightsAggregationRecordsRequest{AggregationId: int32(1), scope: "s1"},
// 			outErr: false,
// 			output: &v1.ListAcqRightsAggregationRecordsResponse{
// 				AcquiredRights: []*v1.AcqRights{
// 					{
// 						SKU:                            "sku",
// 						SwidTag:                        "prod",
// 						ProductName:                    "pname",
// 						Editor:                         "e",
// 						Metric:                         "m",
// 						AcquiredLicensesNumber:         int32(10),
// 						LicensesUnderMaintenanceNumber: int32(10),
// 						AvgLicenesUnitPrice:            float64(1),
// 						AvgMaintenanceUnitPrice:        float64(1),
// 						TotalPurchaseCost:              float64(20),
// 						TotalMaintenanceCost:           float64(20),
// 						TotalCost:                      float64(40),
// 					},
// 				},
// 			},
// 			ctx: ctx,
// 			mock: func(input *v1.ListAcqRightsAggregationRecordsRequest) {
// 				dbObj.EXPECT().ListAcqRightsAggregationIndividual(ctx, db.ListAcqRightsAggregationIndividualParams{
// 					Scope:         input.scope,
// 					AggregationID: input.AggregationId,
// 				}).Return([]db.ListAcqRightsAggregationIndividualRow{
// 					{
// 						Sku:                     "sku",
// 						Swidtag:                 "prod",
// 						ProductName:             "pname",
// 						ProductEditor:           "e",
// 						Metric:                  "m",
// 						NumLicensesAcquired:     int32(10),
// 						NumLicencesMaintainance: int32(10),
// 						AvgUnitPrice:            decimal.NewFromFloat(1),
// 						AvgMaintenanceUnitPrice: decimal.NewFromFloat(1),
// 						TotalPurchaseCost:       decimal.NewFromFloat(20),
// 						TotalMaintenanceCost:    decimal.NewFromFloat(20),
// 						TotalCost:               decimal.NewFromFloat(40),
// 					},
// 				}, nil).Times(1)
// 			},
// 		},
// 		{
// 			name:   "ListAcqRightsAggregationRecordsWithoutContext",
// 			input:  &v1.ListAcqRightsAggregationRecordsRequest{AggregationId: int32(1)},
// 			outErr: true,
// 			ctx:    context.Background(),
// 			mock:   func(input *v1.ListAcqRightsAggregationRecordsRequest) {},
// 		},
// 		{
// 			name: "FAILURE: User does not have access to the scope",
// 			mock: func(*v1.ListAcqRightsAggregationRecordsRequest) {},
// 			ctx:  ctx,
// 			input: &v1.ListAcqRightsAggregationRecordsRequest{
// 				scope: []string{"s4"},
// 			},
// 			outErr: true,
// 		},
// 	}
// 	for _, test := range testSet {
// 		t.Run("", func(t *testing.T) {
// 			test.mock(test.input)
// 			s := NewProductServiceServer(dbObj, qObj, nil, "",nil)
// 			got, err := s.ListAcqRightsAggregationRecords(test.ctx, test.input)
// 			if (err != nil) != test.outErr {
// 				t.Errorf("Failed case [%s]  because expected err is mismatched with actual err ", test.name)
// 				return
// 			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
// 				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)
// 				return
// 			} else {
// 				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
// 			}
// 		})
// 	}
// }

func Test_GetAggregationAcqrightsExpandedView(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GetAggregationAcqrightsExpandedViewRequest
		output *v1.GetAggregationAcqrightsExpandedViewResponse
		mock   func(*v1.GetAggregationAcqrightsExpandedViewRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "GetAggregationAcqrightsExpandedView_correct_data",
			input: &v1.GetAggregationAcqrightsExpandedViewRequest{
				AggregationName: "ag1",
				Scope:           "s1",
				Metric:          "m1",
			},
			output: &v1.GetAggregationAcqrightsExpandedViewResponse{
				TotalRecords: int32(1),
				AcqRights: []*v1.AcqRights{
					{
						SKU:                            "sk1",
						ProductName:                    "p1",
						Metric:                         "m1",
						SwidTag:                        "sw4",
						AcquiredLicensesNumber:         int32(1),
						LicensesUnderMaintenanceNumber: int32(2),
						AvgLicenesUnitPrice:            float64(1.0),
						AvgMaintenanceUnitPrice:        float64(2.0),
						TotalPurchaseCost:              float64(1.0),
						TotalMaintenanceCost:           float64(4.0),
						Editor:                         "e1",
						TotalCost:                      float64(5),
						Version:                        "v1",
						LicensesUnderMaintenance:       "no",
						SoftwareProvider:               "Software",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.GetAggregationAcqrightsExpandedViewRequest) {
				dbObj.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: "ag1",
					Scope:           "s1",
				}).Return(db.Aggregation{
					ID:              int32(1),
					AggregationName: "ag1",
					Swidtags:        []string{"sw1", "sw2"},
				}, nil).Times(1)

				dbObj.EXPECT().GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
					Swidtag:  []string{"sw1", "sw2"},
					Scope:    "s1",
					IsMetric: true,
					Metric:   "m1",
				}).Return([]db.GetAcqBySwidtagsRow{
					{

						ProductEditor:           "e1",
						Metric:                  "m1",
						Sku:                     "sk1",
						Swidtag:                 "sw4",
						ProductName:             "p1",
						NumLicensesAcquired:     int32(1),
						NumLicencesMaintainance: int32(2),
						AvgUnitPrice:            decimal.NewFromFloat(1),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
						TotalPurchaseCost:       decimal.NewFromFloat(1),
						TotalMaintenanceCost:    decimal.NewFromFloat(4),
						TotalCost:               decimal.NewFromFloat(5),
						Version:                 "v1",
						SoftwareProvider:        "Software",
					},
				}, nil).Times(1)
			},
		},
		{
			name: "No Context",
			input: &v1.GetAggregationAcqrightsExpandedViewRequest{
				Scope:           "s1",
				AggregationName: "ag1",
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.GetAggregationAcqrightsExpandedViewRequest) {},
		},
		{
			name: "No Data found",
			input: &v1.GetAggregationAcqrightsExpandedViewRequest{
				Scope:           "s1",
				AggregationName: "ag1",
			},
			output: &v1.GetAggregationAcqrightsExpandedViewResponse{
				AcqRights:    []*v1.AcqRights{},
				TotalRecords: int32(0),
			},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.GetAggregationAcqrightsExpandedViewRequest) {
				dbObj.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: "ag1",
					Scope:           "s1",
				}).Return(db.Aggregation{}, nil).Times(1)
			},
		},
		{
			name: "FAILURE: User does not have access to the Scope",
			mock: func(*v1.GetAggregationAcqrightsExpandedViewRequest) {},
			ctx:  ctx,
			input: &v1.GetAggregationAcqrightsExpandedViewRequest{
				Scope:           "s4",
				AggregationName: "ag1",
			},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.GetAggregationAcqrightsExpandedView(test.ctx, test.input)

			if (err != nil) != test.outErr {
				t.Errorf("Failed case GetAggregationAcqrightsExpandedView [%s]  because expected err is mismatched with actual err [%s] ", test.name, err.Error())
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case GetAggregationAcqrightsExpandedView [%s]  because expected and actual output is mismatched, act [%+v], ex[ [%+v]", test.name, test.output, got)
				return
			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}
