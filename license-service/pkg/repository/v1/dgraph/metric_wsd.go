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
type metricWSD struct {
	ID         string `json:"uid"`
	MetricType string `json:"metric.type"`
	MetricName string `json:"metric.name"`
	Reference  string `json:"metric.wsd.reference"`
	Core       string `json:"metric.wsd.core"`
	CPU        string `json:"metric.wsd.cpu"`
}

// ListMetricWSD implements Licence ListMetricWSD function
func (l *LicenseRepository) ListMetricWSD(ctx context.Context, scopes ...string) ([]*v1.MetricWSD, error) {
	respJSON, err := l.listMetricWithMetricType(ctx, v1.MetricWindowsServerDataCenter, scopes...)
	if err != nil {
		errorParams := map[string]interface{}{
			"status": codes.Internal,
			"error":  err.Error(),
			"scope":  scopes,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD - Error while getting WSD Metric", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD - Error while getting WSD Metric",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// )
		return nil, err
	}
	type Resp struct {
		Data []*metricWSD
	}
	var dataWSD Resp
	if err := json.Unmarshal(respJSON, &dataWSD); err != nil {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    err.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD - Unmarshal failed", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD - Unmarshal failed",
		// 	"error", err.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, errors.New("cannot Unmarshal")
	}
	if len(dataWSD.Data) == 0 {
		errorParams := map[string]interface{}{
			"status":   codes.Internal,
			"error":    v1.ErrNoData.Error(),
			"response": respJSON,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD -"+v1.ErrNoData.Error(), errorParams)

		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD -"+v1.ErrNoData.Error(),
		// 	"error", v1.ErrNoData.Error(),
		// 	"scope", scopes,
		// 	"response", respJSON,
		// )
		return nil, v1.ErrNoData
	}
	return converMetricToModelMetricWSDAll(dataWSD.Data)
}

func converMetricToModelMetricWSDAll(mts []*metricWSD) ([]*v1.MetricWSD, error) {
	mats := make([]*v1.MetricWSD, len(mts))
	for i := range mts {
		m, err := converMetricToModelMetricWSD(mts[i])
		if err != nil {
			return nil, err
		}
		mats[i] = m
	}

	return mats, nil
}

func converMetricToModelMetricWSD(wsdMat *metricWSD) (*v1.MetricWSD, error) {

	if len(wsdMat.Reference) == 0 {
		errorParams := map[string]interface{}{
			"status":    codes.Internal,
			"error":     v1.ErrNoData.Error(),
			"Reference": wsdMat.Reference,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD -converMetricToModelMetric - Reference not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD  - converMetricToModelMetric - Reference not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"Reference", wsdMat.Reference,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AtrrCoreFactor not found")
	}

	if len(wsdMat.Core) == 0 {
		errorParams := map[string]interface{}{
			"status":     codes.Internal,
			"error":      v1.ErrNoData.Error(),
			"AttrNumCPU": wsdMat.Core,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD - converMetricToModelMetric AttrNumCPU not found", errorParams)
		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD  - converMetricToModelMetric AttrNumCPU not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCPU", wsdMat.Core,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCPU not found")
	}

	if len(wsdMat.CPU) == 0 {
		errorParams := map[string]interface{}{
			"status":       codes.Internal,
			"error":        v1.ErrNoData.Error(),
			"AttrNumCores": wsdMat.CPU,
		}
		helper.CustomErrorHandle("Errorw", "dgraph/ListMetricWSD - converMetricToModelMetric AttrNumCores not found", errorParams)

		// logger.Log.Sugar().Errorw("dgraph/ListMetricWSD  - converMetricToModelMetric AttrNumCores not found",
		// 	"error", v1.ErrNoData.Error(),
		// 	"AttrNumCores", wsdMat.CPU,
		// )
		return nil, errors.New("dgraph converMetricToModelMetric - AttrNumCores not found")
	}

	return &v1.MetricWSD{
		ID:         wsdMat.ID,
		MetricName: wsdMat.MetricName,
		MetricType: wsdMat.MetricType,
		Reference:  wsdMat.Reference,
		Core:       wsdMat.Core,
		CPU:        wsdMat.CPU,
	}, nil
}
