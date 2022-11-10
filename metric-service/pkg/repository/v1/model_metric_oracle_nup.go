package v1

// MetricNUPOracle is a representation of oracle.nup.standard
type MetricNUPOracle struct {
	ID                    string
	Name                  string
	NumCoreAttrID         string
	NumCPUAttrID          string
	CoreFactorAttrID      string
	StartEqTypeID         string
	BaseEqTypeID          string
	AggerateLevelEqTypeID string
	EndEqTypeID           string
	NumberOfUsers         uint32
	Transform             bool
	TransformMetricName   string
}

// MetricOPS return metric ops
func (m *MetricNUPOracle) MetricOPS() *MetricOPS {
	return &MetricOPS{
		ID:                    m.ID,
		Name:                  m.Name,
		NumCoreAttrID:         m.NumCoreAttrID,
		NumCPUAttrID:          m.NumCPUAttrID,
		CoreFactorAttrID:      m.CoreFactorAttrID,
		StartEqTypeID:         m.StartEqTypeID,
		BaseEqTypeID:          m.BaseEqTypeID,
		AggerateLevelEqTypeID: m.AggerateLevelEqTypeID,
		EndEqTypeID:           m.EndEqTypeID,
	}
}

// MetricNUPComputed has all the information required to be computed
type MetricNUPComputed struct {
	Name           string
	EqTypeTree     []*EquipmentType
	BaseType       *EquipmentType
	AggregateLevel *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
	NumCPUAttr     *Attribute
	NumOfUsers     uint32
}

// NewMetricNUPComputed  returns NewMetricNUPComputed from MetricOPSComputed and num of users node
func NewMetricNUPComputed(m *MetricOPSComputed, numOfUsers uint32) *MetricNUPComputed {
	return &MetricNUPComputed{
		EqTypeTree:     m.EqTypeTree,
		BaseType:       m.BaseType,
		AggregateLevel: m.AggregateLevel,
		CoreFactorAttr: m.CoreFactorAttr,
		NumCoresAttr:   m.NumCoresAttr,
		NumCPUAttr:     m.NumCPUAttr,
		NumOfUsers:     numOfUsers,
	}
}

// MetricOPSComputed returns MetricOPSComputed for nup metic
func (m MetricNUPComputed) MetricOPSComputed() *MetricOPSComputed {
	return &MetricOPSComputed{
		EqTypeTree:     m.EqTypeTree,
		BaseType:       m.BaseType,
		AggregateLevel: m.AggregateLevel,
		CoreFactorAttr: m.CoreFactorAttr,
		NumCoresAttr:   m.NumCoresAttr,
		NumCPUAttr:     m.NumCPUAttr,
	}
}

// User ...
type User struct {
	ID        string
	UserID    string
	UserCount int64
}

// MetricNUPConfig is a representation of oracle.nup.standard metric configuration
type MetricNUPConfig struct {
	ID                  string
	Name                string
	NumCoreAttr         string
	NumCPUAttr          string
	CoreFactorAttr      string
	StartEqType         string
	BaseEqType          string
	AggerateLevelEqType string
	EndEqType           string
	NumberOfUsers       uint32
	Transform           bool
	TransformMetricName string
}
