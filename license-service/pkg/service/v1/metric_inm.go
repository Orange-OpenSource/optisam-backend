package v1

import (
	"context"
	"strconv"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *licenseServiceServer) computedLicensesINM(ctx context.Context, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricINM(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesINM", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric INM")
	}
	ind := metricNameExistsINM(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}

	mat := computedMetricINM(metrics[ind])
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricINMComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricINMComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesINM - ", zap.String("reason", err.Error()))
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric INM")

	}
	return computedLicenses, "Total instances: " + strconv.FormatUint(computedDetails, 10), nil
}

func computedMetricINM(met *repo.MetricINM) *repo.MetricINMComputed {
	return &repo.MetricINMComputed{
		Name:        met.Name,
		Coefficient: met.Coefficient,
	}
}

func metricNameExistsINM(metrics []*repo.MetricINM, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
