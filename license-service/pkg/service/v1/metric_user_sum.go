package v1

import (
	"context"
	"strconv"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// nolint: unparam
func (s *licenseServiceServer) computedLicensesUserSum(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	prodID, _ := input[ProdID].([]string)
	metrics, err := s.licenseRepo.ListMetricUserSum(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesUserSum - repo/ListMetricUserSum -", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric user sum")
	}
	ind := metricNameExistsUserSum(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUserSumComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), scope...)
	} else {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUserSumComputedLicenses(ctx, prodID, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesUserSum - ", zap.String("reason", err.Error()))
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric USS")

	}
	return computedLicenses, "Sum of users: " + strconv.FormatUint(computedDetails, 10), nil
}

func metricNameExistsUserSum(metrics []*repo.MetricUserSumStand, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
