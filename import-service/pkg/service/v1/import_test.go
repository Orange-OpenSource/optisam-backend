package v1

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	v1Acc "optisam-backend/account-service/pkg/api/v1"
	mockAcc "optisam-backend/account-service/pkg/api/v1/mock"
	"optisam-backend/common/optisam/logger"
	rest_middleware "optisam-backend/common/optisam/middleware/rest"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/dps-service/pkg/api/v1/mock"
	"optisam-backend/import-service/pkg/config"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

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
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
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
			name: "FAILURE - Unable to get scope type info",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
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
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
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
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
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
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
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
			name: "FAILURE - Data Multiple Files with some having incorrect correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `^products_equipments\.csv$`, `product_application\.csv`}}}},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()
			i := &importServiceServer{
				config:    tt.fields.config,
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

func Test_importServiceServer_UploadMetaDataHandler(t *testing.T) {

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
			code: 200,
		},
		{
			name: "FAILURE - Metadata Multiple Files with some having incorrect correct naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
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
			i := &importServiceServer{
				config:    tt.fields.config,
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
			name: "FAILURE - IncorrectFileExtenison",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{RawDataUploadDir: "data"}}},
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

			i := &importServiceServer{
				config:    tt.fields.config,
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

			i := &importServiceServer{
				config:    tt.fields.config,
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

func Test_importServiceServer_DownloadGlobalDataErrors(t *testing.T) {
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
		i      *importServiceServer
		setup  func()
		fields fields
		code   int
	}{
		/*	{name: "FAILURE - File name is missing",
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
			code: 400,
		},*/
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
		/*{name: "Success - DownloadError",
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
				mockDPSClient.EXPECT().GetAnalysisFileInfo(ctx, &v1.GetAnalysisFileInfoRequest{Scope: "scope1", UploadId: int32(123)}).Return(&v1.GetAnalysisFileInfoResponse{FileName: "111_temp.xlsx"}, nil).Times(1)

			},
			code: 200,
		},*/
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
			i := &importServiceServer{
				config:    tt.fields.config,
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
