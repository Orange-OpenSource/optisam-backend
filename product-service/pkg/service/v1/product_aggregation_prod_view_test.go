// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/product-service/pkg/api/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
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
			},
			output: &v1.ListProductAggregationViewResponse{
				TotalRecords: int32(2),
				Aggregations: []*v1.ProductAggregation{
					&v1.ProductAggregation{
						ID:              int32(100),
						Name:            "agg1",
						Editor:          "e1",
						NumApplications: int32(5),
						NumEquipments:   int32(5),
						TotalCost:       int32(25),
						Swidtags:        []string{"p1", "p2"},
					},
					&v1.ProductAggregation{
						ID:              int32(101),
						Name:            "agg2",
						Editor:          "e2",
						NumApplications: int32(10),
						NumEquipments:   int32(10),
						TotalCost:       int32(100),
						Swidtags:        []string{"p3", "p4"},
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductAggregationViewRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListAggregationsView(ctx, db.ListAggregationsViewParams{
					Scope:    userClaims.Socpes,
					PageNum:  input.PageSize * (input.PageNum - 1),
					PageSize: input.PageSize}).Return([]db.ListAggregationsViewRow{
					{
						Totalrecords:      int64(2),
						AggregationID:     int32(100),
						AggregationName:   "agg1",
						ProductEditor:     "e1",
						Swidtags:          []string{"p1", "p2"},
						NumOfApplications: int32(5),
						NumOfEquipments:   int32(5),
						TotalCost:         int32(25),
					},
					{
						Totalrecords:      int64(2),
						AggregationID:     int32(101),
						AggregationName:   "agg2",
						ProductEditor:     "e2",
						Swidtags:          []string{"p3", "p4"},
						NumOfApplications: int32(10),
						NumOfEquipments:   int32(10),
						TotalCost:         int32(100),
					},
				}, nil).Times(1)
			},
		},
		{
			name: "ListProductAggregationViewWithoutContext",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductAggregationViewRequest) {},
		},
		{
			name: "ListProductAggregationViewWithNoResult",
			input: &v1.ListProductAggregationViewRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
			},
			outErr: true,
			output: &v1.ListProductAggregationViewResponse{
				TotalRecords: int32(0),
				Aggregations: []*v1.ProductAggregation{},
			},
			ctx: context.Background(),
			mock: func(input *v1.ListProductAggregationViewRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
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
			s := NewProductServiceServer(dbObj, qObj)
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

func TestProductAggregationProductViewOptions(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductAggregationProductViewOptionsRequest
		output *v1.ProductAggregationProductViewOptionsResponse
		mock   func(*v1.ProductAggregationProductViewOptionsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ProductAggregationProductViewOptionswitCorrectData",
			input: &v1.ProductAggregationProductViewOptionsRequest{ID: int32(1)},
			output: &v1.ProductAggregationProductViewOptionsResponse{
				NumOfOptions: int32(2),
				Optioninfo: []*v1.OptionInfo{
					{
						SwidTag: "p1",
						Name:    "pname1",
						Edition: "ed1",
						Editor:  "e1",
						Version: "v1",
					},
					{
						SwidTag: "p2",
						Name:    "pname2",
						Edition: "ed2",
						Editor:  "e2",
						Version: "v2",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductAggregationProductViewOptionsRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ProductAggregationChildOptions(ctx, db.ProductAggregationChildOptionsParams{
					AggregationID: input.ID,
					Scope:         userClaims.Socpes,
				}).Return([]db.ProductAggregationChildOptionsRow{
					{
						Swidtag:        "p1",
						ProductName:    "pname1",
						ProductEditor:  "e1",
						ProductEdition: "ed1",
						ProductVersion: "v1",
					},
					{
						Swidtag:        "p2",
						ProductName:    "pname2",
						ProductEditor:  "e2",
						ProductEdition: "ed2",
						ProductVersion: "v2",
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ProductAggregationProductViewOptionswithoutContext",
			input:  &v1.ProductAggregationProductViewOptionsRequest{ID: int32(1)},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ProductAggregationProductViewOptionsRequest) {},
		},

		{
			name:  "ProductAggregationProductViewOptionswithNoResult",
			input: &v1.ProductAggregationProductViewOptionsRequest{ID: int32(1)},
			ctx:   context.Background(),
			output: &v1.ProductAggregationProductViewOptionsResponse{
				NumOfOptions: int32(0),
				Optioninfo:   []*v1.OptionInfo{},
			},
			outErr: true,
			mock:   func(input *v1.ProductAggregationProductViewOptionsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ProductAggregationProductViewOptions(test.ctx, test.input)
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

func TestListProductAggregationProductView(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListProductAggregationProductViewRequest
		output *v1.ListProductAggregationProductViewResponse
		mock   func(*v1.ListProductAggregationProductViewRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ListProductAggregationProductViewWithCorrectData",
			input: &v1.ListProductAggregationProductViewRequest{ID: int32(1)},
			output: &v1.ListProductAggregationProductViewResponse{
				Products: []*v1.Product{
					&v1.Product{
						SwidTag:           "p1",
						Name:              "pname1",
						Version:           "v1",
						Edition:           "ed1",
						Editor:            "e1",
						TotalCost:         float64(100.00),
						NumOfApplications: int32(5),
						NumofEquipments:   int32(5),
					},
					&v1.Product{
						SwidTag:           "p2",
						Name:              "pname2",
						Version:           "v2",
						Edition:           "ed2",
						Editor:            "e2",
						TotalCost:         float64(200.00),
						NumOfApplications: int32(10),
						NumofEquipments:   int32(10),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductAggregationProductViewRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListAggregationProductsView(ctx, db.ListAggregationProductsViewParams{
					AggregationID: input.ID,
					Scope:         userClaims.Socpes}).Return([]db.ListAggregationProductsViewRow{
					{
						Swidtag:           "p1",
						ProductName:       "pname1",
						ProductVersion:    "v1",
						ProductCategory:   "c1",
						ProductEditor:     "e1",
						ProductEdition:    "ed1",
						NumOfApplications: int32(5),
						NumOfEquipments:   int32(5),
						Cost:              float64(100.0),
					},
					{
						Swidtag:           "p2",
						ProductName:       "pname2",
						ProductVersion:    "v2",
						ProductCategory:   "c2",
						ProductEditor:     "e2",
						ProductEdition:    "ed2",
						NumOfApplications: int32(10),
						NumOfEquipments:   int32(10),
						Cost:              float64(200.0),
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListProductAggregationProductViewWithoutContext",
			input:  &v1.ListProductAggregationProductViewRequest{ID: int32(1)},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductAggregationProductViewRequest) {},
		},
		{
			name:   "ListProductAggregationProductViewWithnoResultSEt",
			input:  &v1.ListProductAggregationProductViewRequest{ID: int32(1)},
			output: &v1.ListProductAggregationProductViewResponse{Products: []*v1.Product{}},
			outErr: true,
			ctx:    context.Background(),
			mock: func(input *v1.ListProductAggregationProductViewRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListAggregationProductsView(ctx, db.ListAggregationProductsViewParams{
					AggregationID: input.ID,
					Scope:         userClaims.Socpes}).Return([]db.ListAggregationProductsViewRow{}, nil).Times(1)
			},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ListProductAggregationProductView(test.ctx, test.input)
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

func TestProductAggregationProductViewDetails(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ProductAggregationProductViewDetailsRequest
		output *v1.ProductAggregationProductViewDetailsResponse
		mock   func(*v1.ProductAggregationProductViewDetailsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ProductAggregationProductViewDetailsWithCorrectData",
			input: &v1.ProductAggregationProductViewDetailsRequest{ID: int32(1)},
			output: &v1.ProductAggregationProductViewDetailsResponse{
				ID:              int32(1),
				Name:            "agg",
				Editor:          "e",
				NumApplications: int32(5),
				NumEquipments:   int32(5),
				Products:        []string{"p1", "p2", "p3"},
				Editions:        []string{"ed1", "ed2", "ed3"},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductAggregationProductViewDetailsRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ProductAggregationDetails(ctx, db.ProductAggregationDetailsParams{
					AggregationID: input.ID,
					Scope:         userClaims.Socpes,
				}).Return(db.ProductAggregationDetailsRow{
					AggregationID:     int32(1),
					AggregationName:   "agg",
					ProductEditor:     "e",
					Swidtags:          []string{"p1", "p2", "p3"},
					Editions:          []string{"ed1", "ed2", "ed3"},
					NumOfApplications: int32(5),
					NumOfEquipments:   int32(5),
				}, nil).Times(1)
			},
		},
		{
			name:   "ProductAggregationProductViewDetailsWithOutContext",
			input:  &v1.ProductAggregationProductViewDetailsRequest{ID: int32(1)},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ProductAggregationProductViewDetailsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ProductAggregationProductViewDetails(test.ctx, test.input)
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
