package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_GetMetricConfigConcurentUser(t *testing.T) {
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
		want    *v1.MetricUCS
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "UCS",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met1, err := l.CreateMetricUserConcurentStandard(context.Background(), &v1.MetricUCS{
					Name:    "inm",
					Profile: "P1",
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
			want: &v1.MetricUCS{
				Name:    "UCS",
				Profile: "P1",
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
			got, err := tt.l.GetMetricConfigConcurentUser(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigConcurentUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricUCS(t, "MetricRepository.GetMetricConfigConcurentUser", tt.want, got)
			}
		})
	}
}

func compareMetricUCS(t *testing.T, name string, exp, act *v1.MetricUCS) {
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
	assert.Equalf(t, exp.Profile, act.Profile, "%s.Profile should be same", name)
}

func TestMetricRepository_CreateMetricUserConcurentStandard(t *testing.T) {
	type args struct {
		ctx   context.Context
		met   *v1.MetricUCS
		scope string
	}
	tests := []struct {
		name            string
		l               *MetricRepository
		args            args
		wantRetmet      *v1.MetricUCS
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
				met: &v1.MetricUCS{
					Name:    "user.concurrent.standard",
					Profile: "P1",
				},
			},
			wantRetmet: &v1.MetricUCS{
				Name:    "user.concurrent.standard",
				Profile: "P1",
			},
			wantSchemaNodes: []*SchemaNode{
				{
					Predicate: "metric.user_concurrent.profile",
					Type:      "int",
					Index:     false,
					Tokenizer: []string{},
				},
			},
			predicates: []string{
				"metric.user_concurrent.profile",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricUserConcurentStandard(tt.args.ctx, tt.args.met, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricUserConcurentStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
				compareMetricUCS(t, "MetricUCS", tt.wantRetmet, gotRetmet)
				sns, err := querySchema(tt.predicates...)
				if !assert.Emptyf(t, err, "error is not expect while quering schema for predicates: %v", tt.predicates) {
					return
				}
				compareSchemaNodeAll(t, "schemaNodes", tt.wantSchemaNodes, sns)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricUCS(t *testing.T) {
	myoldprofile := "P1"
	type args struct {
		ctx   context.Context
		met   *v1.MetricUCS
		scope string
	}
	tests := []struct {
		name     string
		l        *MetricRepository
		args     args
		setup    func(l *MetricRepository) error
		checking func(l *MetricRepository) bool
	}{
		{
			name: "testname__",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx:   context.Background(),
				scope: "scope1",
				met: &v1.MetricUCS{
					Name:    "UCS",
					Profile: "P2",
				},
			},
			setup: func(l *MetricRepository) error {
				_, err := l.CreateMetricUserConcurentStandard(context.Background(), &v1.MetricUCS{
					Name:    "UCS",
					Profile: myoldprofile,
				}, "scope1")
				if err != nil {
					return errors.New("error while creating metric")
				}
				return nil
			},
			checking: func(l *MetricRepository) bool {
				met, err := l.GetMetricConfigConcurentUser(context.Background(), "UCS", "scope1")
				if err != nil {
					return false
				}
				if met.Profile == "P2" {
					return true
				}
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.setup(tt.l)
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			err = tt.l.UpdateMetricUCS(tt.args.ctx, tt.args.met, tt.args.scope)
			if err != nil {
				//test case fail(db error)
				t.Errorf("MetricRepository.UpdateMetricUCS() error = %v", err)
				return
			}
			if !tt.checking(tt.l) {
				//not updated
				fmt.Println("Metric not updated")
			}
		})
	}
}
