package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricIPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.ips.base"`
	AttrNumCores   []*id  `json:"metric.ips.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.ips.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.ips.attr_core_factor"`
}

// ListMetricIPS implements Licence ListMetricIPS function
func (l *LicenseRepository) ListMetricIPS(ctx context.Context, scopes ...string) ([]*v1.MetricIPS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricIPSIbmPvuStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricIPS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricIPS
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		// fmt.Println(string(resp.Json))
		logger.Log.Error("dgraph/ListMetricIPS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllIPS(data.Data)
}

func converMetricToModelMetricAllIPS(mts []*metricIPS) ([]*v1.MetricIPS, error) {
	mats := make([]*v1.MetricIPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricIPS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricIPS(m *metricIPS) (*v1.MetricIPS, error) {

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - Base equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AttrNumCores not found")
	}

	if len(m.AttrNumCPU) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricIPS - AttrNumCPU not found")
	}

	return &v1.MetricIPS{
		ID:               m.ID,
		Name:             m.Name,
		BaseEqTypeID:     m.Base[0].ID,
		CoreFactorAttrID: m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:    m.AttrNumCores[0].ID,
		NumCPUAttrID:     m.AttrNumCPU[0].ID,
	}, nil
}
