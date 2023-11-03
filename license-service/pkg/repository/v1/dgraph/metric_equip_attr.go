package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"go.uber.org/zap"
)

type metricEquipAttr struct {
	ID          string  `json:"uid"`
	Name        string  `json:"metric.name"`
	EqType      string  `json:"metric.equip_attr.equipment_type"`
	Environment string  `json:"metric.equip_attr.environment"`
	AttrName    string  `json:"metric.equip_attr.attr_name"`
	Value       float64 `json:"metric.equip_attr.value"`
}

// ListMetricEquipAttr implements Licence ListMetricEquipAttr function
func (l *LicenseRepository) ListMetricEquipAttr(ctx context.Context, scopes ...string) ([]*v1.MetricEquipAttrStand, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricEquipAttrStandard, scopes...)
	if err != nil {
		logger.Log.Error("dgraph/ListMetricEquipAttr - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricEquipAttr
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		logger.Log.Error("dgraph/ListMetricEquipAttr - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllEquipAttr(data.Data)
}

func converMetricToModelMetricAllEquipAttr(mets []*metricEquipAttr) ([]*v1.MetricEquipAttrStand, error) {
	modelMets := make([]*v1.MetricEquipAttrStand, len(mets))
	for i := range mets {
		m, err := converMetricToModelMetricEquipAttr(mets[i])
		if err != nil {
			return nil, err
		}
		modelMets[i] = m
	}
	return modelMets, nil
}

func converMetricToModelMetricEquipAttr(m *metricEquipAttr) (*v1.MetricEquipAttrStand, error) {
	if len(m.EqType) == 0 {
		return nil, errors.New("dgraph converMetricToModelMetricEquipAttr - equipment type not found")
	}
	if m.AttrName == "" {
		return nil, errors.New("dgraph converMetricToModelMetricEquipAttr - Attribute not found")
	}
	return &v1.MetricEquipAttrStand{
		ID:            m.ID,
		Name:          m.Name,
		EqType:        m.EqType,
		Environment:   m.Environment,
		AttributeName: m.AttrName,
		Value:         m.Value,
	}, nil
}
