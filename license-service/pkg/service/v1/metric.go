package v1

import (
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/strcomp"

	repo "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"
)

func metricNameExistsAll(metrics []*repo.Metric, name string) int {
	for i, met := range metrics {
		if strcomp.CompareStrings(met.Name, name) {
			return i
		}
	}
	return -1
}
