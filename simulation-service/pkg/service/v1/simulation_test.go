package v1

import (
	"context"
	"errors"
	"fmt"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	ls "optisam-backend/license-service/pkg/api/v1"
	mockls "optisam-backend/license-service/pkg/api/v1/mock"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSimulationService_SimulationByMetric(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	var mockCtrl *gomock.Controller
	var licenseClient ls.LicenseServiceClient
	type args struct {
		ctx context.Context
		req *v1.SimulationByMetricRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.SimulationByMetricResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.SimulationByMetricRequest{
					Editor: "Oracle",
					MetricDetails: []*v1.MetricSimDetails{
						{
							Sku:        "sku1",
							Swidtag:    "swid3",
							MetricName: "ibm_pvu",
							UnitCost:   200,
						},
						{
							Sku:        "sku2",
							Swidtag:    "swid3",
							MetricName: "ibm_pvu",
							UnitCost:   300,
						},
						{
							Sku:             "sku3",
							Swidtag:         "swid2,swid1",
							AggregationName: "aggname1",
							MetricName:      "oracle_processor",
							UnitCost:        300,
						},
						{
							Sku:             "sku5",
							Swidtag:         "swid1",
							AggregationName: "",
							MetricName:      "oracle_processor",
							UnitCost:        50,
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				mockLicenseClient.EXPECT().GetOverAllCompliance(ctx, &ls.GetOverAllComplianceRequest{
					Scope:      "Scope1",
					Editor:     "Oracle",
					Simulation: true,
				}).Times(2).Return(&ls.GetOverAllComplianceResponse{
					AcqRights: []*ls.AggregationAcquiredRights{
						{
							SKU:             "sku3",
							AggregationName: "aggname1",
							SwidTags:        "swid1,swid2",
							Metric:          "oracle_processor",
							NumCptLicences:  10,
							AvgUnitPrice:    10,
						},
						{
							SKU:             "sku5",
							AggregationName: "",
							SwidTags:        "swid1",
							Metric:          "oracle_processor",
							NumCptLicences:  10,
							AvgUnitPrice:    10,
						},
						{
							SKU:             "sku1",
							AggregationName: "",
							SwidTags:        "swid3",
							Metric:          "ibm_pvu",
							NumCptLicences:  10,
							AvgUnitPrice:    20,
						},
						{
							SKU:             "sku2",
							AggregationName: "",
							SwidTags:        "swid3",
							Metric:          "ibm_pvu",
							NumCptLicences:  10,
							AvgUnitPrice:    10,
						},
						{
							SKU:             "sku4",
							AggregationName: "",
							SwidTags:        "swid4",
							Metric:          "ibm",
							NumCptLicences:  10,
							AvgUnitPrice:    10,
						},
					},
				}, nil)
			},
			want: &v1.SimulationByMetricResponse{
				Success: true,
				MetricSimResult: []*v1.MetricSimulationResult{
					{
						Sku:             "sku3",
						Swidtag:         "swid1,swid2",
						AggregationName: "aggname1",
						MetricName:      "oracle_processor",
						NumCptLicences:  10,
						OldTotalCost:    100,
						NewTotalCost:    3000,
					},
					{
						Sku:             "sku5",
						Swidtag:         "swid1",
						AggregationName: "",
						MetricName:      "oracle_processor",
						NumCptLicences:  10,
						OldTotalCost:    100,
						NewTotalCost:    500,
					},
					{
						Sku:             "sku1,sku2",
						Swidtag:         "swid3",
						AggregationName: "",
						MetricName:      "ibm_pvu",
						NumCptLicences:  20,
						OldTotalCost:    300,
						NewTotalCost:    5000,
					},
				},
			},
		},
		{
			name: "Success - With no metrics",
			args: args{
				ctx: ctx,
				req: &v1.SimulationByMetricRequest{
					Editor:        "Oracle",
					MetricDetails: []*v1.MetricSimDetails{},
					Scope:         "Scope1",
				},
			},
			setup: func() {

			},
			want: &v1.SimulationByMetricResponse{
				Success: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationServiceForTest(nil, licenseClient)
			got, err := hcs.SimulationByMetric(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.SimulationByMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricSimulationResultAll(t, "SimulationService.SimulationByMetric", tt.want.MetricSimResult, got.MetricSimResult)
				//t.Errorf("SimulationService.SimulationByMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimulationService_SimulationByHardware(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"Scope1", "Scope2", "Scope3"},
	})
	attributes := []*v1.EquipAttribute{
		{
			ID:         "0x5092e",
			Name:       "server_corenumber",
			DataType:   v1.DataTypes_INT,
			Displayed:  true,
			Searchable: true,
			MappedTo:   "server_coresNumber",
			Val:        &v1.EquipAttribute_IntVal{16},
			OldVal:     &v1.EquipAttribute_IntValOld{16},
			Simulated:  false,
		},
		{
			ID:         "0x5092f",
			Name:       "pvu",
			DataType:   v1.DataTypes_FLOAT,
			Displayed:  true,
			Searchable: true,
			MappedTo:   "pvu",
			Val:        &v1.EquipAttribute_FloatVal{70},
			OldVal:     &v1.EquipAttribute_FloatValOld{100},
			Simulated:  true,
		},
		{
			ID:         "0x50935",
			Name:       "corefactor_oracle",
			DataType:   v1.DataTypes_FLOAT,
			Displayed:  true,
			Searchable: true,
			MappedTo:   "corefactor_oracle",
			Val:        &v1.EquipAttribute_FloatVal{0.625},
			OldVal:     &v1.EquipAttribute_FloatValOld{1},
			Simulated:  true,
		},
		{
			ID:         "0x50934",
			Name:       "serverprocessornumber",
			DataType:   v1.DataTypes_INT,
			Displayed:  true,
			Searchable: true,
			MappedTo:   "server_processorsNumber",
			Val:        &v1.EquipAttribute_IntVal{2},
			OldVal:     &v1.EquipAttribute_IntValOld{2},
			Simulated:  false,
		},
		{
			ID:         "0x5093b",
			Name:       "sag",
			DataType:   v1.DataTypes_FLOAT,
			Displayed:  true,
			Searchable: true,
			MappedTo:   "sag",
			Val:        &v1.EquipAttribute_FloatVal{0.625},
			OldVal:     &v1.EquipAttribute_FloatValOld{0.625},
			Simulated:  false,
		},
	}
	var mockCtrl *gomock.Controller
	var licenseClient ls.LicenseServiceClient
	type args struct {
		ctx context.Context
		req *v1.SimulationByHardwareRequest
	}
	tests := []struct {
		name    string
		hcs     *SimulationService
		args    args
		setup   func()
		want    *v1.SimulationByHardwareResponse
		wantErr bool
	}{
		{
			name: "SUCCESS",
			args: args{
				ctx: ctx,
				req: &v1.SimulationByHardwareRequest{
					EquipType:  "server",
					EquipId:    "30373237-3132-5a43-3336-32364341424d",
					Attributes: attributes,
					MetricDetails: []*v1.SimMetricDetails{
						{
							MetricType: "oracle.processor.standard",
							MetricName: "oracle_processor",
						},
						{
							MetricType: "ibm.pvu.standard",
							MetricName: "ibm_pvu",
						},
					},
					Scope: "Scope1",
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockLicenseClient := mockls.NewMockLicenseServiceClient(mockCtrl)
				licenseClient = mockLicenseClient
				gomock.InOrder(
					mockLicenseClient.EXPECT().LicensesForEquipAndMetric(ctx, gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, req *ls.LicensesForEquipAndMetricRequest) (*ls.LicensesForEquipAndMetricResponse, error) {
						if req.MetricName == "ibm_pvu" {
							return nil, errors.New("Internal")
						}

						return &ls.LicensesForEquipAndMetricResponse{
							Licenses: []*ls.ProductLicenseForEquipAndMetric{
								{
									MetricName:  "oracle_processor",
									OldLicences: 120000,
									NewLicenses: 130000,
									Delta:       10000,
									Product: &ls.Product{
										SwidTag:  "Oracle_Real_Application_Testing_12.1.0.1.0",
										Name:     "Oracle Real Application Testing",
										Version:  "12.1.0.1.0",
										Category: "Other",
										Editor:   "Oracle",
									},
								},
							},
						}, nil
					}),
				)
			},
			want: &v1.SimulationByHardwareResponse{
				SimulationResult: []*v1.SimulatedProductsLicenses{
					{
						Success:    true,
						MetricName: "oracle_processor",
						Licenses: []*v1.SimulatedProductLicense{
							{
								OldLicences: 120000,
								NewLicenses: 130000,
								Delta:       10000,
								SwidTag:     "Oracle_Real_Application_Testing_12.1.0.1.0",
								ProductName: "Oracle Real Application Testing",
								Editor:      "Oracle",
							},
						},
					},
					{
						MetricName:       "ibm_pvu",
						SimFailureReason: "Internal",
					},
				},
			},
		},
		{

			name: "Success - With no metrics",
			args: args{
				ctx: ctx,
				req: &v1.SimulationByHardwareRequest{
					EquipType:     "server",
					EquipId:       "30373237-3132-5a43-3336-32364341424d",
					MetricDetails: []*v1.SimMetricDetails{},
					Attributes:    attributes,
					Scope:         "Scope1",
				},
			},
			setup: func() {

			},
			want: &v1.SimulationByHardwareResponse{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			hcs := NewSimulationServiceForTest(nil, licenseClient)
			got, err := hcs.SimulationByHardware(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SimulationService.SimulationByHardware() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.SimulationByHardware() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func compareMetricSimulationResultAll(t *testing.T, name string, exp []*v1.MetricSimulationResult, act []*v1.MetricSimulationResult) {
	for i := range exp {
		compareMetricSimulationResult(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetricSimulationResult(t *testing.T, name string, exp *v1.MetricSimulationResult, act *v1.MetricSimulationResult) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "resulr is expected to be nil")
	}
	assert.Equalf(t, exp.Swidtag, act.Swidtag, "%s.Swidtag should be same", name)
	assert.Equalf(t, exp.AggregationName, act.AggregationName, "%s.AggregationName should be same", name)
	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.MetricName should be same", name)
	assert.Equalf(t, exp.NumCptLicences, act.NumCptLicences, "%s.NumCptLicences should be same", name)
	assert.Equalf(t, exp.OldTotalCost, act.OldTotalCost, "%s.OldTotalCost should be same", name)
	assert.Equalf(t, exp.NewTotalCost, act.NewTotalCost, "%s.NewTotalCost should be same", name)
	assert.Equalf(t, exp.Sku, act.Sku, "%s.Sku should be same", name)
}
