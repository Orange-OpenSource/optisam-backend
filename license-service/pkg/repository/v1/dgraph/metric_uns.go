package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

type metricUNS struct {
	ID      string `json:"uid"`
	Name    string `json:"metric.name"`
	Profile string `json:"metric.user_nominative.profile"`
}

// ListMetricUNS implements Licence ListMetricUNS function
func (l *LicenseRepository) ListMetricUNS(ctx context.Context, scopes ...string) ([]*v1.MetricUNS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricUserNomStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricUNS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricUNS
	}
	// log.Println("UNS metrics ", string(respJSON))
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricUNS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllUNS(data.Data)
}

func converMetricToModelMetricAllUNS(mts []*metricUNS) ([]*v1.MetricUNS, error) {
	mats := make([]*v1.MetricUNS, len(mts))
	for i := range mts {
		m := converMetricToModelMetricUNS(mts[i])
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricUNS(m *metricUNS) *v1.MetricUNS {

	return &v1.MetricUNS{
		ID:      m.ID,
		Name:    m.Name,
		Profile: m.Profile,
	}
}
