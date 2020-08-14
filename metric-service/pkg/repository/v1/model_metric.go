// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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
	// MetricAttrCounterStandard is attribute.counter.standard
	MetricAttrCounterStandard MetricType = "attribute.counter.standard"
	//MetricInstanceNumberStandard is instance.number.metric
	MetricInstanceNumberStandard MetricType = "instance.number.standard"
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

// MetricTypeId is an alias for int
type MetricTypeId int

const (
	MetricUnknown         MetricTypeId = 0
	MetricOracleProcessor MetricTypeId = 1
	MetricOracleNUP       MetricTypeId = 2
	MetricSAGProcessor    MetricTypeId = 3
	MetricIBMPVU          MetricTypeId = 4
	MetricAttrCounter     MetricTypeId = 5
	MetricInstanceNumber  MetricTypeId = 6
)

// MetricDescription provide description
type MetricDescription string

func (m MetricDescription) String() string {
	return string(m)
}

const (
	// MetricDescriptionOracleProcessorStandard provides description of oracle.processor.standard
	MetricDescriptionOracleProcessorStandard MetricDescription = "Number of processor licenses required = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionSagProcessorStandard provides description of sag.processor.standard
	MetricDescriptionSagProcessorStandard MetricDescription = "Number of processor licenses required = MAX(Prod_licenses, NonProd_licenses) : licenses = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionIbmPvuStandard provides description of ibm.pvu.standard
	MetricDescriptionIbmPvuStandard MetricDescription = "Number of licenses required = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionOracleNUPStandard provides description of oracle.nup.standard
	MetricDescriptionOracleNUPStandard MetricDescription = "Named User Plus licenses required = MAX(A,B) : A = CPU nb x Core(per CPU) nb x CoreFactor x minimum number of NUP per processor, B = total number of current users with access to the Oracle product"
	// MetricDescriptionAttrCounterStandard provides description of attribute.counter.standard
	MetricDescriptionAttrCounterStandard MetricDescription = "Number of licenses required = Number of equipment of a specific type containing a specific atribute set to a specific value."

	// MetricDescriptionAttrCounterStandard provides description of attribute.counter.standard
	MetricDescriptionInstanceNumberStandard MetricDescription = "Number Of instances where product has been installed multiply by a cofficent ,where instances are links between product and equipment(of any kind)"
)

var (
	// MetricTypes is a slice of MetricTypeInfo
	MetricTypes = []*MetricTypeInfo{
		&MetricTypeInfo{
			Name:        MetricOPSOracleProcessorStandard,
			Description: MetricDescriptionOracleProcessorStandard.String(),
			Href:        "/api/v1/metric/ops",
			MetricType:  MetricOracleProcessor,
		},
		&MetricTypeInfo{
			Name:        MetricSPSSagProcessorStandard,
			Description: MetricDescriptionSagProcessorStandard.String(),
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
		},
		&MetricTypeInfo{
			Name:        MetricIPSIbmPvuStandard,
			Description: MetricDescriptionIbmPvuStandard.String(),
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
		},
		&MetricTypeInfo{
			Name:        MetricOracleNUPStandard,
			Description: MetricDescriptionOracleNUPStandard.String(),
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
		},
		&MetricTypeInfo{
			Name:        MetricAttrCounterStandard,
			Description: MetricDescriptionAttrCounterStandard.String(),
			Href:        "/api/v1/metric/acs",
			MetricType:  MetricAttrCounter,
		},
		&MetricTypeInfo{
			Name:        MetricInstanceNumberStandard,
			Description: MetricDescriptionInstanceNumberStandard.String(),
			Href:        "/api/v1/metric/inm",
			MetricType:  MetricInstanceNumber,
		},
	}
)

// MetricTypeInfo has name and description of MetricType
type MetricTypeInfo struct {
	Name        MetricType
	Description string
	Href        string
	MetricType  MetricTypeId
}

// Metric contains name and metric of the metrics
type MetricInfo struct {
	ID   string
	Name string
	Type MetricType
}
