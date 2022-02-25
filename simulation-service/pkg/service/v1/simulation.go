package v1

import (
	"context"
	"log"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"
	grpc_middleware "optisam-backend/common/optisam/middleware/grpc"
	licenseService "optisam-backend/license-service/pkg/api/v1"
	v1 "optisam-backend/simulation-service/pkg/api/v1"
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
	var wg sync.WaitGroup
	var metricResults []*v1.MetricSimulationResult // nolint: prealloc
	for _, simDetails := range req.MetricDetails {
		wg.Add(1)
		var metricRes v1.MetricSimulationResult
		req := licenseService.ProductLicensesForMetricRequest{
			SwidTag:    req.SwidTag,
			MetricName: simDetails.MetricName,
			UnitCost:   simDetails.UnitCost,
			Scope:      req.Scope,
		}
		go func(req *licenseService.ProductLicensesForMetricRequest) {
			log.Printf("Context:%v", ctx)
			cmptLicense, err := hcs.licenseClient.ProductLicensesForMetric(ctx, req)
			if err != nil {
				logger.Log.Error("service/v1 - Simulation - SimulationByMetric - LicenseService - ProductLicensesForMetric", zap.Error(err))
				metricRes.Success = false
				metricRes.MetricName = req.MetricName
				metricRes.SimFailureReason = err.Error()
			} else {
				metricRes.Success = true
				metricRes.NumCptLicences = cmptLicense.NumCptLicences
				metricRes.TotalCost = cmptLicense.TotalCost
				metricRes.MetricName = req.MetricName
			}
			defer wg.Done()
		}(&req)
		metricResults = append(metricResults, &metricRes)
	}
	wg.Wait()

	return &v1.SimulationByMetricResponse{
		MetricSimResult: metricResults,
	}, nil
}

// SimulationByHardware function implements simulation by hardware functionality
func (hcs *SimulationService) SimulationByHardware(ctx context.Context, req *v1.SimulationByHardwareRequest) (*v1.SimulationByHardwareResponse, error) {
	var simulationResults []*v1.SimulatedProductsLicenses // nolint: prealloc
	var wg sync.WaitGroup
	for _, simDetails := range req.MetricDetails {
		wg.Add(1)
		var simulationResult v1.SimulatedProductsLicenses
		req := licenseService.LicensesForEquipAndMetricRequest{
			EquipType:  req.EquipType,
			EquipId:    req.EquipId,
			MetricType: simDetails.MetricType,
			MetricName: simDetails.MetricName,
			Attributes: simulationToLicenseAttributesAll(req.Attributes),
			Scope:      req.Scope,
		}

		go func(req *licenseService.LicensesForEquipAndMetricRequest) {
			response, err := hcs.licenseClient.LicensesForEquipAndMetric(ctx, req)
			if err != nil {
				logger.Log.Error("service/v1 - Simulation - SimulationByHardware - LicenseService - LicensesForEquipAndMetric", zap.Error(err))
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

func licenseServToSimulationServProductLicenseAll(licenses []*licenseService.ProductLicenseForEquipAndMetric) []*v1.SimulatedProductLicense {

	simLicenses := make([]*v1.SimulatedProductLicense, 0)

	for _, productLicense := range licenses {
		simLicense := licenseServToSimulationServProductLicense(productLicense)
		simLicenses = append(simLicenses, simLicense)
	}

	return simLicenses
}

func licenseServToSimulationServProductLicense(license *licenseService.ProductLicenseForEquipAndMetric) *v1.SimulatedProductLicense {
	return &v1.SimulatedProductLicense{
		OldLicences: license.OldLicences,
		NewLicenses: license.NewLicenses,
		Delta:       license.Delta,
		ProductName: license.Product.Name,
		SwidTag:     license.Product.SwidTag,
		Editor:      license.Product.Editor,
	}
}

func simulationToLicenseAttributesAll(attrs []*v1.EquipAttribute) []*licenseService.Attribute {

	resAttrs := make([]*licenseService.Attribute, 0)
	for _, attr := range attrs {
		resAttr := simulationToLicenseAttributes(attr)
		resAttrs = append(resAttrs, resAttr)
	}

	return resAttrs
}

func simulationToLicenseAttributes(attr *v1.EquipAttribute) *licenseService.Attribute {
	lsattr := &licenseService.Attribute{
		ID:               attr.ID,
		Name:             attr.Name,
		PrimaryKey:       attr.PrimaryKey,
		DataType:         licenseService.DataTypes(attr.DataType),
		Displayed:        attr.Displayed,
		Searchable:       attr.Searchable,
		ParentIdentifier: attr.ParentIdentifier,
		MappedTo:         attr.MappedTo,
		Simulated:        attr.Simulated,
	}

	switch attr.DataType {
	case v1.DataTypes_INT:
		lsattr.Val = &licenseService.Attribute_IntVal{IntVal: attr.GetIntVal()}
		lsattr.OldVal = &licenseService.Attribute_IntValOld{IntValOld: attr.GetIntValOld()}
	case v1.DataTypes_FLOAT:
		lsattr.Val = &licenseService.Attribute_FloatVal{FloatVal: attr.GetFloatVal()}
		lsattr.OldVal = &licenseService.Attribute_FloatValOld{FloatValOld: attr.GetFloatValOld()}
	case v1.DataTypes_STRING:
		lsattr.Val = &licenseService.Attribute_StringVal{StringVal: attr.GetStringVal()}
		lsattr.OldVal = &licenseService.Attribute_StringValOld{StringValOld: attr.GetStringValOld()}
	}

	return lsattr
}
