package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricSS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	ReferenceValue int32  `json:"metric.static.reference_value"`
}

// ListMetricSS implements Licence ListMetricSS function
func (l *LicenseRepository) ListMetricSS(ctx context.Context, scopes ...string) ([]*v1.MetricSS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricStaticStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricSS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricSS
	}
	// log.Println("SS metrics ", string(respJSON))
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricSS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllSS(data.Data)
}

func converMetricToModelMetricAllSS(mts []*metricSS) ([]*v1.MetricSS, error) {
	mats := make([]*v1.MetricSS, len(mts))
	for i := range mts {
		m := converMetricToModelMetricSS(mts[i])
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricSS(m *metricSS) *v1.MetricSS {

	return &v1.MetricSS{
		ID:             m.ID,
		Name:           m.Name,
		ReferenceValue: m.ReferenceValue,
	}
}
