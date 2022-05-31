package v1

// MetricIPS is a representation of IBM.pvu.standard
type MetricIPS struct {
	ID               string
	Name             string
	NumCoreAttrID    string
	NumCPUAttrID     string
	CoreFactorAttrID string
	BaseEqTypeID     string
}

// MetricIPSComputed has all the information required to be computed
type MetricIPSComputed struct {
	Name           string
	BaseType       *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
	NumCPUAttr     *Attribute
}
