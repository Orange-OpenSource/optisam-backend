package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

type metricINM struct {
	ID          string `json:"uid"`
	Name        string `json:"metric.name"`
	Coefficient int32  `json:"metric.instancenumber.coefficient"`
}

// ListMetricINM implements Licence ListMetricINM function
func (l *LicenseRepository) ListMetricINM(ctx context.Context, scopes ...string) ([]*v1.MetricINM, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricInstanceNumberStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricINM - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricINM
	}
	// log.Println("INM metrics ", string(respJSON))
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricINM - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllINM(data.Data)
}

func converMetricToModelMetricAllINM(mts []*metricINM) ([]*v1.MetricINM, error) {
	mats := make([]*v1.MetricINM, len(mts))
	for i := range mts {
		m := converMetricToModelMetricINM(mts[i])
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricINM(m *metricINM) *v1.MetricINM {

	return &v1.MetricINM{
		ID:          m.ID,
		Name:        m.Name,
		Coefficient: m.Coefficient,
	}
}
