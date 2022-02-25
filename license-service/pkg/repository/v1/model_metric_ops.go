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

func (m *MetricOPSComputed) Licenses() (new, old float64) {
	var nc, np, cf float64
	var ncOld, npOld, cfOld float64
	if m.CoreFactorAttr != nil {
		cf = m.CoreFactorAttr.ValFloat()
		cfOld = m.CoreFactorAttr.ValFloatOld()
	}
	if m.NumCoresAttr != nil {
		nc = m.NumCoresAttr.ValFloat()
		ncOld = m.NumCoresAttr.ValFloatOld()
	}
	if m.NumCoresAttr != nil {
		np = m.NumCPUAttr.ValFloat()
		npOld = m.NumCPUAttr.ValFloatOld()
	}

	return nc * np * cf, ncOld * npOld * cfOld
}
