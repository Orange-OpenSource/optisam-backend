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
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ListMetrices(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes []string
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		setup   func() (func() error, error)
		want    []*v1.Metric
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx: context.Background(),
			},
			setup: func() (func() error, error) {
				// TODO create two nodes for metrics
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						&api.NQuad{
							Subject:     blankID("met1"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						&api.NQuad{
							Subject:     blankID("met1"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type1"),
						},
						&api.NQuad{
							Subject:     blankID("met1"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						&api.NQuad{
							Subject:     blankID("met2"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						&api.NQuad{
							Subject:     blankID("met2"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type2"),
						},
						&api.NQuad{
							Subject:     blankID("met2"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
					},
				}

				assigned, err := dgClient.NewTxn().Mutate(context.Background(), mu)
				if err != nil {
					return nil, err
				}

				metID1, ok := assigned.Uids["met1"]
				if !ok {
					return nil, errors.New("cannot find metric1 id after mutation in setup")
				}

				metID2, ok := assigned.Uids["met2"]
				if !ok {
					return nil, errors.New("cannot find metric2 id after mutation in setup")
				}
				return func() error {
					return deleteNodes(metID1, metID2)
				}, nil
			},

			want: []*v1.Metric{
				&v1.Metric{
					Name: "Oracle type1",
					Type: v1.MetricOPSOracleProcessorStandard,
				},
				&v1.Metric{
					Name: "Oracle type2",
					Type: v1.MetricOPSOracleProcessorStandard,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			got, err := tt.l.ListMetrices(tt.args.ctx, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("LicenseRepository.ListMetrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricsAll(t, "ListMetrices", tt.want, got)
			}
		})
	}
}

func compareMetricsAll(t *testing.T, name string, exp, act []*v1.Metric) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		idx := getMetricByName(exp[i].Name, act)
		if !assert.NotEqualf(t, -1, idx, "group by Name: %s not found in Metrics ", exp[i].Name) {
			return
		}
		compareMetrics(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[idx])
	}
}

func getMetricByName(name string, met []*v1.Metric) int {
	for i := range met {
		if name == met[i].Name {
			return i
		}
	}
	return -1
}

func compareMetrics(t *testing.T, name string, exp, act *v1.Metric) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metric is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type should be same", name)
}
