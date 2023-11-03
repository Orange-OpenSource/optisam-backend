package v1

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/api/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/config"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1"
	dbmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/dbmock"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/postgres/db"
	queuemock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/pkg/repository/v1/queuemock"
	equipV1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/equipment-service/pkg/api/v1"
	eqmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/equipment-service/pkg/api/v1/mock"
	metv1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/metric-service/pkg/api/v1"
	metmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/metric-service/pkg/api/v1/mock"
	prodV1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/product-service/pkg/api/v1"
	prodmock "gitlab.tech.orange/optisam/optisam-it/optisam-services/dps-service/thirdparty/product-service/pkg/api/v1/mock"

	"github.com/golang/mock/gomock"
)

func Test_DataAnalysis(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	var rep repo.Dps
	qObj := queuemock.NewMockWorkerqueue(mockCtrl)
	var met metv1.MetricServiceClient
	var equip equipV1.EquipmentServiceClient
	var prod prodV1.ProductServiceClient

	tests := []struct {
		name    string
		ctx     context.Context
		input   *v1.DataAnalysisRequest
		setup   func(*v1.DataAnalysisRequest)
		output  *v1.DataAnalysisResponse
		wantErr bool
		s       *dpsServiceServer
	}{
		{
			name: "claims Not found",
			ctx:  context.Background(),
			input: &v1.DataAnalysisRequest{
				Scope: "Scope1",
				File:  "Scope1_applications.xlsx",
			},
			setup:   func(*v1.DataAnalysisRequest) {},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: true,
		},
		{
			name: "Scope Not found",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "Scope10",
				File:  "Scope0_applications.xlsx",
			},
			setup:   func(*v1.DataAnalysisRequest) {},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: true,
		},
		{
			name: "Invalid File Extension",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file1.cv",
			},
			setup: func(*v1.DataAnalysisRequest) {
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "sheet missing",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file1.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "correct sheet",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockEquip := eqmock.NewMockEquipmentServiceClient(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				mockProd := prodmock.NewMockProductServiceClient(mockCtrl)
				equip = mockEquip
				met = mockMetric
				prod = mockProd
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
				mockEquip.EXPECT().GetMetrics(gomock.Any(), gomock.Any()).Return(&equipV1.GetMetricsResponse{Name: []string{"m1"}, Type: []string{"oracle.nup.standard"}}, nil)
				mockRepository.EXPECT().GetCoreFactorList(ctx).Return([]db.CoreFactorReference{{Model: "str", Manufacturer: "string", CoreFactor: "1", ID: int32(1)}}, nil)
				mockProd.EXPECT().GetAllEditorsCatalog(ctx, gomock.Any()).Return(&prodV1.GetAllEditorsCatalogResponse{EditorName: []string{"oracle", "Oracle"}}, nil)
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "correct sheet 2",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file2.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockEquip := eqmock.NewMockEquipmentServiceClient(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				mockProd := prodmock.NewMockProductServiceClient(mockCtrl)
				equip = mockEquip
				met = mockMetric
				prod = mockProd
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
				mockEquip.EXPECT().GetMetrics(gomock.Any(), gomock.Any()).Return(&equipV1.GetMetricsResponse{}, nil)
				mockRepository.EXPECT().GetCoreFactorList(ctx).Return([]db.CoreFactorReference{{Model: "str", Manufacturer: "string", CoreFactor: "1", ID: int32(1)}}, nil).AnyTimes()
				mockProd.EXPECT().GetAllEditorsCatalog(ctx, gomock.Any()).Return(&prodV1.GetAllEditorsCatalogResponse{EditorName: []string{"oracle", "Oracle"}}, nil)
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "correct sheet 3",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file3.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockEquip := eqmock.NewMockEquipmentServiceClient(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				mockProd := prodmock.NewMockProductServiceClient(mockCtrl)
				equip = mockEquip
				met = mockMetric
				prod = mockProd
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
				mockEquip.EXPECT().GetMetrics(gomock.Any(), gomock.Any()).Return(&equipV1.GetMetricsResponse{Name: []string{"m1"}, Type: []string{"oracle.nup.standard"}}, nil)
				mockRepository.EXPECT().GetCoreFactorList(ctx).Return([]db.CoreFactorReference{{Model: "str", Manufacturer: "string", CoreFactor: "1", ID: int32(1)}}, nil).AnyTimes()
				mockProd.EXPECT().GetAllEditorsCatalog(ctx, gomock.Any()).Return(&prodV1.GetAllEditorsCatalogResponse{EditorName: []string{"oracle", "Oracle"}}, nil)
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "correct sheet 4",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file4.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockEquip := eqmock.NewMockEquipmentServiceClient(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				mockProd := prodmock.NewMockProductServiceClient(mockCtrl)
				equip = mockEquip
				met = mockMetric
				prod = mockProd
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
				mockEquip.EXPECT().GetMetrics(gomock.Any(), gomock.Any()).Return(&equipV1.GetMetricsResponse{Name: []string{"m1"}, Type: []string{"type"}}, nil)
				mockRepository.EXPECT().GetCoreFactorList(ctx).Return([]db.CoreFactorReference{{Model: "str", Manufacturer: "string", CoreFactor: "1", ID: int32(1)}}, nil).AnyTimes()
				mockProd.EXPECT().GetAllEditorsCatalog(ctx, gomock.Any()).Return(&prodV1.GetAllEditorsCatalogResponse{EditorName: []string{"oracle", "Oracle"}}, nil)
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
		{
			name: "correct sheet 5 header missing",
			ctx:  ctx,
			input: &v1.DataAnalysisRequest{
				Scope: "AAK",
				File:  "file4.xlsx",
			},
			setup: func(*v1.DataAnalysisRequest) {
				mockCtrl = gomock.NewController(t)
				mockRepository := dbmock.NewMockDps(mockCtrl)
				rep = mockRepository
				mockEquip := eqmock.NewMockEquipmentServiceClient(mockCtrl)
				mockMetric := metmock.NewMockMetricServiceClient(mockCtrl)
				mockProd := prodmock.NewMockProductServiceClient(mockCtrl)
				equip = mockEquip
				met = mockMetric
				prod = mockProd
				config.SetConfig(config.Config{RawdataLocation: "testdata"})
				mockEquip.EXPECT().GetMetrics(gomock.Any(), gomock.Any()).Return(&equipV1.GetMetricsResponse{Name: []string{"m1"}, Type: []string{"type"}}, nil)
				mockRepository.EXPECT().GetCoreFactorList(ctx).Return([]db.CoreFactorReference{{Model: "str", Manufacturer: "string", CoreFactor: "1", ID: int32(1)}}, nil).AnyTimes()
				mockProd.EXPECT().GetAllEditorsCatalog(ctx, gomock.Any()).Return(&prodV1.GetAllEditorsCatalogResponse{EditorName: []string{"oracle", "Oracle"}}, nil)
			},
			output:  &v1.DataAnalysisResponse{Status: FAILED},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.input)
			tt.s = &dpsServiceServer{
				dpsRepo:   rep,
				queue:     qObj,
				equipment: equip,
				metric:    met,
				product:   prod,
			}
			_, err := tt.s.DataAnalysis(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("dpsServiceServer.DataAnalysis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("Test passed ", tt.name)
		})
	}
}

func TestGetErrorResponse(t *testing.T) {
	// Test Case 1: Failed Case
	err := errors.New("analysis: failed to analyze data")
	status, msg := getErrorResponse(err)
	expectedStatus := "FAILED"
	expectedMsg := "analysis: failed to analyze data"
	if status != expectedStatus || msg != expectedMsg {
		t.Errorf("Test case 1 failed. Expected status: %s, Expected msg: %s. Got status: %s, Got msg: %s", expectedStatus, expectedMsg, status, msg)
	}

	// Test Case 2: Partial Case
	err = errors.New("unexpected error occurred")
	status, msg = getErrorResponse(err)
	expectedStatus = "PARTIAL"
	expectedMsg = "InternalError"
	if status != expectedStatus || msg != expectedMsg {
		t.Errorf("Test case 2 failed. Expected status: %s, Expected msg: %s. Got status: %s, Got msg: %s", expectedStatus, expectedMsg, status, msg)
	}

	// Test Case 3: Custom Failed Case
	err = errors.New("something went wrong: analysis: data not found")
	status, msg = getErrorResponse(err)
	expectedStatus = "FAILED"
	expectedMsg = "something went wrong: analysis: data not found"
	if status != expectedStatus || msg != expectedMsg {
		t.Errorf("Test case 3 failed. Expected status: %s, Expected msg: %s. Got status: %s, Got msg: %s", expectedStatus, expectedMsg, status, msg)
	}

	// Test Case 4: Empty Error
	err = errors.New("")
	status, msg = getErrorResponse(err)
	expectedStatus = "PARTIAL"
	expectedMsg = "InternalError"
	if status != expectedStatus || msg != expectedMsg {
		t.Errorf("Test case 4 failed. Expected status: %s, Expected msg: %s. Got status: %s, Got msg: %s", expectedStatus, expectedMsg, status, msg)
	}
}
