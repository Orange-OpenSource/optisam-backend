// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/metric-service/pkg/repository/v1 Metric

//Metric interface
type Metric interface {
	// ListMetricTypeInfo gives a list of supported metric types
	ListMetricTypeInfo(ctx context.Context, scopes []string) ([]*MetricTypeInfo, error)

	// ListMetrices gives a list of supported metric types
	ListMetrices(ctx context.Context, scopes []string) ([]*MetricInfo, error)

	// CreateMetricOPS creates an oracle.processor.standard metric
	CreateMetricOPS(ctx context.Context, mat *MetricOPS, scopes []string) (*MetricOPS, error)

	// ListMetricOPS returns all metrices of type oracle.processor.standard
	ListMetricOPS(ctx context.Context, scopes []string) ([]*MetricOPS, error)

	//ListMetricNUP returns all metrics of type of oracle NUP
	ListMetricNUP(ctx context.Context, scopes []string) ([]*MetricNUPOracle, error)

	// CreateMetricSPS creates an sag.processor.standard metric
	CreateMetricSPS(ctx context.Context, mat *MetricSPS, scopes []string) (*MetricSPS, error)

	// CreateMetricACS creates an attribute.counter.standard metric
	CreateMetricACS(ctx context.Context, mat *MetricACS, attr *Attribute, scopes []string) (*MetricACS, error)

	// ListMetricACS returns all metrices of type attribute.counter.standard
	ListMetricACS(ctx context.Context, scopes []string) ([]*MetricACS, error)

	// ListMetricSPS returns all metrices of type sag.processor.standard
	ListMetricSPS(ctx context.Context, scopes []string) ([]*MetricSPS, error)

	// CreateMetricIPS creates an sag.processor.standard metric
	CreateMetricIPS(ctx context.Context, mat *MetricIPS, scopes []string) (*MetricIPS, error)

	// ListMetricIPS returns all metrices of type ibm.pvu.standard
	ListMetricIPS(ctx context.Context, scopes []string) ([]*MetricIPS, error)

	//CreateMetricOracleNUPStandard creates an oracle.nup.standard metric
	CreateMetricOracleNUPStandard(ctx context.Context, mat *MetricNUPOracle, scopes []string) (*MetricNUPOracle, error)

	// EquipmentTypes fetches all equipment types from database
	EquipmentTypes(ctx context.Context, scopes []string) ([]*EquipmentType, error)

	CreateMetricInstanceNumberStandard(ctx context.Context, mat *MetricINM, scopes []string) (*MetricINM, error)

	// GetMetricConfigOPS return metric configuration of type oracle.processor.standard
	GetMetricConfigOPS(ctx context.Context, metName string, scopes []string) (*MetricOPSConfig, error)

	// GetMetricConfigNUP return metric configuration of type oracle.nup.standard
	GetMetricConfigNUP(ctx context.Context, metName string, scopes []string) (*MetricNUPConfig, error)

	// GetMetricConfigSPS return metric configuration of type sag.processor.standard
	GetMetricConfigSPS(ctx context.Context, metName string, scopes []string) (*MetricSPSConfig, error)

	// GetMetricConfigIPS return metric configuration of type ibm.pvu.standard
	GetMetricConfigIPS(ctx context.Context, metName string, scopes []string) (*MetricIPSConfig, error)

	// GetMetricConfigACS return metric configuration of type attribute.counter.standard
	GetMetricConfigACS(ctx context.Context, metName string, scopes []string) (*MetricACS, error)

	// GetMetricConfigINM return metric configuration of type instance.number.standard
	GetMetricConfigINM(ctx context.Context, metName string, scopes []string) (*MetricINMConfig, error)
}

//Filtertype ...
type Filtertype int32

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
