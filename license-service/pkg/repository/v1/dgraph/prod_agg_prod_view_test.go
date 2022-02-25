package dgraph

import (
	"fmt"
	v1 "optisam-backend/license-service/pkg/repository/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func compareApplicationForProductAggregationAll(t *testing.T, name string, exp []*v1.ApplicationsForProductData, act []*v1.ApplicationsForProductData) {
	if !assert.Lenf(t, act, len(exp), "expected number of elemnts are: %d", len(exp)) {
		return
	}

	for i := range exp {
		compareApplicationForProductAggregation(t, fmt.Sprintf("%s[%d]", name, i), exp[i], act[i])
	}
}

func compareApplicationForProductAggregation(t *testing.T, name string, exp *v1.ApplicationsForProductData, act *v1.ApplicationsForProductData) {
	assert.Equalf(t, exp.Name, act.Name, "%s.Name are not same", name)
	assert.Equalf(t, exp.Owner, act.Owner, "%s.Owner are not same", name)
	assert.Equalf(t, exp.NumOfEquipments, act.NumOfEquipments, "%s.NumOfEquipments are not same", name)
	assert.Equalf(t, exp.NumOfInstances, act.NumOfInstances, "%s.NumOfInstances are not same", name)
}
