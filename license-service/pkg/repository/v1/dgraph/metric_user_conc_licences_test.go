package dgraph

import (
	"context"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricUCSComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     []string
		mat    *v1.MetricUCSComputed
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
				ctx: context.Background(),
				id:  []string{ID},
				mat: &v1.MetricUCSComputed{
					Profile: "p1",
				},
				scopes: []string{"scope1"},
			},
			want:   uint64(5),
			wantcd: uint64(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotcd, err := tt.l.MetricUCSComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricUCSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricUCSComputedLicenses() - computed licenses = %v, want %v", got, tt.want)
			}
			if gotcd != tt.wantcd {
				t.Errorf("LicenseRepository.MetricUCSComputedLicenses() - computed details = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseRepository_MetricUCSComputedLicensesAgg(t *testing.T) {
	type args struct {
		ctx    context.Context
		name   string
		metric string
		mat    *v1.MetricUCSComputed
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
				mat: &v1.MetricUCSComputed{
					Profile: "p1",
				},
				scopes: []string{"scope1"},
			},
			want:   uint64(5),
			wantcd: uint64(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotcd, err := tt.l.MetricUCSComputedLicensesAgg(tt.args.ctx, tt.args.name, tt.args.metric, tt.args.mat, tt.args.scopes...)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricUCSComputedLicensesAgg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricUCSComputedLicensesAgg() - computed licenses = %v, want %v", got, tt.want)
			}
			if gotcd != tt.wantcd {
				t.Errorf("LicenseRepository.MetricUCSComputedLicensesAgg() - computed details = %v, want %v", got, tt.want)
			}
		})
	}
}
