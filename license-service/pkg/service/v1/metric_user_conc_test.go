package v1

import (
	"context"
	"database/sql"
	"testing"

	grpc_middleware "gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/middleware/grpc"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/token/claims"
	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1/mock"
	prov1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1"
	mockpro "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/thirdparty/product-service/pkg/api/v1/mock"

	"github.com/golang/mock/gomock"
)

func Test_licenseServiceServer_computedLicensesUCS(t *testing.T) {
	ctx := grpc_middleware.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"scope1", "scope2"},
	})
	metrics := []*repo.MetricUCS{
		{
			Name: "metric1",
			ID:   "oracle.processor.standard",
		},
	}
	var mockCtrl *gomock.Controller
	var rep repo.License
	var prod prov1.ProductServiceClient
	type args struct {
		ctx     context.Context
		req     map[string]interface{}
		eqTypes []*repo.EquipmentType
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		wantErr bool
	}{
		{name: "SUCCESS - individual",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
				eqTypes: []*repo.EquipmentType{&repo.EquipmentType{ID: "s1", Type: "s1"}},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric2",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
			wantErr: true,
		},
		{name: "SUCCESS - individual iss agg false",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        false,
					"IS_SA":         false,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicenses(ctx, gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), nil)
			},
		},
		{name: "SUCCESS - err",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, sql.ErrNoRows)
			},
			wantErr: true,
		},
		{name: "SUCCESS - not found",
			args: args{
				ctx: ctx,
				req: map[string]interface{}{
					"scopes":        []string{"scope1", "scope2"},
					"METRIC_NAME":   "metric1",
					"PROD_AGG_NAME": "agg",
					"IS_AGG":        true,
					"IS_SA":         true,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockProdClient := mockpro.NewMockProductServiceClient(mockCtrl)
				prod = mockProdClient
				mockRepo.EXPECT().ListMetricUCS(ctx, gomock.Any()).AnyTimes().Return(metrics, nil)
				mockRepo.EXPECT().MetricUCSComputedLicensesAgg(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(uint64(1), uint64(1), sql.ErrNoRows)

			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			s := &licenseServiceServer{
				licenseRepo:   rep,
				productClient: prod,
			}
			_, _, err := s.computedLicensesUCS(tt.args.ctx, tt.args.eqTypes, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListComputationDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// if !tt.wantErr {
			// 	compareListComputationDetailsResponse(t, "licenseServiceServer.ListComputationDetails", tt.want, got)
			// }
		})
	}
}
