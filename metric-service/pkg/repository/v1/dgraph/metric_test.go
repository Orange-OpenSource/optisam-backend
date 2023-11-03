package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
)

func Test_DropMetrics(t *testing.T) {

	tests := []struct {
		name    string
		l       *MetricRepository
		input   string
		setup   func() (func() error, error)
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "SuccessCase",
			ctx:     context.Background(),
			input:   "s1",
			wantErr: false,
			setup: func() (func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     blankID("met1"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("s1"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type1"),
						},
						{
							Subject:     blankID("met1"),
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

				return func() error {
					return deleteNodes(metID1)
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l = NewMetricRepository(dgClient)
			cleanup, err := tt.setup()
			if !assert.Empty(t, err, "not expecting error from setup") {
				return
			}
			defer func() {
				assert.Empty(t, cleanup(), "not expecting error in setup")
			}()
			err = tt.l.DropMetrics(tt.ctx, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.DropMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
func TestMetricRepository_ListMetrices(t *testing.T) {

	type args struct {
		ctx    context.Context
		scopes string
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
				ctx:    context.Background(),
				scopes: "Scope1",
			},
			setup: func() (func() error, error) {
				// TODO create two nodes for metrics
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     blankID("met1"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type1"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type2"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
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
				{
					Name: "Oracle type1",
					Type: v1.MetricOPSOracleProcessorStandard,
				},
				{
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

func TestMetricRepository_DeleteMetric(t *testing.T) {
	type args struct {
		ctx        context.Context
		metricName string
		scope      string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (func() error, error)
		verify  func(*MetricRepository)
		wantErr bool
	}{
		{name: "SUCCESS",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:        context.Background(),
				metricName: "Oracle type1",
				scope:      "Scope1",
			},
			setup: func() (func() error, error) {
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     blankID("met1"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type1"),
						},
						{
							Subject:     blankID("met1"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type2"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     blankID("met2"),
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
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
			verify: func(l *MetricRepository) {
				metrics, err := l.ListMetrices(context.Background(), "Scope1")
				if err != nil {
					t.Errorf("MetricRepository.DeleteMetric(), verify() - Error in getting metrices error = %v", err)
				}
				wantMet := []*v1.MetricInfo{
					{
						Name: "Oracle type2",
						Type: v1.MetricOPSOracleProcessorStandard,
					},
				}
				compareMetricsAll(t, "MetricRepository.DeleteMetric()", wantMet, metrics)
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
			if err := tt.l.DeleteMetric(tt.args.ctx, tt.args.metricName, tt.args.scope); (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.DeleteMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMetricRepository_MetricInfoWithAcqAndAgg(t *testing.T) {
	type args struct {
		ctx        context.Context
		metricName string
		scope      string
	}
	tests := []struct {
		name    string
		l       *MetricRepository
		args    args
		setup   func() (func() error, error)
		want    *v1.MetricInfoFull
		wantErr bool
	}{
		{name: "SUCCESS - linking present",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:        context.Background(),
				metricName: "OracleMet",
				scope:      "Scope1",
			},
			setup: func() (func() error, error) {
				met1 := blankID("met1")
				met2 := blankID("met2")
				agg1 := blankID("agg1")
				acq1 := blankID("acq1")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     met1,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     met1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     met1,
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("OracleMet"),
						},
						{
							Subject:     met1,
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     met2,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     met2,
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type2"),
						},
						{
							Subject:     met2,
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     met2,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     agg1,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("aggregated_rights"),
						},
						{
							Subject:     agg1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     agg1,
							Predicate:   "aggregatedRights.SKU",
							ObjectValue: stringObjectValue("Aggregated right sku"),
						},
						{
							Subject:     agg1,
							Predicate:   "aggregatedRights.metric",
							ObjectValue: stringObjectValue("OracleMet"),
						},
						{
							Subject:     acq1,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("acqRights"),
						},
						{
							Subject:     acq1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     acq1,
							Predicate:   "acqRights.SKU",
							ObjectValue: stringObjectValue("Acq1 SKU"),
						},
						{
							Subject:     acq1,
							Predicate:   "acqRights.metric",
							ObjectValue: stringObjectValue("OracleMet"),
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
				AggID, ok := assigned.Uids["agg1"]
				if !ok {
					return nil, errors.New("cannot find aggregatedright1 id after mutation in setup")
				}
				AcqID, ok := assigned.Uids["acq1"]
				if !ok {
					return nil, errors.New("cannot find acqrights1 id after mutation in setup")
				}
				return func() error {
					return deleteNodes(metID1, metID2, AggID, AcqID)
				}, nil
			},
			want: &v1.MetricInfoFull{
				Name:              "OracleMet",
				Type:              v1.MetricOPSOracleProcessorStandard,
				TotalAggregations: 1,
				TotalAcqRights:    1,
			},
		},
		{name: "SUCCESS - no linking present",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:        context.Background(),
				metricName: "OracleMet1",
				scope:      "Scope1",
			},
			setup: func() (func() error, error) {
				met1 := blankID("met1")
				met2 := blankID("met2")
				mu := &api.Mutation{
					CommitNow: true,
					Set: []*api.NQuad{
						{
							Subject:     met1,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     met1,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
						},
						{
							Subject:     met1,
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("OracleMet1"),
						},
						{
							Subject:     met1,
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     met2,
							Predicate:   "type_name",
							ObjectValue: stringObjectValue("metric"),
						},
						{
							Subject:     met2,
							Predicate:   "metric.name",
							ObjectValue: stringObjectValue("Oracle type2"),
						},
						{
							Subject:     met2,
							Predicate:   "metric.type",
							ObjectValue: stringObjectValue("oracle.processor.standard"),
						},
						{
							Subject:     met2,
							Predicate:   "scopes",
							ObjectValue: stringObjectValue("Scope1"),
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
			want: &v1.MetricInfoFull{
				Name:              "OracleMet1",
				Type:              v1.MetricOPSOracleProcessorStandard,
				TotalAggregations: 0,
				TotalAcqRights:    0,
			},
		},
		{name: "SUCCESS - metric does not exist",
			l: NewMetricRepository(dgClient),
			args: args{
				ctx:        context.Background(),
				metricName: "OracleMet3",
				scope:      "Scope1",
			},
			setup: func() (func() error, error) {
				return func() error {
					return nil
				}, nil
			},
			want: &v1.MetricInfoFull{},
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
			got, err := tt.l.MetricInfoWithAcqAndAgg(tt.args.ctx, tt.args.metricName, tt.args.scope)
			if (err != nil) != tt.wantErr {
				t.Errorf("MetricRepository.MetricInfoWithAcqAndAgg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				compareMetricInfoFull(t, "MetricRepository.MetricInfoWithAcqAndAgg() = %v, want %v", tt.want, got)
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

func compareMetricInfoFull(t *testing.T, name string, exp, act *v1.MetricInfoFull) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metric is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Name should be same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Type should be same", name)
	assert.Equalf(t, exp.TotalAggregations, act.TotalAggregations, "%s.TotalAggregations should be same", name)
	assert.Equalf(t, exp.TotalAcqRights, act.TotalAcqRights, "%s.TotalAcqRights should be same", name)
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
