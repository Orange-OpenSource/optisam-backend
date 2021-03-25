// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/license-service/pkg/repository/v1 License

//License interface
type License interface {
	GetProductInformation(ctx context.Context, swidtag string, scopes ...string) (*ProductAdditionalInfo, error)

	// CreateProductAggregation creates aggregations of a product
	CreateProductAggregation(ctx context.Context, pa *ProductAggregation, scopes []string) (*ProductAggregation, error)
	ProductAggregationDetails(ctx context.Context, name string, params *QueryProductAggregations, scopes ...string) (*ProductAggregation, error)

	// ProductAggregationsByName returns true and product aggregation details if object or node with that name exists
	ProductAggregationsByName(ctx context.Context, name string, scopes []string) (*ProductAggregation, error)
	// ProductIDForSwidtag returns true and unique id assignerd by database if object or node with that id exists
	ProductIDForSwidtag(ctx context.Context, id string, params *QueryProducts, scopes ...string) (string, error)

	// ProductAcquiredRights fets list of acquired rights for the product along with ID of the product
	ProductAcquiredRights(ctx context.Context, swidTag string, scopes ...string) (string, []*ProductAcquiredRight, error)

	// MetadataAllWithType gets metadata for given metadata type
	MetadataAllWithType(ctx context.Context, typ MetadataType, scopes ...string) ([]*Metadata, error)

	// CreateEquipmentType stores equipmentdata and creates schema with required primary key
	// and indexes.
	CreateEquipmentType(ctx context.Context, eqType *EquipmentType, scopes []string) (*EquipmentType, error)

	// EquipmentTypes fetches all equipment types from database
	EquipmentTypes(ctx context.Context, scopes ...string) ([]*EquipmentType, error)

	// ListMetrices gives a list of supported metric types
	ListMetrices(ctx context.Context, scopes ...string) ([]*Metric, error)

	// ListMetricOPS returns all metrices of type oracle.processor.standard
	ListMetricOPS(ctx context.Context, scopes ...string) ([]*MetricOPS, error)

	ListMetricNUP(ctx context.Context, scopes ...string) ([]*MetricNUPOracle, error)

	// MetricOPSComputedLicenses returns the computed licenses
	// for oracle.processor.standard metric
	MetricOPSComputedLicenses(ctx context.Context, id string, mat *MetricOPSComputed, scopes ...string) (uint64, error)

	MetricOPSComputedLicensesAgg(ctx context.Context, name, mertic string, mat *MetricOPSComputed, scopes ...string) (uint64, error)

	MetricNUPComputedLicenses(ctx context.Context, id string, mat *MetricNUPComputed, scopes ...string) (uint64, error)

	MetricNUPComputedLicensesAgg(ctx context.Context, name, mertic string, mat *MetricNUPComputed, scopes ...string) (uint64, error)

	// ListMetricSPS returns all metrices of type sag.processor.standard
	ListMetricSPS(ctx context.Context, scopes ...string) ([]*MetricSPS, error)

	//TODO: consider scope in computation of licenses ? clarify .
	// MetricSPSComputedLicenses returns the computed licenses
	// for sag.processor.standard metric
	MetricSPSComputedLicenses(ctx context.Context, id string, mat *MetricSPSComputed, scopes ...string) (uint64, uint64, error)

	MetricSPSComputedLicensesAgg(ctx context.Context, name, mertic string, mat *MetricSPSComputed, scopes ...string) (uint64, uint64, error)

	// ListMetricIPS returns all metrices of type ibm.pvu.standard
	ListMetricIPS(ctx context.Context, scopes ...string) ([]*MetricIPS, error)

	// MetricIPSComputedLicenses returns the computed licenses for ibm.pvu.standard metric
	MetricIPSComputedLicenses(ctx context.Context, id string, mat *MetricIPSComputed, scopes ...string) (uint64, error)

	MetricIPSComputedLicensesAgg(ctx context.Context, name, metric string, mat *MetricIPSComputed, scopes ...string) (uint64, error)

	// MetricACSComputedLicenses returns the computed licenses for attribute.counter.standard metric
	MetricACSComputedLicenses(ctx context.Context, id string, mat *MetricACSComputed, scopes ...string) (uint64, error)

	// MetricINMComputedLicenses returns the computed licenses for instance.number.standard metric
	MetricINMComputedLicenses(ctx context.Context, id string, mat *MetricINMComputed, scopes ...string) (uint64, error)

	// MetricACSComputedLicensesAgg returns the computed licenses for product aggregation for attribute.counter.standard metric
	MetricACSComputedLicensesAgg(ctx context.Context, name, id string, mat *MetricACSComputed, scopes ...string) (uint64, error)

	//MetricINMComputedLicensesAgg returns the computes licences for prodAgg for instance.number.standard metric
	MetricINMComputedLicensesAgg(ctx context.Context, name, metric string, mat *MetricINMComputed, scopes ...string) (uint64, error)

	// ListMetricACS returns all metrices of type attribute.counter.standard
	ListMetricACS(ctx context.Context, scopes ...string) ([]*MetricACS, error)

	// ListMetricINM returns all metrices of type instance.number.standard
	ListMetricINM(ctx context.Context, scopes ...string) ([]*MetricINM, error)

	// ParentHirearchy gives equipment along with parent hirearchy
	ParentsHirerachyForEquipment(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, scopes ...string) (*Equipment, error)

	// ProductsForEquipmentForMetricOracleProcessorStandard gives products for oracle processor.standard
	ProductsForEquipmentForMetricOracleProcessorStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *MetricOPSComputed, scopes ...string) ([]*ProductData, error)

	// ProductsForEquipmentForMetricOracleProcessorStandard gives products for oracle.nup.standard
	ProductsForEquipmentForMetricOracleNUPStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *MetricNUPComputed, scopes ...string) ([]*ProductData, error)

	// ProductsForEquipmentForMetricIPSStandard gives products for oracle.nup.standard
	ProductsForEquipmentForMetricIPSStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *MetricIPSComputed, scopes ...string) ([]*ProductData, error)

	// ProductsForEquipmentForMetricSAGStandard gives products for oracle.nup.standard
	ProductsForEquipmentForMetricSAGStandard(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, metric *MetricSPSComputed, scopes ...string) ([]*ProductData, error)

	// ComputedLicensesForEquipmentForMetricOracleProcessorStandard gives licenses for product
	ComputedLicensesForEquipmentForMetricOracleProcessorStandard(ctx context.Context, equipID, equipType string, metric *MetricOPSComputed, scopes ...string) (int64, error)

	// ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll return ceiled and unceiled if equipment is at aggregation level or below aggregation level
	ComputedLicensesForEquipmentForMetricOracleProcessorStandardAll(ctx context.Context, equipID, equipType string, mat *MetricOPSComputed, scopes ...string) (int64, float64, error)

	//UsersForEquipmentForMetricOracleNUP gives users details for equipment for oracle nup
	UsersForEquipmentForMetricOracleNUP(ctx context.Context, equipID, equipType, productID string, hirearchyLevel uint8, metric *MetricNUPComputed, scopes ...string) ([]*User, error)

	//ProductExistsForApplication checks if the given product is linked with given application
	ProductExistsForApplication(ctx context.Context, prodID, appID string, scopes ...string) (bool, error)

	//ProductApplicationEquipments gives common equipments of product and applications
	ProductApplicationEquipments(ctx context.Context, prodID, appID string, scopes ...string) ([]*Equipment, error)

	//MetricOPSComputedLicensesForAppProduct gives licenses for application's product
	MetricOPSComputedLicensesForAppProduct(ctx context.Context, prodID, appID string, mat *MetricOPSComputed, scopes ...string) (uint64, error)
}

// Queryable interface provide methods for something that can be queried
type Queryable interface {
	// Key that needed to be queried (coloumn name)
	Key() string
	// Value for key tha we need tio search
	Value() interface{}

	// Values for key tha we need tio search
	Values() []interface{}

	Priority() int32

	Type() Filtertype
}

// SortOrder - type defined for sorting parameters i.e ascending/descending
type SortOrder int32

const (
	// SortASC - sorting in ascending order
	SortASC SortOrder = 0
	// SortDESC - sorting in descending order
	SortDESC SortOrder = 1
)
