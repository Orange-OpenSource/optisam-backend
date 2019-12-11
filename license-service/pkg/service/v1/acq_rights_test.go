// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/ctxmanage"
	"optisam-backend/common/optisam/token/claims"
	v1 "optisam-backend/license-service/pkg/api/v1"
	repo "optisam-backend/license-service/pkg/repository/v1"
	"optisam-backend/license-service/pkg/repository/v1/mock"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_licenseServiceServer_ListAcquiredRights(t *testing.T) {
	ctx := ctxmanage.AddClaims(context.Background(), &claims.Claims{
		UserID: "admin@superuser.com",
		Role:   "Admin",
		Socpes: []string{"A", "B"},
	})

	var mockCtrl *gomock.Controller
	var rep repo.License
	type args struct {
		ctx context.Context
		req *v1.ListAcquiredRightsRequest
	}
	tests := []struct {
		name    string
		args    args
		setup   func()
		want    *v1.ListAcquiredRightsResponse
		wantErr bool
	}{
		{name: "success all filters",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsRequest{
					PageNum:   2,
					PageSize:  10,
					SortOrder: v1.SortOrder_DESC,
					SortBy:    v1.ListAcquiredRightsRequest_ENTITY,
					SearchParams: &v1.AcquiredRightsSearchParams{
						SKU: &v1.StringFilter{
							Filteringkey: "sku",
						},
						SwidTag: &v1.StringFilter{
							Filteringkey: "st",
						},
						ProductName: &v1.StringFilter{
							Filteringkey: "pn",
						},
						Editor: &v1.StringFilter{
							Filteringkey: "ed",
						},
						Metric: &v1.StringFilter{
							Filteringkey: "me",
						},
					},
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo
				mockRepo.EXPECT().AcquiredRights(ctx, &repo.QueryAcquiredRights{
					PageSize:  10,
					Offset:    10,
					SortBy:    repo.AcquiredRightsSortByEntity,
					SortOrder: repo.SortDESC,
					Filter: &repo.AggregateFilter{
						Filters: []repo.Queryable{
							&repo.Filter{
								FilterKey:   repo.AcquiredRightsSearchKeySKU.String(),
								FilterValue: "sku",
							},
							&repo.Filter{
								FilterKey:   repo.AcquiredRightsSearchKeySwidTag.String(),
								FilterValue: "st",
							},
							&repo.Filter{
								FilterKey:   repo.AcquiredRightsSearchKeyProductName.String(),
								FilterValue: "pn",
							},
							&repo.Filter{
								FilterKey:   repo.AcquiredRightsSearchKeyEditor.String(),
								FilterValue: "ed",
							},
							&repo.Filter{
								FilterKey:   repo.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "me",
							},
						},
					},
				}, []string{"A", "B"}).Return(int32(10), []*repo.AcquiredRights{
					&repo.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC001PROC",
						SwidTag:                        "ORAC001",
						ProductName:                    "Oracle Client",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         1016,
						LicensesUnderMaintenanceNumber: 1008,
						AvgLicenesUnitPrice:            2042,
						AvgMaintenanceUnitPrice:        14294,
						TotalPurchaseCost:              2074672,
						TotalMaintenanceCost:           14408352,
						TotalCost:                      35155072,
					},
					&repo.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC002PROC",
						SwidTag:                        "ORAC002",
						ProductName:                    "Oracle XML Development Kit",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         181,
						LicensesUnderMaintenanceNumber: 181,
						AvgLicenesUnitPrice:            1759,
						AvgMaintenanceUnitPrice:        12313,
						TotalPurchaseCost:              318379,
						TotalMaintenanceCost:           2228653,
						TotalCost:                      5412443,
					},
				}, nil).Times(1)
			},
			want: &v1.ListAcquiredRightsResponse{
				TotalRecords: 10,
				AcquiredRights: []*v1.AcquiredRights{
					&v1.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC001PROC",
						SwidTag:                        "ORAC001",
						ProductName:                    "Oracle Client",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         1016,
						LicensesUnderMaintenanceNumber: 1008,
						AvgLicenesUnitPrice:            2042,
						AvgMaintenanceUnitPrice:        14294,
						TotalPurchaseCost:              2074672,
						TotalMaintenanceCost:           14408352,
						TotalCost:                      35155072,
					},
					&v1.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC002PROC",
						SwidTag:                        "ORAC002",
						ProductName:                    "Oracle XML Development Kit",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         181,
						LicensesUnderMaintenanceNumber: 181,
						AvgLicenesUnitPrice:            1759,
						AvgMaintenanceUnitPrice:        12313,
						TotalPurchaseCost:              318379,
						TotalMaintenanceCost:           2228653,
						TotalCost:                      5412443,
					},
				},
			},
		},
		{name: "success no filters",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsRequest{
					PageNum:   2,
					PageSize:  10,
					SortOrder: v1.SortOrder_DESC,
					SortBy:    v1.ListAcquiredRightsRequest_ENTITY,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().AcquiredRights(ctx, &repo.QueryAcquiredRights{
					PageSize:  10,
					Offset:    10,
					SortBy:    repo.AcquiredRightsSortByEntity,
					SortOrder: repo.SortDESC,
				}, []string{"A", "B"}).Return(int32(10), []*repo.AcquiredRights{
					&repo.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC001PROC",
						SwidTag:                        "ORAC001",
						ProductName:                    "Oracle Client",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         1016,
						LicensesUnderMaintenanceNumber: 1008,
						AvgLicenesUnitPrice:            2042,
						AvgMaintenanceUnitPrice:        14294,
						TotalPurchaseCost:              2074672,
						TotalMaintenanceCost:           14408352,
						TotalCost:                      35155072,
					},
					&repo.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC002PROC",
						SwidTag:                        "ORAC002",
						ProductName:                    "Oracle XML Development Kit",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         181,
						LicensesUnderMaintenanceNumber: 181,
						AvgLicenesUnitPrice:            1759,
						AvgMaintenanceUnitPrice:        12313,
						TotalPurchaseCost:              318379,
						TotalMaintenanceCost:           2228653,
						TotalCost:                      5412443,
					},
				}, nil).Times(1)
			},
			want: &v1.ListAcquiredRightsResponse{
				TotalRecords: 10,
				AcquiredRights: []*v1.AcquiredRights{
					&v1.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC001PROC",
						SwidTag:                        "ORAC001",
						ProductName:                    "Oracle Client",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         1016,
						LicensesUnderMaintenanceNumber: 1008,
						AvgLicenesUnitPrice:            2042,
						AvgMaintenanceUnitPrice:        14294,
						TotalPurchaseCost:              2074672,
						TotalMaintenanceCost:           14408352,
						TotalCost:                      35155072,
					},
					&v1.AcquiredRights{
						Entity:                         "",
						SKU:                            "ORAC002PROC",
						SwidTag:                        "ORAC002",
						ProductName:                    "Oracle XML Development Kit",
						Editor:                         "oracle",
						Metric:                         "oracle.processor.standard",
						AcquiredLicensesNumber:         181,
						LicensesUnderMaintenanceNumber: 181,
						AvgLicenesUnitPrice:            1759,
						AvgMaintenanceUnitPrice:        12313,
						TotalPurchaseCost:              318379,
						TotalMaintenanceCost:           2228653,
						TotalCost:                      5412443,
					},
				},
			},
		},
		{name: "failure - can not retrieve claims",
			args: args{
				ctx: context.Background(),
				req: &v1.ListAcquiredRightsRequest{
					PageNum:   2,
					PageSize:  10,
					SortOrder: v1.SortOrder_DESC,
					SortBy:    v1.ListAcquiredRightsRequest_ENTITY,
				},
			},
			setup:   func() {},
			wantErr: true,
		},
		{name: "failure",
			args: args{
				ctx: ctx,
				req: &v1.ListAcquiredRightsRequest{
					PageNum:   2,
					PageSize:  10,
					SortOrder: v1.SortOrder_DESC,
					SortBy:    v1.ListAcquiredRightsRequest_ENTITY,
				},
			},
			setup: func() {
				mockCtrl = gomock.NewController(t)
				mockRepo := mock.NewMockLicense(mockCtrl)
				rep = mockRepo

				mockRepo.EXPECT().AcquiredRights(ctx, &repo.QueryAcquiredRights{
					PageSize:  10,
					Offset:    10,
					SortBy:    repo.AcquiredRightsSortByEntity,
					SortOrder: repo.SortDESC,
				}, []string{"A", "B"}).Return(int32(0), nil, errors.New("test error")).Times(1)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := NewLicenseServiceServer(rep).ListAcquiredRights(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("licenseServiceServer.ListAcquiredRights() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareListAcquiredRightsResponse(t, "ListAcquiredRequestResponse", tt.want, got)
			}
		})
	}
}

func compareListAcquiredRightsResponse(t *testing.T, name string, exp, act *v1.ListAcquiredRightsResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}
	compareAcquiredRightsAll(t, name+".AcquiredRights", exp.AcquiredRights, act.AcquiredRights)
}

func compareAcquiredRightsAll(t *testing.T, name string, exp []*v1.AcquiredRights, act []*v1.AcquiredRights) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareAcquiredRights(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareAcquiredRights(t *testing.T, name string, exp *v1.AcquiredRights, act *v1.AcquiredRights) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	// if exp.ID != "" {
	// 	assert.Equalf(t, exp.ID, act.ID, "%s.ID are not same", name)
	// }
	assert.Equalf(t, exp.Entity, act.Entity, "%s.Entity are not same", name)
	assert.Equalf(t, exp.SKU, act.SKU, "%s.SKU are not same", name)
	assert.Equalf(t, exp.SwidTag, act.SwidTag, "%s.SwidTag are not same", name)
	assert.Equalf(t, exp.ProductName, act.ProductName, "%s.ProductName are not same", name)
	assert.Equalf(t, exp.Editor, act.Editor, "%s.Type are not same", name)
	assert.Equalf(t, exp.Metric, act.Metric, "%s.Metric are not same", name)
	assert.Equalf(t, exp.AcquiredLicensesNumber, act.AcquiredLicensesNumber, "%s.AcquiredLicensesNumber are not same", name)
	assert.Equalf(t, exp.LicensesUnderMaintenanceNumber, act.LicensesUnderMaintenanceNumber, "%s.LicensesUnderMaintenanceNumber are not same", name)
	assert.Equalf(t, exp.AvgLicenesUnitPrice, act.AvgLicenesUnitPrice, "%s.AvgLicenesUnitPrice are not same", name)
	assert.Equalf(t, exp.AvgMaintenanceUnitPrice, act.AvgMaintenanceUnitPrice, "%s.AvgMaintenanceUnitPrice are not same", name)
	assert.Equalf(t, exp.TotalPurchaseCost, act.TotalPurchaseCost, "%s.TotalPurchaseCost are not same", name)
	assert.Equalf(t, exp.TotalMaintenanceCost, act.TotalMaintenanceCost, "%s.TotalMaintenanceCost are not same", name)
	assert.Equalf(t, exp.TotalCost, act.TotalCost, "%s.TotalCost are not same", name)
}
