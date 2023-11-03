package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricUCS struct {
	ID      string `json:"uid"`
	Name    string `json:"metric.name"`
	Profile string `json:"metric.user_concurrent.profile"`
}

// ListMetricUCS implements Licence ListMetricUCS function
func (l *LicenseRepository) ListMetricUCS(ctx context.Context, scopes ...string) ([]*v1.MetricUCS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricUserConcurentStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricUCS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricUCS
	}
	// log.Println("UCS metrics ", string(respJSON))
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricUCS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllUCS(data.Data)
}

func converMetricToModelMetricAllUCS(mts []*metricUCS) ([]*v1.MetricUCS, error) {
	mats := make([]*v1.MetricUCS, len(mts))
	for i := range mts {
		m := converMetricToModelMetricUCS(mts[i])
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricUCS(m *metricUCS) *v1.MetricUCS {

	return &v1.MetricUCS{
		ID:      m.ID,
		Name:    m.Name,
		Profile: m.Profile,
	}
}
