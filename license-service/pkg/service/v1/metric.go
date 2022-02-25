package v1

import (
	"optisam-backend/common/optisam/strcomp"

	repo "optisam-backend/license-service/pkg/repository/v1"
)

func metricNameExistsAll(metrics []*repo.Metric, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
