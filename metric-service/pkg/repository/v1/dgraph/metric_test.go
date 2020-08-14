// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"strings"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func TestMetricRepository_ListMetrices(t *testing.T) {
	type args struct {
		ctx    context.Context
		scopes []string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (func() error, error)
		want    []*v1.MetricInfo
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
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

			want: []*v1.MetricInfo{
				&v1.MetricInfo{
					Name: "Oracle type1",
					Type: v1.MetricOPSOracleProcessorStandard,
				},
				&v1.MetricInfo{
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
				t.Errorf("MetricRepository.ListMetrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				compareMetricsAll(t, "ListMetrices", tt.want, got)
			}
		})
	}
}

func compareMetricsAll(t *testing.T, name string, exp, act []*v1.MetricInfo) {
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

func getMetricByName(name string, met []*v1.MetricInfo) int {
	for i := range met {
		if name == met[i].Name {
			return i
		}
	}
	return -1
}

func compareMetrics(t *testing.T, name string, exp, act *v1.MetricInfo) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metric is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type should be same", name)
}

type SchemaNode struct {
	Predicate string   `json:"predicate,omitempty"`
	Type      string   `json:"type,omitempty"`
	Index     bool     `json:"index,omitempty"`
	Tokenizer []string `json:"tokenizer,omitempty"`
	Reverse   bool     `json:"reverse,omitempty"`
	Count     bool     `json:"count,omitempty"`
	List      bool     `json:"list,omitempty"`
	Upsert    bool     `json:"upsert,omitempty"`
	Lang      bool     `json:"lang,omitempty"`
}

func querySchema(predicates ...string) ([]*SchemaNode, error) {
	if len(predicates) == 0 {
		return nil, nil
	}
	q := `
		schema (pred: [` + strings.Join(predicates, ",") + `]) {
		type
		index
		reverse
		tokenizer
		list
		count
		upsert
		lang
	  }
	`
	//	fmt.Println(q)
	resp, err := dgClient.NewTxn().Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	type data struct {
		Schema []*SchemaNode
	}
	d := &data{}
	if err := json.Unmarshal(resp.Json, d); err != nil {
		return nil, err
	}

	return d.Schema, nil
}

func deleteNodes(ids ...string) error {

	for _, id := range ids {
		if err := deleteNode(id); err != nil {
			return err
		}
	}

	return nil
}

func deleteNode(id string) error {
	mu := &api.Mutation{
		CommitNow:  true,
		DeleteJson: []byte(`{"uid": "` + id + `"}`),
		// Del: []*api.NQuad{
		// 	&api.NQuad{
		// 		Subject:     id,
		// 		Predicate:   "*",
		// 		ObjectValue: deleteAll,
		// 	},
	}

	// delete all the data
	_, err := dgClient.NewTxn().Mutate(context.Background(), mu)
	if err != nil {
		return err
	}

	return nil
}

func compareSchemaNodeAll(t *testing.T, name string, exp []*SchemaNode, act []*SchemaNode) {
	if !assert.Lenf(t, act, len(exp), "expected number of elements are: %d", len(exp)) {
		return
	}

	for i := range exp {
		actIdx := indexForPredicte(exp[i].Predicate, act)
		if assert.NotEqualf(t, -1, "%s.Predicate is not found in expected nodes", fmt.Sprintf("%s[%d]", name, i)) {

		}
		compareSchemaNode(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[actIdx])
	}
}

func indexForPredicte(predicate string, schemas []*SchemaNode) int {
	for i := range schemas {
		if schemas[i].Predicate == predicate {
			return i
		}
	}
	return -1
}

func compareSchemaNode(t *testing.T, name string, exp *SchemaNode, act *SchemaNode) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Predicate, act.Predicate, "%s.Predicate are not same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type are not same", name)
	assert.Equalf(t, exp.Index, act.Index, "%s.Index are not same", name)
	assert.ElementsMatchf(t, exp.Tokenizer, act.Tokenizer, "%s.Tokenizer are not same", name)
	assert.Equalf(t, exp.Reverse, act.Reverse, "%s.Reverse are not same", name)
	assert.Equalf(t, exp.Count, act.Count, "%s.Count are not same", name)
	assert.Equalf(t, exp.List, act.List, "%s.List are not same", name)
	assert.Equalf(t, exp.Upsert, act.Upsert, "%s.Upsert are not same", name)
	assert.Equalf(t, exp.Lang, act.Lang, "%s.Lang are not same", name)
}
