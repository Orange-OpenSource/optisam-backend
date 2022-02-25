package v1

import "errors"

// DataType for database
type DataType uint8

func (d DataType) String() string {
	switch d {
	case DataTypeInt:
		return "int"
	case DataTypeFloat:
		return "float"
	case DataTypeString:
		return "string"
	default:
		return "unsupported"
	}
}

const (
	// DataTypeString for string
	DataTypeString DataType = 1
	// DataTypeInt for int
	DataTypeInt DataType = 2
	// DataTypeFloat for float
	DataTypeFloat DataType = 3
)

// EquipmentProductSortBy - type defined for sorting fields of equipment products
type EquipmentProductSortBy uint8

const (
	// EquipmentProductSortBySwidTag - sorting by product swid tag
	EquipmentProductSortBySwidTag EquipmentProductSortBy = 0
	// EquipmentProductSortByName - sorting by product name
	EquipmentProductSortByName EquipmentProductSortBy = 1
	// EquipmentProductSortByEditor - sorting by product editor
	EquipmentProductSortByEditor EquipmentProductSortBy = 2
	// EquipmentProductSortByVersion - sorting by product version
	EquipmentProductSortByVersion EquipmentProductSortBy = 3
)

// EquipmentProductSearchKey - type defined for searching fields of equipment products
type EquipmentProductSearchKey string

func (e EquipmentProductSearchKey) String() string {
	return string(e)
}

const (
	// EquipmentProductSearchKeySwidTag - searching by product swid tag
	EquipmentProductSearchKeySwidTag EquipmentProductSearchKey = "swidtag"
	// EquipmentProductSearchKeyName - searching by product name
	EquipmentProductSearchKeyName EquipmentProductSearchKey = "name"
	// EquipmentProductSearchKeyEditor - searching by product editor
	EquipmentProductSearchKeyEditor EquipmentProductSearchKey = "editor"
	// EquipmentProductSearchKeyVersion - searching by product version
	EquipmentProductSearchKeyVersion EquipmentProductSearchKey = "release"
)

// EquipmentType for creating equipment type
type EquipmentType struct {
	ID         string
	Type       string
	SourceID   string
	SourceName string
	ParentID   string
	ParentType string
	Scopes     []string
	Attributes []*Attribute
}

// QueryEquipmentProduct has params to query products of an equipment
type QueryEquipmentProduct struct {
	PageSize  int32
	Offset    int32
	SortBy    EquipmentProductSortBy
	SortOrder SortOrder
	Filter    *AggregateFilter
}

// EquipmentProduct represents fields required for equipment
type EquipmentProduct struct {
	SwidTag string
	Name    string
	Editor  string
	Version string
}

// Equipment has generic infor mation about equipment an ancestors
type EquipmentInfo struct {
	ID      string
	EquipID string
	Type    string
	Parent  *EquipmentInfo
}

// PrimaryKeyAttribute returns primary key attribute of equipment type
func (e *EquipmentType) PrimaryKeyAttribute() (*Attribute, error) {
	for _, attr := range e.Attributes {
		if attr.IsIdentifier {
			return attr, nil
		}
	}
	return nil, errors.New("primary key attribute is not found")
}

// ParentKeyAttribute returns primary key attribute of equipment type
func (e *EquipmentType) ParentKeyAttribute() (*Attribute, error) {
	for _, attr := range e.Attributes {
		if attr.IsParentIdentifier {
			return attr, nil
		}
	}
	return nil, errors.New("primary key attribute is not found")
}

// QueryEquipments has parameters for query
type QueryEquipments struct {
	PageSize          int32
	Offset            int32
	SortBy            string
	SortOrder         SortOrder
	Filter            *AggregateFilter
	ProductFilter     *AggregateFilter
	ApplicationFilter *AggregateFilter
	InstanceFilter    *AggregateFilter
}

// UpdateEquipmentRequest ...
type UpdateEquipmentRequest struct {
	ParentID string
	Attr     []*Attribute
}

func GetGenericScopeEquipmentTypes(scope string) map[string]*EquipmentType { //nolint
	var scopes []string
	scopes = append(scopes, scope)
	data := map[string]*EquipmentType{
		"metadata_vcenter.csv": {
			SourceName: "metadata_vcenter.csv",
			Type:       "vcenter",
			Attributes: []*Attribute{
				{
					Name:         "vcenter_name",
					Type:         DataTypeString,
					MappedTo:     "vcenter_name",
					IsIdentifier: true,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "vcenter_version",
					Type:         DataTypeString,
					MappedTo:     "vcenter_version",
					IsDisplayed:  true,
					IsSearchable: true},
			},
			Scopes: scopes,
		},
		"metadata_cluster.csv": {
			SourceName: "metadata_cluster.csv",
			Type:       "cluster",
			Scopes:     scopes,
			ParentType: "vcenter",
			Attributes: []*Attribute{
				{
					Name:         "cluster_name",
					Type:         DataTypeString,
					MappedTo:     "cluster_name",
					IsIdentifier: true,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:               "parent_id",
					Type:               DataTypeString,
					MappedTo:           "parent_id",
					IsParentIdentifier: true,
					IsDisplayed:        true,
					IsSearchable:       true,
				},
			},
		},
		"metadata_server.csv": {
			SourceName: "metadata_server.csv",
			Scopes:     scopes,
			Type:       "server",
			ParentType: "cluster",
			Attributes: []*Attribute{
				{
					Name:         "hyperthreading",
					Type:         DataTypeString,
					MappedTo:     "hyperthreading",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "datacenter_name",
					Type:         DataTypeString,
					MappedTo:     "datacenter_name",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "server_id",
					Type:         DataTypeString,
					MappedTo:     "server_id",
					IsIdentifier: true,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "server_name",
					Type:         DataTypeString,
					MappedTo:     "server_name",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "cores_per_processor",
					Type:         DataTypeInt,
					MappedTo:     "cores_per_processor",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "oracle_core_factor",
					Type:         DataTypeFloat,
					MappedTo:     "oracle_core_factor",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "cpu_manufacturer",
					Type:         DataTypeString,
					MappedTo:     "cpu_manufacturer",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "ibm_pvu",
					Type:         DataTypeFloat,
					MappedTo:     "ibm_pvu",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "sag_uvu",
					Type:         DataTypeInt,
					MappedTo:     "sag_uvu",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "server_type",
					Type:         DataTypeString,
					MappedTo:     "server_type",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:               "parent_id",
					Type:               DataTypeString,
					MappedTo:           "parent_id",
					IsParentIdentifier: true,
					IsDisplayed:        true,
					IsSearchable:       true,
				},
				{
					Name:         "cpu_model",
					Type:         DataTypeString,
					MappedTo:     "cpu_model",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "server_os",
					Type:         DataTypeString,
					MappedTo:     "server_os",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "server_processors_numbers",
					Type:         DataTypeInt,
					MappedTo:     "server_processors_numbers",
					IsDisplayed:  true,
					IsSearchable: true,
				},
			},
		},
		"metadata_softpartition.csv": {
			SourceName: "metadata_softpartition.csv",
			Type:       "softpartition",
			Scopes:     scopes,
			ParentType: "server",
			Attributes: []*Attribute{
				{
					Name:         "softpartition_id",
					Type:         DataTypeString,
					MappedTo:     "softpartition_id",
					IsIdentifier: true,
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:         "softpartition_name",
					Type:         DataTypeString,
					MappedTo:     "softpartition_name",
					IsDisplayed:  true,
					IsSearchable: true,
				},
				{
					Name:               "parent_id",
					Type:               DataTypeString,
					MappedTo:           "parent_id",
					IsParentIdentifier: true,
					IsDisplayed:        true,
					IsSearchable:       true,
				},
			},
		},
		// "metadata_hardpartition.csv": {
		// 	SourceName: "metadata_hardpartition.csv",
		// 	Scopes:     scopes,
		// 	Type:       "hardpartition",
		// 	ParentType: "server",
		// 	Attributes: []*Attribute{
		// 		{
		// 			Name:         "aix_entitlement",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "aix_entitlement",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "sparc_cap",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "sparc_cap",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "aix_lpm",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "aix_lpm",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "aix_sharingmode",
		// 			Type:         DataTypeString,
		// 			MappedTo:     "aix_sharingmode",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "sparc_livemigration",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "sparc_livemigration",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "aix_sharedpool_cpus",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "aix_sharedpool_cpus",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "aix_onlinevirtualcores",
		// 			Type:         DataTypeInt,
		// 			MappedTo:     "aix_onlinevirtualcores",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:         "aix_processormode",
		// 			Type:         DataTypeString,
		// 			MappedTo:     "aix_processormode",
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 		{
		// 			Name:               "parent_id",
		// 			Type:               DataTypeString,
		// 			MappedTo:           "parent_id",
		// 			IsParentIdentifier: true,
		// 			IsDisplayed:        true,
		// 			IsSearchable:       true,
		// 		},
		// 		{
		// 			Name:         "hardpartition_id",
		// 			Type:         DataTypeString,
		// 			MappedTo:     "hardpartition_id",
		// 			IsIdentifier: true,
		// 			IsDisplayed:  true,
		// 			IsSearchable: true,
		// 		},
		// 	},
		// },
	}
	return data
}
