package dgraph

import (
	"context"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricUserSumComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricUserSumStand
		scopes []string
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

	ID, err := getUIDForProductXID("ORAC099", []string{"scope1"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantcd  uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				id:     ID,
				scopes: []string{"scope1"},
			},
			want:   uint64(5),
			wantcd: uint64(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotcd, err := tt.l.MetricUserSumComputedLicenses(tt.args.ctx, tt.args.id, tt.args.scopes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicenses() - computed licenses = %v, want %v", got, tt.want)
			}
			if gotcd != tt.wantcd {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicenses() - computed details = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseRepository_MetricUserSumComputedLicensesAgg(t *testing.T) {
	type args struct {
		ctx    context.Context
		name   string
		metric string
		scopes []string
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

	ID, err := getUIDForProductXID("ORAC098", []string{"scope1"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	ID2, err := getUIDForProductXID("ORAC099", []string{"scope1"})
	if !assert.Empty(t, err, "error is not expected in getUIDforProductXID") {
		return
	}
	metric := "abc"
	aggName := "xyz"
	aggCleanup1, err := aggSetup(metric, ID, aggName, "scope1")
	if !assert.Empty(t, err, "error is not expected in agg setup") {
		return
	}
	aggCleanup2, err := aggSetup(metric, ID2, aggName, "scope1")
	if !assert.Empty(t, err, "error is not expected in agg setup") {
		return
	}
	defer func() {
		if !assert.Empty(t, aggCleanup1(), "error is not expected in aggCleanup") {
			return
		}
		if !assert.Empty(t, aggCleanup2(), "error is not expected in aggCleanup") {
			return
		}
	}()
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		wantcd  uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				name:   "xyz",
				metric: "abc",
				scopes: []string{"scope1"},
			},
			want:   uint64(5),
			wantcd: uint64(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotcd, err := tt.l.MetricUserSumComputedLicensesAgg(tt.args.ctx, tt.args.name, tt.args.metric, tt.args.scopes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicensesAgg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicensesAgg() - computed licenses = %v, want %v", got, tt.want)
			}
			if gotcd != tt.wantcd {
				t.Errorf("LicenseRepository.MetricUserSumComputedLicensesAgg() - computed details = %v, want %v", got, tt.want)
			}
		})
	}
}
