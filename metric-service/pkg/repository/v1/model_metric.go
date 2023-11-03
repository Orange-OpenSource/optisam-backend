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
	// MetricAttrSumStandard is attribute.sum.standard
	MetricAttrSumStandard MetricType = "attribute.sum.standard"
	// MetricUserSumStandard is user.sum.standard
	MetricUserSumStandard MetricType = "user.sum.standard"
	// MetricStaticStandard is static.standard
	MetricStaticStandard MetricType = "static.standard"
	// MetricEquipAttrStandard is equipment.attribute.standard
	MetricEquipAttrStandard MetricType = "equipment.attribute.standard"
	// MetricUserNomStandard is user.nominative.standard
	MetricUserNomStandard MetricType = "user.nominative.standard"
	// MetricUserConcurentStandard is user.concurrent.standard
	MetricUserConcurentStandard MetricType = "user.concurrent.standard"
	//MetricMicrosoftSQLEnterprise is microsoft.sql.enterprise
	MetricMicrosoftSQLEnterprise MetricType = "microsoft.sql.enterprise"
	//MetricWindowsServerDataCenter is windows.server.datacenter
	MetricWindowsServerDataCenter MetricType = "windows.server.datacenter"
	//MetricWindowsServerStandard is windows.server.standard
	MetricWindowsServerStandard MetricType = "windows.server.standard"
	//MetricMicrosoftSQLEnterprise is microsoft.sql.standard
	MetricMicrosoftSQLStandard MetricType = "microsoft.sql.standard"
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
	MetricSQLEnterprise   MetricTypeID = 13
	MetricWSDCenter       MetricTypeID = 14
	MetricSQLStandard     MetricTypeID = 15
	MetricWindowStandard  MetricTypeID = 16
)

// MetricDescription provide description
type MetricDescription string

func (m MetricDescription) String() string {
	return string(m)
}

const (
	// MetricDescriptionOracleProcessorStandard provides description of oracle.processor.standard
	MetricDescriptionOracleProcessorStandard MetricDescription = "Number of licenses required = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionSagProcessorStandard provides description of sag.processor.standard
	MetricDescriptionSagProcessorStandard MetricDescription = "Number of licenses required = MAX(Prod_licenses, NonProd_licenses) : licenses = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionIbmPvuStandard provides description of ibm.pvu.standard
	MetricDescriptionIbmPvuStandard MetricDescription = "Number of licenses required = CPU nb x Core(per CPU) nb x CoreFactor"
	// MetricDescriptionOracleNUPStandard provides description of oracle.nup.standard
	MetricDescriptionOracleNUPStandard MetricDescription = "Number Of licenses required = MAX(CPU nb x Core(per CPU) nb x CoreFactor x given_users, given number of users)"
	// MetricDescriptionAttrCounterStandard provides description of attribute.counter.standard
	MetricDescriptionAttrCounterStandard MetricDescription = "Number of licenses required = Number of equipment of type specific_type with specific_attribute = value."
	// MetricDescriptionAttrCounterStandard provides description of instance.counter.standard
	MetricDescriptionInstanceNumberStandard MetricDescription = "Number of licenses required = Sum of product installations / number_of_deployments_authorized_licenses"
	// MetricDescriptionAttrSumStandard provides description of attribute.sum.standard
	MetricDescriptionAttrSumStandard MetricDescription = "Number of licenses required = Ceil( Sum( on all equipments of type Equipment_type) of attribute_value)/ Reference_value"
	// MetricDescriptionAttrCounterStandard provides description of user.sum.standard
	MetricDescriptionUserSumStandard MetricDescription = "Number of licenses required = Sum of all users using the product."
	// MetricDescriptionStaticStandard provides description of static.standard
	MetricDescriptionStaticStandard MetricDescription = "Number of licenses required = Reference_value"
	// MetricDescriptionEquipAttrStandard provides description of static.standard
	MetricDescriptionEquipAttrStandard MetricDescription = "Number of licenses required = SUM (each [equipment] on environments [environment]) / [number] [attribute]"

	// MetricDescriptionUserNomStandard provides description of user.nominative.standard
	MetricDescriptionUserNomStandard MetricDescription = "sum the number of users with profile = [profile]"
	// MetricDescriptionUserConcurentStandard provides description of user.concurrent.standard
	MetricDescriptionUserConcurentStandard MetricDescription = "latest number of users with profile = [profile]"
	// MetricDescriptionMicrosoftSQLEnterprise
	MetricDescriptionMicrosoftSQLEnterprise MetricDescription = "Number of licenses required = CPU nb x MAX(Core of the server(per CPU), 4)"
	// MetricDescriptionWindowsServerDataCenter
	MetricDescriptionWindowsServerDataCenter MetricDescription = "Number of licenses required = MAX(MAX(cores per processor,8) * number of processors, 16))"
	// MetricDescriptionWindowsServerStandard provides description for window.server.standard
	MetricDescriptionWindowsServerStandard MetricDescription = "Number of licenses required = MAX(MAX(cores per processor,8) * number of processors, 16))"
	// MetricDescriptionMicrosoftSQLStandard provides description for microsoft.sql.standard
	MetricDescriptionMicrosoftSQLStandard MetricDescription = "Number of licenses required = CPU nb x MAX(Core of the server(per CPU), 4)"
)

var (
	// MetricTypesAll is a slice of MetricTypeInfo for all scopes
	MetricTypesAll = []*MetricTypeInfo{
		{
			Name:        MetricOPSOracleProcessorStandard,
			Description: MetricDescriptionOracleProcessorStandard.String(),
			Href:        "/api/v1/metric/ops",
			MetricType:  MetricOracleProcessor,
			Exist:       false,
		},
		{
			Name:        MetricSPSSagProcessorStandard,
			Description: MetricDescriptionSagProcessorStandard.String(),
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
			Exist:       false,
		},
		{
			Name:        MetricIPSIbmPvuStandard,
			Description: MetricDescriptionIbmPvuStandard.String(),
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
			Exist:       false,
		},
		{
			Name:        MetricOracleNUPStandard,
			Description: MetricDescriptionOracleNUPStandard.String(),
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
			Exist:       false,
		},
		{
			Name:        MetricAttrCounterStandard,
			Description: MetricDescriptionAttrCounterStandard.String(),
			Href:        "/api/v1/metric/acs",
			MetricType:  MetricAttrCounter,
			Exist:       false,
		},
		{
			Name:        MetricInstanceNumberStandard,
			Description: MetricDescriptionInstanceNumberStandard.String(),
			Href:        "/api/v1/metric/inm",
			MetricType:  MetricInstanceNumber,
			Exist:       false,
		},
		{
			Name:        MetricAttrSumStandard,
			Description: MetricDescriptionAttrSumStandard.String(),
			Href:        "/api/v1/metric/attr_sum",
			MetricType:  MetricAttrSum,
			Exist:       false,
		},
		{
			Name:        MetricUserSumStandard,
			Description: MetricDescriptionUserSumStandard.String(),
			Href:        "/api/v1/metric/uss",
			MetricType:  MetricUserSum,
			Exist:       false,
		},
		{
			Name:        MetricStaticStandard,
			Description: MetricDescriptionStaticStandard.String(),
			Href:        "/api/v1/metric/ss",
			MetricType:  MetricStatic,
			Exist:       false,
		},
		{
			Name:        MetricEquipAttrStandard,
			Description: MetricDescriptionEquipAttrStandard.String(),
			Href:        "/api/v1/metric/equip_attr",
			MetricType:  MetricEquipAttr,
			Exist:       false,
		},
		{
			Name:        MetricUserNomStandard,
			Description: MetricDescriptionUserNomStandard.String(),
			Href:        "/api/v1/metric/uns",
			MetricType:  MetricNominativeUser,
			Exist:       false,
		},
		{
			Name:        MetricUserConcurentStandard,
			Description: MetricDescriptionUserConcurentStandard.String(),
			Href:        "/api/v1/metric/user_conc",
			MetricType:  MetricConcurentUser,
			Exist:       false,
		},
		// {
		// 	Name:        MetricMicrosoftSQLEnterprise,
		// 	Description: MetricDescriptionMicrosoftSQLEnterprise.String(),
		// 	Href:        "",
		// 	MetricType:  MetricSQLEnterprise,
		// },
		// {
		// 	Name:        MetricWindowsServerDataCenter,
		// 	Description: MetricWindowsServerDataCenter.String(),
		// 	Href:        "",
		// 	MetricType:  MetricWSDCenter,
		// },
		// {
		// 	Name:        MetricMicrosoftSQLStandard,
		// 	Description: MetricDescriptionMicrosoftSQLStandard.String(),
		// 	Href:        "/api/v1/metric/sql_standard",
		// 	MetricType:  MetricSQLStandard,
		// },
		// {
		// 	Name:        MetricWindowsServerStandard,
		// 	Description: MetricWindowsServerStandard.String(),
		// 	Href:        "/api/v1/metric/wind_serv_stand",
		// 	MetricType:  MetricWindowStandard,
		// },
	}
	// MetricTypesGeneric is a slice of MetricTypeInfo for generic scopes
	MetricTypesGeneric = []*MetricTypeInfo{
		{
			Name:        MetricOPSOracleProcessorStandard,
			Description: MetricDescriptionOracleProcessorStandard.String(),
			Href:        "/api/v1/metric/ops",
			MetricType:  MetricOracleProcessor,
			Exist:       false,
		},
		{
			Name:        MetricSPSSagProcessorStandard,
			Description: MetricDescriptionSagProcessorStandard.String(),
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
			Exist:       false,
		},
		{
			Name:        MetricIPSIbmPvuStandard,
			Description: MetricDescriptionIbmPvuStandard.String(),
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
			Exist:       false,
		},
		{
			Name:        MetricOracleNUPStandard,
			Description: MetricDescriptionOracleNUPStandard.String(),
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
			Exist:       false,
		},
		{
			Name:        MetricAttrCounterStandard,
			Description: MetricDescriptionAttrCounterStandard.String(),
			Href:        "/api/v1/metric/acs",
			MetricType:  MetricAttrCounter,
			Exist:       false,
		},
		{
			Name:        MetricInstanceNumberStandard,
			Description: MetricDescriptionInstanceNumberStandard.String(),
			Href:        "/api/v1/metric/inm",
			MetricType:  MetricInstanceNumber,
			Exist:       false,
		},
		{
			Name:        MetricAttrSumStandard,
			Description: MetricDescriptionAttrSumStandard.String(),
			Href:        "/api/v1/metric/attr_sum",
			MetricType:  MetricAttrSum,
			Exist:       false,
		},
		{
			Name:        MetricUserSumStandard,
			Description: MetricDescriptionUserSumStandard.String(),
			Href:        "/api/v1/metric/uss",
			MetricType:  MetricUserSum,
			Exist:       false,
		},
		{
			Name:        MetricStaticStandard,
			Description: MetricDescriptionStaticStandard.String(),
			Href:        "/api/v1/metric/ss",
			MetricType:  MetricStatic,
			Exist:       false,
		},
		{
			Name:        MetricEquipAttrStandard,
			Description: MetricDescriptionEquipAttrStandard.String(),
			Href:        "/api/v1/metric/equip_attr",
			MetricType:  MetricEquipAttr,
			Exist:       false,
		},
		{
			Name:        MetricUserNomStandard,
			Description: MetricDescriptionUserNomStandard.String(),
			Href:        "/api/v1/metric/uns",
			MetricType:  MetricNominativeUser,
			Exist:       false,
		},
		{
			Name:        MetricUserConcurentStandard,
			Description: MetricDescriptionUserConcurentStandard.String(),
			Href:        "/api/v1/metric/user_conc",
			MetricType:  MetricConcurentUser,
			Exist:       false,
		},
		// {
		// 	Name:        MetricMicrosoftSQLStandard,
		// 	Description: MetricDescriptionMicrosoftSQLStandard.String(),
		// 	Href:        "/api/v1/metric/sql_standard",
		// 	MetricType:  MetricSQLStandard,
		// },
		// {
		// 	Name:        MetricWindowsServerStandard,
		// 	Description: MetricWindowsServerStandard.String(),
		// 	Href:        "/api/v1/metric/wind_serv_stand",
		// 	MetricType:  MetricWindowStandard,
		// },
	}
	// MetricTypesSpecific is a slice of MetricTypeInfo for specific scopes
	MetricTypesSpecific = []*MetricTypeInfo{
		{
			Name:        MetricOPSOracleProcessorStandard,
			Description: MetricDescriptionOracleProcessorStandard.String(),
			Href:        "/api/v1/metric/ops",
			MetricType:  MetricOracleProcessor,
			Exist:       false,
		},
		{
			Name:        MetricSPSSagProcessorStandard,
			Description: MetricDescriptionSagProcessorStandard.String(),
			Href:        "/api/v1/metric/sps",
			MetricType:  MetricSAGProcessor,
			Exist:       false,
		},
		{
			Name:        MetricIPSIbmPvuStandard,
			Description: MetricDescriptionIbmPvuStandard.String(),
			Href:        "/api/v1/metric/ips",
			MetricType:  MetricIBMPVU,
			Exist:       false,
		},
		{
			Name:        MetricOracleNUPStandard,
			Description: MetricDescriptionOracleNUPStandard.String(),
			Href:        "/api/v1/metric/oracle_nup",
			MetricType:  MetricOracleNUP,
			Exist:       false,
		},
		{
			Name:        MetricAttrCounterStandard,
			Description: MetricDescriptionAttrCounterStandard.String(),
			Href:        "/api/v1/metric/acs",
			MetricType:  MetricAttrCounter,
			Exist:       false,
		},
		{
			Name:        MetricInstanceNumberStandard,
			Description: MetricDescriptionInstanceNumberStandard.String(),
			Href:        "/api/v1/metric/inm",
			MetricType:  MetricInstanceNumber,
			Exist:       false,
		},
		{
			Name:        MetricAttrSumStandard,
			Description: MetricDescriptionAttrSumStandard.String(),
			Href:        "/api/v1/metric/attr_sum",
			MetricType:  MetricAttrSum,
			Exist:       false,
		},
		{
			Name:        MetricStaticStandard,
			Description: MetricDescriptionStaticStandard.String(),
			Href:        "/api/v1/metric/ss",
			MetricType:  MetricStatic,
			Exist:       false,
		},
		{
			Name:        MetricEquipAttrStandard,
			Description: MetricDescriptionEquipAttrStandard.String(),
			Href:        "/api/v1/metric/equip_attr",
			MetricType:  MetricEquipAttr,
			Exist:       false,
		},
		{
			Name:        MetricUserNomStandard,
			Description: MetricDescriptionUserNomStandard.String(),
			Href:        "/api/v1/metric/uns",
			MetricType:  MetricNominativeUser,
			Exist:       false,
		},
		{
			Name:        MetricUserConcurentStandard,
			Description: MetricDescriptionUserConcurentStandard.String(),
			Href:        "/api/v1/metric/user_conc",
			MetricType:  MetricConcurentUser,
			Exist:       false,
		},
		// {
		// 	Name:        MetricMicrosoftSQLStandard,
		// 	Description: MetricDescriptionMicrosoftSQLStandard.String(),
		// 	Href:        "/api/v1/metric/sql_standard",
		// 	MetricType:  MetricSQLStandard,
		// },
		// {
		// 	Name:        MetricWindowsServerStandard,
		// 	Description: MetricWindowsServerStandard.String(),
		// 	Href:        "/api/v1/metric/wind_serv_stand",
		// 	MetricType:  MetricWindowStandard,
		// },
	}
)

// MetricTypeInfo has name and description of MetricType
type MetricTypeInfo struct {
	Name        MetricType
	Description string
	Href        string
	MetricType  MetricTypeID
	Exist       bool
}

// Metric contains name and metric of the metrics
type MetricInfo struct {
	ID      string
	Name    string
	Type    MetricType
	Default bool
}

// MetricInfoFull contains metric info with linking of aggregation and acqrights
type MetricInfoFull struct {
	ID                string
	Name              string
	Type              MetricType
	Default           bool
	TotalAggregations int32
	TotalAcqRights    int32
}

// ScopeType is the types of scopes available in optisam
type ScopeType string

func (st ScopeType) String() string {
	return string(st)
}

func GetScopeType(st string) ScopeType {
	switch st {
	case "GENERIC":
		return ScopeTypeGeneric
	case "SPECIFIC":
		return ScopeTypeSpecific
	default:
		return ScopeTypeGeneric
	}
}

const (
	ScopeTypeGeneric  ScopeType = "GENERIC"
	ScopeTypeSpecific ScopeType = "SPECIFIC"
)

func (st ScopeType) ListMetricTypes(flag bool) []*MetricTypeInfo {
	newMetrics := []*MetricTypeInfo{
		{
			Name:        MetricMicrosoftSQLEnterprise,
			Description: MetricDescriptionMicrosoftSQLEnterprise.String(),
			Href:        "",
			MetricType:  MetricSQLEnterprise,
			Exist:       false,
		},
		{
			Name:        MetricWindowsServerDataCenter,
			Description: MetricDescriptionWindowsServerDataCenter.String(),
			Href:        "",
			MetricType:  MetricWSDCenter,
			Exist:       false,
		},
		{
			Name:        MetricMicrosoftSQLStandard,
			Description: MetricDescriptionMicrosoftSQLStandard.String(),
			Href:        "/api/v1/metric/sql_standard",
			MetricType:  MetricSQLStandard,
			Exist:       false,
		},
		{
			Name:        MetricWindowsServerStandard,
			Description: MetricDescriptionWindowsServerStandard.String(),
			Href:        "/api/v1/metric/wind_serv_stand",
			MetricType:  MetricWindowStandard,
			Exist:       false,
		},
	}
	switch st {
	case ScopeTypeGeneric:
		var MetricGeneric []*MetricTypeInfo
		if flag {
			MetricGeneric = append(MetricTypesGeneric, newMetrics...)
		} else {
			return MetricTypesGeneric
		}
		return MetricGeneric
	case ScopeTypeSpecific:
		var MetricSpecific []*MetricTypeInfo
		if flag {
			MetricSpecific = append(MetricTypesSpecific, newMetrics...)
		} else {
			return MetricTypesSpecific
		}
		return MetricSpecific
	default:
		var MetricAll []*MetricTypeInfo
		if flag {
			MetricAll = append(MetricTypesAll, newMetrics...)
		} else {
			return MetricTypesAll
		}
		return MetricAll
	}
}
