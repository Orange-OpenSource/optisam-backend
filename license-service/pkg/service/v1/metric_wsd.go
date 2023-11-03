package v1

import (
	"context"
	"errors"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesWSD(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricWSD(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		errorParams := map[string]interface{}{
			"status": codes.Internal,
			"error":  err.Error(),
			"scope":  scope,
		}
		helper.CustomErrorHandle("Errorw", "computedLicensesWSD - Error while getting WSD Metrics", errorParams)
		//logger.Log.Sugar().Errorw("computedLicensesWSD - Error while getting WSD Metrics")
		return 0, status.Error(codes.Internal, "cannot fetch metric WSD")

	}
	ind := 0
	if ind = metricNameExistsWSD(metrics, input[MetricName].(string)); ind == -1 {
		errorParams := map[string]interface{}{
			"status":     codes.Internal,
			"error":      errors.New("metric name doesnot exists"),
			"scope":      scope,
			"metricName": input[MetricName].(string),
			"metrics":    metrics,
		}
		helper.CustomErrorHandle("Errorw", "computedLicensesWSD - Error metric name not exists from WSD metrics of scopes", errorParams)

		// logger.Log.Sugar().Errorw("computedLicensesWSD - Error metric name not exists from WSD metrics of scopes",
		// 	"status", codes.Internal,
		// 	"error", errors.New("metric name doesnot exists"),
		// 	"scope", scope,
		// 	"metricName", input[MetricName].(string),
		// 	"metrics", metrics,
		// )
		return 0, status.Error(codes.Internal, "metric name doesnot exists")
	}
	childEquipments := getChildEquipmentsByParentType(metrics[ind].Reference, eqTypes)

	mat := &repo.MetricWSDComputed{
		Name:          input[MetricName].(string),
		BaseType:      childEquipments,
		ReferenceType: metrics[ind].Reference,
		NumCoresAttr:  metrics[ind].Core,
		NumCPUAttr:    metrics[ind].CPU,
		IsSA:          input[IsSa].(bool),
	}

	computedLicenses := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, err = s.licenseRepo.MetricWSDComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricWSDComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		errorParams := map[string]interface{}{
			"status":      codes.Internal,
			"error":       errors.New("metric name doesnot exists"),
			"scope":       scope,
			"requestData": input,
		}
		helper.CustomErrorHandle("Errorw", "computedLicensesWSD - Error while computing licences for WSD Metric", errorParams)

		// logger.Log.Sugar().Errorw("computedLicensesWSD - Error while computing licences for WSD Metric",
		// 	"status", codes.Internal,
		// 	"error", err.Error(),
		// 	"scope", scope,
		// 	"requestData", input,
		// )
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric WSD")
	}

	return computedLicenses, nil
}

func metricNameExistsWSD(metrics []*repo.MetricWSD, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.MetricName, name) {
			return i
		}
	}
	return -1
}
