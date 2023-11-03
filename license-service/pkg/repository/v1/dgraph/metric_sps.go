package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricSPS struct {
	ID             string `json:"uid"`
	Name           string `json:"metric.name"`
	Base           []*id  `json:"metric.sps.base"`
	AttrNumCores   []*id  `json:"metric.sps.attr_num_cores"`
	AttrNumCPU     []*id  `json:"metric.sps.attr_num_cpu"`
	AtrrCoreFactor []*id  `json:"metric.sps.attr_core_factor"`
}

// ListMetricSPS implements Licence ListMetricSPS function
func (l *LicenseRepository) ListMetricSPS(ctx context.Context, scopes ...string) ([]*v1.MetricSPS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricSPSSagProcessorStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricSPS - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricSPS
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricSPS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllSPS(data.Data)
}

func converMetricToModelMetricAllSPS(mts []*metricSPS) ([]*v1.MetricSPS, error) {
	mats := make([]*v1.MetricSPS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricSPS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricSPS(m *metricSPS) (*v1.MetricSPS, error) {

	if len(m.Base) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - Base equipment level not found")
	}

	if len(m.AtrrCoreFactor) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - AtrrCoreFactor not found")
	}

	if len(m.AttrNumCores) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - AttrNumCores not found")
	}

	if len(m.AttrNumCPU) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricSPS - AttrNumCPU not found")
	}

	return &v1.MetricSPS{
		ID:               m.ID,
		Name:             m.Name,
		BaseEqTypeID:     m.Base[0].ID,
		CoreFactorAttrID: m.AtrrCoreFactor[0].ID,
		NumCoreAttrID:    m.AttrNumCores[0].ID,
		NumCPUAttrID:     m.AttrNumCPU[0].ID,
	}, nil
}
