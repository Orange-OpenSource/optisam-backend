package v1

// MetricOPS is a representation of oracle.processor.standard
type MetricOPS struct {
	ID                    string
	Name                  string
	NumCoreAttrID         string
	NumCPUAttrID          string
	CoreFactorAttrID      string
	StartEqTypeID         string
	BaseEqTypeID          string
	AggerateLevelEqTypeID string
	EndEqTypeID           string
}

// MetricOPSComputed has all the information required to be computed
type MetricOPSComputed struct {
	Name           string
	EqTypeTree     []*EquipmentType
	BaseType       *EquipmentType
	AggregateLevel *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
	NumCPUAttr     *Attribute
}

// MetricOPSConfig is a representation of oracle.processor.standard metric configuration
type MetricOPSConfig struct {
	ID                  string
	Name                string
	NumCoreAttr         string
	NumCPUAttr          string
	CoreFactorAttr      string
	StartEqType         string
	BaseEqType          string
	AggerateLevelEqType string
	EndEqType           string
}
