package v1

// MetricACS is a representation of sag.processor.standard
type MetricACS struct {
	ID            string
	Name          string
	EqType        string
	AttributeName string
	Value         string
}

// MetricACSComputed has all the information required to be computed
type MetricACSComputed struct {
	Name      string
	BaseType  *EquipmentType
	Attribute *Attribute
	Value     string
}
