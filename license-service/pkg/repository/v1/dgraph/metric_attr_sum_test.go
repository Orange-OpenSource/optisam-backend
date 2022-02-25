package dgraph

import (
	"context"
	"errors"
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicenseRepository_ListMetricAttrSum(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes string
	}
	tests := []struct {
		name    string
		l       *LicenseRepository
		args    args
		setup   func(*LicenseRepository) ([]*v1.MetricAttrSumStand, func() error, error)
		want    []*v1.MetricAttrSumStand
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewLicenseRepository(dgClient),
			args: args{
				ctx:    context.Background(),
				scopes: "scope1",
			},
			setup: func(l *LicenseRepository) (retMat []*v1.MetricAttrSumStand, cleanup func() error, retErr error) {
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
