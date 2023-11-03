package dgraph

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricAttrSum(t *testing.T) {
	type args struct {
		ctx       context.Context
		met       *v1.MetricAttrSumStand
		attribute *v1.Attribute
		scope     string
	}
	tests := []struct {
		name       string
		l          *MetricRepository
		args       args
		wantRetmet *v1.MetricAttrSumStand
		wantErr    bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx: context.Background(),
				met: &v1.MetricAttrSumStand{
					Name:           "attribute.sum.standard",
					EqType:         "MyType1",
					AttributeName:  "attr1",
					ReferenceValue: 2,
				},
				attribute: &v1.Attribute{
					Name:         "attr1",
					Type:         v1.DataTypeInt,
					IsSearchable: true,
					IntVal:       5,
				},
				scope: "scope1",
			},
			wantRetmet: &v1.MetricAttrSumStand{
				Name:           "attribute.sum.standard",
				EqType:         "MyType1",
				AttributeName:  "attr1",
				ReferenceValue: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricAttrSum(tt.args.ctx, tt.args.met, tt.args.attribute, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricAttrSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
				compareMetricAttrSum(t, "CreateMetricAttrSum", tt.wantRetmet, gotRetmet)
			}
		})
	}
}

func TestMetricRepository_ListMetricAttrSum(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func(*MetricRepository) ([]*v1.MetricAttrSumStand, func() error, error)
		want    []*v1.MetricAttrSumStand
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func(l *MetricRepository) (retMat []*v1.MetricAttrSumStand, cleanup func() error, retErr error) {
				retMat = []*v1.MetricAttrSumStand{}
				gotRetmet1, err := l.CreateMetricAttrSum(context.Background(), &v1.MetricAttrSumStand{
					Name:           "attribute.counter.standard",
					EqType:         "server",
					AttributeName:  "corefactor",
					ReferenceValue: 2,
				}, &v1.Attribute{
					Name:         "corefactor",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
					FloatVal:     2,
				}, "scope1")
				if err != nil {
					return nil, nil, errors.New("error while creating metric 1")
				}
				gotRetmet2, err := l.CreateMetricAttrSum(context.Background(), &v1.MetricAttrSumStand{
					Name:           "AttrSum1",
					EqType:         "server",
					AttributeName:  "cpu",
					ReferenceValue: 2,
				}, &v1.Attribute{
					Name:         "cpu",
					Type:         v1.DataTypeInt,
					IsSearchable: true,
					IntVal:       3,
				}, "scope1")
				if err != nil {
					return nil, nil, errors.New("error while creating metric 1")
				}
				retMat = append(retMat, gotRetmet1, gotRetmet2)
				return retMat, func() error {
					assert.Empty(t, deleteNode(gotRetmet1.ID), "error not expected in deleting metric type")
					assert.Empty(t, deleteNode(gotRetmet2.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wantMet, cleanup, err := tt.setup(tt.l)
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			got, err := tt.l.ListMetricAttrSum(tt.args.ctx, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.ListMetricAttrSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricAttrSumAll(t, "ListMetricAttrSum", got, wantMet)
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigAttrSum(t *testing.T) {
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
		want    *v1.MetricAttrSumStand
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:     context.Background(),
				metName: "acs",
				scopes:  "scope1",
			},
			setup: func(l *MetricRepository) (func() error, error) {
				gotRetmet1, err := l.CreateMetricAttrSum(context.Background(), &v1.MetricAttrSumStand{
					Name:           "acs",
					EqType:         "server",
					AttributeName:  "corefactor",
					ReferenceValue: 4,
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
			want: &v1.MetricAttrSumStand{
				Name:           "acs",
				EqType:         "server",
				AttributeName:  "corefactor",
				ReferenceValue: 4,
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
			got, err := tt.l.GetMetricConfigAttrSum(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigAttrSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricAttrSum(t, "MetricRepository.GetMetricConfigAttrSum", tt.want, got)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricAttrSum(t *testing.T) {
	type args struct {
		ctx   context.Context
		met   *v1.MetricAttrSumStand
		scope string
	}
	tests := []struct {
		name     string
		l        *MetricRepository
		args     args
		setup    func(l *MetricRepository) (func() error, error)
		checking func(l *MetricRepository) (*v1.MetricAttrSumStand, error)
		want     *v1.MetricAttrSumStand
		wantErr  bool
	}{
		{
			name: "testname__",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx:   context.Background(),
				scope: "scope1",
				met: &v1.MetricAttrSumStand{
					Name:           "att",
					EqType:         "zyx",
					AttributeName:  "A2",
					ReferenceValue: 0.88,
				},
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met, err := l.CreateMetricAttrSum(context.Background(), &v1.MetricAttrSumStand{
					Name:           "att",
					EqType:         "abc",
					AttributeName:  "A1",
					ReferenceValue: 0.55,
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
			checking: func(l *MetricRepository) (*v1.MetricAttrSumStand, error) {
				actmet, err := l.GetMetricConfigAttrSum(context.Background(), "att", "scope1")
				if err != nil {
					return nil, err
				}

				return actmet, nil
			},
			want: &v1.MetricAttrSumStand{
				Name:           "att",
				EqType:         "zyx",
				AttributeName:  "A2",
				ReferenceValue: 0.88,
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
			err = tt.l.UpdateMetricAttrSum(tt.args.ctx, tt.args.met, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.UpdateMetricAttrSum() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := tt.checking(tt.l)
				if !assert.Empty(t, err, "not expecting error from checking") {
					return
				}
				compareMetricAttrSum(t, "MetricRepository.UpdateMetricAttrSum", tt.want, got)
			}
		})
	}
}

func compareMetricAttrSumAll(t *testing.T, name string, act, exp []*v1.MetricAttrSumStand) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetricAttrSum(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetricAttrSum(t *testing.T, name string, exp, act *v1.MetricAttrSumStand) {
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
	assert.Equalf(t, exp.ReferenceValue, act.ReferenceValue, "%s.ReferenceValue should be same", name)
}
