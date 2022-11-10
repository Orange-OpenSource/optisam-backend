package v1

// MetricEquipAttrStand is a representation of equipment.attribute.standard
type MetricEquipAttrStand struct {
	ID            string
	Name          string
	EqType        string
	Environment   string
	AttributeName string
	Value         float64
}

// MetriccEquipAttrStandComputed has all the information required to be computed
type MetricEquipAttrStandComputed struct {
	Name        string
	EqTypeTree  []*EquipmentType
	BaseType    *EquipmentType
	Environment string
	Attribute   *Attribute
	Value       float64
}
