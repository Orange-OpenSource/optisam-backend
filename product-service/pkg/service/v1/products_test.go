package v1

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	appv1 "optisam-backend/application-service/pkg/api/v1"
	appmock "optisam-backend/application-service/pkg/api/v1/mock"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	"optisam-backend/common/optisam/workerqueue"
	"optisam-backend/common/optisam/workerqueue/job"
	metv1 "optisam-backend/metric-service/pkg/api/v1"
	metmock "optisam-backend/metric-service/pkg/api/v1/mock"
	v1 "optisam-backend/product-service/pkg/api/v1"
	repo "optisam-backend/product-service/pkg/repository/v1"
	dbmock "optisam-backend/product-service/pkg/repository/v1/dbmock"
	"optisam-backend/product-service/pkg/repository/v1/postgres/db"
	queuemock "optisam-backend/product-service/pkg/repository/v1/queuemock"
	dgworker "optisam-backend/product-service/pkg/worker/dgraph"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ProductRequest) {
				dbObj.EXPECT().GetProductInformation(ctx, db.GetProductInformationParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationRow{}, sql.ErrNoRows).Times(1)
				dbObj.EXPECT().GetProductInformationFromAcqright(ctx, db.GetProductInformationFromAcqrightParams{
					Swidtag: input.SwidTag,
					Scope:   input.Scope}).Return(db.GetProductInformationFromAcqrightRow{
					Swidtag:       "p",
					ProductName:   "pn",
					ProductEditor: "e",
					Version:       "v",
					Metrics:       []string{"m1", "m2", "m3"},
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
	}
	for _, test := range testSet {
		t.Run("", func(t *testing.T) {
			test.mock(test.input)
			s := &productServiceServer{
				productRepo: dbObj,
				queue:       qObj,
				metric:      metObj,
			}
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
			s := NewProductServiceServer(dbObj, qObj, nil, "")
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
				Scopes: []string{"s1"},
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
						NumOfApplications: int32(1),
						NumofEquipments:   int32(2),
					},
				},
			},
			outErr: false,
			ctx:    ctx,
			mock: func(input *v1.ListProductsRequest) {
				mockApp := appmock.NewMockApplicationServiceClient(mockCtrl)
				app = mockApp
				mockApp.EXPECT().GetEquipmentsByApplication(ctx, &appv1.GetEquipmentsByApplicationRequest{
					Scope:         "s1",
					ApplicationId: "app",
				}).Times(1).Return(&appv1.GetEquipmentsByApplicationResponse{
					EquipmentId: []string{"eq1", "eq2", "eq3"},
				}, nil)
				dbObj.EXPECT().ListProductsViewRedirectedApplication(ctx, db.ListProductsViewRedirectedApplicationParams{
					Scope:         input.Scopes,
					PageNum:       input.PageSize * (input.PageNum - 1),
					ApplicationID: "app",
					IsEquipmentID: true,
					EquipmentIds:  []string{"eq1", "eq2", "eq3"},
					PageSize:      input.PageSize}).Return([]db.ListProductsViewRedirectedApplicationRow{
					{
						Totalrecords:      int64(1),
						Swidtag:           "p",
						ProductName:       "n",
						ProductVersion:    "v",
						ProductCategory:   "c",
						ProductEditor:     "e",
						ProductEdition:    "ed",
						NumOfApplications: int32(1),
						NumOfEquipments:   int32(2),
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
			s := &productServiceServer{
				productRepo: dbObj,
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
			s := NewProductServiceServer(dbObj, qObj, nil, "")
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

func Test_productServiceServer_DropProductData(t *testing.T) {
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
		s       *productServiceServer
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
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.DropProductData(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.DropProductData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.DropProductData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_productServiceServer_GetEquipmentsByProduct(t *testing.T) {
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
		s       *productServiceServer
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
			tt.s = &productServiceServer{
				productRepo: rep,
				queue:       queue,
			}
			got, err := tt.s.GetEquipmentsByProduct(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("productServiceServer.GetEquipmentsByProduct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("productServiceServer.GetEquipmentsByProduct() = %v, want %v", got, tt.want)
			}
		})
	}
}
