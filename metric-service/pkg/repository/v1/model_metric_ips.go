package v1

// MetricIPS is a representation of IBM.pvu.standard
type MetricIPS struct {
	ID               string
	Name             string
	NumCoreAttrID    string
	CoreFactorAttrID string
	BaseEqTypeID     string
}

// MetricIPSComputed has all the information required to be computed
type MetricIPSComputed struct {
	Name           string
	BaseType       *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
}

// EquipmentType for creating equipment type
type EquipmentType struct {
	ID         string
	Type       string
	SourceID   string
	SourceName string
	ParentID   string
	ParentType string
	Attributes []*Attribute
}

// MetricIPSConfig is a representation of IBM.pvu.standard metric configuration
type MetricIPSConfig struct {
	ID             string
	Name           string
	NumCoreAttr    string
	CoreFactorAttr string
	BaseEqType     string
}
