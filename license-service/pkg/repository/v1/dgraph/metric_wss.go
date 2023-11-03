package dgraph

import (
	"context"
	"encoding/json"
	"errors"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
	"google.golang.org/grpc/codes"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
)

// Metadata for injectors
type metricWSS struct {
	ID         string `json:"uid"`
	MetricType string `json:"metric.type"`
	MetricName string `json:"metric.name"`
	Reference  string `json:"metric.wss.reference"`
	Core       string `json:"metric.wss.core"`
	CPU        string `json:"metric.wss.cpu"`
}

// ListMetricWSS implements Licence ListMetricWSS function
func (l *LicenseRepository) ListMetricWSS(ctx context.Context, scopes ...string) ([]*v1.MetricWSS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricWindowsServerStandard, scopes...)
	if err != nil {
		errorParams := map[string]interface{}{
			"status": codes.Internal,
			"error":  err.Error(),
			"scope":  scopes,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD - Error while getting WSD Metric", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS - Error while getting WSS Metric",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// )
		return nil, err
	}
	type Resp struct {
		Data []*metricWSS
	}
	var dataWss Resp
	if err := json.Unmarshal(respJSON, &dataWss); err != nil {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    err.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSS - Unmarshal failed", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS - Unmarshal failed",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, errors.New("cannot Unmarshal")
	}
	if len(dataWss.Data) == 0 {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    v1.ErrNoData.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSS -"+v1.ErrNoData.Error(), errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS -"+v1.ErrNoData.Error(),
		// 	"error", v1.ErrNoData.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricWSSAll(dataWss.Data)
}

func converMetricToModelMetricWSSAll(mts []*metricWSS) ([]*v1.MetricWSS, error) {
	mats := make([]*v1.MetricWSS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricWSS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricWSS(wssMetric *metricWSS) (*v1.MetricWSS, error) {

	if len(wssMetric.Reference) == 0 {
		errorParams := map[string]interface{}{
			"status":    codes.Internal,
			"error":     v1.ErrNoData.Error(),
			"Reference": wssMetric.Reference,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSS -converMetricToModelMetric - Reference not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS  - converMetricToModelMetric - Reference not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"Reference", wssMetric.Reference,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - Reference not found")
	}

	if len(wssMetric.Core) == 0 {
		errorParams := map[string]interface{}{
			"status":     codes.Internal,
			"error":      v1.ErrNoData.Error(),
			"AttrNumCPU": wssMetric.Core,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSS - converMetricToModelMetric AttrNumCPU not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS  - converMetricToModelMetric AttrNumCPU not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCPU", wssMetric.Core,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCPU not found")
	}

	if len(wssMetric.CPU) == 0 {
		errorParams := map[string]interface{}{
			"status":       codes.Internal,
			"error":        v1.ErrNoData.Error(),
			"AttrNumCores": wssMetric.CPU,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSS - converMetricToModelMetric AttrNumCores not found", errorParams)

		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSS  - converMetricToModelMetric AttrNumCores not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCores", wssMetric.CPU,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCores not found")
	}

	return &v1.MetricWSS{
		ID:         wssMetric.ID,
		MetricName: wssMetric.MetricName,
		MetricType: wssMetric.MetricType,
		Reference:  wssMetric.Reference,
		Core:       wssMetric.Core,
		CPU:        wssMetric.CPU,
	}, nil
}
