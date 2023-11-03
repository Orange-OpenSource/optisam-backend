package dgraph

import (
	"context"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricMSEComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     []string
		mat    *v1.MetricMSEComputed
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
		want1   uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  []string{ID},
				mat: &v1.MetricMSEComputed{
					Reference: "Server",
					Core:      "ServerCoresNumber",
					CPU:       "ServerCPUNumber",
				},
			},
			want:  uint64(10),
			want1: uint64(3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricMSEComputedLicenses(tt.args.ctx, true, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricMSEComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricMSEComputedLicenses() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLicenseRepository_MetricMSEComputedLicensesAgg(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricMSEComputed
		scopes string
	}
	cleanup, err := setup()
	if !assert.Empty(t, err, "error is not expected in setup") {
		return
	}
	defer func() {
		//return
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
		//return
		if !assert.Empty(t, aggCleanup(), "error is not expected in aggCleanup") {
			return
		}
		//	time.Sleep(10 * time.Minute)
	}()

	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		want    uint64
		want1   uint64
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
				id:  ID,
				mat: &v1.MetricMSEComputed{
					Reference: "Server",
					Core:      "ServerCoresNumber",
					CPU:       "ServerCPUNumber",
				},
			},
			want:  uint64(10),
			want1: uint64(3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.MetricMSEComputedLicensesAgg(tt.args.ctx, true, aggName, metric, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricMSEComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricMSEComputedLicenses() got = %v, want %v", got, tt.want)
			}
		})
	}
}
