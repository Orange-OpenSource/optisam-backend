package v1

import (
	"context"
	"errors"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesWSS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricWSS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Sugar().Errorw("computedLicensesWSS - Error while getting WSS Metrics",
			"status", codes.Internal,
			"error", err.Error(),
			"scope", scope,
		)
		return 0, status.Error(codes.Internal, "cannot fetch metric WSS")

	}
	ind := 0
	if ind = metricNameExistsWSS(metrics, input[MetricName].(string)); ind == -1 {
		logger.Log.Sugar().Errorw("computedLicensesWSS - Error metric name not exists from WSS metrics of scopes",
			"status", codes.Internal,
			"error", errors.New("metric name doesnot exists"),
			"scope", scope,
			"metricName", input[MetricName].(string),
			"metrics", metrics,
		)
		return 0, status.Error(codes.Internal, "metric name doesnot exists")
	}
	childEquipments := getChildEquipmentsByParentType(metrics[ind].Reference, eqTypes)

	mat := &repo.MetricWSSComputed{
		Name:          input[MetricName].(string),
		BaseType:      childEquipments,
		ReferenceType: metrics[ind].Reference,
		NumCoresAttr:  metrics[ind].Core,
		NumCPUAttr:    metrics[ind].CPU,
		IsSA:          input[IsSa].(bool),
	}

	computedLicenses := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, err = s.licenseRepo.MetricWSSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricWSSComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Sugar().Errorw("computedLicensesWSS - Error while computing licences for WSS Metric",
			"status", codes.Internal,
			"error", err.Error(),
			"scope", scope,
			"requestData", input,
		)
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric WSS")
	}

	return computedLicenses, nil
}

func metricNameExistsWSS(metrics []*repo.MetricWSS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.MetricName, name) {
			return i
		}
	}
	return -1
}
