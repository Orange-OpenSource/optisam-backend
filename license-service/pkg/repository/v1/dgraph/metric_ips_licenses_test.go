package dgraph

import (
	"context"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricIPSComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricIPSComputed
		scopes string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()

	ID, err := getUIDForProductXID("ORAC098", []string{"scope2"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricIPSComputed{
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "PVU",
					},
				},
			},
			want: uint64(13),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricIPSComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricIPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricIPSComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseRepository_MetricIPSComputedLicensesAgg(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricIPSComputed
		scopes string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, cleanup(), "error is not expected in cleanup") {
			return
		}
	}()

	ID, err := getUIDForProductXID("ORAC098", []string{"scope2"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}

	metric := "abc"
	aggName := "xyz"
	aggCleanup, err := aggSetup(metric, ID, aggName, "scope2")
	if !assert.Empty(t, err, "error is not expected in agg setup") {
		return
	}

	defer func() {
		if !assert.Empty(t, aggCleanup(), "error is not expected in aggCleanup") {
			return
		}
	}()

	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricIPSComputed{
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "PVU",
					},
				},
			},
			want: uint64(13),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricIPSComputedLicensesAgg(tt.args.ctx, aggName, metric, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricIPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricIPSComputedLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}
