package dgraph

import (
	"testing"

	v1 "gitlab.tech.orange/optisam/optisam-it/optisam-services/license-service/pkg/repository/v1"

	"github.com/stretchr/testify/assert"
)

func compareMetricOracleNUP(t *testing.T, name string, exp, act *v1.MetricNUPOracle) {
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
	assert.Equalf(t, exp.StartEqTypeID, act.StartEqTypeID, "%s.StartEqTypeID should be same", name)
	assert.Equalf(t, exp.BaseEqTypeID, act.BaseEqTypeID, "%s.BaseEqTypeID should be same", name)
	assert.Equalf(t, exp.AggerateLevelEqTypeID, act.AggerateLevelEqTypeID, "%s.AggerateLevelEqTypeID should be same", name)
	assert.Equalf(t, exp.EndEqTypeID, act.EndEqTypeID, "%s.EndEqTypeID should be same", name)
	assert.Equalf(t, exp.CoreFactorAttrID, act.CoreFactorAttrID, "%s.CoreFactorAttrID should be same", name)
	assert.Equalf(t, exp.NumCoreAttrID, act.NumCoreAttrID, "%s.NumCoreAttrID should be same", name)
	assert.Equalf(t, exp.NumCPUAttrID, act.NumCPUAttrID, "%s.NumCPUAttrID should be same", name)
	assert.Equalf(t, exp.NumberOfUsers, act.NumberOfUsers, "%s.NumUsersAttrID should be same", name)
}
