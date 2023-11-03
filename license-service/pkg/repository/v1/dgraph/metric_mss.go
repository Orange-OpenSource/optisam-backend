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
type metricMSS struct {
	ID         string `json:"uid"`
	MetricType string `json:"metric.type"`
	MetricName string `json:"metric.name"`
	Reference  string `json:"metric.sql.standard.reference"`
	Core       string `json:"metric.sql.standard.core"`
	CPU        string `json:"metric.sql.standard.cpu"`
}

// ListMetricMSS implements Licence ListMetricMSS function
func (l *LicenseRepository) ListMetricMSS(ctx context.Context, scopes ...string) ([]*v1.MetricMSS, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricMicrosoftSqlStandard, scopes...)
	if err != nil {
		errorParams := map[string]interface{}{
			"status": codes.Internal,
			"error":  err.Error(),
			"scope":  scopes,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS - Error while getting MSS Metric", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS - Error while getting MSS Metric",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// )
		return nil, err
	}
	type Resp struct {
		Data []*metricMSS
	}
	var data Resp
	if err := json.Unmarshal(respJSON, &data); err != nil {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    err.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS - Unmarshal failed", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS - Unmarshal failed",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, errors.New("cannot Unmarshal")
	}
	if len(data.Data) == 0 {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    v1.ErrNoData.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS -"+v1.ErrNoData.Error(), errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS -"+v1.ErrNoData.Error(),
		// 	"error", v1.ErrNoData.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricMSSAll(data.Data)
}

func converMetricToModelMetricMSSAll(mts []*metricMSS) ([]*v1.MetricMSS, error) {
	mats := make([]*v1.MetricMSS, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricMSS(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricMSS(mssMetric *metricMSS) (*v1.MetricMSS, error) {

	if len(mssMetric.Reference) == 0 {
		errorParams := map[string]interface{}{
			"status":    codes.Internal,
			"error":     v1.ErrNoData.Error(),
			"Reference": mssMetric.Reference,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS -converMetricToModelMetric - Reference not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS  - converMetricToModelMetric - Reference not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"Reference", mssMetric.Reference,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AtrrCoreFactor not found")
	}

	if len(mssMetric.Core) == 0 {
		errorParams := map[string]interface{}{
			"status":     codes.Internal,
			"error":      v1.ErrNoData.Error(),
			"AttrNumCPU": mssMetric.Core,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS - converMetricToModelMetric AttrNumCPU not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS  - converMetricToModelMetric AttrNumCPU not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCPU", mssMetric.Core,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCPU not found")
	}

	if len(mssMetric.CPU) == 0 {
		errorParams := map[string]interface{}{
			"status":       codes.Internal,
			"error":        v1.ErrNoData.Error(),
			"AttrNumCores": mssMetric.CPU,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricMSS - converMetricToModelMetric AttrNumCores not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricMSS  - converMetricToModelMetric AttrNumCores not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCores", mssMetric.CPU,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCores not found")
	}

	return &v1.MetricMSS{
		ID:         mssMetric.ID,
		MetricName: mssMetric.MetricName,
		MetricType: mssMetric.MetricType,
		Reference:  mssMetric.Reference,
		Core:       mssMetric.Core,
		CPU:        mssMetric.CPU,
	}, nil
}
