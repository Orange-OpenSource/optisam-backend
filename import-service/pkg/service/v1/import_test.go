// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/dps-service/pkg/api/v1"
	"optisam-backend/import-service/pkg/config"
	"optisam-backend/import-service/pkg/service/v1/mock"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

func Test_importServiceServer_UploadDataHandler(t *testing.T) {

	var dpsClient v1.DpsServiceClient
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
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
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
			name: "FAILURE - Data Single file with incorrect naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", DataFileAllowedRegex: []string{`^products\.csv$`, `products_equipments\.csv`, `product_application\.csv`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products_1.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
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
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/products_equipments.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
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
				request, err = newfileUploadRequest("/api/v1/import/data", "France", "files", []string{"testdata/products.csv", "testdata/applications.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
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
				request, err = newfileUploadRequest("/api/v1/import/metadata", "", "files", []string{"testdata/metadata_laptop.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "", Files: []string{"metadata_laptop.csv"}, Type: "metadata", UploadedBy: "TestUser",
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
			name: "FAILURE - Metadata Single file with incorrect naming",
			// args:   args{res: httptest.NewRecorder(), req: request, param: httprouter.Params{}},
			fields: fields{&config.Config{Upload: config.UploadConfig{UploadDir: "data", MetaDatafileAllowedRegex: []string{`^metadata_[a-zA-Z]*\.csv$`}}}},
			setup: func() {
				request, err = newfileUploadRequest("/api/v1/import/metadata", "", "files", []string{"testdata/metadata_laptop1.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "", Files: []string{"metadata_laptop1.csv"}, Type: "metadata", UploadedBy: "TestUser",
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
				request, err = newfileUploadRequest("/api/v1/import/metadata", "", "files", []string{"testdata/metadata_laptop.csv", "testdata/metadata_desktop.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "", Files: []string{"metadata_laptop.csv", "metadata_desktop.csv"}, Type: "metadata", UploadedBy: "TestUser",
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
				request, err = newfileUploadRequest("/api/v1/import/metadata", "", "files", []string{"testdata/metadata_laptop.csv", "testdata/metadata_desktop1.csv"})
				if err != nil {
					logger.Log.Error("Failed creating request", zap.Error(err))
					t.Fatal(err)
				}
				mockDPSClient := mock.NewMockDpsServiceClient(mockCtrl)
				dpsClient = mockDPSClient
				// Gotcha Need to check
				mockDPSClient.EXPECT().NotifyUpload(request.Context(), &v1.NotifyUploadRequest{
					Scope: "", Files: []string{"metadata_laptop.csv", "metadata_desktop1.csv"}, Type: "metadata", UploadedBy: "TestUser",
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
			}
			rec := httptest.NewRecorder()
			i.UploadMetaDataHandler(rec, request, httprouter.Params{})
			if rec.Code != tt.code {
				t.Errorf("Failed = got %v, want %v", rec.Code, tt.code)
			}
		})
	}
}

func newfileUploadRequest(uri string, scope string, paramName string, files []string) (*http.Request, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("scope", scope)
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
	//Add Context for User Information
	ctx := req.Context()
	ctx = ctxmanage.AddClaims(ctx, &claims.Claims{UserID: "TestUser"})
	req = req.WithContext(ctx)
	return req, err
}
