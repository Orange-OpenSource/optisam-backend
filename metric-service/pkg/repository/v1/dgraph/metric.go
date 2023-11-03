package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/metric-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	dgo "github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"go.uber.org/zap"
)

// ListMetricTypeInfo implements Licence ListMetricTypeInfo function
func (l *MetricRepository) ListMetricTypeInfo(ctx context.Context, scopetype v1.ScopeType, scope string, flag bool) ([]*v1.MetricTypeInfo, error) {
	return scopetype.ListMetricTypes(flag), nil
}

// DropMetrics deletes all metrics in a scope
func (l *MetricRepository) DropMetrics(ctx context.Context, scope string) error {
	query := `query {
		 var(func: eq(type_name,metric)) @filter(eq(scopes,` + scope + `)){
			 metricId as uid
		}
		`
	delete := `
			uid(metricId) * * .
	`
	set := `
			uid(metricId) <Recycle> "true" .

	`
	query += `
	}`
	muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
	logger.Log.Info(query)
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{muDelete},
		CommitNow: true,
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, err := l.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DropMetrics - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf(" DropMetrics - cannot complete query transaction")
	}
	return nil
}

// ListMetrices implements Licence ListMetrices function
func (l *MetricRepository) ListMetrices(ctx context.Context, scope string) ([]*v1.MetricInfo, error) {

	q := `   {
		Metrics(func:eq(type_name,"metric"))@filter(eq(scopes,` + scope + `)){
			   ID  : uid
			   Name: metric.name
			   Type: metric.type
			   Default : metric.default
		   }
		}
		  `

	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("listMetrices - cannot complete query transaction")
	}

	type Data struct {
		Metrics []*v1.MetricInfo
	}
	var metricList Data
	if err := json.Unmarshal(resp.GetJson(), &metricList); err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("listMetrices - cannot unmarshal Json object")
	}

	return metricList.Metrics, nil
}

// MetricRepository for Dgraph
type MetricRepository struct {
	dg *dgo.Dgraph
	mu sync.Mutex
}

// NewMetricRepository creates new Repository
func NewMetricRepository(dg *dgo.Dgraph) *MetricRepository {
	return &MetricRepository{
		dg: dg,
	}
}

// NewMetricRepositoryWithTemplates creates new Repository with templates
func NewMetricRepositoryWithTemplates(dg *dgo.Dgraph) (*MetricRepository, error) {
	return NewMetricRepository(dg), nil
}

func (l *MetricRepository) MetricInfoWithAcqAndAgg(ctx context.Context, metricName, scope string) (*v1.MetricInfoFull, error) {
	q := `{ 
			Metric(func:eq(metric.name,"` + metricName + `")) @filter(eq(scopes,"` + scope + `")){
				ID  : uid
				Name: metric.name
				Type: metric.type
				Default: metric.default
		 	}
   			var(func:eq(aggregatedRights.metric,["` + metricName + `"]))@filter(eq(scopes,"` + scope + `")){
	 			tagg as count(aggregatedRights.SKU)
		 	}
	   		var(func:eq(acqRights.metric,["` + metricName + `"]))@filter(eq(scopes,"` + scope + `")){
				tacq as count(acqRights.SKU)
		 	}
   		  	AggregationCount(){
				TotalAggregations: sum(val(tagg))
 			}
			AcqrightCount(){
				TotalAcqRights: sum(val(tacq))
			}
   		}`
	resp, err := l.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Debug("GetMetric - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("getMetric - cannot complete query transaction")
	}
	type metricAggcount struct {
		TotalAggregations float64
	}
	type metricAcqCount struct {
		TotalAcqRights float64
	}
	type Data struct {
		Metric           []*v1.MetricInfo
		AggregationCount []*metricAggcount
		AcqrightCount    []*metricAcqCount
	}
	var metricInfo Data
	if err := json.Unmarshal(resp.GetJson(), &metricInfo); err != nil {
		logger.Log.Debug("GetMetric - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("getMetric - cannot unmarshal Json object")
	}
	retMetric := &v1.MetricInfoFull{}
	if len(metricInfo.Metric) != 0 {
		retMetric.ID = metricInfo.Metric[0].ID
		retMetric.Name = metricInfo.Metric[0].Name
		retMetric.Type = metricInfo.Metric[0].Type
		retMetric.Default = metricInfo.Metric[0].Default
	}
	if len(metricInfo.AggregationCount) != 0 {
		retMetric.TotalAggregations = int32(metricInfo.AggregationCount[0].TotalAggregations)
	}
	if len(metricInfo.AcqrightCount) != 0 {
		retMetric.TotalAcqRights = int32(metricInfo.AcqrightCount[0].TotalAcqRights)
	}
	return retMetric, nil
}

func (l *MetricRepository) DeleteMetric(ctx context.Context, metricName, scope string) error {
	query := `query {
		var(func: eq(metric.name,"` + metricName + `"))@filter(eq(scopes,"` + scope + `")){
			metric as metric.name
		}
		`
	delete := `
			uid(metric) * * .
	`
	set := `
			uid(metric) <Recycle> "true" .
	`
	query += `
	}`
	muDelete := &api.Mutation{DelNquads: []byte(delete), SetNquads: []byte(set)}
	req := &api.Request{
		Query:     query,
		Mutations: []*api.Mutation{muDelete},
		CommitNow: true,
	}
	if _, err := l.dg.NewTxn().Do(ctx, req); err != nil {
		logger.Log.Error("DeleteMetric - ", zap.String("reason", err.Error()), zap.String("query", query))
		return fmt.Errorf("deleteMetric - cannot complete query transaction")
	}
	return nil
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
		return nil, fmt.Errorf("cannot get metrics of %s", metType.String())
	}
	return resp.Json, nil
}
