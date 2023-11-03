package dgraph

import (
	"context"
	"errors"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

// CreateMetricDataCenterForScope handles datacenter scope metric creation
func TestMetricRepository_CreateMetricDataCenterForScope(t *testing.T) {
	type args struct {
		ctx context.Context
		met *v1.ScopeMetric
	}
	tests := []struct {
		name       string
		l          *MetricRepository
		args       args
		wantRetmet *v1.ScopeMetric
		wantErr    bool
	}{
		{
			name: "sucess",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx: context.Background(),
				met: &v1.ScopeMetric{
					MetricType: "windows.server.datacenter",
					MetricName: "windows.server.datacenter.2016",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
				},
			},
			wantRetmet: &v1.ScopeMetric{
				MetricType: "windows.server.datacenter",
				MetricName: "windows.server.datacenter.2016",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Scope:      "scope1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricDataCenterForScope(tt.args.ctx, tt.args.met)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricUSS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigDataCenterForScope(t *testing.T) {
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
		want    *v1.ScopeMetric
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "windows.server.datacenter.2016",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met1, err := l.CreateMetricDataCenterForScope(context.Background(), &v1.ScopeMetric{
					MetricType: "windows.server.datacenter",
					MetricName: "windows.server.datacenter.2016",
					Reference:  "server",
					Core:       "cores_per_processor",
					CPU:        "server_processors_numbers",
					Scope:      "scope1",
					Default:    true,
				})
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
			want: &v1.ScopeMetric{
				MetricType: "windows.server.datacenter",
				MetricName: "windows.server.datacenter.2016",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Scope:      "scope1",
				Default:    true,
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
			got, err := tt.l.GetMetricConfigDataCenterForScope(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigDataCenterForScope() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricWinDCenter(t, "MetricRepository.GetMetricConfigDataCenterForScope", tt.want, got)
			}
		})
	}
}

func compareMetricWinDCenter(t *testing.T, name string, exp, act *v1.ScopeMetric) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.Equalf(t, exp.MetricName, act.MetricName, "%s.Source should be same", name)
	assert.Equalf(t, exp.Reference, act.Reference, "%s.Reference should be same", name)
	assert.Equalf(t, exp.Core, act.Core, "%s.core should be same", name)
	assert.Equalf(t, exp.CPU, act.CPU, "%s.CPU should be same", name)
	assert.Equalf(t, exp.Default, act.Default, "%s.CPU should be same", name)
}
