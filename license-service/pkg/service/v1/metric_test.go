package v1

import (
	"fmt"
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/api/v1"

	"github.com/stretchr/testify/assert"
)

func compareListMetricResponse(t *testing.T, name string, exp *v1.ListMetricResponse, act *v1.ListMetricResponse) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	compareMetricAll(t, name+".Metrices", exp.Metrices, act.Metrices)
}

func compareMetricAll(t *testing.T, name string, exp []*v1.Metric, act []*v1.Metric) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareMetric(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareMetric(t *testing.T, name string, exp *v1.Metric, act *v1.Metric) {
	if exp == nil && act == nil {
		return
	}
	if exp == nil {
		assert.Nil(t, act, "attribute is expected to be nil")
	}

	assert.Equalf(t, exp.Name, act.Name, "%s.Names are not same", name)
	assert.Equalf(t, exp.Type, act.Type, "%s.Types are not same", name)
	assert.Equalf(t, exp.Description, act.Description, "%s.Descriptions are not same", name)

}
