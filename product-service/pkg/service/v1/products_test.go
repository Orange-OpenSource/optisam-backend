package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	appv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/application-service/pkg/api/v1"

	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/thirdparty/metric-service/pkg/api/v1/mock"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/repository/v1/queuemock"
	dgworker "gitlab.tech.orange/optisam/optisam-it/optisam-services/product-service/pkg/worker/dgraph"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/workerqueue/job"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetProductDetail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	metObj := metmock.NewMockMetricServiceClient(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductRequest
		output *v1.ProductResponse
		mock   func(*v1.ProductRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "GetProductDetailWithCorrectData",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:         "p",
				ProductName:     "pn",
				Editor:          "e",
				Version:         "v",
				NumApplications: 1,
				NumEquipments:   3,
				DefinedMetrics:  []string{"m1", "m2"},
				NotDeployed:     false,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{
					Swidtag:           "p",
					ProductName:       "pn",
					ProductEditor:     "e",
					ProductVersion:    "v",
					NumOfApplications: 1,
					NumOfEquipments:   3,
					Metrics:           []string{"m1", "m2", "m3"},
				}, nil).Times(1)
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "OPS",
							Name:        "m1",
							Description: "metric description",
							Default:     false,
						},
						{
							Type:        "NUP",
							Name:        "m2",
							Description: "metricNup description",
						},
					}}, nil)
			},
		},
		{
			name:  "GetProductDetailWithCorrectData - produc does not exist",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:        "p",
				ProductName:    "pn",
				Editor:         "e",
				Version:        "v",
				DefinedMetrics: []string{"m1", "m2"},
				NotDeployed:    true,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{}, sql.ErrNoRows).Times(1)
				dbObj.EXPECT().GetProductInformationFromAcqright(ctx, gomock.Any()).Return(db.GetProductInformationFromAcqrightRow{
					Swidtag:       "p",
					ProductName:   "pn",
					ProductEditor: "e",
					Version:       "v",
					Metrics:       []string{"m1", "m2", "m3"},
				}, nil).Times(1).AnyTimes()
				dbObj.EXPECT().GetProductInformationFromAcqrightForAll(ctx, gomock.Any()).Return(db.GetProductInformationFromAcqrightForAllRow{
					Swidtag:       "p",
					ProductName:   "pn",
					ProductEditor: "e",
					Version:       "v",
					Metrics:       []string{"m1", "m2", "m3"},
				}, nil).Times(1).AnyTimes()
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "OPS",
							Name:        "m1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "m2",
							Description: "metricNup description",
						},
					}}, nil)
			},
		},
		{
			name:   "GetProductDetailWithoutContext",
			input:  &v1.ProductRequest{SwidTag: "p1", Scope: "s1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ProductRequest) {},
		},
		{
			name:   "FAILURE: No access to scopes",
			input:  &v1.ProductRequest{SwidTag: "p1", Scope: "s4"},
			ctx:    ctx,
			outErr: true,
			mock:   func(*v1.ProductRequest) {},
		},
		{
			name:  "GetProductDetailWithCorrectData",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:         "p",
				ProductName:     "pn",
				Editor:          "e",
				Version:         "v",
				NumApplications: 1,
				NumEquipments:   3,
				DefinedMetrics:  []string{"m1", "m2"},
				NotDeployed:     false,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				// Mocking the database call
				dbObj.EXPECT().GetProductInformation(ctx, gomock.Any()).Return(db.GetProductInformationRow{
					Swidtag:           "p",
					ProductName:       "pn",
					ProductEditor:     "e",
					ProductVersion:    "v",
					NumOfApplications: 1,
					NumOfEquipments:   3,
					Metrics:           []string{"m1", "m2", "m3"},
				}, nil).Times(1)
				// Mocking the metric service call
				metObj.EXPECT().ListMetrices(ctx, gomock.Any()).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "OPS",
							Name:        "m1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "m2",
							Description: "metricNup description",
						},
					},
				}, nil)
			},
		},
		{
			name:  "GetProductDetailWithNoMetrics",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:         "p",
				ProductName:     "pn",
				Editor:          "e",
				Version:         "v",
				NumApplications: 1,
				NumEquipments:   3,
				DefinedMetrics:  []string{},
				NotDeployed:     false,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				// Mocking the database call
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope,
				}).Return(db.GetProductInformationRow{
					Swidtag:           "p",
					ProductName:       "pn",
					ProductEditor:     "e",
					ProductVersion:    "v",
					NumOfApplications: 1,
					NumOfEquipments:   3,
					Metrics:           nil,
				}, nil).Times(1)
				// Mocking the metric service call
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: nil,
				}, nil)
			},
		},
		{
			name:   "GetProductDetailWithDatabaseError",
			input:  &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: nil,
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				// Mocking the database call
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope,
				}).Return(db.GetProductInformationRow{}, errors.New("database error")).Times(1)
			},
		},
		{
			name:   "GetProductDetailWithMetricServiceError",
			input:  &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: nil,
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				// Mocking the database call
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope,
				}).Return(db.GetProductInformationRow{
					Swidtag:           "p",
					ProductName:       "pn",
					ProductEditor:     "e",
					ProductVersion:    "v",
					NumOfApplications: 1,
					NumOfEquipments:   3,
					Metrics:           []string{"m1", "m2", "m3"},
				}, nil).Times(1)
				// Mocking the metric service call
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(nil, errors.New("metric service error"))
			},
		},
		{
			name:  "GetProductDetailWithCorrectData - acq rights",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:        "p",
				ProductName:    "pn",
				Editor:         "e",
				Version:        "v",
				DefinedMetrics: []string{"m1", "m2"},
				NotDeployed:    true,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{}, sql.ErrNoRows).Times(1)
				dbObj.EXPECT().GetProductInformationFromAcqright(ctx, gomock.Any()).Return(db.GetProductInformationFromAcqrightRow{
					Swidtag:       "p",
					ProductName:   "pn",
					ProductEditor: "e",
					Version:       "v",
					Metrics:       []string{"m1", "m2", "m3"},
				}, errors.New("some error")).Times(1).AnyTimes()
				dbObj.EXPECT().GetProductInformationFromAcqright(ctx, gomock.Any()).Return(db.GetProductInformationFromAcqrightRow{}, errors.New("some error")).Times(1).AnyTimes()
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "OPS",
							Name:        "m1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "m2",
							Description: "metricNup description",
						},
					}}, nil)
			},
		},
		{
			name:  "sql no roes",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductResponse{
				SwidTag:        "p",
				ProductName:    "pn",
				Editor:         "e",
				Version:        "v",
				DefinedMetrics: []string{"m1", "m2"},
				NotDeployed:    true,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{}, sql.ErrNoRows).Times(1)
				dbObj.EXPECT().GetProductInformationFromAcqright(ctx, gomock.Any()).Return(db.GetProductInformationFromAcqrightRow{
					Swidtag:       "p",
					ProductName:   "pn",
					ProductEditor: "e",
					Version:       "v",
					Metrics:       []string{"m1", "m2", "m3"},
				}, sql.ErrNoRows).Times(1).AnyTimes()
				metObj.EXPECT().ListMetrices(ctx, &metv1.ListMetricRequest{
					Scopes: []string{"s1"},
				}).Times(1).Return(&metv1.ListMetricResponse{
					Metrices: []*metv1.Metric{
						{
							Type:        "OPS",
							Name:        "m1",
							Description: "metric description",
						},
						{
							Type:        "NUP",
							Name:        "m2",
							Description: "metricNup description",
						},
					}}, nil).AnyTimes()
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := &ProductServiceServer{
				ProductRepo: dbObj,
				queue:       qObj,
				metric:      metObj,
			}
			_, err := s.GetProductDetail(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestGetProductOptions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductRequest
		output *v1.ProductOptionsResponse
		mock   func(*v1.ProductRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "GetProductOptionsWithCorrectData",
			input: &v1.ProductRequest{SwidTag: "p", Scope: "s1"},
			output: &v1.ProductOptionsResponse{
				NumOfOptions: int32(2),
				Optioninfo: []*v1.OptionInfo{
					{
						SwidTag: "p1",
						Name:    "n1",
						Edition: "ed1",
						Editor:  "e1",
						Version: "v1",
					},
					{
						SwidTag: "p2",
						Name:    "n2",
						Edition: "ed2",
						Editor:  "e2",
						Version: "v2",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductOptions(ctx, db.GetProductOptionsParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope,
				}).Return([]db.GetProductOptionsRow{
					{
						Swidtag:        "p1",
						ProductName:    "n1",
						ProductEdition: "ed1",
						ProductEditor:  "e1",
						ProductVersion: "v1",
					},
					{
						Swidtag:        "p2",
						ProductName:    "n2",
						ProductEdition: "ed2",
						ProductEditor:  "e2",
						ProductVersion: "v2",
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "GetProductOptionsWithoutContext",
			input:  &v1.ProductRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ProductRequest) {},
		},
		{
			name:   "FAILURE: No access to scopes",
			input:  &v1.ProductRequest{SwidTag: "p1", Scope: "s4"},
			outErr: true,
			ctx:    ctx,
			mock:   func(*v1.ProductRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.GetProductOptions(test.ctx, test.input)
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

func TestListProducts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	var app appv1.ApplicationServiceClient
	testSet := []struct {
		name   string
		input  *v1.ListProductsRequest
		output *v1.ListProductsResponse
		mock   func(*v1.ListProductsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListProductsRequestWithoutappIdandEquipId",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1"},
			},
			output: &v1.ListProductsResponse{
				TotalRecords: int32(1),
				Products: []*v1.Product{
					{
						SwidTag:           "p",
						Name:              "n",
						Version:           "v",
						Category:          "c",
						Edition:           "ed",
						Editor:            "e",
						TotalCost:         float64(100.0),
						NumOfApplications: int32(10),
						NumofEquipments:   int32(10),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductsRequest) {
				dbObj.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize}).Return([]db.ListProductsViewRow{
					{
						Totalrecords:      int64(1),
						Swidtag:           "p",
						ProductName:       "n",
						ProductVersion:    "v",
						ProductCategory:   "c",
						ProductEditor:     "e",
						ProductEdition:    "ed",
						NumOfApplications: int32(10),
						NumOfEquipments:   int32(10),
						Cost:              float64(100.00),
						ConcurrentUsers:   int32(1),
						NominativeUsers:   int32(0),
					},
				}, nil).AnyTimes()
			},
		},
		{
			name: "ListProductsRequestWithAppId",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.ProductSearchParams{
					ApplicationId: &v1.StringFilter{
						Filteringkey: "app",
					},
				},
				Scopes: []string{"s1"},
			},
			output: &v1.ListProductsResponse{
				TotalRecords: int32(1),
				Products: []*v1.Product{
					{
						SwidTag:   "p",
						Name:      "n",
						Version:   "v",
						Editor:    "e",
						TotalCost: float64(100.0),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductsRequest) {
				// mockApp := appmock.NewMockApplicationServiceClient(mockCtrl)
				// app = mockApp
				// mockApp.EXPECT().GetEquipmentsByApplication(ctx, &appv1.GetEquipmentsByApplicationRequest{
				// 	Scope:         "s1",
				// 	ApplicationId: "app",
				// }).Times(1).Return(&appv1.GetEquipmentsByApplicationResponse{
				// 	EquipmentId: []string{"eq1", "eq2", "eq3"},
				// }, nil)
				// dbObj.EXPECT().ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
				// 	Scope:         input.Scopes,
				// 	PageNum:       input.PageSize * (input.PageNum - 1),
				// 	ApplicationID: "app",
				// 	IsEquipmentID: true,
				// 	EquipmentIds:  []string{"eq1", "eq2", "eq3"},
				// 	PageSize:      input.PageSize}).Return([]db.ListProductsViewRedirectedApplicationRow{
				// 	{
				// 		Totalrecords:      int64(1),
				// 		Swidtag:           "p",
				// 		ProductName:       "n",
				// 		ProductVersion:    "v",
				// 		ProductCategory:   "c",
				// 		ProductEditor:     "e",
				// 		ProductEdition:    "ed",
				// 		NumOfApplications: int32(1),
				// 		NumOfEquipments:   int32(2),
				// 		Cost:              float64(100.00),
				// 	},
				// }, nil).Times(1)
				dbObj.EXPECT().GetProductsByApplicationID(ctx, db.GetProductsByApplicationIDParams{
					Scope:         input.Scopes[0],
					ApplicationID: input.SearchParams.ApplicationId.Filteringkey,
				}).Times(1).Return([]string{"eq1", "eq2", "eq3"}, nil)

				dbObj.EXPECT().ListProductsByApplication(ctx, db.ListProductsByApplicationParams{
					Scope:    input.Scopes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					Swidtag:  []string{"eq1", "eq2", "eq3"},
					PageSize: input.PageSize}).Return([]db.ListProductsByApplicationRow{
					{
						Totalrecords:   int64(1),
						Swidtag:        "p",
						ProductName:    "n",
						ProductVersion: "v",
						ProductEditor:  "e",
						TotalCost:      float64(100.00),
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListProductsRequestWithEquipID",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.ProductSearchParams{
					EquipmentId: &v1.StringFilter{
						Filteringkey: "equip",
					},
				},
				Scopes: []string{"s1", "s2"},
			},
			output: &v1.ListProductsResponse{
				TotalRecords: int32(1),
				Products: []*v1.Product{
					{
						SwidTag:           "p",
						Name:              "n",
						Version:           "v",
						Category:          "c",
						Edition:           "ed",
						Editor:            "e",
						TotalCost:         float64(100.0),
						NumOfApplications: int32(10),
						NumofEquipments:   int32(10),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductsRequest) {
				dbObj.EXPECT().ListProductsViewRedirectedEquipment(ctx, db.ListProductsViewRedirectedEquipmentParams{
					Scope:       input.Scopes,
					PageNum:     input.PageSize * (input.PageNum - 1),
					EquipmentID: "equip",
					PageSize:    input.PageSize}).Return([]db.ListProductsViewRedirectedEquipmentRow{
					{
						Totalrecords:      int64(1),
						Swidtag:           "p",
						ProductName:       "n",
						ProductVersion:    "v",
						ProductCategory:   "c",
						ProductEditor:     "e",
						ProductEdition:    "ed",
						NumOfApplications: int32(10),
						NumOfEquipments:   int32(10),
						Cost:              float64(100.00),
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListProductsRequestWithoutContext",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s1", "s2"},
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductsRequest) {},
		},
		{
			name: "FAILURE: No scope access",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				Scopes:   []string{"s4"},
			},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListProductsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := &ProductServiceServer{
				ProductRepo: dbObj,
				queue:       qObj,
				application: app,
			}
			got, err := s.ListProducts(test.ctx, test.input)
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

func TestUpsertProduct(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.UpsertProductRequest
		output *v1.UpsertProductResponse
		mock   func(*v1.UpsertProductRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "UpsertProductWithCorrectData",
			input: &v1.UpsertProductRequest{
				SwidTag:  "p",
				Name:     " n",
				Category: "c",
				Edition:  "ed",
				Editor:   "e",
				Version:  "v",
				OptionOf: "temp",
				Scope:    "s1",
				Applications: &v1.UpsertProductRequestApplication{
					Operation:     "add",
					ApplicationId: []string{"app1", "app2"},
				},
				Equipments: &v1.UpsertProductRequestEquipment{
					Operation: "add",
					Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
						{
							EquipmentId:    "e1",
							AllocatedUsers: int32(1),
						},
					},
				},
			},
			output: &v1.UpsertProductResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertProductRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				fcall := dbObj.EXPECT().UpsertProductTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				jsonData, err := json.Marshal(input)
				if err != nil {
					t.Errorf("Failed to do json marshalling")
				}
				e := dgworker.Envelope{Type: dgworker.UpsertProductRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Error("Failed to do json marshalling")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData}, "aw").Return(int32(1), nil).After(fcall).AnyTimes()
			},
		},
		{
			name:   "UpsertProductWithoutContext",
			input:  &v1.UpsertProductRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.UpsertProductRequest) {},
		},
		{
			name: "UpsertProductWithCorrectData TX ERROR",
			input: &v1.UpsertProductRequest{
				SwidTag:  "p",
				Name:     " n",
				Category: "c",
				Edition:  "ed",
				Editor:   "e",
				Version:  "v",
				OptionOf: "temp",
				Scope:    "s1",
				Applications: &v1.UpsertProductRequestApplication{
					Operation:     "add",
					ApplicationId: []string{"app1", "app2"},
				},
				Equipments: &v1.UpsertProductRequestEquipment{
					Operation: "add",
					Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
						{
							EquipmentId:    "e1",
							AllocatedUsers: int32(1),
						},
					},
				},
			},
			output: &v1.UpsertProductResponse{Success: false},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.UpsertProductRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				fcall := dbObj.EXPECT().UpsertProductTx(ctx, input, userClaims.UserID).Return(errors.New("some error")).Times(1)
				jsonData, err := json.Marshal(input)
				if err != nil {
					t.Errorf("Failed to do json marshalling")
				}
				e := dgworker.Envelope{Type: dgworker.UpsertProductRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Error("Failed to do json marshalling")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData}, "aw").Return(int32(1), nil).After(fcall).AnyTimes()
			},
		},
		// {
		// 	name: "UpsertProductWithCorrectData job ERROR",
		// 	input: &v1.UpsertProductRequest{
		// 		SwidTag:  "p",
		// 		Name:     " n",
		// 		Category: "c",
		// 		Edition:  "ed",
		// 		Editor:   "e",
		// 		Version:  "v",
		// 		OptionOf: "temp",
		// 		Scope:    "s1",
		// 		Applications: &v1.UpsertProductRequestApplication{
		// 			Operation:     "add",
		// 			ApplicationId: []string{"app1", "app2"},
		// 		},
		// 		Equipments: &v1.UpsertProductRequestEquipment{
		// 			Operation: "add",
		// 			Equipmentusers: []*v1.UpsertProductRequestEquipmentEquipmentuser{
		// 				{
		// 					EquipmentId:    "e1",
		// 					AllocatedUsers: int32(1),
		// 				},
		// 			},
		// 		},
		// 	},
		// 	output: &v1.UpsertProductResponse{Success: false},
		// 	outErr: true,
		// 	ctx:    ctx,
		// 	mock: func(input *v1.UpsertProductRequest) {
		// 		userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
		// 		if !ok {
		// 			t.Errorf("cannot find claims in context")
		// 		}
		// 		fcall := dbObj.EXPECT().UpsertProductTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
		// 		jsonData, err := json.Marshal(input)
		// 		if err != nil {
		// 			t.Errorf("Failed to do json marshalling")
		// 		}
		// 		e := dgworker.Envelope{Type: dgworker.UpsertProductRequest, JSON: jsonData}

		// 		envolveData, err := json.Marshal(e)
		// 		if err != nil {
		// 			t.Error("Failed to do json marshalling")
		// 		}
		// 		qObj.EXPECT().PushJob(ctx, job.Job{
		// 			Type:   sql.NullString{String: "aw"},
		// 			Status: job.JobStatusPENDING,
		// 			Data:   envolveData}, "aw").Return(int32(1), errors.New("job error")).After(fcall)
		// 	},
		// },
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.UpsertProduct(test.ctx, test.input)
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

func Test_ProductServiceServer_DropProductData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DropProductDataRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.DropProductDataResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DropProductDataRequest{
					Scope:        "Scope1",
					DeletionType: v1.DropProductDataRequest_FULL,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DropProductDataTx(ctx, "Scope1", v1.DropProductDataRequest_FULL).Times(1).Return(nil)
				jsonData, err := json.Marshal(&v1.DropProductDataRequest{
					Scope:        "Scope1",
					DeletionType: v1.DropProductDataRequest_FULL,
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.DropProductDataRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				job := job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}
				mockQueue.EXPECT().PushJob(ctx, job, "aw").Return(int32(1000), nil)
			},
			want: &v1.DropProductDataResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.DropProductDataRequest{
					Scope:        "Scope1",
					DeletionType: v1.DropProductDataRequest_FULL,
				},
			},
			setup: func() {},
			want: &v1.DropProductDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DropProductDataRequest{
					Scope:        "Scope4",
					DeletionType: v1.DropProductDataRequest_FULL,
				},
			},
			setup: func() {},
			want: &v1.DropProductDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DropProductDataTx - DBError",
			args: args{
				ctx: ctx,
				req: &v1.DropProductDataRequest{
					Scope:        "Scope1",
					DeletionType: v1.DropProductDataRequest_FULL,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DropProductDataTx(ctx, "Scope1", v1.DropProductDataRequest_FULL).Times(1).Return(errors.New("Internal"))
			},
			want: &v1.DropProductDataResponse{
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
			got, err := tt.s.DropProductData(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DropProductData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DropProductData() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_ProductServiceServer_DropAggregationData(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.DropAggregationDataRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.DropAggregationDataResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DropAggregationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DeleteAggregationByScope(ctx, "Scope1").Times(1).Return(nil)
				jsonData, err := json.Marshal(&v1.DropAggregationDataRequest{
					Scope: "Scope1",
				})
				if err != nil {
					t.Errorf("Failed to do json marshalling in test %v", err)
				}
				e := dgworker.Envelope{Type: dgworker.DropAggregationData, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Errorf("Failed to do json marshalling in test  %v", err)
				}
				job := job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData,
				}
				mockQueue.EXPECT().PushJob(ctx, job, "aw").Return(int32(1000), nil)
			},
			want: &v1.DropAggregationDataResponse{
				Success: true,
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.DropAggregationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {},
			want: &v1.DropAggregationDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.DropAggregationDataRequest{
					Scope: "Scope4",
				},
			},
			setup: func() {},
			want: &v1.DropAggregationDataResponse{
				Success: false,
			},
			wantErr: true,
		},
		{name: "FAILURE - DeleteAggregationByScope - DBError",
			args: args{
				ctx: ctx,
				req: &v1.DropAggregationDataRequest{
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().DeleteAggregationByScope(ctx, "Scope1").Times(1).Return(errors.New("Internal"))
			},
			want: &v1.DropAggregationDataResponse{
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
			got, err := tt.s.DropAggregationData(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.DropAggregationData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.DropAggregationData() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_ProductServiceServer_GetEquipmentsByProduct(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var rep repo.Product
	var queue workerqueue.Workerqueue
	type args struct {
		ctx context.Context
		req *v1.GetEquipmentsByProductRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetEquipmentsByProductResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByProductRequest{
					Scope:   "Scope1",
					SwidTag: "prod_1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetEquipmentsBySwidtag(ctx, db.GetEquipmentsBySwidtagParams{
					Scope:   "Scope1",
					Swidtag: "prod_1",
				}).Times(1).Return([]string{"Eq1", "Eq2", "Eq3"}, nil)
			},
			want: &v1.GetEquipmentsByProductResponse{
				EquipmentId: []string{"Eq1", "Eq2", "Eq3"},
			},
			wantErr: false,
		},
		{name: "FAILURE - ClaimsNotFound",
			args: args{
				ctx: context.Background(),
				req: &v1.GetEquipmentsByProductRequest{
					Scope:   "Scope1",
					SwidTag: "prod_1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - ScopeValidationError",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByProductRequest{
					Scope:   "Scope4",
					SwidTag: "prod_1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE - GetEquipmentsBySwidtag - DBError",
			args: args{
				ctx: ctx,
				req: &v1.GetEquipmentsByProductRequest{
					Scope:   "Scope1",
					SwidTag: "prod_1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := dbmock.NewMockProduct(mockCtrl)
				mockQueue := queuemock.NewMockWorkerqueue(mockCtrl)
				rep = mockRepo
				queue = mockQueue
				mockRepo.EXPECT().GetEquipmentsBySwidtag(ctx, db.GetEquipmentsBySwidtagParams{
					Scope:   "Scope1",
					Swidtag: "prod_1",
				}).Times(1).Return([]string{}, errors.New("Internal"))
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
			got, err := tt.s.GetEquipmentsByProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProductServiceServer.GetEquipmentsByProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProductServiceServer.GetEquipmentsByProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestUpsertNominativeUser(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	dbObj := dbmock.NewMockProduct(mockCtrl)
// 	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
// 	testSet := []struct {
// 		name   string
// 		input  *v1.UpserNominativeUserRequest
// 		output *v1.UpserNominativeUserResponse
// 		mock   func(*v1.UpserNominativeUserRequest)
// 		ctx    context.Context
// 		outErr bool
// 	}{
// 		{
// 			name: "UpsertNominativeUserWithCorrectData",
// 			input: &v1.UpserNominativeUserRequest{
// 				AggregationId: 12,
// 				Scope:         "s1",
// 				UserDetails: []*v1.NominativeUserDetails{
// 					{
// 						UserName:  "u1",
// 						FirstName: "f1",
// 						Email:     "email1",
// 						Profile:   "p1",
// 					},
// 				},
// 			},
// 			output: &v1.UpserNominativeUserResponse{
// 				Status: true,
// 			},
// 			outErr: false,
// 			ctx:    ctx,
// 			mock: func(input *v1.UpserNominativeUserRequest) {
// 				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 				if !ok {
// 					t.Errorf("cannot find claims in context")
// 				}
// 				if input.AggregationId > 0 {
// 					dbObj.EXPECT().GetAggregationByID(ctx, db.GetAggregationByIDParams{
// 						ID:    input.AggregationId,
// 						Scope: "s1",
// 					}).Times(1).Return(db.Aggregation{ID: input.AggregationId, AggregationName: "n1",
// 						Scope: "s1", ProductEditor: "e1", Products: []string{"a"}, Swidtags: []string{"swid1"}, CreatedOn: time.Now(), CreatedBy: userClaims.UserID}, nil)
// 				}
// 				nominativeUsers := make([]*v1.NominativeUser, 5000)
// 				for i := range nominativeUsers {
// 					user := &v1.NominativeUser{
// 						Editor:          fmt.Sprintf("User %d", i+1),
// 						ProductName:     fmt.Sprintf("User %d", i+1),
// 						AggregationName: "a1",
// 						ActivationDate:  timestamppb.New(time.Now()),
// 						ProductVersion:  fmt.Sprintf("User %d", i+1),
// 						UserName:        fmt.Sprintf("User %d", i+1),
// 					}
// 					nominativeUsers[i] = user
// 				}

// 				fcall := dbObj.EXPECT().UpsertNominativeUsersTx(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nominativeUsers).Times(1)
// 				// dReq := prepairUpsertNominativeUserDgraphRequest(input, "", userClaims.UserID, "n1")
// 				// jsonData, err := json.Marshal(dReq)
// 				// if err != nil {
// 				// 	t.Errorf("Failed to do json marshalling")
// 				// }
// 				// e := dgworker.Envelope{Type: dgworker.UpsertNominativeUserRequest, JSON: jsonData}

// 				// envolveData, err := json.Marshal(e)
// 				// if err != nil {
// 				// 	t.Error("Failed to do json marshalling")
// 				// }
// 				qObj.EXPECT().PushJob(ctx, gomock.Any(), "aw").Return(int32(1), nil).After(fcall).AnyTimes()
// 			},
// 		},
// 		{
// 			name: "UpsertNominativeUserWithOutContext",
// 			input: &v1.UpserNominativeUserRequest{
// 				AggregationId: 12,
// 				Scope:         "s1",
// 				UserDetails: []*v1.NominativeUserDetails{
// 					{
// 						UserName:  "u1",
// 						FirstName: "f1",
// 						Email:     "email1",
// 						Profile:   "p1",
// 					},
// 				},
// 			},
// 			ctx:    context.Background(),
// 			outErr: true,
// 			mock:   func(input *v1.UpserNominativeUserRequest) {},
// 		},
// 		{
// 			name: "FAILURE: No access to scopes",
// 			input: &v1.UpserNominativeUserRequest{
// 				AggregationId: 12,
// 				Scope:         "s1",
// 				UserDetails: []*v1.NominativeUserDetails{
// 					{
// 						UserName:  "u1",
// 						FirstName: "f1",
// 						Email:     "email1",
// 						Profile:   "p1",
// 					},
// 				},
// 			},
// 			ctx:    context.Background(),
// 			outErr: true,
// 			mock:   func(input *v1.UpserNominativeUserRequest) {},
// 		},
// 		{
// 			name:   "UpsertNominativeUserWithoutContext",
// 			input:  &v1.UpserNominativeUserRequest{},
// 			outErr: true,
// 			ctx:    context.Background(),
// 			mock:   func(input *v1.UpserNominativeUserRequest) {},
// 		},
// 	}
// 	for _, test := range testSet {
// 		t.Run("", func(t *testing.T) {
// 			test.mock(test.input)
// 			s := NewProductServiceServer(dbObj, qObj, nil, "", nil,nil,,&config.Config{})
// 			got, err := s.UpsertNominativeUser(test.ctx, test.input)
// 			if (err != nil) != test.outErr {
// 				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
// 				return
// 			} else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
// 				t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

// 			} else {
// 				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
// 			}
// 		})
// 	}
// }

func TestListNominativeUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.ListNominativeUsersRequest
		output *v1.ListNominativeUsersResponse
		mock   func(*v1.ListNominativeUsersRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "ListNominativeUserSuccess",
			input: &v1.ListNominativeUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			output: &v1.ListNominativeUsersResponse{
				TotalRecords: 1,
				NominativeUser: []*v1.NominativeUser{
					{
						ProductName:     "p1",
						AggregationName: "a1",
						UserName:        "u1",
						FirstName:       "f1",
						UserEmail:       "e1",
						Profile:         "p1",
						AggregationId:   12,
						ActivationDate:  timestamppb.New(timeNow),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListNominativeUsersRequest) {
				dbObj.EXPECT().ListNominativeUsersAggregation(ctx, db.ListNominativeUsersAggregationParams{Scope: []string{"s1"},
					UserEmailAsc: true, PageNum: 0, PageSize: 20}).Return([]db.ListNominativeUsersAggregationRow{
					{
						Totalrecords:    1,
						AggregationName: sql.NullString{String: "a1", Valid: true},
						UserName:        sql.NullString{String: "u1", Valid: true},
						FirstName:       sql.NullString{String: "f1", Valid: true},
						UserEmail:       "e1",
						Profile:         sql.NullString{String: "p1", Valid: true},
						AggregationsID:  sql.NullInt32{Int32: 12, Valid: true},
						ActivationDate:  sql.NullTime{Time: timeNow, Valid: true},
					},
				}, nil).AnyTimes()
			},
		},
		{
			name: "ListNominativeUserWithOutContext",
			input: &v1.ListNominativeUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ListNominativeUsersRequest) {},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.ListNominativeUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "SSS!",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ListNominativeUsersRequest) {},
		},
		{
			name:   "ListNominativeUserWithoutContext",
			input:  &v1.ListNominativeUsersRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListNominativeUsersRequest) {},
		},
		{
			name: "ListNominativeUserSuccess2",
			input: &v1.ListNominativeUsersRequest{
				PageNum:   1,
				PageSize:  20,
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: true,
			},
			output: &v1.ListNominativeUsersResponse{
				TotalRecords: 1,
				NominativeUser: []*v1.NominativeUser{
					{
						ProductName:     "p1",
						AggregationName: "a1",
						UserName:        "u1",
						FirstName:       "f1",
						UserEmail:       "e1",
						Profile:         "p1",
						AggregationId:   12,
						ActivationDate:  timestamppb.New(timeNow),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListNominativeUsersRequest) {
				dbObj.EXPECT().ListNominativeUsersAggregation(ctx, db.ListNominativeUsersAggregationParams{Scope: []string{"s1"},
					UserEmailAsc: true, PageNum: 0, PageSize: 20}).Return([]db.ListNominativeUsersAggregationRow{
					{
						Totalrecords:    1,
						AggregationName: sql.NullString{String: "a1", Valid: true},
						UserName:        sql.NullString{String: "u1", Valid: true},
						FirstName:       sql.NullString{String: "f1", Valid: true},
						UserEmail:       "e1",
						Profile:         sql.NullString{String: "p1", Valid: true},
						AggregationsID:  sql.NullInt32{Int32: 12, Valid: true},
						ActivationDate:  sql.NullTime{Time: timeNow, Valid: true},
					},
				}, nil).AnyTimes()
				dbObj.EXPECT().ListNominativeUsersProducts(ctx, db.ListNominativeUsersProductsParams{Scope: []string{"s1"},
					UserEmailAsc: true, PageNum: 0, PageSize: 20}).Return([]db.ListNominativeUsersProductsRow{
					{
						Totalrecords:   1,
						UserName:       sql.NullString{String: "u1", Valid: true},
						FirstName:      sql.NullString{String: "f1", Valid: true},
						UserEmail:      "e1",
						Profile:        sql.NullString{String: "p1", Valid: true},
						ActivationDate: sql.NullTime{Time: timeNow, Valid: true},
					},
				}, nil).AnyTimes()
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.ListNominativeUser(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func TestDeleteNominativeUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.DeleteNominativeUserRequest
		output *v1.DeleteNominativeUserResponse
		mock   func(*v1.DeleteNominativeUserRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "DeleteNominativeUsersSuccess",
			input: &v1.DeleteNominativeUserRequest{
				Id:    1,
				Scope: "s1",
			},
			output: &v1.DeleteNominativeUserResponse{
				Success: true,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.DeleteNominativeUserRequest) {
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				if !helper.Contains(userClaims.Socpes, input.Scope) {
					t.Errorf("ScopeValidationError")
				}
				dbNomUser := db.NominativeUser{
					UserID:         32,
					AggregationsID: sql.NullInt32{Int32: 20, Valid: true},
					Swidtag:        sql.NullString{String: "ABC_abc_14", Valid: true},
					Profile:        sql.NullString{String: "123", Valid: true},
					UserName:       sql.NullString{String: "p1", Valid: true},
					UserEmail:      "test@abc.com",
					Scope:          input.Scope,
					ActivationDate: sql.NullTime{Time: timeNow},
				}
				dbObj.EXPECT().GetNominativeUserByID(ctx, db.GetNominativeUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(dbNomUser, nil).Times(1)
				fcall := dbObj.EXPECT().DeleteNominativeUserByID(ctx, db.DeleteNominativeUserByIDParams{Scope: input.Scope, ID: input.Id}).Return(nil).Times(1)
				deleteNominativeReqDgraph := DeleteNominativeUserRequest(dbNomUser)
				deleteNominativeReqDgraph.Scope = input.Scope

				jsonData, err := json.Marshal(deleteNominativeReqDgraph)
				if err != nil {
					t.Errorf("Failed to do json marshalling")
				}
				e := dgworker.Envelope{Type: dgworker.DeleteNominativeUserRequest, JSON: jsonData}

				envolveData, err := json.Marshal(e)
				if err != nil {
					t.Error("Failed to do json marshalling")
				}
				qObj.EXPECT().PushJob(ctx, job.Job{
					Type:   sql.NullString{String: "aw"},
					Status: job.JobStatusPENDING,
					Data:   envolveData}, "aw").Return(int32(1), nil).After(fcall)

			},
		},
		{
			name: "DeleteNominativeUserWithOutContext",
			input: &v1.DeleteNominativeUserRequest{
				Id:    1,
				Scope: "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.DeleteNominativeUserRequest) {},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.DeleteNominativeUserRequest{
				Id:    1,
				Scope: "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.DeleteNominativeUserRequest) {},
		},
		{
			name:   "DeleteNominativeUserWithoutContext",
			input:  &v1.DeleteNominativeUserRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.DeleteNominativeUserRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			got, err := s.DeleteNominativeUsers(test.ctx, test.input)
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

func TestGetEditorProductExpensesByScope(t *testing.T) {
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
		req *v1.EditorProductsExpensesByScopeRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.EditorProductExpensesByScopeResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.EditorProductsExpensesByScopeRequest{
					Scope:  "scope1",
					Editor: "Oracle",
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

				mockRepo.EXPECT().GetEditorProductExpensesByScopeData(ctx, db.GetEditorProductExpensesByScopeDataParams{Scope: []string{"scope1"}, Reqeditor: "Oracle"}).Times(1).Return([]db.GetEditorProductExpensesByScopeDataRow{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
					},
				}, nil)
				mockRepo.EXPECT().GetComputedCostEditorProducts(ctx, db.GetComputedCostEditorProductsParams{Scope: []string{"scope1"}, Editor: "Oracle"}).Times(1).Return([]db.GetComputedCostEditorProductsRow{
					{
						ProductNames: "Oracle1",
						Cost:         5.0,
					},
					{
						ProductNames: "Oracle2",
						Cost:         25.0,
					},
				}, nil)
			},
			want: &v1.EditorProductExpensesByScopeResponse{
				EditorProductExpensesByScope: []*v1.EditorProductExpensesByScopeData{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
						TotalComputedCost:    5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
						TotalComputedCost:    25.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "GetComputedCostEditorProducts err",
			args: args{
				ctx: ctx,
				req: &v1.EditorProductsExpensesByScopeRequest{
					Scope:  "scope1",
					Editor: "Oracle",
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

				mockRepo.EXPECT().GetEditorProductExpensesByScopeData(ctx, db.GetEditorProductExpensesByScopeDataParams{Scope: []string{"scope1"}, Reqeditor: "Oracle"}).Times(1).Return([]db.GetEditorProductExpensesByScopeDataRow{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
					},
				}, nil)
				mockRepo.EXPECT().GetComputedCostEditorProducts(ctx, db.GetComputedCostEditorProductsParams{Scope: []string{"scope1"}, Editor: "Oracle"}).Times(1).Return([]db.GetComputedCostEditorProductsRow{
					{
						ProductNames: "Oracle1",
						Cost:         5.0,
					},
					{
						ProductNames: "Oracle2",
						Cost:         25.0,
					},
				}, errors.New("err"))
			},
			want: &v1.EditorProductExpensesByScopeResponse{
				EditorProductExpensesByScope: []*v1.EditorProductExpensesByScopeData{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
						TotalComputedCost:    5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
						TotalComputedCost:    25.0,
					},
				},
			},
			wantErr: true,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.EditorProductsExpensesByScopeRequest{
					Scope:  "scope1",
					Editor: "Oracle",
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

				mockRepo.EXPECT().GetEditorProductExpensesByScopeData(ctx, db.GetEditorProductExpensesByScopeDataParams{Scope: []string{"scope1"}, Reqeditor: "Oracle"}).Times(1).Return([]db.GetEditorProductExpensesByScopeDataRow{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
					},
				}, errors.New("error"))
			},
			want: &v1.EditorProductExpensesByScopeResponse{
				EditorProductExpensesByScope: []*v1.EditorProductExpensesByScopeData{
					{
						Name:                 "Oracle1",
						TotalPurchaseCost:    2.0,
						TotalMaintenanceCost: 3.0,
						TotalCost:            5.0,
					},
					{
						Name:                 "Oracle2",
						TotalPurchaseCost:    12.0,
						TotalMaintenanceCost: 13.0,
						TotalCost:            25.0,
					},
				},
			},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.EditorProductsExpensesByScopeRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.EditorProductsExpensesByScopeRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.EditorProductsExpensesByScopeRequest{
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
				mockRepo.EXPECT().GetEditorProductExpensesByScopeData(ctx, gomock.Any()).Return([]db.GetEditorProductExpensesByScopeDataRow{}, errors.New("internal")).AnyTimes()
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
			_, err := tt.s.GetEditorProductExpensesByScope(tt.args.ctx, tt.args.req)
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

func TestUpsertAllocatedMetricEquipment(t *testing.T) {
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
		req *v1.UpsertAllocateMetricEquipementRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.UpsertAllocateMetricEquipementResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.UpsertAllocateMetricEquipementRequest{
					Scope:   "scope1",
					Swidtag: "swidtag",
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

				mockRepo.EXPECT().UpsertProductEquipments(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.UpsertAllocateMetricEquipementResponse{Success: true},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.UpsertAllocateMetricEquipementRequest{
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
				mockRepo.EXPECT().UpsertProductEquipments(ctx, gomock.Any()).Times(1).Return(errors.New("error"))

			},
			want:    &v1.UpsertAllocateMetricEquipementResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.UpsertAllocateMetricEquipementRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.UpsertAllocateMetricEquipementRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.UpsertAllocateMetricEquipementRequest{
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
				mockRepo.EXPECT().UpsertProductEquipments(ctx, gomock.Any()).Times(1).Return(errors.New("error"))
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
			_, err := tt.s.UpsertAllocatedMetricEquipment(tt.args.ctx, tt.args.req)
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

func TestDeleteAllocatedMetricEquipment(t *testing.T) {
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
		req *v1.DropAllocateMetricEquipementRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.UpsertAllocateMetricEquipementResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.DropAllocateMetricEquipementRequest{
					Scope:   "scope1",
					Swidtag: "swidtag",
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

				mockRepo.EXPECT().DropAllocatedMetricFromEquipment(ctx, gomock.Any()).Times(1).Return(nil)
			},
			want:    &v1.UpsertAllocateMetricEquipementResponse{Success: true},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.DropAllocateMetricEquipementRequest{
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
				mockRepo.EXPECT().DropAllocatedMetricFromEquipment(ctx, gomock.Any()).Times(1).Return(errors.New("error"))

			},
			want:    &v1.UpsertAllocateMetricEquipementResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.DropAllocateMetricEquipementRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.DropAllocateMetricEquipementRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.DropAllocateMetricEquipementRequest{
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
				mockRepo.EXPECT().DropAllocatedMetricFromEquipment(ctx, gomock.Any()).Times(1).Return(errors.New("error"))
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
			_, err := tt.s.DeleteAllocatedMetricEquipment(tt.args.ctx, tt.args.req)
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

func TestGetProductCountByApp(t *testing.T) {
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
		req *v1.GetProductCountByAppRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetProductCountByAppResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetProductCountByAppRequest{
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

				mockRepo.EXPECT().GetProductCount(ctx, gomock.Any()).Times(1).Return([]db.GetProductCountRow{}, nil)
			},
			want:    &v1.GetProductCountByAppResponse{AppData: []*v1.GetProductCountByAppResponseApplications{{}}},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.GetProductCountByAppRequest{
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
				mockRepo.EXPECT().GetProductCount(ctx, gomock.Any()).Times(1).Return([]db.GetProductCountRow{}, errors.New("error"))

			},
			want:    &v1.GetProductCountByAppResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetProductCountByAppRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetProductCountByAppRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.GetProductCountByAppRequest{
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
				mockRepo.EXPECT().GetProductCount(ctx, gomock.Any()).Times(1).Return([]db.GetProductCountRow{}, errors.New("error"))
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
			_, err := tt.s.GetProductCountByApp(tt.args.ctx, tt.args.req)
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

func TestGetApplicationsByProduct(t *testing.T) {
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
		req *v1.GetApplicationsByProductRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetApplicationsByProductResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetApplicationsByProductRequest{
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

				mockRepo.EXPECT().GetApplicationsByProductID(ctx, gomock.Any()).Times(1).Return([]string{}, nil)
			},
			want:    &v1.GetApplicationsByProductResponse{},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.GetApplicationsByProductRequest{
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
				mockRepo.EXPECT().GetApplicationsByProductID(ctx, gomock.Any()).Times(1).Return([]string{}, errors.New("error"))

			},
			want:    &v1.GetApplicationsByProductResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetApplicationsByProductRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetApplicationsByProductRequest{
					Scope: "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.GetApplicationsByProductRequest{
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
				mockRepo.EXPECT().GetApplicationsByProductID(ctx, gomock.Any()).Times(1).Return([]string{}, errors.New("error"))
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
			_, err := tt.s.GetApplicationsByProduct(tt.args.ctx, tt.args.req)
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

func TestNominativeUserExport(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	timeNow := time.Now()
	testSet := []struct {
		name   string
		input  *v1.NominativeUsersExportRequest
		output *v1.ListNominativeUsersExportResponse
		mock   func(*v1.NominativeUsersExportRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "NominativeUserExportSuccess",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			output: &v1.ListNominativeUsersExportResponse{
				TotalRecords: 1,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersProducts(ctx, gomock.Any()).Return([]db.ExportNominativeUsersProductsRow{}, nil).AnyTimes()
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return([]db.ExportNominativeUsersAggregationRow{}, nil).Times(2)
			},
		},
		{
			name: "NominativeUserExportWithOutContext",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.NominativeUsersExportRequest) {},
		},
		{
			name: "FAILURE: No access to scopes",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "SSS!",
			},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.NominativeUsersExportRequest) {},
		},
		{
			name:   "NominativeUserExportWithoutContext",
			input:  &v1.NominativeUsersExportRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.NominativeUsersExportRequest) {},
		},
		{
			name: "NominativeUserExportSuccess2",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: true,
			},
			output: &v1.ListNominativeUsersExportResponse{
				TotalRecords: 1,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return([]db.ExportNominativeUsersAggregationRow{
					{
						Totalrecords:    1,
						AggregationName: sql.NullString{String: "a1", Valid: true},
						UserName:        sql.NullString{String: "u1", Valid: true},
						FirstName:       sql.NullString{String: "f1", Valid: true},
						UserEmail:       "e1",
						Profile:         sql.NullString{String: "p1", Valid: true},
						AggregationsID:  sql.NullInt32{Int32: 12, Valid: true},
						ActivationDate:  sql.NullTime{Time: timeNow, Valid: true},
					},
				}, nil).Times(2)

			},
		},
		{
			name: "NominativeUserExportFailureDB",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			ctx:    ctx,
			outErr: false,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersProducts(ctx, gomock.Any()).Return(nil, errors.New("database error")).AnyTimes()
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return(nil, errors.New("database error")).Times(1)
			},
		},
		{
			name: "NominativeUserExportEmptyResult",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			output: &v1.ListNominativeUsersExportResponse{
				TotalRecords: 0,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersProducts(ctx, gomock.Any()).Return([]db.ExportNominativeUsersProductsRow{}, nil).AnyTimes()
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return([]db.ExportNominativeUsersAggregationRow{}, nil).AnyTimes()
			},
		},
		{
			name: "NominativeUserExportPartialFailure",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			ctx:    ctx,
			outErr: false,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersProducts(ctx, gomock.Any()).Return([]db.ExportNominativeUsersProductsRow{}, nil).AnyTimes()
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return(nil, errors.New("database error")).AnyTimes()
			},
		},
		{
			name: "NominativeUserExportNonEmptyResult",
			input: &v1.NominativeUsersExportRequest{
				SortBy:    "user_email",
				SortOrder: v1.SortOrder_asc,
				Scopes:    "s1",
				IsProduct: false,
			},
			output: &v1.ListNominativeUsersExportResponse{
				TotalRecords: 2,
			},
			outErr: true,
			ctx:    ctx,
			mock: func(input *v1.NominativeUsersExportRequest) {
				dbObj.EXPECT().ExportNominativeUsersProducts(ctx, gomock.Any()).Return([]db.ExportNominativeUsersProductsRow{
					// Add test data here
				}, nil).AnyTimes()
				dbObj.EXPECT().ExportNominativeUsersAggregation(ctx, gomock.Any()).Return([]db.ExportNominativeUsersAggregationRow{
					// Add test data here
				}, nil).AnyTimes()
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "", nil, nil, &config.Config{})
			_, err := s.NominativeUserExport(test.ctx, test.input)
			if (err != nil) != test.outErr {
				t.Errorf("Failed case [%s]  because expected err [%v] is mismatched with actual err [%v]", test.name, test.outErr, err)
				return
				// } else if (got != nil && test.output != nil) && !assert.Equal(t, *got, *(test.output)) {
				// 	t.Errorf("Failed case [%s]  because expected and actual output is mismatched, act [%v], ex[ [%v]", test.name, test.output, got)

			} else {
				logger.Log.Info(" passed : ", zap.String(" test : ", test.name))
			}
		})
	}
}

func Test_GetProductInformationBySwidTag(t *testing.T) {
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
		req *v1.GetProductInformationBySwidTagRequest
	}
	tests := []struct {
		name    string
		s       *ProductServiceServer
		args    args
		setup   func()
		want    *v1.GetProductInformationBySwidTagResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetProductInformationBySwidTagRequest{
					Scope:   "scope1",
					SwidTag: "swid1",
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

				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any()).Times(1).Return(db.GetProductInformationRow{
					Swidtag:        "swid1",
					ProductName:    "p1",
					ProductEditor:  "e1",
					ProductVersion: "v1",
				}, nil)
			},
			want:    &v1.GetProductInformationBySwidTagResponse{Swidtag: "swid1", ProductName: "p1", ProductEditor: "e1", ProductVersion: "v1"},
			wantErr: false,
		},
		{name: "Db error",
			args: args{
				ctx: ctx,
				req: &v1.GetProductInformationBySwidTagRequest{
					Scope:   "scope1",
					SwidTag: "swid1",
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
				mockRepo.EXPECT().GetProductInformation(ctx, gomock.Any()).Times(1).Return(db.GetProductInformationRow{}, errors.New("error"))

			},
			want:    &v1.GetProductInformationBySwidTagResponse{},
			wantErr: true,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetProductInformationBySwidTagRequest{
					Scope:   "scope1",
					SwidTag: "swid1",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetProductInformationBySwidTagRequest{
					Scope:   "scope1233",
					SwidTag: "swid1",
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
			_, err := tt.s.GetProductInformationBySwidTag(tt.args.ctx, tt.args.req)
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
