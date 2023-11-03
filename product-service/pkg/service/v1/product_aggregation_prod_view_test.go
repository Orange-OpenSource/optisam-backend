package v1

import (
	"context"
	"database/sql"
	"errors"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1/mock"

	"testing"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"

	"github.com/golang/mock/gomock"

	"go.uber.org/zap"
)

func TestListProductAggregationView(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListProductAggregationViewRequest
		output *v1.ListProductAggregationViewResponse
		mock   func(*v1.ListProductAggregationViewRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListProductAggregationViewWithCorrectInfo",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1", "s2"},
			},
			output: &v1.ListProductAggregationViewResponse{
				TotalRecords: int32(2),
				Aggregations: []*v1.ProductAggregationView{
					{
						AggregationName: "agg1",
						Editor:          "e1",
						NumApplications: int32(5),
						NumEquipments:   int32(5),
						Swidtags:        []string{"p1", "p2"},
					},
					{
						AggregationName: "agg2",
						Editor:          "e2",
						NumApplications: int32(10),
						NumEquipments:   int32(10),
						Swidtags:        []string{"p3", "p4"},
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductAggregationViewRequest) {
				dbObj.EXPECT().ListProductAggregation(ctx, db.ListProductAggregationParams{
					Scope:    input.GetScopes()[0],
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize}).Return([]db.ListProductAggregationRow{
					{
						AggregationName:   "agg1",
						ProductEditor:     "e1",
						Swidtags:          []string{"p1", "p2"},
						NumOfApplications: int32(5),
						NumOfEquipments:   int32(5),
					},
					{
						AggregationName:   "agg2",
						ProductEditor:     "e2",
						Swidtags:          []string{"p3", "p4"},
						NumOfApplications: int32(10),
						NumOfEquipments:   int32(10),
					},
				}, nil).Times(1)
				dbObj.EXPECT().GetIndividualProductForAggregationCount(ctx, gomock.Any()).Return(int64(1), nil).AnyTimes()
			},
		},
		{
			name: "ListProductAggregationViewWithoutContext",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1", "s2"},
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductAggregationViewRequest) {},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s4"},
			},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListProductAggregationViewRequest) {},
		},
		{
			name: "ListProductAggregationViewWithNoResult",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1", "s2"},
			},
			outErr: true,
			output: &v1.ListProductAggregationViewResponse{
				TotalRecords: int32(0),
				Aggregations: []*v1.ProductAggregationView{},
			},
			ctx: context.Background(),
			mock: func(input *v1.ListProductAggregationViewRequest) {
				dbObj.EXPECT().ListProductAggregation(ctx, []db.ListProductAggregationRow{}).Return([]db.ListProductAggregationRow{}, nil).Times(1)
			},
		},
		{
			name: "FAILURE: Database error",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1", "s2"},
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ListProductAggregationViewRequest) {
				dbObj.EXPECT().ListProductAggregation(ctx, gomock.Any()).Return(nil, errors.New("DB error")).Times(1)
			},
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.ListProductAggregationView(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestGetAggregationProductsExpandedView(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.GetAggregationProductsExpandedViewRequest
		output *v1.GetAggregationProductsExpandedViewResponse
		mock   func(*v1.GetAggregationProductsExpandedViewRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ProductAggregationProductViewOptionswitCorrectData",
			input: &v1.GetAggregationProductsExpandedViewRequest{AggregationName: "agg1", Scope: "s1"},
			output: &v1.GetAggregationProductsExpandedViewResponse{
				TotalRecords: int32(2),
				Products: []*v1.ProductExpand{
					{
						SwidTag: "p1",
						Name:    "pname1",
						Editor:  "e1",
						Version: "v1",
					},
					{
						SwidTag: "p2",
						Name:    "pname2",
						Editor:  "e2",
						Version: "v2",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.GetAggregationProductsExpandedViewRequest) {
				dbObj.EXPECT().GetIndividualProductDetailByAggregation(ctx, gomock.Any()).Return([]db.GetIndividualProductDetailByAggregationRow{{}}, nil).Times(1)
			},
		},
		{
			name:   "ProductAggregationProductViewOptionswithoutContext",
			input:  &v1.GetAggregationProductsExpandedViewRequest{AggregationName: "agg1", Scope: "s1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.GetAggregationProductsExpandedViewRequest) {},
		},
		{
			name:   "FAILURE - No access to Scopes",
			input:  &v1.GetAggregationProductsExpandedViewRequest{AggregationName: "agg1", Scope: "s4"},
			ctx:    ctx,
			outErr: true,
			mock:   func(input *v1.GetAggregationProductsExpandedViewRequest) {},
		},

		{
			name:  "ProductAggregationProductViewOptionswithNoResult",
			input: &v1.GetAggregationProductsExpandedViewRequest{AggregationName: "agg1", Scope: "s1"},
			ctx:   context.Background(),
			output: &v1.GetAggregationProductsExpandedViewResponse{
				TotalRecords: int32(0),
				Products:     []*v1.ProductExpand{},
			},
			outErr: true,
			mock:   func(input *v1.GetAggregationProductsExpandedViewRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.GetAggregationProductsExpandedView(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestAggregatedRightDetails(t *testing.T) {
	timeStart := time.Now()
	timeEnd := timeStart.Add(10 * time.Hour)
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
	rep := dbObj
	met := mockMetric

	testSet := []struct {
		name   string
		input  *v1.AggregatedRightDetailsRequest
		output *v1.AggregatedRightDetailsResponse
		s      *ProductServiceServer
		mock   func(*v1.AggregatedRightDetailsRequest, *time.Time, *time.Time)
		outErr bool
		ctx    context.Context
	}{
		{
			name:   "AggregatedRightDetailsWithCorrectData",
			input:  &v1.AggregatedRightDetailsRequest{Scope: "s1"},
			output: &v1.AggregatedRightDetailsResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.AggregatedRightDetailsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().AggregatedRightDetails(ctx, gomock.Any()).Return(db.AggregatedRightDetailsRow{Metrics: []string{"m1"}}, nil).Times(1)
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(1), nil).AnyTimes()
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Return(&metv1.ListMetricResponse{Metrices: []*metv1.Metric{{Name: "m1"}}}, nil)
			},
		},
		{
			name:   "ctx err",
			input:  &v1.AggregatedRightDetailsRequest{Scope: "s1"},
			output: &v1.AggregatedRightDetailsResponse{},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.AggregatedRightDetailsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().AggregatedRightDetails(ctx, gomock.Any()).Return(db.AggregatedRightDetailsRow{Metrics: []string{"m1"}}, nil).Times(1)
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(1), nil).AnyTimes()
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Return(&metv1.ListMetricResponse{Metrices: []*metv1.Metric{{Name: "m1"}}}, nil)
			},
		},
		{
			name:   "scope err",
			input:  &v1.AggregatedRightDetailsRequest{Scope: "na"},
			output: &v1.AggregatedRightDetailsResponse{},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.AggregatedRightDetailsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().AggregatedRightDetails(ctx, gomock.Any()).Return(db.AggregatedRightDetailsRow{Metrics: []string{"m1"}}, nil).Times(1)
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(1), nil).AnyTimes()
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Return(&metv1.ListMetricResponse{Metrices: []*metv1.Metric{{Name: "m1"}}}, nil)
			},
		},
		{
			name:   "sql no rows",
			input:  &v1.AggregatedRightDetailsRequest{Scope: "s1"},
			output: &v1.AggregatedRightDetailsResponse{},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.AggregatedRightDetailsRequest, s *time.Time, e *time.Time) {
				dbObj.EXPECT().AggregatedRightDetails(ctx, gomock.Any()).Return(db.AggregatedRightDetailsRow{Metrics: []string{"m1"}}, sql.ErrNoRows).AnyTimes()
				dbObj.EXPECT().GetAvailableAcqLicenses(ctx, gomock.Any()).Return(int32(1), nil).AnyTimes()
				mockMetric.EXPECT().ListMetrices(ctx, gomock.Any()).Return(&metv1.ListMetricResponse{Metrices: []*metv1.Metric{{Name: "m1"}}}, nil)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input, &timeStart, &timeEnd)
			test.s = &ProductServiceServer{
				ProductRepo: rep,
				metric:      met,
			}
			_, err := test.s.AggregatedRightDetails(test.ctx, test.input)
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
