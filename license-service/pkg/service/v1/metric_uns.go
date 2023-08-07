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

// nolint: unparam
func (s *licenseServiceServer) computedLicensesUNS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricUNS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesUNS - repo/ListMetricUNS -", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric user sum")
	}
	ind := metricNameExistsUNS(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}
	mat := computedMetricUNS(metrics[ind])
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUNSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUNSComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesUNS - ", zap.String("reason", err.Error()))
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric USS")

	}
	return computedLicenses, "Sum of users: " + strconv.FormatUint(computedDetails, 10), nil
}

func computedMetricUNS(met *repo.MetricUNS) *repo.MetricUNSComputed {
	return &repo.MetricUNSComputed{
		Name:    met.Name,
		Profile: met.Profile,
	}
}

func metricNameExistsUNS(metrics []*repo.MetricUNS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
