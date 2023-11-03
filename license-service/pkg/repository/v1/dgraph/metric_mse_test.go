package dgraph

import (
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func compareMetricMSE(t *testing.T, name string, exp, act *v1.MetricMSE) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "metadata is expected to be nil")
	}

	if exp.ID != "" {
		assert.Equalf(t, exp.ID, act.ID, "%s.ID should be same", name)
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Source should be same", name)
	assert.Equalf(t, exp.Reference, act.Reference, "%s.Reference should be same", name)
	assert.Equalf(t, exp.Core, act.Core, "%s.Core should be same", name)
	assert.Equalf(t, exp.CPU, act.CPU, "%s.CPU should be same", name)
}
