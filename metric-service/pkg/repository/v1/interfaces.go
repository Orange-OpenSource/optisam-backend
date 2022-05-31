package v1

import (
	"context"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/metric-service/pkg/repository/v1 Metric

// Metric interface
type Metric interface {
	// ListMetricTypeInfo gives a list of supported metric types
	ListMetricTypeInfo(ctx context.Context, scopetype ScopeType, scope string) ([]*MetricTypeInfo, error)

	// ListMetrices gives a list of supported metric types
	ListMetrices(ctx context.Context, scopes string) ([]*MetricInfo, error)

	// MetricInfoWithAcqAndAgg gives a metric info with aggregation and acqrights
	MetricInfoWithAcqAndAgg(ctx context.Context, metricName, scope string) (*MetricInfoFull, error)

	// DeleteMetric deletes metric with given metric name and scope
	DeleteMetric(ctx context.Context, metricName, scope string) error

	// CreateMetricOPS creates an oracle.processor.standard metric
	CreateMetricOPS(ctx context.Context, mat *MetricOPS, scopes string) (*MetricOPS, error)

	// ListMetricOPS returns all metrics of type oracle.processor.standard
	ListMetricOPS(ctx context.Context, scopes string) ([]*MetricOPS, error)

	// ListMetricNUP returns all metrics of type of oracle NUP
	ListMetricNUP(ctx context.Context, scopes string) ([]*MetricNUPOracle, error)

	// CreateMetricSPS creates an sag.processor.standard metric
	CreateMetricSPS(ctx context.Context, mat *MetricSPS, scopes string) (*MetricSPS, error)

	// CreateMetricACS creates an attribute.counter.standard metric
	CreateMetricACS(ctx context.Context, mat *MetricACS, attr *Attribute, scopes string) (*MetricACS, error)

	// ListMetricACS returns all metrics of type attribute.counter.standard
	ListMetricACS(ctx context.Context, scopes string) ([]*MetricACS, error)

	// CreateMetricAttrSum creates an attribute.sum.standard metric
	CreateMetricAttrSum(ctx context.Context, mat *MetricAttrSumStand, attr *Attribute, scopes string) (*MetricAttrSumStand, error)

	// ListMetricAttrSum returns all metrics of type attribute.sum.standard metric
	ListMetricAttrSum(ctx context.Context, scopes string) ([]*MetricAttrSumStand, error)

	// ListMetricSPS returns all metrics of type sag.processor.standard
	ListMetricSPS(ctx context.Context, scopes string) ([]*MetricSPS, error)

	// CreateMetricIPS creates an sag.processor.standard metric
	CreateMetricIPS(ctx context.Context, mat *MetricIPS, scopes string) (*MetricIPS, error)

	// ListMetricIPS returns all metrics of type ibm.pvu.standard
	ListMetricIPS(ctx context.Context, scopes string) ([]*MetricIPS, error)

	// ListMetricSS returns all metrics of type static.standard
	// ListMetricSS(ctx context.Context, scopes string) ([]*MetricSS, error)

	// CreateMetricOracleNUPStandard creates an oracle.nup.standard metric
	CreateMetricOracleNUPStandard(ctx context.Context, mat *MetricNUPOracle, scopes string) (*MetricNUPOracle, error)

	// EquipmentTypes fetches all equipment types from database
	EquipmentTypes(ctx context.Context, scopes string) ([]*EquipmentType, error)

	CreateMetricInstanceNumberStandard(ctx context.Context, mat *MetricINM, scopes string) (*MetricINM, error)

	// CreateMetricUSS creates an User.sum.standard metric
	CreateMetricUSS(ctx context.Context, met *MetricUSS, scope string) (*MetricUSS, error)

	// CreateMetricStaticStandard creates an static.standard metric
	CreateMetricStaticStandard(ctx context.Context, met *MetricSS, scope string) (*MetricSS, error)

	// GetMetricConfigUSS return metric configuration of type User.sum.standard
	GetMetricConfigUSS(ctx context.Context, metName string, scope string) (*MetricUSS, error)

	// GetMetricConfigOPS return metric configuration of type oracle.processor.standard
	GetMetricConfigOPS(ctx context.Context, metName string, scopes string) (*MetricOPSConfig, error)

	// GetMetricConfigOPSID return metric configuration of type oracle.processor.standard
	GetMetricConfigOPSID(ctx context.Context, metName string, scope string) (*MetricOPS, error)

	// GetMetricConfigNUP return metric configuration of type oracle.nup.standard
	GetMetricConfigNUP(ctx context.Context, metName string, scopes string) (*MetricNUPConfig, error)

	// GetMetricConfigNUPID return metric configuration of type oracle.nup.standard
	GetMetricConfigNUPID(ctx context.Context, metName string, scope string) (*MetricNUPOracle, error)

	// GetMetricConfigSPS return metric configuration of type sag.processor.standard
	GetMetricConfigSPS(ctx context.Context, metName string, scopes string) (*MetricSPSConfig, error)

	// GetMetricConfigSPSID return metric configuration of type sag.processor.standard
	GetMetricConfigSPSID(ctx context.Context, metName string, scope string) (*MetricSPS, error)

	// GetMetricConfigIPS return metric configuration of type ibm.pvu.standard
	GetMetricConfigIPS(ctx context.Context, metName string, scopes string) (*MetricIPSConfig, error)

	// GetMetricConfigIPSID return metric configuration of type ibm.pvu.standard
	GetMetricConfigIPSID(ctx context.Context, metName string, scope string) (*MetricIPS, error)

	// GetMetricConfigACS return metric configuration of type attribute.counter.standard
	GetMetricConfigACS(ctx context.Context, metName string, scopes string) (*MetricACS, error)

	// GetMetricConfigAttrSum return metric configuration of type attribute.sum.standard
	GetMetricConfigAttrSum(ctx context.Context, metName string, scopes string) (*MetricAttrSumStand, error)

	// GetMetricConfigINM return metric configuration of type instance.number.standard
	GetMetricConfigINM(ctx context.Context, metName string, scopes string) (*MetricINM, error)

	// GetMetricConfigSS return metric configuration of type static.standard
	GetMetricConfigSS(ctx context.Context, metName string, scopes string) (*MetricSS, error)

	// DropMetrics delete the all metrics of particular scope
	DropMetrics(ctx context.Context, scope string) error

	// UpdateMetricINM updates parameter(coeffitient) of the metric
	UpdateMetricINM(ctx context.Context, met *MetricINM, scope string) error

	// UpdateMetricSS updates parameter(Reference Value) of the metric
	UpdateMetricSS(ctx context.Context, met *MetricSS, scope string) error

	// UpdateMetricAttrSum updates parameter(metric Reference Value, EqType, AttributeName) of the metric
	UpdateMetricAttrSum(ctx context.Context, met *MetricAttrSumStand, scope string) error

	// UpdateMetricACS updates parameter(metric Value, EqType, AttributeName) of the metric
	UpdateMetricACS(ctx context.Context, met *MetricACS, scope string) error

	// UpdateMetricIPS updates parameter(NumCoreAttrID, BaseEqTypeID, CoreFactorAttrID) of the metric
	UpdateMetricIPS(ctx context.Context, met *MetricIPS, scope string) error

	// UpdateMetricSPS updates parameter(NumCoreAttrID, BaseEqTypeID, CoreFactorAttrID) of the metric
	UpdateMetricSPS(ctx context.Context, met *MetricSPS, scope string) error

	// UpdateMetricOPS updates parameter(StartEqTypeID, AggerateLevelEqTypeID, EndEqTypeID, NumCPUAttrID, NumCoreAttrID, BaseEqTypeID, CoreFactorAttrID) of the metric
	UpdateMetricOPS(ctx context.Context, met *MetricOPS, scope string) error

	// UpdateMetricNUP updates parameter(StartEqTypeID, AggerateLevelEqTypeID, EndEqTypeID, NumCPUAttrID, NumCoreAttrID, BaseEqTypeID, CoreFactorAttrID, NumberOfUsers) of the metric
	UpdateMetricNUP(ctx context.Context, met *MetricNUPOracle, scope string) error
}

// Filtertype ...
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
