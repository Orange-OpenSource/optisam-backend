package v1

// MetricAttrSumStand is a representation of attribute.sum.standard
type MetricAttrSumStand struct {
	ID             string
	Name           string
	EqType         string
	AttributeName  string
	ReferenceValue float64
	Default        bool
}

// MetricAttrSumStandComputed has all the information required to be computed
type MetricAttrSumStandComputed struct {
	Name           string
	BaseType       *EquipmentType
	Attribute      *Attribute
	ReferenceValue float64
}
