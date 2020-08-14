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
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/product-service/pkg/api/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	ctx = ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "SuperAdmin",
		Socpes: []string{"s1", "s2", "s3"},
	})
)

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	os.Exit(m.Run())
}

func TestListEditors(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListEditorsRequest
		output *v1.ListEditorsResponse
		mock   func(*v1.ListEditorsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:   "ListEditorsWithCorrectData",
			input:  &v1.ListEditorsRequest{Scopes: []string{"s1", "s2", "s3"}},
			output: &v1.ListEditorsResponse{Editors: []string{"e1", "e2", "e3"}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListEditorsRequest) {
				dbObj.EXPECT().ListEditors(ctx, input.Scopes).Return([]string{"e1", "e2", "e3"}, nil).Times(1)
			},
		},
		{
			//This test case need to change in future
			name:   "ListEditorsWithScopeMismatch",
			input:  &v1.ListEditorsRequest{Scopes: []string{"s5", "s6"}},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListEditorsRequest) {
				var scopes []string
				dbObj.EXPECT().ListEditors(ctx, scopes).Return(scopes, nil).Times(1)
			},
		},
		{
			name:   "ListEditorsWithoutContext",
			input:  &v1.ListEditorsRequest{},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListEditorsRequest) {},
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ListEditors(test.ctx, test.input)
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

func TestListEditorProducts(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockProduct(mockCtrl)
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	testSet := []struct {
		name   string
		input  *v1.ListEditorProductsRequest
		output *v1.ListEditorProductsResponse
		mock   func(*v1.ListEditorProductsRequest)
		ctx    context.Context
		outErr bool
	}{
		{
			name:  "ListEditorProductsWithCorrectData",
			input: &v1.ListEditorProductsRequest{Editor: "e1", Scopes: []string{"s1", "s2", "s3"}},
			output: &v1.ListEditorProductsResponse{
				Products: []*v1.Product{
					&v1.Product{
						SwidTag: "swid1",
						Name:    "p1",
					},
					&v1.Product{
						SwidTag: "swid2",
						Name:    "p2",
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListEditorProductsRequest) {
				dbObj.EXPECT().GetProductsByEditor(ctx, db.GetProductsByEditorParams{
					ProductEditor: input.Editor,
					Scopes:        input.Scopes}).Return([]db.GetProductsByEditorRow{
					{
						Swidtag:     "swid1",
						ProductName: "p1",
					},
					{
						Swidtag:     "swid2",
						ProductName: "p2",
					},
				}, nil).Times(1)
			},
		},
		{
			name:   "ListEditorProductsWithoutContext",
			input:  &v1.ListEditorProductsRequest{Scopes: []string{"s4", "s5"}, Editor: "e1"},
			outErr: true,
			ctx:    context.Background(),
			mock:   func(input *v1.ListEditorProductsRequest) {},
		},
	}

	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := NewProductServiceServer(dbObj, qObj)
			got, err := s.ListEditorProducts(test.ctx, test.input)
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
