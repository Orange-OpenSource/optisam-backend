// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_GetMetricConfigINM(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  []string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func(l *MetricRepository) (func() error, error)
		want    *v1.MetricINMConfig
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "inm",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met1, err := l.CreateMetricInstanceNumberStandard(context.Background(), &v1.MetricINM{
					Name:        "inm",
					Coefficient: 5.6,
				}, []string{"scope1"})
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric 1")
				}
				fmt.Println(met1.ID)
				return func() error {
					assert.Empty(t, deleteNode(met1.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			want: &v1.MetricINMConfig{
				Name:        "inm",
				Coefficient: 5.6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup(tt.l)
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			got, err := tt.l.GetMetricConfigINM(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigINM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricINM(t, "MetricRepository.GetMetricConfigINM", tt.want, got)
			}
		})
	}
}

func compareMetricINM(t *testing.T, name string, exp, act *v1.MetricINMConfig) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Source should be same", name)
	assert.Equalf(t, exp.Coefficient, act.Coefficient, "%s.Coefficient should be same", name)
}
