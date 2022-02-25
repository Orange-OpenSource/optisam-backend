package v1

// MetricSPS is a representation of sag.processor.standard
type MetricSPS struct {
	ID               string
	Name             string
	NumCoreAttrID    string
	CoreFactorAttrID string
	BaseEqTypeID     string
}

// MetricSPSComputed has all the information required to be computed
type MetricSPSComputed struct {
	Name           string
	BaseType       *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
}
