// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/license-service/pkg/repository/v1 License

//License interface
type License interface {
	GetProducts(ctx context.Context, params *QueryProducts, scopes []string) (*ProductInfo, error)
	GetProductInformation(ctx context.Context, swidtag string, scopes []string) (*ProductAdditionalInfo, error)

	// CreateProductAggregation creates aggregations of a product
	CreateProductAggregation(ctx context.Context, pa *ProductAggregation, scopes []string) (*ProductAggregation, error)

	// List aggrations of product
	ListProductAggregations(ctx context.Context, scopes []string) ([]*ProductAggregation, error)

	// ProductAggregationsByName returns true and product aggregation details if object or node with that name exists
	ProductAggregationsByName(ctx context.Context, name string, scopes []string) (*ProductAggregation, error)

	// DeleteProductAggregation deletes an aggregation
	DeleteProductAggregation(ctx context.Context, id string, scopes []string) ([]*ProductAggregation, error)

	// ProductIDForSwidtag returns true and unique id assignerd by database if object or node with that id exists
	ProductIDForSwidtag(ctx context.Context, id string, params *QueryProducts, scopes []string) (string, error)

	GetApplications(ctx context.Context, params *QueryApplications, scopes []string) (*ApplicationInfo, error)

	// GetApplication gives the details of a perticular application with given id
	GetApplication(ctx context.Context, appID string, scopes []string) (*ApplicationDetails, error)
	GetProductsForApplication(ctx context.Context, id string, scopes []string) (*ProductsForApplication, error)
	GetApplicationsForProduct(ctx context.Context, params *QueryApplicationsForProduct, scopes []string) (*ApplicationsForProduct, error)
	GetInstancesForApplicationsProduct(ctx context.Context, params *QueryInstancesForApplicationProduct, scopes []string) (*InstancesForApplicationProduct, error)
	// ProductAcquiredRights fets list of acquired rights for the product along with ID of the product
	ProductAcquiredRights(ctx context.Context, swidTag string, scopes []string) (string, []*ProductAcquiredRight, error)

	// ProductEquipments list all the equipments for a product for given equipment type
	ProductEquipments(ctx context.Context, swidTag string, eqType *EquipmentType, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)
	// MetadataAllWithType gets metadata for given metadata type
	MetadataAllWithType(ctx context.Context, typ MetadataType, scopes []string) ([]*Metadata, error)

	// MetadataWithID gets metadata for given id
	MetadataWithID(ctx context.Context, id string, scopes []string) (*Metadata, error)

	// CreateEquipmentType stores equipmentdata and creates schema with required primary key
	// and indexes.
	CreateEquipmentType(ctx context.Context, eqType *EquipmentType, scopes []string) (*EquipmentType, error)

	// EquipmentTypes fetches all equipment types from database
	EquipmentTypes(ctx context.Context, scopes []string) ([]*EquipmentType, error)

	// AcquiredRights gets acquried rights based on query params.
	AcquiredRights(ctx context.Context, params *QueryAcquiredRights, scopes []string) (int32, []*AcquiredRights, error)

	EquipmentWithID(ctx context.Context, id string, scopes []string) (*EquipmentType, error)

	UpdateEquipmentType(ctx context.Context, id string, typ string, req *UpdateEquipmentRequest, scopes []string) (retType []*Attribute, retErr error)
	Equipments(ctx context.Context, eqType *EquipmentType, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)

	// Equipment gets equipmet for given type and id if exists,if not exist then ErrNotFound
	Equipment(ctx context.Context, eqType *EquipmentType, id string, scopes []string) (json.RawMessage, error)

	// EquipmentParents return parent of the given equipment
	EquipmentParents(ctx context.Context, eqType, parentEqType *EquipmentType, id string, scopes []string) (int32, json.RawMessage, error)

	// EquipmentChildren return children of the given equipment id for child type
	EquipmentChildren(ctx context.Context, eqType, childEqType *EquipmentType, id string, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)

	// EquipmentProducts return all the prioducts associted with given equipment
	EquipmentProducts(ctx context.Context, eqType *EquipmentType, id string, params *QueryEquipmentProduct, scopes []string) (int32, []*EquipmentProduct, error)

	// ListMetricTypeInfo gives a list of supported metric types
	ListMetricTypeInfo(ctx context.Context, scopes []string) ([]*MetricTypeInfo, error)

	// ListMetrices gives a list of supported metric types
	ListMetrices(ctx context.Context, scopes []string) ([]*Metric, error)

	// CreateMetricOPS creates an oracle.processor.standard metric
	CreateMetricOPS(ctx context.Context, mat *MetricOPS, scopes []string) (*MetricOPS, error)

	// ListMetricOPS returns all metrices of type oracle.processor.standard
	ListMetricOPS(ctx context.Context, scopes []string) ([]*MetricOPS, error)

	// MetricOPSComputedLicenses returns the computed licenses
	// for oracle.processor.standard metric
	MetricOPSComputedLicenses(ctx context.Context, id string, mat *MetricOPSComputed, scopes []string) (uint64, error)

	// CreateMetricSPS creates an sag.processor.standard metric
	CreateMetricSPS(ctx context.Context, mat *MetricSPS, scopes []string) (*MetricSPS, error)

	// ListMetricSPS returns all metrices of type sag.processor.standard
	ListMetricSPS(ctx context.Context, scopes []string) ([]*MetricSPS, error)

	//TODO: consider scope in computation of licenses ? clarify .
	// MetricSPSComputedLicenses returns the computed licenses
	// for sag.processor.standard metric
	MetricSPSComputedLicenses(ctx context.Context, id string, mat *MetricSPSComputed, scopes []string) (uint64, uint64, error)

	// CreateMetricIPS creates an sag.processor.standard metric
	CreateMetricIPS(ctx context.Context, mat *MetricIPS, scopes []string) (*MetricIPS, error)

	// ListMetricIPS returns all metrices of type sag.processor.standard
	ListMetricIPS(ctx context.Context, scopes []string) ([]*MetricIPS, error)

	// MetricIPSComputedLicenses returns the computed licenses
	// for ibm.pvu.standard metric
	MetricIPSComputedLicenses(ctx context.Context, id string, mat *MetricIPSComputed, scopes []string) (uint64, error)

	ListEditors(ctx context.Context, params *EditorQueryParams, scopes []string) ([]*Editor, error)
}

// Queryable interface provide methods for something that can be queried
type Queryable interface {
	// Key that needed to be queried (coloumn name)
	Key() string
	// Value for key tha we need tio search
	Value() interface{}

	Priority() int32
}

// SortOrder - type defined for sorting parameters i.e ascending/descending
type SortOrder int32

const (
	// SortASC - sorting in ascending order
	SortASC SortOrder = 0
	// SortDESC - sorting in descending order
	SortDESC SortOrder = 1
)
