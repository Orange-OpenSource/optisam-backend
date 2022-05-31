package v1

import (
	"context"

	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/strcomp"
	repo "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint: unparam
func (s *licenseServiceServer) computedLicensesSS(ctx context.Context, input map[string]interface{}) (uint64, string, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricSS(ctx, scope...)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Error("service/v1 computedLicensesSS", zap.Error(err))
		return 0, "", status.Error(codes.Internal, "cannot fetch metric SS")
	}
	ind := metricNameExistsSS(metrics, input[MetricName].(string))
	if ind == -1 {
		return 0, "", status.Error(codes.NotFound, "cannot find metric name")
	}

	return uint64(metrics[ind].ReferenceValue), "", nil
}

func metricNameExistsSS(metrics []*repo.MetricSS, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
