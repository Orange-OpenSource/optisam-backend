// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
	repo "optisam-backend/license-service/pkg/repository/v1"
)

//func takes input and output as map[string]interface to handle different types of input and ouput
var MetricCalculation map[repo.MetricType]func(*licenseServiceServer, context.Context, []*repo.EquipmentType, map[string]interface{}) (map[string]interface{}, error)

//Short abbreviation as key for map for output handling with uniqnuess
var (
	COMPUTED_LICENCES string = "COMPUTED_LICENCES"
	PROD_ID           string = "PROD_ID"
	METRIC_NAME       string = "METRIC_NAME"
	PROD_AGG_NAME     string = "PROD_AGG_NAME"
	IS_AGG            string = "IS_AGG"
	SCOPES            string = "SCOPES"
)

/* Map based metric handling , define new metric and handle here
it will call it self whenever new kind of metrics come into system */

func init() {
	MetricCalculation = make(map[repo.MetricType]func(*licenseServiceServer, context.Context, []*repo.EquipmentType, map[string]interface{}) (map[string]interface{}, error))
	MetricCalculation[repo.MetricOPSOracleProcessorStandard] = opsMetricCalulation
	MetricCalculation[repo.MetricSPSSagProcessorStandard] = spsMetricCalulation
	MetricCalculation[repo.MetricIPSIbmPvuStandard] = ipsMetricCalulation
	MetricCalculation[repo.MetricOracleNUPStandard] = nupMetricCalulation
	MetricCalculation[repo.MetricAttrCounterStandard] = acsMetricCalulation
	MetricCalculation[repo.MetricInstanceNumberStandard] = insMetricCalulation
}

func opsMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesOPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[COMPUTED_LICENCES] = computedLicences
	return resp, nil
}

func spsMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicencesProd, computedLicencesNoProd, err := s.computedLicensesSPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	if computedLicencesProd > computedLicencesNoProd {
		resp[COMPUTED_LICENCES] = computedLicencesProd
	} else {
		resp[COMPUTED_LICENCES] = computedLicencesNoProd
	}
	return resp, nil
}

func ipsMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesIPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[COMPUTED_LICENCES] = computedLicences
	return resp, nil
}

func nupMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesNUP(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[COMPUTED_LICENCES] = computedLicences
	return resp, nil
}

func acsMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesACS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[COMPUTED_LICENCES] = computedLicences
	return resp, nil
}

func insMetricCalulation(s *licenseServiceServer, ctx context.Context, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesINM(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[COMPUTED_LICENCES] = computedLicences
	return resp, nil
}
