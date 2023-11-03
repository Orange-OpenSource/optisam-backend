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
func (s *licenseServiceServer) computedLicensesUCS(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricUCS(ctx, scope...)
	prodID, _ := input[ProdID].([]string)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesUCS - repo/ListMetricUCS -", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric user sum")
	}
	ind := metricNameExistsUCS(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}
	mat := computedMetricUCS(metrics[ind])
	computedLicenses := uint64(0)
	computedDetails := uint64(0)
	if input[IsAgg].(bool) {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUCSComputedLicensesAgg(ctx, input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicenses, computedDetails, err = s.licenseRepo.MetricUCSComputedLicenses(ctx, prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Error("service/v1 - computedLicensesUCS - ", zap.String("reason", err.Error()))
		return 0, "", status.Error(codes.Internal, "cannot compute licenses for metric UCS")

	}
	return computedLicenses, "Sum of users: " + strconv.FormatUint(computedDetails, 10), nil
}

func computedMetricUCS(met *repo.MetricUCS) *repo.MetricUCSComputed {
	return &repo.MetricUCSComputed{
		Name:    met.Name,
		Profile: met.Profile,
	}
}

func metricNameExistsUCS(metrics []*repo.MetricUCS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
