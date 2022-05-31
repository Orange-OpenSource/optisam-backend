package v1

import (
	"context"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	l_v1 "optisam-backend/license-service/pkg/api/v1"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SimulationByMetric function implements simulation by metric functionality
func (hcs *SimulationService) SimulationByMetric(ctx context.Context, req *v1.SimulationByMetricRequest) (*v1.SimulationByMetricResponse, error) {
	userClaims, ok := grpc_middleware.RetrieveClaims(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "cannot find claims in context")
	}
	if !helper.Contains(userClaims.Socpes, req.GetScope()) {
		return nil, status.Error(codes.PermissionDenied, "Do not have access to the scope")
	}
	if len(req.MetricDetails) == 0 {
		return &v1.SimulationByMetricResponse{
			Success: true,
		}, nil
	}
	compResp, err := hcs.licenseClient.GetOverAllCompliance(ctx, &l_v1.GetOverAllComplianceRequest{
		Editor:     req.Editor,
		Scope:      req.Scope,
		Simulation: true,
	})
	if err != nil {
		logger.Log.Error("service/v1 - Simulation - SimulationByMetric - l_v1 - GetOverAllCompliance", zap.Error(err))
		return &v1.SimulationByMetricResponse{
			Success:          false,
			SimFailureReason: err.Error(),
		}, nil
	}
	return &v1.SimulationByMetricResponse{
		Success:         true,
		MetricSimResult: convertCompToMetricSimResponse(compResp.AcqRights, req.MetricDetails),
	}, nil
}

// SimulationByHardware function implements simulation by hardware functionality
func (hcs *SimulationService) SimulationByHardware(ctx context.Context, req *v1.SimulationByHardwareRequest) (*v1.SimulationByHardwareResponse, error) {
	var simulationResults []*v1.SimulatedProductsLicenses // nolint: prealloc
	var wg sync.WaitGroup
	for _, simDetails := range req.MetricDetails {
		wg.Add(1)
		var simulationResult v1.SimulatedProductsLicenses
		req := l_v1.LicensesForEquipAndMetricRequest{
			EquipType:  req.EquipType,
			EquipId:    req.EquipId,
			MetricType: simDetails.MetricType,
			MetricName: simDetails.MetricName,
			Attributes: simulationToLicenseAttributesAll(req.Attributes),
			Scope:      req.Scope,
		}

		go func(req *l_v1.LicensesForEquipAndMetricRequest) {
			response, err := hcs.licenseClient.LicensesForEquipAndMetric(ctx, req)
			if err != nil {
				logger.Log.Error("service/v1 - Simulation - SimulationByHardware - l_v1 - LicensesForEquipAndMetric", zap.Error(err))
				simulationResult.Success = false
				simulationResult.MetricName = req.MetricName
				simulationResult.SimFailureReason = err.Error()
			} else {
				simulationResult.Success = true
				simulationResult.MetricName = req.MetricName
				simulationResult.Licenses = licenseServToSimulationServProductLicenseAll(response.Licenses)
			}

			defer wg.Done()
		}(&req)

		simulationResults = append(simulationResults, &simulationResult)
	}
	wg.Wait()
	return &v1.SimulationByHardwareResponse{
		SimulationResult: simulationResults,
	}, nil
}

func licenseServToSimulationServProductLicenseAll(licenses []*l_v1.ProductLicenseForEquipAndMetric) []*v1.SimulatedProductLicense {

	simLicenses := make([]*v1.SimulatedProductLicense, 0)

	for _, productLicense := range licenses {
		simLicense := licenseServToSimulationServProductLicense(productLicense)
		simLicenses = append(simLicenses, simLicense)
	}

	return simLicenses
}

func licenseServToSimulationServProductLicense(license *l_v1.ProductLicenseForEquipAndMetric) *v1.SimulatedProductLicense {
	return &v1.SimulatedProductLicense{
		OldLicences: license.OldLicences,
		NewLicenses: license.NewLicenses,
		Delta:       license.Delta,
		ProductName: license.Product.Name,
		SwidTag:     license.Product.SwidTag,
		Editor:      license.Product.Editor,
	}
}

func simulationToLicenseAttributesAll(attrs []*v1.EquipAttribute) []*l_v1.Attribute {

	resAttrs := make([]*l_v1.Attribute, 0)
	for _, attr := range attrs {
		resAttr := simulationToLicenseAttributes(attr)
		resAttrs = append(resAttrs, resAttr)
	}

	return resAttrs
}

func simulationToLicenseAttributes(attr *v1.EquipAttribute) *l_v1.Attribute {
	lsattr := &l_v1.Attribute{
		ID:               attr.ID,
		Name:             attr.Name,
		PrimaryKey:       attr.PrimaryKey,
		DataType:         l_v1.DataTypes(attr.DataType),
		Displayed:        attr.Displayed,
		Searchable:       attr.Searchable,
		ParentIdentifier: attr.ParentIdentifier,
		MappedTo:         attr.MappedTo,
		Simulated:        attr.Simulated,
	}

	switch attr.DataType {
	case v1.DataTypes_INT:
		lsattr.Val = &l_v1.Attribute_IntVal{IntVal: attr.GetIntVal()}
		lsattr.OldVal = &l_v1.Attribute_IntValOld{IntValOld: attr.GetIntValOld()}
	case v1.DataTypes_FLOAT:
		lsattr.Val = &l_v1.Attribute_FloatVal{FloatVal: attr.GetFloatVal()}
		lsattr.OldVal = &l_v1.Attribute_FloatValOld{FloatValOld: attr.GetFloatValOld()}
	case v1.DataTypes_STRING:
		lsattr.Val = &l_v1.Attribute_StringVal{StringVal: attr.GetStringVal()}
		lsattr.OldVal = &l_v1.Attribute_StringValOld{StringValOld: attr.GetStringValOld()}
	}

	return lsattr
}

func convertCompToMetricSimResponse(compliance []*l_v1.AggregationAcquiredRights, metsim []*v1.MetricSimDetails) []*v1.MetricSimulationResult {
	simresp := []*v1.MetricSimulationResult{}
	for _, comp := range compliance {
		for _, ms := range metsim {
			sliceCompSwidtags := strings.Split(comp.SwidTags, ",")
			if len(sliceCompSwidtags) > 1 {
				sort.Strings(sliceCompSwidtags)
			}
			sliceMsSwidtags := strings.Split(ms.Swidtag, ",")
			if len(sliceMsSwidtags) > 1 {
				sort.Strings(sliceMsSwidtags)
			}
			if comp.Metric == ms.MetricName && comp.SKU == ms.Sku && strings.Join(sliceCompSwidtags, ",") == strings.Join(sliceMsSwidtags, ",") && comp.AggregationName == ms.AggregationName {
				simresp = append(simresp, &v1.MetricSimulationResult{
					Swidtag:          strings.Join(sliceCompSwidtags, ","),
					AggregationName:  comp.AggregationName,
					MetricName:       ms.MetricName,
					NumCptLicences:   uint64(comp.NumCptLicences),
					OldTotalCost:     float64(comp.NumCptLicences) * comp.AvgUnitPrice,
					NewTotalCost:     float64(comp.NumCptLicences) * ms.UnitCost,
					Sku:              comp.SKU,
					NotDeployed:      comp.NotDeployed,
					MetricNotDefined: comp.MetricNotDefined,
				})
			}
		}
	}
	return concatMetricSimResultForSameSwidtag(simresp)
}

func concatMetricSimResultForSameSwidtag(simMet []*v1.MetricSimulationResult) []*v1.MetricSimulationResult {
	resSimMetric := make([]*v1.MetricSimulationResult, 0, len(simMet))
	encountered := map[string]int{}
	for i := range simMet {
		idx, ok := encountered[simMet[i].Swidtag+":"+simMet[i].MetricName]
		if ok {
			// Add values to original.
			resSimMetric[idx].Sku = strings.Join([]string{resSimMetric[idx].Sku, simMet[i].Sku}, ",")
			resSimMetric[idx].OldTotalCost += simMet[i].OldTotalCost
			resSimMetric[idx].NewTotalCost += simMet[i].NewTotalCost
		} else {
			// Record this element as an encountered element.
			encountered[simMet[i].Swidtag+":"+simMet[i].MetricName] = len(resSimMetric)
			// Append to result slice.
			resSimMetric = append(resSimMetric, &v1.MetricSimulationResult{
				Sku:              simMet[i].Sku,
				MetricName:       simMet[i].MetricName,
				NumCptLicences:   simMet[i].NumCptLicences,
				OldTotalCost:     simMet[i].OldTotalCost,
				NewTotalCost:     simMet[i].NewTotalCost,
				AggregationName:  simMet[i].AggregationName,
				Swidtag:          simMet[i].Swidtag,
				NotDeployed:      simMet[i].NotDeployed,
				MetricNotDefined: simMet[i].MetricNotDefined,
			})
		}
	}
	return resSimMetric
}
