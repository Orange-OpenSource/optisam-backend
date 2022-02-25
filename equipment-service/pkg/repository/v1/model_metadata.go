package v1

// MetadataType of metadata
type MetadataType uint8

const (
	// MetadataTypeProduct is for product
	MetadataTypeProduct MetadataType = 0
	// MetadataTypeApplication is for application
	MetadataTypeApplication MetadataType = 1
	// MetadataTypeInstance is for instance
	MetadataTypeInstance MetadataType = 2
	// MetadataTypeEquipment is for equipment
	MetadataTypeEquipment MetadataType = 3
	// MetadataTypeMetadata is for metadata
	MetadataTypeMetadata MetadataType = 4
)

// Metadata for injectors
type Metadata struct {
	ID           string
	Type         MetadataType
	Source       string
	MetadataType string
	Scope        string
	Attributes   []string
}

// Attribute for attribute of data
type Attribute struct {
	Type               DataType
	IsIdentifier       bool
	IsDisplayed        bool
	IsSearchable       bool
	IsParentIdentifier bool
	IsSimulated        bool
	IntVal             int
	IntValOld          int
	FloatVal           float32
	FloatValOld        float32
	ID                 string
	Name               string
	MappedTo           string
	StringVal          string
	StringValOld       string
}

func (a *Attribute) Val() interface{} {
	switch a.Type {
	case DataTypeInt:
		return a.IntVal
	case DataTypeFloat:
		return a.FloatVal
	case DataTypeString:
		return a.StringVal
	default:
		return a.StringVal
	}
}

func GetGenericScopeMetadata(scope string) []Metadata {
	resp := []Metadata{
		{
			MetadataType: "equipment",
			Source:       "metadata_vcenter.csv",
			Attributes:   []string{"vcenter_name", "vcenter_version"},
			Scope:        scope,
		},
		{
			MetadataType: "equipment",
			Source:       "metadata_cluster.csv",
			Attributes:   []string{"cluster_name", "parent_id"},
			Scope:        scope,
		},
		{
			MetadataType: "equipment",
			Source:       "metadata_server.csv",
			Attributes:   []string{"hyperthreading", "datacenter_name", "server_id", "server_name", "cores_per_processor", "oracle_core_factor", "cpu_manufacturer", "ibm_pvu", "sag_uvu", "server_type", "parent_id", "cpu_model", "server_os", "server_processors_numbers"},
			Scope:        scope,
		},
		{
			MetadataType: "equipment",
			Source:       "metadata_softpartition.csv",
			Attributes:   []string{"softpartition_id", "softpartition_name", "parent_id"},
			Scope:        scope,
		},
		// {
		// 	MetadataType: "equipment",
		// 	Source:       "metadata_hardpartition.csv",
		// 	Attributes:   []string{"aix_entitlement", "sparc_cap", "aix_lpm", "aix_sharingmode", "sparc_livemigration", "aix_sharedpool_cpus", "aix_onlinevirtualcores", "aix_processormode", "parent_id", "hardpartition_id"},
		// 	Scope:        scope,
		// },
	}

	return resp
}
