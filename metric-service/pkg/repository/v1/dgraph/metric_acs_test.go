package dgraph

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_CreateMetricACS(t *testing.T) {
	type args struct {
		ctx       context.Context
		met       *v1.MetricACS
		attribute *v1.Attribute
		scopes    string
	}
	tests := []struct {
		name            string
		l               *MetricRepository
		args            args
		wantRetmet      *v1.MetricACS
		wantSchemaNodes []*SchemaNode
		predicates      []string
		wantErr         bool
	}{
		{name: "sucess",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx: context.Background(),
				met: &v1.MetricACS{
					Name:          "attribute.counter.standard",
					EqType:        "MyType1",
					AttributeName: "attr1",
					Value:         "attrvalue",
				},
				attribute: &v1.Attribute{
					Name:         "attr1",
					Type:         v1.DataTypeString,
					IsSearchable: true,
				},
				scopes: "scope1",
			},
			wantRetmet: &v1.MetricACS{
				Name:          "attribute.counter.standard",
				EqType:        "MyType1",
				AttributeName: "attr1",
				Value:         "attrvalue",
			},
			wantSchemaNodes: []*SchemaNode{
				{
					Predicate: "equipment.MyType1.attr1",
					Type:      "string",
					Index:     true,
					Tokenizer: []string{"trigram", "exact"},
				},
			},
			predicates: []string{
				"equipment.MyType1.attr1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetmet, err := tt.l.CreateMetricACS(tt.args.ctx, tt.args.met, tt.args.attribute, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.CreateMetricACS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer func() {
					assert.Empty(t, deleteNode(gotRetmet.ID), "error not expected in deleting metric type")
				}()
				compareMetricACS(t, "MetricACS", tt.wantRetmet, gotRetmet)
				sns, err := querySchema(tt.predicates...)
				if !assert.Emptyf(t, err, "error is not expect while quering schema for predicates: %v", tt.predicates) {
					return
				}
				compareSchemaNodeAll(t, "schemaNodes", tt.wantSchemaNodes, sns)
			}
		})
	}
}

func TestMetricRepository_ListMetricACS(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func(*MetricRepository) ([]*v1.MetricACS, func() error, error)
		want    []*v1.MetricACS
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func(l *MetricRepository) (retMat []*v1.MetricACS, cleanup func() error, retErr error) {
				retMat = []*v1.MetricACS{}
				gotRetmet1, err := l.CreateMetricACS(context.Background(), &v1.MetricACS{
					Name:          "attribute.counter.standard",
					EqType:        "server",
					AttributeName: "corefactor",
					Value:         "2",
				}, &v1.Attribute{
					Name:         "corefactor",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
				}, "scope1")
				if err != nil {
					return nil, nil, errors.New("error while creating metric 1")
				}
				gotRetmet2, err := l.CreateMetricACS(context.Background(), &v1.MetricACS{
					Name:          "ACS1",
					EqType:        "server",
					AttributeName: "cpu",
					Value:         "2",
				}, &v1.Attribute{
					Name:         "cpu",
					Type:         v1.DataTypeInt,
					IsSearchable: true,
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
			got, err := tt.l.ListMetricACS(tt.args.ctx, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.ListMetricACS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricACSAll(t, "ListMetricACS", got, wantMet)
			}
		})
	}
}

func TestMetricRepository_GetMetricConfigACS(t *testing.T) {
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
		want    *v1.MetricACS
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
				gotRetmet1, err := l.CreateMetricACS(context.Background(), &v1.MetricACS{
					Name:          "acs",
					EqType:        "server",
					AttributeName: "corefactor",
					Value:         "4",
				}, &v1.Attribute{
					Name:         "corefactor",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
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
			want: &v1.MetricACS{
				Name:          "acs",
				EqType:        "server",
				AttributeName: "corefactor",
				Value:         "4",
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
			got, err := tt.l.GetMetricConfigACS(tt.args.ctx, tt.args.metName, tt.args.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.GetMetricConfigACS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricACS(t, "MetricRepository.GetMetricConfigACS", tt.want, got)
			}
		})
	}
}

func TestMetricRepository_UpdateMetricACS(t *testing.T) {
	type args struct {
		ctx   context.Context
		met   *v1.MetricACS
		scope string
	}
	tests := []struct {
		name     string
		l        *MetricRepository
		args     args
		setup    func(l *MetricRepository) (func() error, error)
		checking func(l *MetricRepository) (*v1.MetricACS, error)
		want     *v1.MetricACS
		wantErr  bool
	}{
		{
			name: "testname__",
			l:    NewMetricRepository(dgClient),
			args: args{
				ctx:   context.Background(),
				scope: "scope1",
				met: &v1.MetricACS{
					Name:          "acs",
					EqType:        "zyx",
					AttributeName: "A2",
					Value:         "8",
				},
			},
			setup: func(l *MetricRepository) (func() error, error) {
				met, err := l.CreateMetricACS(context.Background(), &v1.MetricACS{
					Name:          "acs",
					EqType:        "abc",
					AttributeName: "A1",
					Value:         "5",
				}, &v1.Attribute{
					Name:         "A1",
					Type:         v1.DataTypeFloat,
					IsSearchable: true,
				}, "scope1")
				if err != nil {
					return func() error {
						return nil
					}, errors.New("error while creating metric acs")
				}
				return func() error {
					assert.Empty(t, deleteNode(met.ID), "error not expected in deleting metric type")
					return nil
				}, nil
			},
			checking: func(l *MetricRepository) (*v1.MetricACS, error) {
				actmet, err := l.GetMetricConfigACS(context.Background(), "acs", "scope1")
				if err != nil {
					return nil, err
				}

				return actmet, nil
			},
			want: &v1.MetricACS{
				Name:          "acs",
				EqType:        "zyx",
				AttributeName: "A2",
				Value:         "8",
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
			err = tt.l.UpdateMetricACS(tt.args.ctx, tt.args.met, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.UpdateMetricACS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := tt.checking(tt.l)
				if !assert.Empty(t, err, "not expecting error from checking") {
					return
				}
				compareMetricACS(t, "MetricRepository.UpdateMetricACS", tt.want, got)
			}
		})
	}
}
func compareMetricACSAll(t *testing.T, name string, act, exp []*v1.MetricACS) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetricACS(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetricACS(t *testing.T, name string, exp, act *v1.MetricACS) {
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
	assert.Equalf(t, exp.Value, act.Value, "%s.Value should be same", name)
}
