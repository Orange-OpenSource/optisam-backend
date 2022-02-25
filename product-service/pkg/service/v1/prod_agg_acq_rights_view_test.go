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

func TestListAggregatedAcqrights(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListAggregatedAcqRightsRequest
		output *v1.ListAggregatedAcqRightsResponse
		mock   func(*v1.ListAggregatedAcqRightsRequest)
		outErr bool
		ctx    context.Context
	}{
		{
			name: "List_AggregatedAcqrights_CorrectData",
			input: &v1.ListAggregatedAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1"},
			},
			output: &v1.ListAggregatedAcqRightsResponse{
				TotalRecords: int32(1),
				Aggregations: []*v1.AggregatedRightsView{
					{
						ID:                      int32(1),
						AggregationName:         "x1",
						Sku:                     "sk1",
						ProductNames:            []string{"p1", "p2"},
						MetricName:              "m1",
						Swidtags:                []string{"sw1", "sw2"},
						NumLicensesAcquired:     int32(1),
						NumLicencesMaintainance: int32(2),
						AvgUnitPrice:            float64(1.0),
						AvgMaintenanceUnitPrice: float64(2.0),
						Scope:                   "s1",
						TotalPurchaseCost:       float64(1.0),
						TotalMaintenanceCost:    float64(4.0),
						ProductEditor:           "e1",
						TotalCost:               float64(5),
						IsIndividualRightExists: true,
						LicenceUnderMaintenance: "no",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListAggregatedAcqRightsRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					Scope:    input.Scopes,
				}).Return([]db.ListAcqRightsAggregationRow{
					{
						ID:                      int32(1),
						Totalrecords:            int64(1),
						AggregationName:         "x1",
						ProductEditor:           "e1",
						Metric:                  "m1",
						Sku:                     "sk1",
						Scope:                   "s1",
						Swidtags:                []string{"sw1", "sw2"},
						Products:                []string{"p1", "p2"},
						NumLicensesAcquired:     int32(1),
						NumLicencesMaintainance: int32(2),
						AvgUnitPrice:            decimal.NewFromFloat(1),
						AvgMaintenanceUnitPrice: decimal.NewFromFloat(2),
						TotalPurchaseCost:       decimal.NewFromFloat(1),
						TotalMaintenanceCost:    decimal.NewFromFloat(4),
						TotalCost:               decimal.NewFromFloat(5),
					},
				}, nil).Times(1)

				dbObj.EXPECT().GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
					Swidtag: []string{"sw1", "sw2"},
					Scope:   "s1",
				}).Return([]db.Acqright{
					{
						Sku:     "s5",
						Swidtag: "sw4",
					},
				}, nil).Times(1)
			},
		},
		{
			name: "Context No found",
			input: &v1.ListAggregatedAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListAggregatedAcqRightsRequest) {},
		},
		{
			name: "no data found",
			input: &v1.ListAggregatedAcqRightsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			output: &v1.ListAggregatedAcqRightsResponse{
				Aggregations: []*v1.AggregatedRightsView{},
				TotalRecords: int32(0),
			},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListAggregatedAcqRightsRequest) {
				dbObj.EXPECT().ListAcqRightsAggregation(ctx, db.ListAcqRightsAggregationParams{
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize,
					Scope:    []string{},
				}).Return([]db.ListAcqRightsAggregationRow{}, nil).Times(1)
			},
		},
		{
			name: "FAILURE: User does not have access to the scopes",
			mock: func(*v1.ListAggregatedAcqRightsRequest) {},
			ctx:  ctx,
			input: &v1.ListAggregatedAcqRightsRequest{
				Scopes: []string{"s4"},
			},
			outErr: true,
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.ListAggregatedAcqRights(test.ctx, test.input)

			if (err != nil) != test.outErr {
				t.Errorf("Failed case  ListAggregatedAcqRights [%s]  because expected err is mismatched with actual err [%s] ", test.name, err.Error())
				return
			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				t.Errorf("Failed case ListAggregatedAcqRights [%s]  because expected and actual output is mismatched, act [%+v], ex[ [%+v]", test.name, test.output, got)
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
					{
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
			s := NewProductServiceServer(dbObj, qObj, nil, "")
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
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.GetAggregationAcqrightsExpandedViewRequest) {
				dbObj.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					AggregationName: "ag1",
					Scope:           "s1",
				}).Return(db.AggregatedRight{
					ID:              int32(1),
					AggregationName: "ag1",
					Swidtags:        []string{"sw1", "sw2"},
				}, nil).Times(1)

				dbObj.EXPECT().GetAcqBySwidtags(ctx, db.GetAcqBySwidtagsParams{
					Swidtag: []string{"sw1", "sw2"},
					Scope:   "s1",
				}).Return([]db.Acqright{
					{

						ProductEditor:           "e1",
						Metric:                  "m1",
						Sku:                     "sk1",
						Scope:                   "s1",
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
				}).Return(db.AggregatedRight{}, nil).Times(1)
			},
		},
		{
			name: "FAILURE: User does not have access to the scopes",
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
