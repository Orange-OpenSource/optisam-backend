package v1

import (
	"context"
	repo "optisam-backend/license-service/pkg/repository/v1"
)

// MetricCalculation takes input and output as map[string]interface to handle different types of input and output
var MetricCalculation map[repo.MetricType]func(context.Context, *licenseServiceServer, []*repo.EquipmentType, map[string]interface{}) (map[string]interface{}, error)

// Short abbreviation as key for map for output handling with uniqnuess
var (
	ComputedLicenses string = "COMPUTED_LICENCES"
	ComputedDetails  string = "COMPUTED_DETAILS"
	ProdID           string = "PROD_ID"
	MetricName       string = "METRIC_NAME"
	ProdAggName      string = "PROD_AGG_NAME"
	IsAgg            string = "IS_AGG"
	SCOPES           string = "SCOPES"
	SWIDTAG          string = "SWIDTAG"
)

/* Map based metric handling , define new metric and handle here
it will call it self whenever new kind of metrics come into system */

func init() {
	MetricCalculation = make(map[repo.MetricType]func(context.Context, *licenseServiceServer, []*repo.EquipmentType, map[string]interface{}) (map[string]interface{}, error))
	MetricCalculation[repo.MetricOPSOracleProcessorStandard] = opsMetricCalulation
	MetricCalculation[repo.MetricSPSSagProcessorStandard] = spsMetricCalulation
	MetricCalculation[repo.MetricIPSIbmPvuStandard] = ipsMetricCalulation
	MetricCalculation[repo.MetricOracleNUPStandard] = nupMetricCalulation
	MetricCalculation[repo.MetricAttrCounterStandard] = acsMetricCalulation
	MetricCalculation[repo.MetricInstanceNumberStandard] = insMetricCalulation
	MetricCalculation[repo.MetricAttrSumStandard] = attrSumMetricCalulation
	MetricCalculation[repo.MetricUserSumStandard] = userSumMetricCalulation
	MetricCalculation[repo.MetricStaticStandard] = ssMetricCalulation
	MetricCalculation[repo.MetricEquipAttrStandard] = equipAttrMetricCalulation
	MetricCalculation[repo.MetricUserNomStandard] = unsMetricCalulation
	MetricCalculation[repo.MetricUserConcurentStandard] = ucsMetricCalulation

}

func opsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesOPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	return resp, nil
}

func spsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicencesProd, computedLicencesNoProd, err := s.computedLicensesSPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	if computedLicencesProd > computedLicencesNoProd {
		resp[ComputedLicenses] = computedLicencesProd
	} else {
		resp[ComputedLicenses] = computedLicencesNoProd
	}
	return resp, nil
}

func ipsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesIPS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	return resp, nil
}

func nupMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesNUP(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func acsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesACS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	return resp, nil
}

func insMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) { //nolint:unparam
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesINM(ctx, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func ssMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) { //nolint:unparam
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesSS(ctx, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func attrSumMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesAttrSum(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func userSumMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesUserSum(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func equipAttrMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) {
	resp := make(map[string]interface{})
	computedLicences, err := s.computedLicensesEquipAttr(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	return resp, nil
}

func unsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) { //nolint:unparam
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesUNS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}

func ucsMetricCalulation(ctx context.Context, s *licenseServiceServer, eqTypes []*repo.EquipmentType, input map[string]interface{}) (map[string]interface{}, error) { //nolint:unparam
	resp := make(map[string]interface{})
	computedLicences, computedDetails, err := s.computedLicensesUCS(ctx, eqTypes, input)
	if err != nil {
		return resp, err
	}
	resp[ComputedLicenses] = computedLicences
	resp[ComputedDetails] = computedDetails
	return resp, nil
}
