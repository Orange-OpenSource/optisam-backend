package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/metric-service/pkg/repository/v1"
	"sync"

	"github.com/dgraph-io/dgo/v2"
	"go.uber.org/zap"
)

type ProductRepository struct {
	dg *dgo.Dgraph
	mu sync.Mutex
}

func NewProductRepository(dg *dgo.Dgraph) *ProductRepository {
	return &ProductRepository{
		dg: dg,
	}
}

// NewMetricRepositoryWithTemplates creates new Repository with templates
func NewProductRepositoryWithTemplates(dg *dgo.Dgraph) (*ProductRepository, error) {
	return NewProductRepository(dg), nil
}

func (p *ProductRepository) ListMetrices(ctx context.Context, scope string) error {

	q := `   {
             Metrics(func:eq(type_name,"metric"))@filter(eq(scopes,` + scope + `)){
			   ID  : uid
			   Name: metric.name
			   Type: metric.type
		   }
		}
		  `

	resp, err := p.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return errors.New("listMetrices - cannot complete query transaction")
	}
	type Data struct {
		Metrics []*v1.MetricInfo
	}
	var metricList Data
	if err := json.Unmarshal(resp.GetJson(), &metricList); err != nil {
		logger.Log.Error("ListMetrices - ", zap.String("reason", err.Error()), zap.String("query", q))
		return errors.New("listMetrices - cannot unmarshal Json object")
	}
	fmt.Println(metricList)
	return nil
}
