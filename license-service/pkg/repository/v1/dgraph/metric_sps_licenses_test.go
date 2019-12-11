// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_MetricSPSComputedLicenses(t *testing.T) {
	type args struct {
		ctx    context.Context
		id     string
		mat    *v1.MetricSPSComputed
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

	ID, err := getUIDForProductXID("ORAC098")
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
				id:  ID,
				mat: &v1.MetricSPSComputed{
					BaseType: &v1.EquipmentType{
						Type: "Server",
					},
					NumCoresAttr: &v1.Attribute{
						Name: "ServerCoresNumber",
					},
					CoreFactorAttr: &v1.Attribute{
						Name: "SAG",
					},
				},
			},
			want:  uint64(10),
			want1: uint64(3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.l.MetricSPSComputedLicenses(tt.args.ctx, tt.args.id, tt.args.mat, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.MetricSPSComputedLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LicenseRepository.MetricSPSComputedLicenses() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("LicenseRepository.MetricSPSComputedLicenses() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
