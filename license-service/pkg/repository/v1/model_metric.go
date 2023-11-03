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
	// MetricInstanceNumberStandard is instance.number.metric
	MetricInstanceNumberStandard MetricType = "instance.number.standard"
	// MetricStaticStandard is static.metric
	MetricStaticStandard MetricType = "static.standard"
	// MetricAttrSumStandard is attribute.sum.standard
	MetricAttrSumStandard MetricType = "attribute.sum.standard"
	// MetricUserSumStandard is user.sum.standard
	MetricUserSumStandard MetricType = "user.sum.standard"
	// MetricEquipAttrStandard is equipment.attribute.standard
	MetricEquipAttrStandard MetricType = "equipment.attribute.standard"
	// MetricUserNomStandard is user.nominative.standard
	MetricUserNomStandard MetricType = "user.nominative.standard"
	// MetricUserConcurentStandard is user.concurrent.standard
	MetricUserConcurentStandard MetricType = "user.concurrent.standard"
	// MetricMicrosoftSqlEnterprise is microsoft.sql.enterprise
	MetricMicrosoftSqlEnterprise MetricType = "microsoft.sql.enterprise"
	// MetricWindowsServerDataCenter is windows.server.datacenter
	MetricWindowsServerDataCenter MetricType = "windows.server.datacenter"
	// MetricMicrosoftSqlStandard is microsoft.sql.standard
	MetricMicrosoftSqlStandard MetricType = "microsoft.sql.standard"
	// MetricWindowsServerStandard is windows.server.standard
	MetricWindowsServerStandard MetricType = "windows.server.standard"
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
	MetricAttrCounter     MetricTypeID = 5
	MetricInstanceNumber  MetricTypeID = 6
	MetricAttrSum         MetricTypeID = 7
	MetricUserSum         MetricTypeID = 8
	MetricStatic          MetricTypeID = 9
	MetricEquipAttr       MetricTypeID = 10
	MetricNominativeUser  MetricTypeID = 11
	MetricConcurentUser   MetricTypeID = 12
	MetricMicrosoftSE     MetricTypeID = 13
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
	MetricDescriptionAttrCounterStandard MetricDescription = "Number of licenses required = Number of equipment of a specific type containing a specific attribute set to a specific value."
	// MetricDescriptionInstanceNumberStandard provides description of attribute.counter.standard
	MetricDescriptionInstanceNumberStandard MetricDescription = "Number of licenses required = Number of instances where product has been installed divided by the number of deployments authorized per license."
	// MetricDescriptionAttrSumStandard provides description of attribute.sum.standard
	MetricDescriptionAttrSumStandard MetricDescription = "Number of licenses required = Ceil( Sum( on all equipments of the chosen Equipment type) of the attribute values)/Reference value (present in the metric)"
	// MetricDescriptionAttrCounterStandard provides description of user.sum.standard
	MetricDescriptionUserSumStandard MetricDescription = "Number of licenses required = Sum of all users using the product."
	// MetricDescriptionStaticStandard provides description of static.standard
	MetricDescriptionStaticStandard MetricDescription = "Number of licenses required = Reference Value."
	// MetricDescriptionUserNomStandard provides description of user.nominative.standard
	MetricDescriptionUserNomStandard MetricDescription = "sum the number of users with profile = [profile]"
	// MetricDescriptionUserConcurentStandard provides description of user.concurrent.standard
	MetricDescriptionUserConcurentStandard MetricDescription = "latest number of users with profile = [profile]"
	// MetricDescriptionMicrosoftSqlEnterprise provides description of microsoft.sql.enterprise
	MetricDescriptionMicrosoftSqlEnterprise MetricDescription = "Number of licenses required = CPU nb x MAX(Core of the server(per CPU), 4)"
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
			Description: MetricDescriptionSagProcessorStandard.String(),
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
		},
		{
			Name:        MetricIPSIbmPvuStandard,
			Description: MetricDescriptionIbmPvuStandard.String(),
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
		},
		{
			Name:        MetricOracleNUPStandard,
			Description: MetricDescriptionOracleNUPStandard.String(),
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
		},
		{
			Name:        MetricAttrCounterStandard,
			Description: MetricDescriptionAttrCounterStandard.String(),
			Href:        "/api/v1/metric/acs",
			MetricType:  MetricAttrCounter,
		},
		{
			Name:        MetricInstanceNumberStandard,
			Description: MetricDescriptionInstanceNumberStandard.String(),
			Href:        "/api/v1/metric/inm",
			MetricType:  MetricInstanceNumber,
		},
		{
			Name:        MetricAttrSumStandard,
			Description: MetricDescriptionAttrSumStandard.String(),
			Href:        "/api/v1/metric/attr_sum",
			MetricType:  MetricAttrSum,
		},
		{
			Name:        MetricUserSumStandard,
			Description: MetricDescriptionUserSumStandard.String(),
			Href:        "/api/v1/metric/uss",
			MetricType:  MetricUserSum,
		},
		{
			Name:        MetricStaticStandard,
			Description: MetricDescriptionStaticStandard.String(),
			Href:        "/api/v1/metric/ss",
			MetricType:  MetricStatic,
		},
		{
			Name:        MetricUserNomStandard,
			Description: MetricDescriptionUserNomStandard.String(),
			Href:        "/api/v1/metric/uns",
			MetricType:  MetricNominativeUser,
		},
		{
			Name:        MetricUserConcurentStandard,
			Description: MetricDescriptionUserConcurentStandard.String(),
			Href:        "/api/v1/metric/user_conc",
			MetricType:  MetricConcurentUser,
		},
		{
			Name:        MetricMicrosoftSqlEnterprise,
			Description: MetricDescriptionMicrosoftSqlEnterprise.String(),
			Href:        "",
			MetricType:  MetricMicrosoftSE,
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
