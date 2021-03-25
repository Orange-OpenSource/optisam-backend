// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ParentsHirerachyForEquipment(t *testing.T) {
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()
	type args struct {
		ctx            context.Context
		equipID        string
		equipType      string
		hirearchyLevel uint8
		scopes         string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    *v1.Equipment
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "SERV_0001",
				equipType:      "Vcenter",
				hirearchyLevel: 5,
				scopes:         "scope1",
			},
			want: &v1.Equipment{
				EquipID: "SERV_0001",
				Type:    "Server",
				Parent: &v1.Equipment{
					EquipID: "CL_001",
					Type:    "Cluster",
					Parent: &v1.Equipment{
						EquipID: "VC_001",
						Type:    "Vcenter",
						Parent: &v1.Equipment{
							EquipID: "DT_01",
							Type:    "Datacenter",
							Parent:  nil,
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.ParentsHirerachyForEquipment(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.hirearchyLevel, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ParentsHirerachyForEquipment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareEquipmentList(t, "ParentsHirerachyForEquipment", tt.want, got)
			}
		})
	}
}

func TestLicenseRepository_ProductsForEquipmentForMetricOracleProcessorStandard(t *testing.T) {
	eqTypes, cleanup, err := equipmentSetup(t)
	if !assert.Empty(t, err, "error not expected as cleanup") {
		return
	}

	if !assert.Empty(t, loadEquipments("badger", "testdata", []string{"scope1", "scope2", "scope3"}, []string{
		"equip_3.csv",
		"equip_4.csv",
	}...), "error not expected in loading equipments") {
		return
	}

	defer func() {
		assert.Empty(t, cleanup(), "error  not expected from clean up")
	}()
	eqType := eqTypes[1]
	type args struct {
		ctx            context.Context
		equipID        string
		equipType      string
		hirearchyLevel uint8
		metric         *v1.MetricOPSComputed
		scopes         string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    []*v1.ProductData
		wantErr bool
	}{
		{name: "SUCCESS",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "equip4_1",
				equipType:      eqType.Type,
				hirearchyLevel: 5,
				metric: &v1.MetricOPSComputed{
					Name: "oracle.processor.standard",
				},
				scopes: "scope1",
			},
			want: []*v1.ProductData{
				&v1.ProductData{
					Name:     "Oracle Instant Client",
					Version:  "9.2.0.8.0",
					Category: "Other",
					Editor:   "oracle",
					Swidtag:  "ORAC001",
				},
				&v1.ProductData{
					Name:     "ORACLE SGBD Enterprise",
					Version:  "9.2.0.8.0",
					Category: "Database",
					Editor:   "oracle",
					Swidtag:  "ORAC003",
				},
				&v1.ProductData{
					Name:     "Oracle SGBD Noyau",
					Version:  "9.2.0.8.0",
					Category: "Database",
					Editor:   "oracle",
					Swidtag:  "ORAC002",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.ProductsForEquipmentForMetricOracleProcessorStandard(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.hirearchyLevel, tt.args.metric, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ProductsForEquipmentForMetricOracleProcessorStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareProductsForEquipment(t, "ProductsForEquipmentForMetricOracleProcessorStandard", tt.want, got)
			}
		})
	}
}

func TestLicenseRepository_ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(t *testing.T) {

	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()
	repo, err := NewLicenseRepositoryWithTemplates(dgClient)
	if !assert.Emptyf(t, err, "err is not expected from NewLicenseRepositoryWithTemplates") {
		return
	}
	type args struct {
		ctx       context.Context
		equipID   string
		equipType string
		mat       *v1.MetricOPSComputed
		scopes    string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    int64
		want1   float64
		wantErr bool
	}{
		{name: "SUCCESS - server without simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "SERV_0023",
				equipType: "Server",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
				scopes: "scope1",
			},
			want:  int64(2),
			want1: 2,
		},
		{name: "SUCCESS - cluster without simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "CL_011",
				equipType: "Cluster",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
				scopes: "scope1",
			},
			want:  int64(8),
			want1: 8,
		},
		{name: "SUCCESS - vcenter without simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "VC_005",
				equipType: "Vcenter",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
				scopes: "scope1",
			},
			want: int64(36),
		},
		{name: "SUCCESS - datacenter without simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "DT_02",
				equipType: "Datacenter",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					NumCPUAttr: &v1.Attribute{
						Name: "ServerProcessorsNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "OracleCoreFactor",
					},
				},
				scopes: "scope1",
			},
			want: int64(44),
		},
		{name: "SUCCESS - server with simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "SERV_0023",
				equipType: "Server",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name:        "ServerCoresNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      2,
					},
					NumCPUAttr: &v1.Attribute{
						Name:        "ServerProcessorsNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      4,
					},
					CoreFactorAttr: &v1.Attribute{
						Name:        "OracleCoreFactor",
						IsSimulated: true,
						Type:        v1.DataTypeFloat,
						FloatVal:    2,
					},
				},
				scopes: "scope1",
			},
			want:  int64(16),
			want1: 16,
		},
		{name: "SUCCESS - cluster with simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "CL_011",
				equipType: "Cluster",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name:        "ServerCoresNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      2,
					},
					NumCPUAttr: &v1.Attribute{
						Name:        "ServerProcessorsNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      4,
					},
					CoreFactorAttr: &v1.Attribute{
						Name:        "OracleCoreFactor",
						IsSimulated: true,
						Type:        v1.DataTypeFloat,
						FloatVal:    2,
					},
				},
				scopes: "scope1",
			},
			want:  int64(32),
			want1: 32,
		},
		{name: "SUCCESS - vcenter with simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "VC_005",
				equipType: "Vcenter",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name:        "ServerCoresNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      2,
					},
					NumCPUAttr: &v1.Attribute{
						Name:        "ServerProcessorsNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      4,
					},
					CoreFactorAttr: &v1.Attribute{
						Name:        "OracleCoreFactor",
						IsSimulated: true,
						Type:        v1.DataTypeFloat,
						FloatVal:    2,
					},
				},
				scopes: "scope1",
			},
			want: int64(64),
		},
		{name: "SUCCESS - datacenter with simulation",
			r: repo,
			args: args{
				ctx:       context.Background(),
				equipID:   "DT_02",
				equipType: "Datacenter",
				mat: &v1.MetricOPSComputed{
					EqTypeTree: []*v1.EquipmentType{
						&v1.EquipmentType{
							Type: "Partition",
						},
						&v1.EquipmentType{
							Type: "Server",
						},
						&v1.EquipmentType{
							Type: "Cluster",
						},
						&v1.EquipmentType{
							Type: "Vcenter",
						},
						&v1.EquipmentType{
							Type: "Datacenter",
						},
					},
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					AggregateLevel: &v1.EquipmentType{
						Type: "Cluster",
					},
					NumCoresAttr: &v1.Attribute{
						Name:        "ServerCoresNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      4,
					},
					NumCPUAttr: &v1.Attribute{
						Name:        "ServerProcessorsNumber",
						IsSimulated: true,
						Type:        v1.DataTypeInt,
						IntVal:      4,
					},
					CoreFactorAttr: &v1.Attribute{
						Name:        "OracleCoreFactor",
						IsSimulated: true,
						Type:        v1.DataTypeFloat,
						FloatVal:    2,
					},
				},
				scopes: "scope1",
			},
			want: int64(256),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.r.ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ComputedLicensesForEquipmentForMetricOracleProcessorStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.ComputedLicensesForEquipmentForMetricOracleProcessorStandard() = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("LicenseRepository.ComputedLicensesForEquipmentForMetricOracleProcessorStandard() = %v, want1 %v", got1, tt.want1)
			}
		})
	}
}

func TestLicenseRepository_UsersForEquipmentForMetricOracleNUP(t *testing.T) {
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()
	type args struct {
		ctx            context.Context
		equipID        string
		equipType      string
		productID      string
		hirearchyLevel uint8
		metric         *v1.MetricNUPComputed
		scopes         string
	}
	tests := []struct {
		name    string
		r       *LicenseRepository
		args    args
		want    []*v1.User
		wantErr bool
	}{
		{name: "SUCCESS- partition",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "PA_001",
				equipType:      "Server",
				productID:      "ORAC099",
				hirearchyLevel: 1,
				scopes:         "scope1",
			},
			want: []*v1.User{
				&v1.User{
					UserID:    "user_ORAC099PA_001",
					UserCount: 1,
				},
			},
		},
		{name: "SUCCESS- server",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "SERV_0001",
				equipType:      "Server",
				productID:      "ORAC099",
				hirearchyLevel: 2,
				scopes:         "scope1",
			},
			want: []*v1.User{
				&v1.User{
					UserID:    "user_ORAC099PA_001",
					UserCount: 1,
				},
			},
		},
		{name: "SUCCESS - cluster",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "CL_001",
				equipType:      "Cluster",
				productID:      "ORAC099",
				hirearchyLevel: 3,
				scopes:         "scope1",
			},
			want: []*v1.User{
				&v1.User{
					UserID:    "user_ORAC099PA_001",
					UserCount: 1,
				},
			},
		},
		{name: "SUCCESS - vcenter",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "VC_001",
				equipType:      "Vcenter",
				productID:      "ORAC099",
				hirearchyLevel: 4,
				scopes:         "scope1",
			},
			want: []*v1.User{
				&v1.User{
					UserID:    "user_ORAC099PA_001",
					UserCount: 1,
				},
			},
		},
		{name: "SUCCESS - datacenter",
			r: NewLicenseRepository(dgClient),
			args: args{
				ctx:            context.Background(),
				equipID:        "DT_01",
				equipType:      "DataCenter",
				productID:      "ORAC099",
				hirearchyLevel: 5,
				scopes:         "scope1",
			},
			want: []*v1.User{
				&v1.User{
					UserID:    "user_ORAC099PA_001",
					UserCount: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.UsersForEquipmentForMetricOracleNUP(tt.args.ctx, tt.args.equipID, tt.args.equipType, tt.args.productID, tt.args.hirearchyLevel, tt.args.metric, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.UsersForEquipmentForMetricOracleNUP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareUsersForEquipmentForMetricOracleNUP(t, "UsersForEquipmentForMetricOracleNUP", tt.want, got)
			}
		})
	}
}

func compareEquipmentList(t *testing.T, name string, exp, act *v1.Equipment) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "Equipmet is expected to be nil")
	}
	assert.Equalf(t, exp.EquipID, act.EquipID, "%s.EquipID should be same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type should be same", name)
	compareEquipmentList(t, name+".Parent", exp.Parent, act.Parent)
}

func compareProductsForEquipment(t *testing.T, name string, exp, act []*v1.ProductData) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "Equipment is expected to be nil")
	}
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}
	for i := range exp {
		if idx := productDataIndex(exp[i], act); idx != -1 {
			compareProductData(t, name, exp[i], act[idx])
		}
	}
}

func compareUsersForEquipmentForMetricOracleNUP(t *testing.T, name string, exp, act []*v1.User) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "User is expected to be nil")
	}
	if !assert.Lenf(t, act, len(exp), "expected number of users are: %d", len(exp)) {
		return
	}
	for i := range exp {
		if idx := userIndex(exp[i], act); idx != -1 {
			compareUser(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
		}
	}
}

func compareUser(t *testing.T, name string, exp, act *v1.User) {
	assert.Equalf(t, exp.UserID, act.UserID, "%s.UserID should be same", name)
	assert.Equalf(t, exp.UserCount, act.UserCount, "%s.UserCount should be same", name)
}

func userIndex(exp *v1.User, act []*v1.User) int {
	for i := range act {
		if exp.UserID == act[i].UserID {
			return i
		}
	}
	return -1
}
