package dgraph

import (
	"context"
	"errors"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricEquipAttrStandard(t *testing.T) {
	type args struct {
		ctx       context.Context
		met       *v1.MetricEquipAttrStand
		attribute *v1.Attribute
		scope     string
	}
	tests := []struct {
		name       string
		l          *MetricRepository
		args       args
		wantRetmet *v1.MetricEquipAttrStand
		wantErr    bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx: context.Background(),
				met: &v1.MetricEquipAttrStand{
					Name:          "equipment.attribute.standard",
					EqType:        "MyType1",
					AttributeName: "attr1",
					Environment:   "e1",
					Value:         2,
				},
				attribute: &v1.Attribute{
					Name:         "attr1",
					Type:         v1.DataTypeInt,
					IsSearchable: true,
					IntVal:       5,
				},
				scope: "scope1",
			},
			wantRetmet: &v1.MetricEquipAttrStand{
				Name:          "equipment.attribute.standard",
				EqType:        "MyType1",
				AttributeName: "attr1",
				Environment:   "e1",
				Value:         2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricEquipAttrStandard(tt.args.ctx, tt.args.met, tt.args.attribute, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricEquipAttrStandard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
				compareMetricEquipAtt(t, "CreateMetricEquipAttrStandard", tt.wantRetmet, gotRetmet)
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigEquipAttr(t *testing.T) {
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
		want    *v1.MetricEquipAttrStand
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "eqp",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				gotRetmet1, err := l.CreateMetricEquipAttrStandard(context.Background(), &v1.MetricEquipAttrStand{
					Name:          "eqp",
					EqType:        "server",
					AttributeName: "corefactor",
					Environment:   "environment",
					Value:         4,
				}, &v1.Attribute{
					Name:         "corefactor",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
					FloatVal:     2,
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric 1")
				}
				return func() error {
					assert.Empty(t, deleteNode(gotRetmet1.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			want: &v1.MetricEquipAttrStand{
				Name:          "eqp",
				EqType:        "server",
				AttributeName: "corefactor",
				Environment:   "environment",
				Value:         4,
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
			got, err := tt.l.GetMetricConfigEquipAttr(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigEquipAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricEquipAtt(t, "MetricRepository.GetMetricConfigEquipAttr", tt.want, got)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricEquipAttr(t *testing.T) {
	type args struct {
		ctx   context.Context
		met   *v1.MetricEquipAttrStand
		scope string
	}
	tests := []struct {
		name     string
		l        *MetricRepository
		args     args
		setup    func(l *MetricRepository) (func() error, error)
		checking func(l *MetricRepository) (*v1.MetricEquipAttrStand, error)
		want     *v1.MetricEquipAttrStand
		wantErr  bool
	}{
		{
			name: "testname__",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx:   context.Background(),
				scope: "scope1",
				met: &v1.MetricEquipAttrStand{
					Name:          "att",
					EqType:        "zyx",
					AttributeName: "A2",
					Environment:   "env",
					Value:         8,
				},
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met, err := l.CreateMetricEquipAttrStandard(context.Background(), &v1.MetricEquipAttrStand{
					Name:          "att",
					EqType:        "abc",
					AttributeName: "A1",
					Environment:   "env1",
					Value:         5,
				}, &v1.Attribute{
					Name:         "A1",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
					FloatVal:     2,
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric att")
				}
				return func() error {
					assert.Empty(t, deleteNode(met.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			checking: func(l *MetricRepository) (*v1.MetricEquipAttrStand, error) {
				actmet, err := l.GetMetricConfigEquipAttr(context.Background(), "att", "scope1")
				if err != nil {
					return nil, err
				}

				return actmet, nil
			},
			want: &v1.MetricEquipAttrStand{
				Name:          "att",
				EqType:        "zyx",
				AttributeName: "A2",
				Environment:   "env",
				Value:         8,
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
			err = tt.l.UpdateMetricEquipAttr(tt.args.ctx, tt.args.met, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.UpdateMetricEquipAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := tt.checking(tt.l)
				if !assert.Empty(t, err, "not expecting error from checking") {
					return
				}
				compareMetricEquipAtt(t, "MetricRepository.UpdateMetricEquipAttr", tt.want, got)
			}
		})
	}
}

func compareMetricEquipAtt(t *testing.T, name string, exp, act *v1.MetricEquipAttrStand) {
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
	assert.Equalf(t, exp.EqType, act.EqType, "%s.EqType should be same", name)
	assert.Equalf(t, exp.AttributeName, act.AttributeName, "%s.AttributeName should be same", name)
	assert.Equalf(t, exp.Environment, act.Environment, "%s.Environment should be same", name)
	assert.Equalf(t, exp.Value, act.Value, "%s.Value should be same", name)
}
