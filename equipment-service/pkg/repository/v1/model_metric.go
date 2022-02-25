package v1

// MetricType is an alias for string
type MetricType string

const (
	// MetricOPSOracleProcessorStandard is oracle.processor.standard
	MetricOPSOracleProcessorStandard MetricType = "oracle.processor.standard"
	// MetricSPSSagProcessorStandard is sag.processor.standard
	MetricSPSSagProcessorStandard MetricType = "sag.processor.standard"
	// MetricIPSIbmPvuStandard is ibm.pvu.standard
	MetricIPSIbmPvuStandard MetricType = "ibm.pvu.standard"
	// MetricOracleNUPStandard is oracle.nup.standard
	MetricOracleNUPStandard MetricType = "oracle.nup.standard"
)

// String implements Stringer interface
func (m MetricType) String() string {
	return string(m)
}

// MetricSearchKey is type to represent search keys string
type MetricSearchKey string

const (
	// MetricSearchKeyName ...
	MetricSearchKeyName MetricSearchKey = "Name"
)

// String implements Stringer interface
func (m MetricSearchKey) String() string {
	return string(m)
}

// MetricTypeID is an alias for int
type MetricTypeID int

const (
	MetricUnknown         MetricTypeID = 0
	MetricOracleProcessor MetricTypeID = 1
	MetricOracleNUP       MetricTypeID = 2
	MetricSAGProcessor    MetricTypeID = 3
	MetricIBMPVU          MetricTypeID = 4
)

// MetricDescription provide description
type MetricDescription string

func (m MetricDescription) String() string {
	return string(m)
}

const (
	// MetricDescriptionOracleProcessorStandard provides description of oracle.processor.standard
	MetricDescriptionOracleProcessorStandard MetricDescription = "xyz"
	// MetricDescriptionSagProcessorStandard provides description of sag.processor.standard
	MetricDescriptionSagProcessorStandard MetricDescription = "abc"
	// MetricDescriptionIbmPvuStandard provides description of ibm.pvu.standard
	MetricDescriptionIbmPvuStandard MetricDescription = "pqr"
	// MetricDescriptionOracleNUPStandard provides description of oracle.nup.standard
	MetricDescriptionOracleNUPStandard MetricDescription = "uvw"
)

var (
	// MetricTypes is a slice of MetricTypeInfo
	MetricTypes = []*MetricTypeInfo{
		{
			Name:        MetricOPSOracleProcessorStandard,
			Description: MetricDescriptionOracleProcessorStandard.String(),
			Href:        "/api/v1/metric/ops",
			MetricType:  MetricOracleProcessor,
		},
		{
			Name:        MetricSPSSagProcessorStandard,
			Description: "abc",
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
		},
		{
			Name:        MetricIPSIbmPvuStandard,
			Description: "pqr",
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
		},
		{
			Name:        MetricOracleNUPStandard,
			Description: "uvw",
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
		},
	}
)

// MetricTypeInfo has name and description of MetricType
type MetricTypeInfo struct {
	Name        MetricType
	Description string
	Href        string
	MetricType  MetricTypeID
}

// Metric contains name and metric of the metrics
type Metric struct {
	ID   string
	Name string
	Type MetricType
}
