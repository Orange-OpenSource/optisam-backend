package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"google.golang.org/grpc/codes"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
)

type metricMSE struct {
	ID        string `json:"uid"`
	Name      string `json:"metric.name"`
	Reference string `json:"metric.mse.reference"`
	Cores     string `json:"metric.mse.core"`
	CPU       string `json:"metric.mse.cpu"`
	Default   bool   `json:"metric.default"`
}

// ListMetricMSE implements Licence ListMetricMSE function
func (l *LicenseRepository) ListMetricMSE(ctx context.Context, scopes ...string) ([]*v1.MetricMSE, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricMicrosoftSqlEnterprise, scopes...)
	if err != nil {
		errorParams := map[string]interface{}{
			"status": codes.Internal,
			"error":  err.Error(),
			"scope":  scopes,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSE - Error while getting MSE Metric", errorParams)
		//logger.Log.Error("dgraph/ListMetricMSE - listMetricWithMetricType", zap.Error(err))
		return nil, err
	}
	type Resp struct {
		Data []*metricMSE
	}
	var dataMss Resp
	if err := json.Unmarshal(respJSON, &dataMss); err != nil {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    err.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSE - Unmarshal failed", errorParams)
		//logger.Log.Error("dgraph/ListMetricSPS - Unmarshal failed", zap.Error(err))
		return nil, errors.New("cannot Unmarshal")
	}
	if len(dataMss.Data) == 0 {
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricAllMSE(dataMss.Data)
}

func converMetricToModelMetricAllMSE(mts []*metricMSE) ([]*v1.MetricMSE, error) {
	mats := make([]*v1.MetricMSE, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricMSE(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricMSE(mse *metricMSE) (*v1.MetricMSE, error) {

	return &v1.MetricMSE{
		ID:        mse.ID,
		Name:      mse.Name,
		Reference: mse.Reference,
		Core:      mse.Cores,
		CPU:       mse.CPU,
		Default:   mse.Default,
	}, nil
}
