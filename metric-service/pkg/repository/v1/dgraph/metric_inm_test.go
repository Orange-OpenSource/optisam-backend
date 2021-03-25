// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"errors"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_GetMetricConfigINM(t *testing.T) {
	type args struct {
		ctx     context.Context
		metName string
		scopes  string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func(l *MetricRepository) (func() error, error)
		want    *v1.MetricINM
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "inm",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met1, err := l.CreateMetricInstanceNumberStandard(context.Background(), &v1.MetricINM{
					Name:        "inm",
					Coefficient: 5.6,
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric 1")
				}
				return func() error {
					assert.Empty(t, deleteNode(met1.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			want: &v1.MetricINM{
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

func compareMetricINM(t *testing.T, name string, exp, act *v1.MetricINM) {
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

func TestMetricRepository_CreateMetricInstanceNumberStandard(t *testing.T) {
	type args struct {
		ctx   context.Context
		met   *v1.MetricINM
		scope string
	}
	tests := []struct {
		name            string
		l               *MetricRepository
		args            args
		wantRetmet      *v1.MetricINM
		wantSchemaNodes []*SchemaNode
		predicates      []string
		wantErr         bool
	}{
		{
			name: "sucess",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx:   context.Background(),
				scope: "scope1",
				met: &v1.MetricINM{
					Name:        "instance.number.standard",
					Coefficient: 1.0,
				},
			},
			wantRetmet: &v1.MetricINM{
				Name:        "instance.number.standard",
				Coefficient: 1.0,
			},
			wantSchemaNodes: []*SchemaNode{
				&SchemaNode{
					Predicate: "metric.instancenumber.coefficient",
					Type:      "float",
					Index:     false,
					Tokenizer: []string{},
				},
			},
			predicates: []string{
				"metric.instancenumber.coefficient",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricInstanceNumberStandard(tt.args.ctx, tt.args.met, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricACS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
				compareMetricINM(t, "MetricACS", tt.wantRetmet, gotRetmet)
				sns, err := querySchema(tt.predicates...)
				if !assert.Emptyf(t, err, "error is not expect while quering schema for predicates: %v", tt.predicates) {
					return
				}
				compareSchemaNodeAll(t, "schemaNodes", tt.wantSchemaNodes, sns)
			}
		})
	}
}
