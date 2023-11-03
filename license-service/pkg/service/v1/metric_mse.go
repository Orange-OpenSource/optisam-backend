package v1

import (
	"context"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func metricNameExistsMSE(metrics []*repo.MetricMSE, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}

func (s *licenseServiceServer) computedLicensesMSE(ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (uint64, error) {
	scope, _ := input[SCOPES].([]string)
	metrics, err := s.licenseRepo.ListMetricMSE(ctx, scope...)
	prodID, _ := input[ProdID].([]string)
	if err != nil && err != repo.ErrNoData {
		logger.Log.Sugar().Errorf("service/v1 - computedLicensesMSE - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot fetch metric MSE")
	}
	ind := 0
	if ind = metricNameExistsMSE(metrics, input[MetricName].(string)); ind == -1 {
		return 0, status.Error(codes.Internal, "cannot find metric name")
	}
	mat := &repo.MetricMSEComputed{
		Reference: metrics[ind].Reference,
		Core:      metrics[ind].Core,
		CPU:       metrics[ind].CPU,
		Name:      metrics[ind].Name,
	}
	computedLicensesProd := uint64(0)
	if input[IsAgg].(bool) {
		computedLicensesProd, err = s.licenseRepo.MetricMSEComputedLicensesAgg(ctx, input[IsSa].(bool), input[ProdAggName].(string), input[MetricName].(string), mat, scope...)
	} else {
		computedLicensesProd, err = s.licenseRepo.MetricMSEComputedLicenses(ctx, input[IsSa].(bool), prodID, mat, scope...)
	}
	if err != nil {
		logger.Log.Sugar().Errorf("service/v1 - computedLicensesMSE - ", zap.String("reason", err.Error()))
		return 0, status.Error(codes.Internal, "cannot compute licenses for metric MSE")

	}
	return computedLicensesProd, nil

}

// func computedMetricMSEWithName(met *repo.MetricMSE, eqTypes []*repo.EquipmentType, name string) (*repo.MetricMSEComputed, error) {
// 	metric, err := computedMetricMSE(met, eqTypes)
// 	if err != nil {
// 		logger.Log.Sugar().Errorf("service/v1 - computedMetricMSEWithName - ", zap.String("reason", err.Error()))
// 		return nil, status.Error(codes.Internal, "cannot compute MSE metric")
// 	}
// 	metric.Name = name
// 	return metric, nil
// }

// func computedMetricMSE(met *repo.MetricMSE, eqTypes []*repo.EquipmentType) *repo.MetricMSEComputed {
// 	return &repo.MetricMSEComputed{
// 		Reference: met.Reference,
// 		Core:      met.Core,
// 		CPU:       met.CPU,
// 		Name:      met.Name,
// 	}
// }
