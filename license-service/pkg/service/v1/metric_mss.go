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

func (s *licenseServiceServer) computedLicensesMSS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricMSS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Sugar().Errorw("computedLicensesMSS - Error while getting MSS Metrics",
			"status", codes.Internal,
			"error", err.Error(),
			"scope", scope,
		)
		return 0, status.Error(codes.Internal, "cannot fetch metric MSS")

	}
	ind := 0
	if ind = metricNameExistsMSS(metrics, input[MetricName].(string)); ind == -1 {
		logger.Log.Sugar().Errorw("computedLicensesMSS - Error metric name not exists from MSS metrics of scopes",
			"status", codes.Internal,
			"error", errors.New("metric name doesnot exists"),
			"scope", scope,
			"metricName", input[MetricName].(string),
			"metrics", metrics,
		)
		return 0, status.Error(codes.Internal, "metric name doesnot exists")
	}
	childEquipments := getChildEquipmentsByParentType(metrics[ind].Reference, eqTypes)

	mat := &repo.MetricMSSComputed{
		Name:          input[MetricName].(string),
		BaseType:      childEquipments,
		ReferenceType: metrics[ind].Reference,
		NumCoresAttr:  metrics[ind].Core,
		NumCPUAttr:    metrics[ind].CPU,
		IsSA:          input[IsSa].(bool),
	}

	computedLicenses := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, err = s.licenseRepo.MetricMSSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, err = s.licenseRepo.MetricMSSComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Sugar().Errorw("computedLicensesMSS - Error while computing licences for MSS Metric",
			"status", codes.Internal,
			"error", err.Error(),
			"scope", scope,
			"requestData", input,
		)
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric MSS")
	}

	return computedLicenses, nil
}

func metricNameExistsMSS(metrics []*repo.MetricMSS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.MetricName, name) {
			return i
		}
	}
	return -1
}
