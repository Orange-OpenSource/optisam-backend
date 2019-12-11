// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_AcquiredRights(t *testing.T) {
	acquiredRights := []*v1.AcquiredRights{
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
		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "ORAC003PROC",
			SwidTag:                        "ORAC003",
			ProductName:                    "Oracle Instant Client",
			Editor:                         "oracle",
			Metric:                         "oracle.processor.standard",
			AcquiredLicensesNumber:         967,
			LicensesUnderMaintenanceNumber: 954,
			AvgLicenesUnitPrice:            1426,
			AvgMaintenanceUnitPrice:        9982,
			TotalPurchaseCost:              1378942,
			TotalMaintenanceCost:           9522828,
			TotalCost:                      23312248,
		},

		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "SENT001PROC",
			SwidTag:                        "SENT001",
			ProductName:                    "Sentry - KM oracle",
			Editor:                         "oracle",
			Metric:                         "oracle.processor.standard",
			AcquiredLicensesNumber:         75,
			LicensesUnderMaintenanceNumber: 70,
			AvgLicenesUnitPrice:            2556,
			AvgMaintenanceUnitPrice:        17892,
			TotalPurchaseCost:              191700,
			TotalMaintenanceCost:           125244,
			TotalCost:                      316944,
		},
		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "ORAC004PROC",
			SwidTag:                        "ORAC004",
			ProductName:                    "ORACLE SGBD Enterprise",
			Editor:                         "oracle",
			Metric:                         "oracle.processor.standard",
			AcquiredLicensesNumber:         749,
			LicensesUnderMaintenanceNumber: 746,
			AvgLicenesUnitPrice:            1254,
			AvgMaintenanceUnitPrice:        8778,
			TotalPurchaseCost:              939246,
			TotalMaintenanceCost:           6548388,
			TotalCost:                      15940848,
		},
		&v1.AcquiredRights{
			Entity:                         "",
			SKU:                            "WIN1PROC",
			SwidTag:                        "WIN1",
			ProductName:                    "Windows Client",
			Editor:                         "Windows",
			Metric:                         "Windows.processor.standard",
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
			SKU:                            "WIN2PROC",
			SwidTag:                        "WIN2",
			ProductName:                    "Windows XML Development Kit",
			Editor:                         "Windows",
			Metric:                         "Windows.processor.standard",
			AcquiredLicensesNumber:         181,
			LicensesUnderMaintenanceNumber: 181,
			AvgLicenesUnitPrice:            1759,
			AvgMaintenanceUnitPrice:        12313,
			TotalPurchaseCost:              318379,
			TotalMaintenanceCost:           2228653,
			TotalCost:                      5412443,
		},
	}
	type args struct {
		ctx    context.Context
		params *v1.QueryAcquiredRights
		scopes []string
	}
	tests := []struct {
		name    string
		lr      *LicenseRepository
		args    args
		want    int32
		wantAR  []*v1.AcquiredRights
		wantErr bool
	}{
		{name: "success,sortby SKU",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortBySKU,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[1],
			},
		},
		{name: "success,sortby SwidTag",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortBySwidTag,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[1],
			},
		},
		{name: "success,sortby product name",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortByProductName,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[4],
				acquiredRights[0],
			},
		},

		{name: "success,sortby acq licenses number",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByAcquiredLicensesNumber,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[2],
			},
		},
		{name: "success,sortby licenses under maintenance number",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByLicensesUnderMaintenanceNumber,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[2],
			},
		},
		{name: "success,sortby average unit price",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByAvgLicenseUnitPrice,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[3],
				acquiredRights[0],
			},
		},
		{name: "success,sortby average maintenance unit price",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortByAvgMaintenanceUnitPrice,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[4],
				acquiredRights[2],
			},
		},
		{name: "success,sortby total purchase cost",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByTotalPurchaseCost,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[2],
			},
		},
		{name: "success,sortby total maintenance cost",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByTotalMaintenanceCost,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[2],
			},
		},
		{name: "success,sortby total cost",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortDESC,
					SortBy:    v1.AcquiredRightsSortByTotalCost,
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[2],
			},
		},
		{name: "success,sortby unknowns",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortOrder(67),            // Unknown sort order
					SortBy:    v1.AcquiredRightsSortBy(89), // unknown sort by
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[0],
				acquiredRights[1],
			},
		},
		{name: "success,sortby entity search by all one match",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortByEntity,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeySKU.String(),
								FilterValue: "sent",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeySwidTag.String(),
								FilterValue: "sent",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyProductName.String(),
								FilterValue: "Sent",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyEditor.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   "unmatched case",
								FilterValue: "orac",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
			want: 1,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[3],
			},
		},
		{name: "success,sortby entity search by all no match",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortByEntity,
					Filter: &v1.AggregateFilter{
						Filters: []v1.Queryable{
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeySKU.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeySwidTag.String(),
								FilterValue: "sent",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyProductName.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyEditor.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   v1.AcquiredRightsSearchKeyMetric.String(),
								FilterValue: "orac",
							},
							&v1.Filter{
								FilterKey:   "unmatched case",
								FilterValue: "orac",
							},
						},
					},
				},
				scopes: []string{"scope1", "scope2"},
			},
		},
		{name: "success,sortby swid tag - scope 3",
			lr: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				params: &v1.QueryAcquiredRights{
					PageSize:  2,
					Offset:    0,
					SortOrder: v1.SortASC,
					SortBy:    v1.AcquiredRightsSortBySwidTag,
				},
				scopes: []string{"scope3"},
			},
			want: 5,
			wantAR: []*v1.AcquiredRights{
				acquiredRights[5],
				acquiredRights[6],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.lr.AcquiredRights(tt.args.ctx, tt.args.params, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				assert.Equal(t, tt.want, got, "number of records should be equal")
				t.Errorf("LicenseRepository.AcquiredRights() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, tt.want, got, "number of records should be equal")
				compareAcquiredRightsAll(t, "AcquiredRights", tt.wantAR, got1)
			}
		})
	}
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
