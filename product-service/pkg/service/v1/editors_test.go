package v1

import (
	"context"
	"errors"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	metmock "optisam-backend/metric-service/pkg/api/v1/mock"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	ctx = grpc_middleware.AddClaims(context.Background(), &claims.Claims{
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
			name:   "ListEditorsWithScopeMismatch",
			input:  &v1.ListEditorsRequest{Scopes: []string{"s5", "s6"}},
			outErr: true,
			ctx:    ctx,
			mock:   func(input *v1.ListEditorsRequest) {},
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
			s := NewProductServiceServer(dbObj, qObj, nil, "")
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
					{
						SwidTag: "swid1",
						Name:    "p1",
						Version: "v1",
					},
					{
						SwidTag: "swid2",
						Name:    "p2",
						Version: "v2",
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
						Swidtag:        "swid1",
						ProductName:    "p1",
						ProductVersion: "v1",
					},
					{
						Swidtag:        "swid2",
						ProductName:    "p2",
						ProductVersion: "v2",
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
			s := NewProductServiceServer(dbObj, qObj, nil, "")
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

func TestGetRightsInfoByEditor(t *testing.T) {
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
		req *v1.GetRightsInfoByEditorRequest
	}
	tests := []struct {
		name    string
		s       *productServiceServer
		args    args
		setup   func()
		want    *v1.GetRightsInfoByEditorResponse
		wantErr bool
	}{
		{name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.GetRightsInfoByEditorRequest{
					Editor: "editor1",
					Scope:  "scope1",
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
				mockRepo.EXPECT().GetAcqRightsByEditor(ctx, db.GetAcqRightsByEditorParams{
					ProductEditor: "editor1",
					Scope:         "scope1",
				}).Times(1).Return([]db.GetAcqRightsByEditorRow{
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						Metric:              "met1",
						AvgUnitPrice:        2.0,
						NumLicensesAcquired: 3,
					},
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						Metric:              "met1",
						AvgUnitPrice:        3.0,
						NumLicensesAcquired: 3,
					},
				}, nil)
				mockRepo.EXPECT().GetAggregationByEditor(ctx, db.GetAggregationByEditorParams{
					ProductEditor: "editor1",
					Scope:         "scope1",
				}).Times(1).Return([]db.GetAggregationByEditorRow{
					{
						AggregationName:     "agg2",
						Swidtags:            "swid1",
						Sku:                 "sku1",
						Metric:              "met1",
						AvgUnitPrice:        2.0,
						NumLicensesAcquired: 3,
					},
					{
						AggregationName:     "agg2",
						Swidtags:            "swid1",
						Sku:                 "sku1",
						Metric:              "met1",
						AvgUnitPrice:        3.0,
						NumLicensesAcquired: 3,
					},
				}, nil)
			},
			want: &v1.GetRightsInfoByEditorResponse{
				EditorRights: []*v1.RightsInfoByEditor{
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						MetricName:          "met1",
						AvgUnitPrice:        2.0,
						NumLicensesAcquired: 3,
					},
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						MetricName:          "met1",
						AvgUnitPrice:        3.0,
						NumLicensesAcquired: 3,
					},
					{
						AggregationName:     "agg2",
						Swidtag:             "swid1",
						Sku:                 "sku1",
						MetricName:          "met1",
						AvgUnitPrice:        2.0,
						NumLicensesAcquired: 3,
					},
					{
						AggregationName:     "agg2",
						Swidtag:             "swid1",
						Sku:                 "sku1",
						MetricName:          "met1",
						AvgUnitPrice:        3.0,
						NumLicensesAcquired: 3,
					},
				},
			},
			wantErr: false,
		},
		{name: "FAILURE-can not find claims in context",
			args: args{
				ctx: context.Background(),
				req: &v1.GetRightsInfoByEditorRequest{
					Editor: "editor",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-Scope Validation error",
			args: args{
				ctx: ctx,
				req: &v1.GetRightsInfoByEditorRequest{
					Editor: "editor",
					Scope:  "scope4",
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAcqRightsByEditor",
			args: args{
				ctx: ctx,
				req: &v1.GetRightsInfoByEditorRequest{
					Editor: "editor",
					Scope:  "scope1",
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
				mockRepo.EXPECT().GetAcqRightsByEditor(ctx, db.GetAcqRightsByEditorParams{
					ProductEditor: "editor",
					Scope:         "scope1",
				}).Times(1).Return([]db.GetAcqRightsByEditorRow{}, errors.New("internal"))
			},
			wantErr: true,
		},
		{name: "FAILURE-db/GetAggregationByEditor",
			args: args{
				ctx: ctx,
				req: &v1.GetRightsInfoByEditorRequest{
					Editor: "editor",
					Scope:  "scope1",
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
				mockRepo.EXPECT().GetAcqRightsByEditor(ctx, db.GetAcqRightsByEditorParams{
					ProductEditor: "editor",
					Scope:         "scope1",
				}).Times(1).Return([]db.GetAcqRightsByEditorRow{
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						Metric:              "met1",
						AvgUnitPrice:        2.0,
						NumLicensesAcquired: 3,
					},
					{
						Sku:                 "sku1",
						Swidtag:             "swid1",
						Metric:              "met1",
						AvgUnitPrice:        3.0,
						NumLicensesAcquired: 3,
					},
				}, nil)
				mockRepo.EXPECT().GetAggregationByEditor(ctx, db.GetAggregationByEditorParams{
					ProductEditor: "editor",
					Scope:         "scope1",
				}).Times(1).Return([]db.GetAggregationByEditorRow{}, errors.New("internal"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
				metric:      met,
			}
			got, err := tt.s.GetRightsInfoByEditor(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.GetRightsInfoByEditor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.GetRightsInfoByEditor() = %v, want %v", got, tt.want)
			}
		})
	}
}
