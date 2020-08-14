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
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/workerqueue/job"
	v1 "optisam-backend/product-service/pkg/api/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"optisam-backend/product-service/pkg/worker"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetProductDetail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
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
			input: &v1.ProductRequest{SwidTag: "p"},
			output: &v1.ProductResponse{
				SwidTag: "p",
				Editor:  "e",
				Edition: "ed",
				Release: "v",
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   userClaims.Socpes}).Return(db.GetProductInformationRow{
					Swidtag:        "p",
					ProductEditor:  "e",
					ProductEdition: "ed",
					ProductVersion: "v",
				}, nil).Times(1)
			},
		},
		{
			name:   "GetProductDetailWithoutContext",
			input:  &v1.ProductRequest{SwidTag: "p1"},
			ctx:    context.Background(),
			outErr: true,
			mock:   func(input *v1.ProductRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.GetProductDetail(test.ctx, test.input)
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
			input: &v1.ProductRequest{SwidTag: "p"},
			output: &v1.ProductOptionsResponse{
				NumOfOptions: int32(2),
				Optioninfo: []*v1.OptionInfo{
					&v1.OptionInfo{
						SwidTag: "p1",
						Edition: "ed1",
						Editor:  "e1",
						Version: "v1",
					},
					&v1.OptionInfo{
						SwidTag: "p2",
						Edition: "ed2",
						Editor:  "e2",
						Version: "v2",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().GetProductOptions(ctx, db.GetProductOptionsParams{
					Swidtag: input.SwidTag,
					Scope:   userClaims.Socpes,
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
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
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
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListProductsView(ctx, db.ListProductsViewParams{
					Scope:    userClaims.Socpes,
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
					},
				}, nil).Times(2)
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
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
					Scope:         userClaims.Socpes,
					PageNum:       input.PageSize * (input.PageNum - 1),
					ApplicationID: "app",
					PageSize:      input.PageSize}).Return([]db.ListProductsViewRedirectedApplicationRow{
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
			name: "ListProductsRequestWithEquipID",
			input: &v1.ListProductsRequest{
				PageNum:  int32(1),
				PageSize: int32(10),
				SearchParams: &v1.ProductSearchParams{
					EquipmentId: &v1.StringFilter{
						Filteringkey: "equip",
					},
				},
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
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				dbObj.EXPECT().ListProductsViewRedirectedEquipment(ctx, db.ListProductsViewRedirectedEquipmentParams{
					Scope:       userClaims.Socpes,
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
			},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListProductsRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
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
						&v1.UpsertProductRequestEquipmentEquipmentuser{
							EquipmentId: "e1",
							NumUser:     int32(1),
						},
					},
				},
			},
			output: &v1.UpsertProductResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertProductRequest) {
				userClaims, ok := ctxmanage.RetrieveClaims(ctx)
				if !ok {
					t.Errorf("cannot find claims in context")
				}
				fcall := dbObj.EXPECT().UpsertProductTx(ctx, input, userClaims.UserID).Return(nil).Times(1)
				jsonData, err := json.Marshal(input)
				if err != nil {
					t.Errorf("Failed to do json marshalling")
				}
				e := worker.Envelope{Type: worker.UpsertProductRequest, Json: jsonData}

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
			name:   "UpsertProductWithoutContext",
			input:  &v1.UpsertProductRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.UpsertProductRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
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

func TestUpsertProductAggregation(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.UpsertAggregationRequest
		output *v1.UpsertAggregationResponse
		mock   func(*v1.UpsertAggregationRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name: "UpsertAggregationRequestWithActionAdd",
			input: &v1.UpsertAggregationRequest{
				AggregationId:   int32(1),
				AggregationName: "agg",
				ActionType:      "add",
				Swidtags:        []string{"p1", "p2"},
			},
			output: &v1.UpsertAggregationResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAggregationRequest) {
				dbObj.EXPECT().UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
					AggregationID:   input.AggregationId,
					AggregationName: input.AggregationName,
					Swidtags:        input.Swidtags}).Return(nil).Times(1)
			},
		},
		{
			name: "UpsertAggregationRequestWithActiondel",
			input: &v1.UpsertAggregationRequest{
				AggregationId: int32(1),
				ActionType:    "delete",
			},
			output: &v1.UpsertAggregationResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAggregationRequest) {
				dbObj.EXPECT().DeleteProductAggregation(ctx, db.DeleteProductAggregationParams{AggregationID_2: input.AggregationId}).Return(nil).Times(1)
			},
		},
		{
			name: "UpsertAggregationRequestWithActionUpsert",
			input: &v1.UpsertAggregationRequest{
				AggregationId:   int32(1),
				ActionType:      "upsert",
				AggregationName: "agg",
				Swidtags:        []string{"p1", "p3", "p4"},
			},
			output: &v1.UpsertAggregationResponse{Success: true},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.UpsertAggregationRequest) {
				dbObj.EXPECT().GetProductAggregation(ctx, db.GetProductAggregationParams{
					AggregationID:   input.AggregationId,
					AggregationName: input.AggregationName}).Return([]string{"p1", "p2"}, nil).Times(1)

				dbObj.EXPECT().UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
					AggregationID:   0,
					AggregationName: "",
					Swidtags:        []string{"p2"}}).Return(nil).Times(1)

				dbObj.EXPECT().UpsertProductAggregation(ctx, db.UpsertProductAggregationParams{
					AggregationID:   input.AggregationId,
					AggregationName: input.AggregationName,
					Swidtags:        []string{"p1", "p3", "p4"}}).Return(nil).Times(1)
			},
		},
		{
			name: "UpsertAggregationRequestWithUnknownAction",
			input: &v1.UpsertAggregationRequest{
				AggregationId: int32(1),
				ActionType:    "abc",
			},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.UpsertAggregationRequest) {},
		},
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.UpsertProductAggregation(test.ctx, test.input)
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
