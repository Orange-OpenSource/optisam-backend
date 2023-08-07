package v1

import (
	"context"
	"optisam-backend/common/optisam/logger"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	metmock "optisam-backend/metric-service/pkg/api/v1/mock"
	v1 "optisam-backend/product-service/pkg/api/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

/*
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
						Totalrecords:      int64(2),
						AggregationName:   "agg1",
						ProductEditor:     "e1",
						Swidtags:          []string{"p1", "p2"},
						NumOfApplications: int64(5),
						NumOfEquipments:   int64(5),
					},
					{
						Totalrecords:      int64(2),
						AggregationName:   "agg2",
						ProductEditor:     "e2",
						Swidtags:          []string{"p3", "p4"},
						NumOfApplications: int64(10),
						NumOfEquipments:   int64(10),
					},
				}, nil).Times(1)
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
				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListAggregationsView(ctx, db.ListAggregationsViewParams{
					Scope:    userClaims.Socpes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize}).Return([]db.ListAggregationsViewRow{}, nil).Times(1)
			},
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "",nil)
			got, err := s.ListProductAggregationView(test.ctx, test.input)
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
				dbObj.EXPECT().GetAggregationByName(ctx, db.GetAggregationByNameParams{
					Scope:           input.Scope,
					AggregationName: input.AggregationName,
				}).Return(db.AggregatedRight{
					Swidtags: []string{"p1", "p2"},
				}, nil).Times(1)
				dbObj.EXPECT().GetProductBySwidtags(ctx, db.GetProductBySwidtagsParams{
					Scope:   input.Scope,
					Swidtag: []string{"p1", "p2"},
				}).Return([]db.GetProductBySwidtagsRow{
					{
						Swidtag:        "p1",
						ProductName:    "pname1",
						ProductEditor:  "e1",
						ProductVersion: "v1",
					},
					{
						Swidtag:        "p2",
						ProductName:    "pname2",
						ProductEditor:  "e2",
						ProductVersion: "v2",
					},
				}, nil).Times(1)
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
			s := NewProductServiceServer(dbObj, qObj, nil, "",nil)
			got, err := s.GetAggregationProductsExpandedView(test.ctx, test.input)
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
*/

// func TestListProductAggregationRecords(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	dbObj := dbmock.NewMockProduct(mockCtrl)
// 	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
// 	testSet := []struct {
// 		name   string
// 		input  *v1.ListProductAggregationRecordsRequest
// 		output *v1.ListProductAggregationRecordsResponse
// 		mock   func(*v1.ListProductAggregationRecordsRequest)
// 		ctx    context.Context
// 		outErr bool
// 	}{
// 		{
// 			name:  "ListProductAggregationRecordsWithCorrectData",
// 			input: &v1.ListProductAggregationRecordsRequest{AggregationId: int32(1), Scopes: []string{"s1", "s2"}},
// 			output: &v1.ListProductAggregationRecordsResponse{
// 				ProdAggRecord: []*v1.ProductAggRecord{
// 					{
// 						SwidTag:         "p1",
// 						Name:            "pname1",
// 						Version:         "v1",
// 						Edition:         "ed1",
// 						Editor:          "e1",
// 						NumApplications: int32(5),
// 						NumEquipments:   int32(5),
// 					},
// 					{
// 						SwidTag:         "p2",
// 						Name:            "pname2",
// 						Version:         "v2",
// 						Edition:         "ed2",
// 						Editor:          "e2",
// 						NumApplications: int32(10),
// 						NumEquipments:   int32(10),
// 					},
// 				},
// 			},
// 			outErr: false,
// 			ctx:    ctx,
// 			mock: func(input *v1.ListProductAggregationRecordsRequest) {
// 				dbObj.EXPECT().ListProductsAggregationIndividual(ctx, db.ListProductsAggregationIndividualParams{
// 					AggregationName: input.AggregationName,
// 					Scope:           input.Scopes}).Return([]db.ListProductsAggregationIndividualRow{
// 					{
// 						Swidtag:           "p1",
// 						ProductName:       "pname1",
// 						ProductVersion:    "v1",
// 						ProductCategory:   "c1",
// 						ProductEditor:     "e1",
// 						ProductEdition:    "ed1",
// 						NumOfApplications: int64(5),
// 						NumOfEquipments:   int64(5),
// 					},
// 					{
// 						Swidtag:           "p2",
// 						ProductName:       "pname2",
// 						ProductVersion:    "v2",
// 						ProductCategory:   "c2",
// 						ProductEditor:     "e2",
// 						ProductEdition:    "ed2",
// 						NumOfApplications: int64(10),
// 						NumOfEquipments:   int64(10),
// 					},
// 				}, nil).Times(1)
// 			},
// 		},
// 		{
// 			name:   "ListProductAggregationRecordsWithoutContext",
// 			input:  &v1.ListProductAggregationRecordsRequest{AggregationId: int32(1), Scopes: []string{"s1", "s2"}},
// 			outErr: true,
// 			ctx:    context.Background(),
// 			mock:   func(input *v1.ListProductAggregationRecordsRequest) {},
// 		},
// 		{
// 			name:   "FAILURE: No access to Scopes",
// 			input:  &v1.ListProductAggregationRecordsRequest{AggregationId: int32(1), Scopes: []string{"s4"}},
// 			outErr: true,
// 			ctx:    ctx,
// 			mock:   func(input *v1.ListProductAggregationRecordsRequest) {},
// 		},
// 		{
// 			name:   "ListProductAggregationRecordsWithnoResultSEt",
// 			input:  &v1.ListProductAggregationRecordsRequest{AggregationId: int32(1), Scopes: []string{"s1", "s2"}},
// 			output: &v1.ListProductAggregationRecordsResponse{ProdAggRecord: []*v1.ProductAggRecord{}},
// 			outErr: true,
// 			ctx:    context.Background(),
// 			mock: func(input *v1.ListProductAggregationRecordsRequest) {
// 				userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
// 				if !ok {
// 					t.Errorf("cannot find claims in context")
// 				}
// 				dbObj.EXPECT().ListAggregationProductsView(ctx, db.ListAggregationProductsViewParams{
// 					AggregationID: input.AggregationId,
// 					Scope:         userClaims.Socpes}).Return([]db.ListAggregationProductsViewRow{}, nil).Times(1)
// 			},
// 		},
// 	}
// 	for _, test := range testSet {
// 		t.Run("", func(t *testing.T) {
// 			test.mock(test.input)
// 			s := NewProductServiceServer(dbObj, qObj, nil, "",nil)
// 			got, err := s.ListProductAggregationRecords(test.ctx, test.input)
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

func Test_productServiceServer_AggregatedRightDetails(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	metObj := metmock.NewMockMetricServiceClient(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.AggregatedRightDetailsRequest
		output *v1.AggregatedRightDetailsResponse
		mock   func(*v1.AggregatedRightDetailsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ProductAggregationProductViewDetailsWithCorrectData",
			input: &v1.AggregatedRightDetailsRequest{ID: int32(1), Scope: "s1"},
			output: &v1.AggregatedRightDetailsResponse{
				ID:              int32(1),
				Name:            "agg",
				Editor:          "e",
				NumApplications: int32(5),
				NumEquipments:   int32(5),
				Products:        []string{"p1", "p2", "p3"},
				Versions:        []string{"v1", "v2", "v3"},
				ProductNames:    []string{"pn1", "pn2"},
				NotDeployed:     false,
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.AggregatedRightDetailsRequest) {
				dbObj.EXPECT().AggregatedRightDetails(ctx, db.AggregatedRightDetailsParams{
					ID:    input.ID,
					Scope: input.Scope,
				}).Return(db.AggregatedRightDetailsRow{
					AggregationName:   "agg",
					ProductEditor:     "e",
					ProductSwidtags:   []string{"p1", "p2", "p3"},
					ProductVersions:   []string{"v1", "v2", "v3"},
					NumOfApplications: int32(5),
					NumOfEquipments:   int32(5),
					ProductNames:      []string{"pn1", "pn2"},
				}, nil).Times(1)
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
			name:   "ProductAggregationProductViewDetailsWithOutContext",
			input:  &v1.AggregatedRightDetailsRequest{ID: int32(1), Scope: "s1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.AggregatedRightDetailsRequest) {},
		},
		{
			name:   "FAILURE: No access to Scopes",
			input:  &v1.AggregatedRightDetailsRequest{ID: int32(1), Scope: "s4"},
			ctx:    ctx,
			outErr: true,
			mock:   func(input *v1.AggregatedRightDetailsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj, nil, "")
			got, err := s.AggregatedRightDetails(test.ctx, test.input)
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
