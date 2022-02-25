package v1

import (
	"context"
	"errors"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	"optisam-backend/common/optisam/token/claims"
	ls "optisam-backend/license-service/pkg/api/v1"
	mockls "optisam-backend/license-service/pkg/api/v1/mock"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
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
					SwidTag: "Oracle_Database_11g_Enterprise_Edition_10.3",
					MetricDetails: []*v1.MetricSimDetails{
						&v1.MetricSimDetails{
							MetricName: "ibm_pvu",
							UnitCost:   200,
						},
						&v1.MetricSimDetails{
							MetricName: "oracle_processor",
							UnitCost:   300,
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
					mockLicenseClient.EXPECT().ProductLicensesForMetric(ctx, gomock.Any()).Times(2).DoAndReturn(func(ctx context.Context, req *ls.ProductLicensesForMetricRequest) (*ls.ProductLicensesForMetricResponse, error) {
						if req.MetricName == "ibm_pvu" {
							return &ls.ProductLicensesForMetricResponse{
								NumCptLicences: 1200,
								TotalCost:      240000,
								MetricName:     "ibm_pvu",
							}, nil
						}
						return nil, errors.New("Internal")
					}),
				)
			},
			want: &v1.SimulationByMetricResponse{
				MetricSimResult: []*v1.MetricSimulationResult{
					&v1.MetricSimulationResult{
						Success:        true,
						NumCptLicences: 1200,
						TotalCost:      240000,
						MetricName:     "ibm_pvu",
					},
					&v1.MetricSimulationResult{
						MetricName:       "oracle_processor",
						SimFailureReason: "Internal",
					},
				},
			},
		},
		{
			name: "Success - With no metrics",
			args: args{
				ctx: ctx,
				req: &v1.SimulationByMetricRequest{
					SwidTag:       "Oracle_Database_11g_Enterprise_Edition_10.3",
					MetricDetails: []*v1.MetricSimDetails{},
					Scope:         "Scope1",
				},
			},
			setup: func() {

			},
			want: &v1.SimulationByMetricResponse{},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SimulationService.SimulationByMetric() = %v, want %v", got, tt.want)
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
		&v1.EquipAttribute{
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
		&v1.EquipAttribute{
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
		&v1.EquipAttribute{
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
		&v1.EquipAttribute{
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
		&v1.EquipAttribute{
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
						&v1.SimMetricDetails{
							MetricType: "oracle.processor.standard",
							MetricName: "oracle_processor",
						},
						&v1.SimMetricDetails{
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
								&ls.ProductLicenseForEquipAndMetric{
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
					&v1.SimulatedProductsLicenses{
						Success:    true,
						MetricName: "oracle_processor",
						Licenses: []*v1.SimulatedProductLicense{
							&v1.SimulatedProductLicense{
								OldLicences: 120000,
								NewLicenses: 130000,
								Delta:       10000,
								SwidTag:     "Oracle_Real_Application_Testing_12.1.0.1.0",
								ProductName: "Oracle Real Application Testing",
								Editor:      "Oracle",
							},
						},
					},
					&v1.SimulatedProductsLicenses{
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
