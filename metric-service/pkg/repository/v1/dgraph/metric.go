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
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"go.uber.org/zap"
)

type predMetric string

// String implements string.Stringer
func (p predMetric) String() string {
	return string(p)
}

const (
	predMetricName predMetric = "metric.name"
)

// ListMetricTypeInfo implements Licence ListMetricTypeInfo function
func (l *MetricRepository) ListMetricTypeInfo(ctx context.Context, scopes string) ([]*v1.MetricTypeInfo, error) {
	return v1.MetricTypes, nil
}

// ListMetrices implements Licence ListMetrices function
func (l *MetricRepository) ListMetrices(ctx context.Context, scope string) ([]*v1.MetricInfo, error) {

	q := `   {
             Metrics(func:eq(type_name,"metric"))@filter(eq(scopes,` + scope + `)){
			   ID  : uid
			   Name: metric.name
			   Type: metric.type
		   }
		}
		  `

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListMetrices - cannot complete query transaction")
	}

	type Data struct {
		Metrics []*v1.MetricInfo
	}
	var metricList Data
	if err := json.Unmarshal(resp.GetJson(), &metricList); err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("ListMetrices - cannot unmarshal Json object")
	}

	return metricList.Metrics, nil
}

//MetricRepository for Dgraph
type MetricRepository struct {
	dg *dgo.Dgraph
}

//NewMetricRepository creates new Repository
func NewMetricRepository(dg *dgo.Dgraph) *MetricRepository {
	return &MetricRepository{
		dg: dg,
	}
}

//NewMetricRepositoryWithTemplates creates new Repository with templates
func NewMetricRepositoryWithTemplates(dg *dgo.Dgraph) (*MetricRepository, error) {
	return NewMetricRepository(dg), nil
}

func (l *MetricRepository) listMetricWithMetricType(ctx context.Context, metType v1.MetricType, scope string) (json.RawMessage, error) {
	q := `{
		Data(func: eq(metric.type,` + metType.String() + `)) @filter(eq(scopes,` + scope + `)){
		 uid
		 expand(_all_){
		  uid
		} 
		}
	  }`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/listMetricWithMetricType - query failed", zap.Error(err), zap.String("query", q))
		return nil, errors.New(fmt.Sprintf("cannot get metrices of %s", metType.String()))
	}
	return resp.Json, nil
}

func scopesNquad(scp []string, blankID string) []*api.NQuad {
	nquads := []*api.NQuad{}
	for _, sID := range scp {
		nquads = append(nquads, scopeNquad(sID, blankID)...)
	}
	return nquads
}

func scopeNquad(scope, uid string) []*api.NQuad {
	return []*api.NQuad{
		&api.NQuad{
			Subject:     uid,
			Predicate:   "scopes",
			ObjectValue: stringObjectValue(scope),
		},
	}
}
