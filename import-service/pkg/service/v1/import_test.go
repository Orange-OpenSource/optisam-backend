package v1

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bou.ke/monkey"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/config"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/pkg/repository/v1/postgres/db"
	v1Acc "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/account-service/pkg/api/v1"
	mockAcc "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/account-service/pkg/api/v1/mock"
	v1Catalog "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/catalog-service/pkg/api/v1"
	catmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/catalog-service/pkg/api/v1/mock"
	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/dps-service/pkg/api/v1"
	mock "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/dps-service/pkg/api/v1/mock"
	v1Product "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/product-service/pkg/api/v1"
	prodmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/product-service/pkg/api/v1/mock"
	v1Sim "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/simulation-service/pkg/api/v1"
	simmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/import-service/thirdparty/simulation-service/pkg/api/v1/mock"
	"google.golang.org/grpc"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	rest_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/rest"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_UploadFiles(t *testing.T) {
	var dpsClient v1.DpsServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name:   "Valid scenario",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().DataAnalysis(gomock.Any(), gomock.Any()).Return(&v1.DataAnalysisResponse{}, nil).AnyTimes()

			},
			code: 200,
		},
		{
			name:   "upload type not supported",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "na")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().DataAnalysis(gomock.Any(), gomock.Any()).Return(&v1.DataAnalysisResponse{}, nil).AnyTimes()

			},
			code: 400,
		},
		{
			name:   "data anylysis err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().DataAnalysis(gomock.Any(), gomock.Any()).Return(&v1.DataAnalysisResponse{}, errors.New("err")).AnyTimes()

			},
			code: 500,
		},
		{
			name:   "fail os.mkdir err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return fmt.Errorf("simulated error")
				})
			},
			code: 500,
		},
		{
			name:   "fail  io.create err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockError := fmt.Errorf("simulated error")

				monkey.Patch(os.Create, func(fn string) (*os.File, error) {
					return &os.File{}, mockError
				})
			},
			code: 500,
		},

		{
			name:   "ScopeMissing",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/upload", "", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "MissingUploadType",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name: "ScopeValidationFailure",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN2", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 401,
		},
		{
			name: "Claims not found",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN2", "files", []string{"testdata/temp2.xlsx"}, "GENERIC", "analysis")
				request = request.WithContext(context.Background())

				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "File size err",
			fields: fields{&config.Config{MaxFileSize: int64(0), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/upload", "GEN", "files", []string{"testdata/catalog/data1.xlsx"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name: "FAILURE - IncorrectFileExtenison",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(20), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "Scope1", "files", []string{"testdata/temp.xls"}, "GENERIC", "analysis")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
	}
	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
			}
			rec := httptest.NewRecorder()
			i.UploadFiles(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func Test_UploadFiles1(t *testing.T) {
	var dpsClient v1.DpsServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name:   "Valid scenario",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().StoreCoreFactorReference(gomock.Any(), gomock.Any()).Return(&v1.StoreReferenceDataResponse{}, nil).AnyTimes()
			},
			code: 200,
		},
		{
			name:   "Valid scenario",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().StoreCoreFactorReference(gomock.Any(), gomock.Any()).Return(&v1.StoreReferenceDataResponse{}, nil).AnyTimes()
			},
			code: 200,
		},
		{
			name:   "data anylysis err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient.EXPECT().StoreCoreFactorReference(gomock.Any(), gomock.Any()).Return(&v1.StoreReferenceDataResponse{}, errors.New("err")).AnyTimes()

			},
			code: 500,
		},
		{
			name:   "File size err",
			fields: fields{&config.Config{MaxFileSize: int64(0), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "File sheet err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023-1.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name:   "File sheet err-2",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newCorefileUploadRequest("/api/v1/import/upload", "GEN", "file", "testdata/reference_1.2.0_2023-2.xlsx", "GENERIC", "corefactor")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 500,
		},
	}
	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
			}
			rec := httptest.NewRecorder()
			i.UploadFiles(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}
func Test_importServiceServer_UploadDataHandler(t *testing.T) {

	var dpsClient v1.DpsServiceClient
	var accClient v1Acc.AccountServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name    string
		fields  fields
		setup   func()
		cleanup func()
		code    int
	}{
		{
			name: "SUCCESS - Data Single file with correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{}, nil).AnyTimes()
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil).AnyTimes()
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		// {
		// 	name: "SUCCESS - os.create err",
		// 	// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
		// 	fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
		// 	setup: func() {
		// 		request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
		// 		if err != nil {
		// 			logger.Log.Error("Failed creating request", zap.Error(err))
		// 			t.Fatal(err)
		// 		}
		// 		mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
		// 		mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
		// 		accClient = mockAccClient
		// 		dpsClient = mockDPSClient
		// 		// Gotcha Need to check
		// 		mockAccClient.EXPECT().GetScope(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{}, nil).AnyTimes()
		// 		mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil).AnyTimes()
		// 		monkey.Patch(os.Create, func(fn string) (*os.File, error) {
		// 			return &os.File{}, fmt.Errorf("err")
		// 		})
		// 	},
		// 	cleanup: func() {
		// 		err = os.RemoveAll("data")
		// 		if err != nil {
		// 			fmt.Println(err)
		// 			t.Fatal(err)
		// 		}
		// 	},
		// 	code: 500,
		// },
		{
			name: "SUCCES- size err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(-1), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{}, nil).AnyTimes()
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil).AnyTimes()
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name: "SUCCESS - user role",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"Scope1", "France", "GEN"}})
				request = request.WithContext(ctx)
				// mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				// mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				// accClient = mockAccClient
				// dpsClient = mockDPSClient
				// // Gotcha Need to check
				// mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
				// 	ScopeCode: "France",
				// 	ScopeName: "France scope",
				// 	ScopeType: "SPECIFIC",
				// }, nil)
				// mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
				// 	Scope: "France", Files: []string{"France_products.csv"}, Type: "data", UploadedBy: "TestUser",
				// }).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "SUCCESS - ctx bg",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := context.Background()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				request = request.WithContext(ctx)
				// mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				// mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				// accClient = mockAccClient
				// dpsClient = mockDPSClient
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "Scope blank",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "", "files", []string{"testdata/products.csv"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				// mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				// mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				// accClient = mockAccClient
				// dpsClient = mockDPSClient
				// // Gotcha Need to check
				// mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
				// 	ScopeCode: "France",
				// 	ScopeName: "France scope",
				// 	ScopeType: "SPECIFIC",
				// }, nil)
				// mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
				// 	Scope: "France", Files: []string{"France_products.csv"}, Type: "data", UploadedBy: "TestUser",
				// }).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "Scope not matching",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "FRA", "files", []string{"testdata/products.csv"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				// mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				// mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				// accClient = mockAccClient
				// dpsClient = mockDPSClient
				// // Gotcha Need to check
				// mockAccClient.EXPECT().GetScope(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
				// 	ScopeCode: "France",
				// 	ScopeName: "France scope",
				// 	ScopeType: "SPECIFIC",
				// }, nil)
				// mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "FAILURE - Unable to get scope type info",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(nil, errors.New("internal"))
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "FAILURE - Data upload not allowed for generic scope type",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "GENERIC",
				}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "FAILURE - Data Single file with incorrect naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products_1.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - Data Multiple Files with correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/products_equipments.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv", "France_products_equipments.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "SUCCESS - monkeypatch os mkdir",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/products_equipments.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv", "France_products_equipments.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return fmt.Errorf("simulated error")
				})
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - monkeypatch os hdr open dir",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/products_equipments.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv", "France_products_equipments.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return fmt.Errorf("simulated error")
				})
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - monkeypatch os Create",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/products_equipments.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv", "France_products_equipments.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
				monkey.Patch(os.Create, func(name string) (*os.File, error) {
					return nil, errors.New("failed to create file")
				})
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "FAILURE - Data Multiple Files with some having incorrect correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/applications.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv", "France_products_equipments.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - Notify upload err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(100), Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_products.csv"}, Type: "data", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: false}, errors.New("Injection is already running"))
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
				accClient: accClient,
			}
			rec := httptest.NewRecorder()
			i.UploadDataHandler(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func Test_ImportServiceServer_UploadMetaDataHandler(t *testing.T) {

	var dpsClient v1.DpsServiceClient
	var accClient v1Acc.AccountServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name    string
		fields  fields
		setup   func()
		cleanup func()
		code    int
	}{
		{
			name: "SUCCESS - Metadata Single file with correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_metadata_laptop.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "SUCCESS - notify upload",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_metadata_laptop.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: false}, errors.New("err"))
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "SUCCESS - file size",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(-1), Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_metadata_laptop.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name: "SUCCESS - os.mkdir",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return fmt.Errorf("simulated error")
				})
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS large file err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/test"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				// Gotcha Need to check
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - scope not in claims",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "NotINClaims", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "SUCCESS - ctx user",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"Scope1", "France", "GEN"}})
				request = request.WithContext(ctx)

			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "SUCCESS - ctx bg",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := context.Background()
				request = request.WithContext(ctx)

			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - scope missing",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name: "FAILURE - Unable to get scope info",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(nil, errors.New("internal"))
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "FAILURE - Metadata upload is not allowed for generic scope",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadataa_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "GENERIC",
				}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			name: "FAILURE - Metadata Single file with incorrect naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop1.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"metadata_laptop1.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "SUCCESS - Metadata Multiple Files with correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv", "testdata/metadata_desktop.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"France_metadata_laptop.csv", "France_metadata_desktop.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name: "FAILURE - Metadata Multiple Files with some having incorrect correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadataa_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "France", "files", []string{"testdata/metadata_laptop.csv", "testdata/metadata_desktop1.csv"}, "", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				mockAccClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = mockAccClient
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockAccClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "France"}).Times(1).Return(&v1Acc.Scope{
					ScopeCode: "France",
					ScopeName: "France scope",
					ScopeType: "SPECIFIC",
				}, nil)
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "France", Files: []string{"metadata_laptop.csv", "metadata_desktop1.csv"}, Type: "metadata", UploadedBy: "TestUser",
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			cleanup: func() {
				err = os.RemoveAll("data")
				if err != nil {
					fmt.Println(err)
					t.Fatal(err)
				}
			},
			code: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
				accClient: accClient,
			}
			rec := httptest.NewRecorder()
			i.UploadMetaDataHandler(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func newfileUploadRequest(uri string, scope string, paramName string, files []string, scopeType, uploadType string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("scope", scope)
	if scopeType == GENERIC {
		_ = writer.WriteField("file", strings.Split(files[0], "/")[1])
	}

	_ = writer.WriteField("uploadType", uploadType)
	_ = writer.WriteField("scopeType", scopeType)
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			logger.Log.Error("Failed opening file", zap.Error(err))
			return nil, err
		}
		defer file.Close()
		part, err := writer.CreateFormFile(paramName, filepath.Base(f))
		_, err = io.Copy(part, file)
		if err != nil {
			logger.Log.Error("Failed copying file", zap.Error(err))
			return nil, err
		}
	}
	err := writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing Writer", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Add Context for User Information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"Scope1", "France", "GEN"}})
	req = req.WithContext(ctx)
	return req, err
}

func newCorefileUploadRequest(uri string, scope string, paramName string, f string, scopeType, uploadType string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("scope", scope)
	_ = writer.WriteField("uploadType", uploadType)
	_ = writer.WriteField("scopeType", scopeType)
	file, err := os.Open(f)
	if err != nil {
		logger.Log.Error("Failed opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()
	part, err := writer.CreateFormFile("file", filepath.Base(f))
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Log.Error("Failed copying file", zap.Error(err))
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing Writer", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Add Context for User Information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"Scope1", "France", "GEN"}})
	req = req.WithContext(ctx)
	return req, err

}

func Test_UploadGlobalDataHandler(t *testing.T) {
	var dpsClient v1.DpsServiceClient
	var accClient v1Acc.AccountServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name: "SUCCESS - CorrectDataFile",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/good_1234_temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/good_1234_temp2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "GEN"}).Return(&v1Acc.Scope{
					ScopeCode: "GEN",
					ScopeType: "GENERIC",
				}, nil).Times(1)

				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					AnalysisId: "1234",
					Scope:      "GEN", Files: []string{"temp2.xlsx"}, Type: "globaldata", UploadedBy: "TestUser", ScopeType: v1.NotifyUploadRequest_GENERIC,
				}).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			code: 200,
		},
		{
			name: "SUCCESS - specific scopetype",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/applications.csv"}, "SPECIFIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
					ScopeCode: "SPC",
					ScopeType: "SPECIFIC",
				}, nil).Times(1)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				})
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, errors.New("Injection is already running"))
			},
			code: 500,
		},
		{
			name: "SUCCESS - specific scopetype size err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(-1), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/applications.csv"}, "SPECIFIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
					ScopeCode: "SPC",
					ScopeType: "SPECIFIC",
				}, nil).Times(1)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				})
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, errors.New("Injection is already running"))
			},
			code: 400,
		},
		{
			name: "SUCCESS - specific r",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/applications.csv"}, "SPECIFIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
					ScopeCode: "SPC",
					ScopeType: "SPECIFIC",
				}, nil).Times(1)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				})
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, errors.New("Injection is already running"))
			},
			code: 500,
		},
		{
			name: "SUCCESS - specific scopetype not csv",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/good_1234_temp2.xlsx"}, "SPECIFIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
					ScopeCode: "SPC",
					ScopeType: "SPECIFIC",
				}, nil).Times(1)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				})
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, errors.New("Injection is already running"))
			},
			code: 400,
		},
		{
			name: "SUCCESS - specific scopetype success",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/applications.csv"}, "SPECIFIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
					ScopeCode: "SPC",
					ScopeType: "SPECIFIC",
				}, nil).Times(1)
				monkey.Patch(os.MkdirAll, func(path string, perm os.FileMode) error {
					return nil
				})

				// Gotcha Need to check
				myMap := map[string]int32{}

				// Add key-value pairs to the map
				myMap["key1"] = 10

				mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true, FileUploadId: myMap}, nil)
				monkey.Patch(os.Create, func(fn string) (*os.File, error) {
					return &os.File{}, nil
				})

			},
			code: 500,
		},
		{
			name: "SUCCESS - get scope err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/good_1234_temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/good_1234_temp2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "GEN"}).Return(&v1Acc.Scope{
					ScopeCode: "GEN",
					ScopeType: "GENERIC",
				}, errors.New("err")).Times(1)

				// Gotcha Need to check
			},
			code: 500,
		},
		{
			name: "SUCCESS - scope forbidden",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GENNOTFOUND", "files", []string{"testdata/good_1234_temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/good_1234_temp2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()

				// Gotcha Need to check
			},
			code: 403,
		},
		{
			name: "SUCCESS - scope err",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "", "files", []string{"testdata/good_1234_temp2.xlsx"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp, err := os.Create("testdata/GEN/analysis/good_1234_temp2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				fp.Close()
				// mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				// dpsClient = mockDPSClient
				// accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				// accClient = accMockClient
				// accMockClient.EXPECT().GetScope(gomock.Any(), gomock.Any()).Return(&v1Acc.Scope{
				// 	ScopeCode: "GEN",
				// 	ScopeType: "GENERIC",
				// }, nil).Times(1)

				// Gotcha Need to check
				// mockDPSClient.EXPECT().NotifyUpload(gomock.Any(), gomock.Any()).AnyTimes().Return(&v1.NotifyUploadResponse{Success: true}, nil)
			},
			code: 400,
		},
		{
			name: "FAILURE - IncorrectFileExtenison",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/globaldata", "GEN", "files", []string{"testdata/temp.xls"}, "GENERIC", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				accMockClient := mockAcc.NewMockAccountServiceClient(mockCtrl)
				accClient = accMockClient
				accMockClient.EXPECT().GetScope(request.Context(), &v1Acc.GetScopeRequest{Scope: "GEN"}).Return(&v1Acc.Scope{
					ScopeCode: "GEN",
					ScopeType: "GENERIC",
				}, nil).Times(1)
			},
			code: 400,
		},
	}
	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
				accClient: accClient,
			}
			rec := httptest.NewRecorder()
			i.UploadGlobalDataHandler(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

type queryParam struct {
	Key   string
	Value string
}

func Test_ImportServiceServer_DownloadGlobalDataErrors(t *testing.T) {
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	var dpsClient v1.DpsServiceClient
	defer mockCtrl.Finish()
	type fields struct {
		config *config.Config
	}
	tests := []struct {
		name   string
		i      *ImportServiceServer
		setup  func()
		fields fields
		code   int
	}{
		{name: "FAILURE - File name is missing",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download?downloadType=analysis&scope=scope1&fileName=", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 500,
		},
		{name: "FAILURE - ClaimsNotFound",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "error")
				q.Add("scope", "GEN")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				request = req
			},
			code: 500,
		},
		{name: "FAILURE - RoleValidationFailed",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "error")
				q.Add("scope", "scope1")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 403,
		},
		{name: "FAILURE - BLANK SCOPE",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "error")
				q.Add("scope", "")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "FAILURE - DOWNLOD TYPE blank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "")
				q.Add("scope", "france")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 403,
		},
		{name: "FAILURE - DOWNLOD TYPE blank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "")
				q.Add("scope", "france")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 403,
		},
		{name: "FAILURE - DOWNLOD TYPE",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "/api/v1/import/download", payload)
				q := req.URL.Query()
				q.Add("fileName", "1_scope1_error_temp.zip")
				q.Add("downloadType", "")
				q.Add("scope", "france")
				req.URL.RawQuery = q.Encode()
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 403,
		},
		{name: "Success - AnalysisReport",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "analysis"}, {Key: "scope", Value: "scope1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 200,
		},
		{name: "Success - upload id err",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "error"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "uid"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "Success - upload id err blank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "error"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: ""}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "Success - upload rpc err dps",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "error"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{}, errors.New("err")).AnyTimes()
			},
			code: 500,
		},
		{name: "Success - upload ",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "source"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{ScopeType: "GENERIC", IsOlderGeneric: true}, nil).AnyTimes()
			},
			code: 404,
		},
		{name: "Success - upload 2",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "source"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{ScopeType: "GENERIC", IsOlderGeneric: false}, nil).AnyTimes()
			},
			code: 500,
		},
		{name: "Success - upload 2",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "source"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{ScopeType: "NEW", IsOlderGeneric: false}, nil).AnyTimes()
			},
			code: 404,
		},
		{name: "Success - upload 3",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: "def"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{ScopeType: "NEW", IsOlderGeneric: false}, nil).AnyTimes()
			},
			code: 400,
		},
		{name: "Success - analysis file name",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: ""}, {Key: "downloadType", Value: "analysis"}, {Key: "scope", Value: "scope1"}, {Key: "uploadId", Value: "1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{}, errors.New("err")).AnyTimes()
			},
			code: 500,
		},
		{name: "Success - download type blank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temp.xlsx"}, {Key: "downloadType", Value: ""}, {Key: "scope", Value: "scope1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "Success - DownloadError",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				req, err := newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "uploadId", Value: "123"}, {Key: "downloadType", Value: "error"}, {Key: "scope", Value: "scope1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				mockDPSClient.EXPECT().GetAnalysisFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v1.GetAnalysisFileInfoResponse{}, nil).AnyTimes()

			},
			code: 500,
		},
		{name: "FAILURE - ScopeValidationError",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "1_scope3_error_temp.xlsx"}, {Key: "downloadType", Value: "error"}, {Key: "scope", Value: "scope12"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 401,
		},
		{name: "FAILURE - File does not exist",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "testdata"}}},
			setup: func() {
				request, err = newfileDownloadRequest("/api/v1/import/download", "scope1", []*queryParam{{Key: "fileName", Value: "scope1_temo.xlsx"}, {Key: "downloadType", Value: "analysis"}, {Key: "scope", Value: "scope1"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				dpsClient: dpsClient,
			}
			rec := httptest.NewRecorder()
			i.DownloadFile(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
			if tt.code == 200 {
				expfileData, err := ioutil.ReadFile("testdata/scope1/errors/1_scope1_error_temp.xlsx")
				if err != nil {
					t.Errorf("Error in reading expected file. err: %v", err)
					return
				}
				if string(rec.Body.Bytes()) != string(expfileData) {
					t.Error("response is not same")
					return
				}
			}
		})
	}
}

func newfileDownloadRequest(uri string, scope string, qParams []*queryParam) (*http.Request, error) {
	payload := &bytes.Buffer{}
	req, err := http.NewRequest("GET", uri, payload)
	if qParams != nil {
		q := req.URL.Query()
		for _, qp := range qParams {
			q.Add(qp.Key, qp.Value)
		}
		req.URL.RawQuery = q.Encode()
	}
	// Add Context for User Information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{scope, "france"}})
	req = req.WithContext(ctx)
	return req, err
}

func Test_UploadCatalogData(t *testing.T) {
	var catalogClient v1Catalog.ProductCatalogClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name: "Success - success case",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data1.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 200,
		},
		{
			name: "Success - ctx bg",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data1.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := context.Background()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				request = request.WithContext(ctx)
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				// catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 400,
		},
		{
			name: "Success - no permission",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data1.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "Admin", Socpes: []string{"Scope1", "France", "GEN"}})
				request = request.WithContext(ctx)
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				// catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 400,
		},
		{
			name: "Success - file2",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				// catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 200,
		},
		{
			name: "Success - file3",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data3.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 200,
		},
		{
			name: "Success - file4",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data4.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				// catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 500,
		},
		{
			name: "Success - file 5",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(100)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data5.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 200,
		},
		{
			name: "Success - file size",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "catalog"}, MaxFileSize: int64(-1)}},
			setup: func() {
				request, err = newCatalogUploadRequest("api/v1/import/uploadcatalogdata", "testdata/catalog/data1.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				catalogmock := catmock.NewMockProductCatalogClient(mockCtrl)
				catalogClient = catalogmock
				// catalogmock.EXPECT().BulkFileUpload(gomock.Any(), gomock.Any()).Return(&v1Catalog.UploadResponse{}, nil).Times(1)

			},
			code: 400,
		},
	}
	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			i := &ImportServiceServer{
				Config:        tt.fields.config,
				catalogClient: catalogClient,
			}
			rec := httptest.NewRecorder()
			i.UploadCatalogData(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}
func newCatalogUploadRequest(uri string, f string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, err := os.Open(f)
	if err != nil {
		logger.Log.Error("Failed opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()
	part, err := writer.CreateFormFile("file", filepath.Base(f))
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Log.Error("Failed copying file", zap.Error(err))
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing Writer", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Add Context for User Information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"Scope1", "France", "GEN"}})
	req = req.WithContext(ctx)
	return req, err
}

func TestGetglobalFileExtension(t *testing.T) {
	testCases := []struct {
		fileName string
		expected string
	}{
		{"", ""},                           // Empty file name
		{"file", ""},                       // File name without extension
		{"file.txt", ".txt"},               // File name with extension
		{"file.tar.gz", ".gz"},             // File name with multiple extensions
		{"file.with.multiple.ext", ".ext"}, // File name with multiple extensions, testing last extension retrieval
	}

	for _, tc := range testCases {
		result := getglobalFileExtension(tc.fileName)
		assert.Equal(t, tc.expected, result)
	}
}

func Test_CreateConfigHandler(t *testing.T) {
	var simClient v1Sim.SimulationServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name:   "Valid scenario",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 200,
		},
		{
			name:   "Valid scenario csv",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 200,
		},
		{
			name:   "err scenario csv1",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu1.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 404,
		},
		{
			name:   "err scenario csv2",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu2.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 422,
		},
		{
			name:   "create conf err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, errors.New("err")).AnyTimes()
			},
			code: 500,
		},
		{
			name:   "ScopeMissing",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GENNA", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 401,
		},
		{
			name:   "ScopeMissingBlank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().CreateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.CreateConfigResponse{}, nil).AnyTimes()
			},
			code: 400,
		},
		{
			name:   "MissingConf name",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "Config greater than 50",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "123456789012345678901234567890123456789012345678901234567890", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "equip type blank",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 400,
		},
		{
			name:   "ctx bg",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "")
				request = request.WithContext(context.Background())
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			name:   "user role",
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"Scope1", "France", "GEN"}})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 200,
		},
		{
			name:   "FAILURE - IncorrectFileExtenison",
			fields: fields{&config.Config{MaxFileSize: int64(20), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				request, err = newConfigfileUploadRequest("/api/v1/import/config", "cpu_model", "testdata/temp.xls", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 422,
		},
	}

	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				simClient: simClient,
			}
			rec := httptest.NewRecorder()
			i.CreateConfigHandler(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func newConfigfileUploadRequest(uri, paramName, f, scope, config_name, equipment_type string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("config_name", config_name)
	_ = writer.WriteField("equipment_type", equipment_type)
	_ = writer.WriteField("scope", scope)
	file, err := os.Open(f)
	if err != nil {
		logger.Log.Error("Failed opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()
	part, err := writer.CreateFormFile(paramName, filepath.Base(f))
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Log.Error("Failed copying file", zap.Error(err))
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing Writer", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, payload)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// Add Context for User Information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"Scope1", "France", "GEN"}})
	req = req.WithContext(ctx)
	return req, err
}

func Test_UpdateConfigHandler(t *testing.T) {
	var simClient v1Sim.SimulationServiceClient
	var request *http.Request
	var err error
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config *config.Config
	}
	params := httprouter.Params{}
	params1 := httprouter.Params{}
	params2 := httprouter.Params{}
	params = append(params, httprouter.Param{Key: "config_id", Value: "123"})
	params1 = append(params1, httprouter.Param{Key: "config_id", Value: ""})
	params2 = append(params2, httprouter.Param{Key: "config_id", Value: "abc"})

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
		params httprouter.Params
	}{
		{
			params: params,
			name:   "Valid scenario",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, nil).AnyTimes()
			},
			code: 200,
		},
		{
			params: params1,
			name:   "cong id err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, nil).AnyTimes()
			},
			code: 400,
		},
		{
			params: params2,
			name:   "cong id err str",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "abc", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, nil).AnyTimes()
			},
			code: 500,
		},
		{
			params: params,
			name:   "Valid scenario scope not in req",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "na", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, nil).AnyTimes()
			},
			code: 401,
		},
		{
			params: params,
			name:   "Valid scenario scope not in req blank",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, nil).AnyTimes()
			},
			code: 400,
		},
		{
			params: params,
			name:   "user role",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "TestUser",
					Role:   "User",
					Socpes: []string{"Scope1", "France", "GEN"},
				})
				request = request.WithContext(ctx)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 403,
		},
		{
			params: params,
			name:   "ctx bg",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				ctx := context.Background()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				request = request.WithContext(ctx)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
			},
			code: 500,
		},
		{
			params: params,
			name:   "Valid rpc err",
			fields: fields{&config.Config{MaxFileSize: int64(10), Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
			setup: func() {
				mockSClient := simmock.NewMockSimulationServiceClient(mockCtrl)
				simClient = mockSClient
				request, err = newConfigfileUpdateRequest("/api/v1/import/config", "1", "1,2,3", "cpu_model", "testdata/sim-config-cpu.csv", "GEN", "config1", "server")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockSClient.EXPECT().UpdateConfig(gomock.Any(), gomock.Any()).Return(&v1Sim.UpdateConfigResponse{}, errors.New("err")).AnyTimes()
			},
			code: 500,
		},
	}

	defer func() {
		err = os.RemoveAll("data")
		if err != nil {
			fmt.Println(err)
			t.Fatal(err)
		}
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:    tt.fields.config,
				simClient: simClient,
			}
			rec := httptest.NewRecorder()
			i.UpdateConfigHandler(rec, request, tt.params)
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func newConfigfileUpdateRequest(uri, configID, deletedMetadataIDs, paramName, filePath, scope, configName, equipmentType string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	// Add form fields
	_ = writer.WriteField("scope", scope)
	_ = writer.WriteField("config_name", configName)
	_ = writer.WriteField("equipment_type", equipmentType)
	_ = writer.WriteField("deletedMetadataIDs", deletedMetadataIDs)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Log.Error("Failed opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	// Create a form file part
	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		logger.Log.Error("Failed creating form file", zap.Error(err))
		return nil, err
	}

	// Copy the file contents to the form file part
	_, err = io.Copy(part, file)
	if err != nil {
		logger.Log.Error("Failed copying file", zap.Error(err))
		return nil, err
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing writer", zap.Error(err))
		return nil, err
	}

	// Create the request
	req, err := http.NewRequest("PUT", uri, payload)
	if err != nil {
		logger.Log.Error("Failed creating request", zap.Error(err))
		return nil, err
	}

	// Set the Content-Type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add context for user information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{
		UserID: "TestUser",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "France", "GEN"},
	})
	req = req.WithContext(ctx)

	req = mux.SetURLVars(req, map[string]string{"config_id": configID})

	return req, nil
}

func Test_ImportNominativeUser(t *testing.T) {
	var productClient v1Product.ProductServiceClient
	var request *http.Request
	mockCtrl := gomock.NewController(t)
	dbObj := dbmock.NewMockImport(mockCtrl)
	mockCluster, _ := kafka.NewMockCluster(1)
	defer mockCluster.Close()
	broker := mockCluster.BootstrapServers()
	p, _ := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	var err error
	//mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	type fields struct {
		config     *config.Config
		producer   *kafka.Producer
		importNock *dbmock.MockImport
	}

	tests := []struct {
		name   string
		fields fields
		setup  func()
		code   int
	}{
		{
			name: "Valid scenario",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominative/user", "Scope1", "ProductA", "1.0", "", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: http.StatusOK,
		},
		{
			name: "Valid scenario with aggrigation",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominative/user", "Scope1", "", "", "23", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockPClient.EXPECT().GetAggregationById(gomock.Any(), gomock.Any()).Return(&v1Product.GetAggregationByIdResponse{
					Id:              23,
					AggregationName: "testAggr",
					Scope:           "Scope1",
					ProductEditor:   "editor1",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: http.StatusOK,
		},
		{
			name: "Valid cols err 1",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "", "editor1", "testdata/NominativeUsersTemplate-1.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid cols order 2",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "", "editor1", "testdata/NominativeUsersTemplate-2.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid rpc fail",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, errors.New("err")).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid rpc wrong file",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "", "editor1", "testdata/applications.csv")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 500,
		},
		{
			name: "Valid user role",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "123", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := request.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{
					UserID: "TestUser",
					Role:   "User",
					Socpes: []string{"Scope1", "France", "GEN"},
				})
				request = request.WithContext(ctx)

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 403,
		},
		{
			name: "Valid user ct bg case",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "ProductA", "1.0", "123", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := context.Background()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				request = request.WithContext(ctx)

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 500,
		},
		{
			name: "Valid agg id not num",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "", "1.0", "abc", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid editor blank",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "prod1", "1.0", "", "", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid prod and agg blank",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1", "", "1.0", "", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
		{
			name: "Valid scope na",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "Scope1NA", "ProductA", "1.0", "123", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 401,
		},
		{
			name: "Valid scope blank",
			fields: fields{
				config:     &config.Config{},
				producer:   p,
				importNock: dbObj,
			},
			setup: func() {
				mockPClient := prodmock.NewMockProductServiceClient(mockCtrl)
				productClient = mockPClient
				request, err = newImportNominativeUserRequest("/api/v1/import/nominativeuser", "", "ProductA", "1.0", "123", "editor1", "testdata/NominativeUsersTemplate.xlsx")
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}

				mockPClient.EXPECT().GetProductInformationBySwidTag(gomock.Any(), gomock.Any()).Return(&v1Product.GetProductInformationBySwidTagResponse{
					Swidtag:        "swid",
					ProductName:    "ProductA",
					ProductEditor:  "editor1",
					ProductVersion: "1.0",
				}, nil).AnyTimes()
				dbObj.EXPECT().InsertNominativeUserRequestTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			code: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			tt.fields.config.NoOfPartitions = 3
			defer monkey.UnpatchAll()
			i := &ImportServiceServer{
				Config:        tt.fields.config,
				productClient: productClient,
				KafkaProducer: tt.fields.producer,
				ImportRepo:    tt.fields.importNock,
			}
			rec := httptest.NewRecorder()
			params := httprouter.Params{} // Empty Params instance
			i.ImportNominativeUser(rec, request, params)
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func newImportNominativeUserRequest(uri, scope, productName, productVersion, aggregationId, editor, filePath string) (*http.Request, error) {
	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Set the form values
	_ = writer.WriteField("scope", scope)
	_ = writer.WriteField("product_name", productName)
	_ = writer.WriteField("product_version", productVersion)
	_ = writer.WriteField("aggregation_id", aggregationId)
	_ = writer.WriteField("editor", editor)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		logger.Log.Error("Failed opening file", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	// Create a new form file field
	filePart, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		logger.Log.Error("Failed creating form file", zap.Error(err))
		return nil, err
	}

	// Copy the file content to the form file field
	_, err = io.Copy(filePart, file)
	if err != nil {
		logger.Log.Error("Failed copying file", zap.Error(err))
		return nil, err
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		logger.Log.Error("Failed closing multipart writer", zap.Error(err))
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		logger.Log.Error("Failed creating request", zap.Error(err))
		return nil, err
	}

	// Set the content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add context for user information
	ctx := req.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{
		UserID: "TestUser",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "France", "GEN"},
	})
	req = req.WithContext(ctx)
	return req, nil
}

func Test_SaveNominativeUserFile(t *testing.T) {
	type testInput struct {
		details *v1Product.ListNominativeUsersFileUploadResponse
		path    string
		scope   string
	}

	tests := []struct {
		name        string
		input       testInput
		expectedErr error
	}{
		{
			name: "Valid details",
			input: testInput{
				details: &v1Product.ListNominativeUsersFileUploadResponse{
					FileDetails: []*v1Product.ListNominativeUsersFileUpload{
						{
							FileName:  "test.xlsx",
							SheetName: "Sheet1",
							NominativeUsersDetails: []*v1Product.NominativeUserDetails{
								{
									FirstName:      "John",
									Email:          "john@example.com",
									UserName:       "john123",
									Profile:        "user",
									ActivationDate: "2023-07-09",
									Comments:       "Test",
								},
							},
						},
					},
				},
				path:  "/testdata",
				scope: "Scope1",
			},
			expectedErr: nil,
		},
		// Add more test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err, _, _ := saveNominativeUserFile(tt.input.details, tt.input.path, tt.input.scope)

			if err != tt.expectedErr {
				t.Errorf("Unexpected error. Got: %v, Want: %v", err, tt.expectedErr)
			}

			// Assert the file path and file name as per your requirements
			// You can use testing utilities like `os.Stat` to verify the existence of the file

			// Example:
			// _, statErr := os.Stat(filePath)
			// if statErr != nil {
			// 	t.Errorf("File does not exist. Path: %s", filePath)
			// }
			// if fileName != "test.xlsx" {
			// 	t.Errorf("Unexpected file name. Got: %s, Want: %s", fileName, "test.xlsx")
			// }
		})
	}
}

func TestDownloadFileNominativeUser(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockProductClient := prodmock.NewMockProductServiceClient(mockCtrl)

	i := &ImportServiceServer{
		productClient: mockProductClient,
		Config:        &config.Config{Upload: config.UploadConfig{UploadDir: "/uploads"}},
	}

	// Test case: Claims not found
	req1, _ := http.NewRequest("GET", "/download", nil)
	res1 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res1, req1, httprouter.Params{})
	assert.Equal(t, http.StatusInternalServerError, res1.Code)

	// Test case: Scope is missing
	req2, _ := http.NewRequest("GET", "/download?scope=", nil)
	ctx := req2.Context()
	ctx = rest_middleware.AddLogCtxKey(ctx)
	ctx = rest_middleware.AddClaims(ctx, &claims.Claims{
		UserID: "TestUser",
		Role:   "SuperAdmin",
		Socpes: []string{"Scope1", "France", "GEN"},
	})
	req2 = req2.WithContext(ctx)

	res2 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res2, req2, httprouter.Params{})
	assert.Equal(t, http.StatusBadRequest, res2.Code)

	// Test case: Scope validation failed
	req3, _ := http.NewRequest("GET", "/download?scope=invalid", nil)
	req3 = req3.WithContext(ctx)
	res3 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res3, req3, httprouter.Params{})
	assert.Equal(t, http.StatusUnauthorized, res3.Code)

	// Test case: Role validation failed
	req4, _ := http.NewRequest("GET", "/download?scope=GEN", nil)
	ctx2 := req2.Context()
	ctx2 = rest_middleware.AddLogCtxKey(ctx)
	ctx2 = rest_middleware.AddClaims(ctx, &claims.Claims{
		UserID: "TestUser",
		Role:   "User",
		Socpes: []string{"Scope1", "France", "GEN"},
	})
	req4 = req4.WithContext(ctx2)
	res4 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res4, req4, httprouter.Params{})
	assert.Equal(t, http.StatusForbidden, res4.Code)

	// Test case: Failed to get file id
	req5, _ := http.NewRequest("GET", "/download?scope=GEN&id=abc", nil)
	req5 = req5.WithContext(ctx)
	res5 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res5, req5, httprouter.Params{})
	assert.Equal(t, http.StatusInternalServerError, res5.Code)
	// assert.Equal(t, "Failed to get file id", res5.Body.String())

	// Test case: Invalid file type
	req6, _ := http.NewRequest("GET", "/download?scope=GEN&id=1&type=invalid", nil)
	req6 = req6.WithContext(ctx)
	res6 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res6, req6, httprouter.Params{})
	assert.Equal(t, http.StatusInternalServerError, res6.Code)
	// assert.Equal(t, "invalid file type", res6.Body.String())

	// Test case: Download actual file
	mockProductClient.EXPECT().ListNominativeUserFileUpload(gomock.Any(), gomock.Any()).Return(&v1Product.ListNominativeUsersFileUploadResponse{
		FileDetails: []*v1Product.ListNominativeUsersFileUpload{
			{
				FileName:  "test.xlsx",
				SheetName: "Sheet1",
				UploadId:  "123",
			},
		},
	}, nil).AnyTimes()
	req8, _ := http.NewRequest("GET", "/download?scope=GEN&id=1&type=actual", nil)
	req8 = req8.WithContext(ctx)
	res8 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res8, req8, httprouter.Params{})
	assert.Equal(t, http.StatusInternalServerError, res8.Code)
	// assert.Equal(t, "application/octet-stream", res8.Header().Get("Content-Type"))

	// Test case: Download error file
	// mockProductClient.EXPECT().ListNominativeUserFileUpload(gomock.Any(), gomock.Any()).Return(&v1Product.ListNominativeUsersFileUploadResponse{
	// 	FileDetails: []*v1Product.ListNominativeUsersFileUpload{
	// 		{
	// 			FileName:  "test.xlsx",
	// 			SheetName: "Sheet1",
	// 			UploadId:  "123",
	// 		},
	// 	},
	// }, errors.New("err"))
	req9, _ := http.NewRequest("GET", "/download?scope=GEN&id=1&type=error", nil)
	req9 = req9.WithContext(ctx)
	res9 := httptest.NewRecorder()
	i.DownloadFileNominativeUser(res9, req9, httprouter.Params{})
	assert.Equal(t, http.StatusOK, res9.Code)
}

func TestNewImportServiceServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGRPCServers := map[string]*grpc.ClientConn{
		"dps":        nil,
		"simulation": nil,
		"account":    nil,
		"product":    nil,
		"catalog":    nil,
	}

	config := &config.Config{} // Provide a sample configuration

	server := NewImportServiceServer(mockGRPCServers, config, nil, &kafka.Producer{}, &kafka.Consumer{})

	// Assert that the server is not nil
	assert.NotNil(t, server)

	// Assert that the server is of type *importServiceServer
	//_, ok := server.(*ImportServiceServer)
	//assert.True(t, ok)
}

func TestRemoveFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "removeFilesTest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create some test files
	files := []string{
		filepath.Join(tempDir, "myscope_1.txt"),
		filepath.Join(tempDir, "myscope_2.txt"),
		filepath.Join(tempDir, "metadata_1.txt"),
		filepath.Join(tempDir, "otherfile.txt"),
	}
	for _, file := range files {
		if _, err := os.Create(file); err != nil {
			t.Fatal(err)
		}
	}

	// Test case 1: Removing data files
	removeFiles("myscope", tempDir, "data")
	remainingFiles, _ := filepath.Glob(filepath.Join(tempDir, "*"))
	if len(remainingFiles) != 4 {
		t.Errorf("Expected 2 remaining files, got %d", len(remainingFiles))
	}

	// Test case 2: Removing global data files
	removeFiles("myscope", tempDir, "globaldata")
	remainingFiles, _ = filepath.Glob(filepath.Join(tempDir, "*"))
	if len(remainingFiles) != 4 {
		t.Errorf("Expected 0 remaining files, got %d", len(remainingFiles))
	}

	// Test case 3: Removing metadata files
	removeFiles("myscope", tempDir, "metadata")
	remainingFiles, _ = filepath.Glob(filepath.Join(tempDir, "*"))
	if len(remainingFiles) != 4 {
		t.Errorf("Expected 1 remaining file, got %d", len(remainingFiles))
	}

	// Test case 4: Empty directory
	removeFiles("myscope", tempDir, "data")
	remainingFiles, _ = filepath.Glob(filepath.Join(tempDir, "*"))
	if len(remainingFiles) != 4 {
		t.Errorf("Expected 1 remaining file, got %d", len(remainingFiles))
	}

	// Test case 5: Non-matching files
	removeFiles("other", tempDir, "data")
	remainingFiles, _ = filepath.Glob(filepath.Join(tempDir, "*"))
	if len(remainingFiles) != 4 {
		t.Errorf("Expected 1 remaining file, got %d", len(remainingFiles))
	}

	// Test case 6: Error listing files
	invalidDir := filepath.Join(tempDir, "nonexistent")
	removeFiles("myscope", invalidDir, "data")

	// Test case 7: Error removing files
	unremovableFile := filepath.Join(tempDir, "unremovable.txt")
	if _, err := os.Create(unremovableFile); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unremovableFile, 0444); err != nil {
		t.Fatal(err)
	}
	removeFiles("myscope", tempDir, "data")
}

func Test_ListNominativeUserFileUploads(t *testing.T) {
	var request *http.Request
	//var err error
	mockCtrl := gomock.NewController(t)
	importMock := dbmock.NewMockImport(mockCtrl)
	defer mockCtrl.Finish()
	type fields struct {
		config     *config.Config
		importMock *dbmock.MockImport
	}
	tests := []struct {
		name   string
		i      *ImportServiceServer
		setup  func()
		fields fields
		code   int
	}{
		{name: "missing claims",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=API&page_num=1&page_size=50&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				request = req
			},
			code: 500,
		},
		{name: "scope not found",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=&page_num=1&page_size=50&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "ScopeValidationFailed",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=failed&page_num=1&page_size=50&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 401,
		},
		{name: "RoleValidationFailed",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=50&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "User", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 403,
		},
		{name: "Falied to get page number",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=abc&page_size=50&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "Falied to get page size",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=abc&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "Failed to get file id",
			fields: fields{},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=1&sort_by=name&sort_order=asc&id=abc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req
			},
			code: 400,
		},
		{name: "err",
			fields: fields{
				config:     &config.Config{},
				importMock: importMock,
			},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=1&id=1&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req

				importMock.EXPECT().ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
					Scope:        []string{"scope1"},
					PageNum:      1 * (1 - 1),
					FileUploadID: true,
					ID:           1,
					PageSize:     1,
					NameAsc:      true,
				}).Return([]db.ListNominativeUsersUploadedFilesRow{
					{
						Totalrecords:  2,
						RequestID:     1,
						UploadID:      "u1",
						Scope:         "scope1",
						Swidtag:       sql.NullString{String: "swid1", Valid: true},
						AggregationID: sql.NullString{String: "1", Valid: true},
						Editor:        sql.NullString{String: "editor1", Valid: true},
					},
				}, errors.New("err"))
			},
			code: 500,
		},
		{name: "err without agg id",
			fields: fields{
				config:     &config.Config{},
				importMock: importMock,
			},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=1&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req

				importMock.EXPECT().ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
					Scope:    []string{"scope1"},
					PageNum:  1 * (1 - 1),
					PageSize: 1,
					NameAsc:  true,
				}).Return([]db.ListNominativeUsersUploadedFilesRow{
					{
						Totalrecords:  2,
						UploadID:      "u1",
						Scope:         "scope1",
						Swidtag:       sql.NullString{String: "swid1", Valid: true},
						AggregationID: sql.NullString{String: "1", Valid: true},
						Editor:        sql.NullString{String: "editor1", Valid: true},
					},
				}, errors.New("err"))
			},
			code: 500,
		},
		{name: "unmarshal err",
			fields: fields{
				config:     &config.Config{},
				importMock: importMock,
			},
			setup: func() {
				payload := &bytes.Buffer{}
				req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=1&id=1&sort_by=name&sort_order=asc", payload)
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				ctx := req.Context()
				ctx = rest_middleware.AddLogCtxKey(ctx)
				ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
				req = req.WithContext(ctx)
				request = req

				importMock.EXPECT().ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
					Scope:        []string{"scope1"},
					PageNum:      1 * (1 - 1),
					FileUploadID: true,
					ID:           1,
					PageSize:     1,
					NameAsc:      true,
				}).Return([]db.ListNominativeUsersUploadedFilesRow{
					{
						Totalrecords:   2,
						UploadID:       "u1",
						Scope:          "scope1",
						Swidtag:        sql.NullString{String: "swid1", Valid: true},
						AggregationID:  sql.NullString{String: "1", Valid: true},
						Editor:         sql.NullString{String: "editor1", Valid: true},
						RecordFailed:   []byte(`{"first_name": "Vion4","user_name": "Anne-Lucie4","email": "test3@orange.com","profile": "PRO",	"activation_date": "43515","comments": "duplicate entry"}`),
						RecordSucceed:  23,
						RecordFailed_2: 23,
						Status:         "partial",
						Pname:          "product",
						Nametype:       "aggr",
					},
				}, nil)
			},
			code: 500,
		},
		// {name: "success",
		// 	fields: fields{
		// 		config:     &config.Config{},
		// 		importMock: importMock,
		// 	},
		// 	setup: func() {
		// 		payload := &bytes.Buffer{}
		// 		req, err := http.NewRequest("GET", "api/v1/import/nominative/users/fileupload?scope=scope1&page_num=1&page_size=1&id=1&sort_by=name&sort_order=asc", payload)
		// 		if err != nil {
		// 			logger.Log.Error("Failed creating request", zap.Error(err))
		// 			t.Fatal(err)
		// 		}
		// 		ctx := req.Context()
		// 		ctx = rest_middleware.AddLogCtxKey(ctx)
		// 		ctx = rest_middleware.AddClaims(ctx, &claims.Claims{UserID: "TestUser", Role: "SuperAdmin", Socpes: []string{"scope1", "france"}})
		// 		req = req.WithContext(ctx)
		// 		request = req

		// 		importMock.EXPECT().ListNominativeUsersUploadedFiles(context.Background(), db.ListNominativeUsersUploadedFilesParams{
		// 			Scope:        []string{"scope1"},
		// 			PageNum:      1 * (1 - 1),
		// 			FileUploadID: true,
		// 			ID:           1,
		// 			PageSize:     1,
		// 			NameAsc:      true,
		// 		}).Return([]db.ListNominativeUsersUploadedFilesRow{
		// 			{
		// 				Totalrecords:   2,
		// 				UploadID:       "u1",
		// 				Scope:          "scope1",
		// 				Swidtag:        sql.NullString{String: "swid1", Valid: true},
		// 				AggregationID:  sql.NullString{String: "1", Valid: true},
		// 				Editor:         sql.NullString{String: "editor1", Valid: true},
		// 				RecordFailed:   []byte(`{}`),
		// 				RecordSucceed:  23,
		// 				RecordFailed_2: 23,
		// 				Status:         "partial",
		// 				Pname:          "product",
		// 				Nametype:       "aggr",
		// 			},
		// 		}, nil)
		// 	},
		// 	code: 200,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			i := &ImportServiceServer{
				Config:     tt.fields.config,
				ImportRepo: importMock,
				//dpsClient: dpsClient,
			}
			rec := httptest.NewRecorder()
			i.ListNominativeUserFileUploads(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}
