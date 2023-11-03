package v1

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetadataType of metadata
type MetadataType uint8
type DataType uint8

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
	ID     string
	Type   MetadataType
	Source string
	// Attributes
	//  example: headers of csv files
	Attributes []string
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

const (
	// DataTypeString for string
	DataTypeString DataType = 1
	// DataTypeInt for int
	DataTypeInt DataType = 2
	// DataTypeFloat for float
	DataTypeFloat DataType = 3
)

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

type MetadataOPS struct {
	ID                       string
	Name                     string
	Num_core_attr_id         string
	NumCPU_attr_id           string
	Core_factor_attr_id      string
	Start_eq_type_id         string
	Base_eq_type_id          string
	AggerateLevel_eq_type_id string
	End_eq_type_id           string
	Scopes                   []string
}

type MetadataNUP struct {
	ID                       string
	Name                     string
	Num_core_attr_id         string
	NumCPU_attr_id           string
	Core_factor_attr_id      string
	Start_eq_type_id         string
	Base_eq_type_id          string
	AggerateLevel_eq_type_id string
	End_eq_type_id           string
	Number_of_users          uint32
	Transform                bool
	Transform_metric_name    string
	Scopes                   []string
}

type MetadataINM struct {
	ID                 string
	Name               string
	Num_Of_Deployments int32
	Scopes             []string
}

type MetadataUSS struct {
	ID     string
	Name   string
	Scopes []string
}

type MetadataSS struct {
	ID        string
	Name      string
	Reference int64
	Scopes    []string
}
type MetadataSPS struct {
	ID                  string
	Name                string
	Num_core_attr_id    string
	NumCPU_attr_id      string
	Core_factor_attr_id string
	Base_eq_type_id     string
	Scopes              []string
}

type MetadataUNS struct {
	ID      string
	Name    string
	Profile string
	Scopes  []string
}

type MetadataSQL struct {
	ID         string
	MetricName string
	Reference  string
	Core       string
	CPU        string
	Default    bool
	Scopes     []string
}

type MetadataEquipAttr struct {
	ID             string
	Name           string
	Eq_type        string
	Attribute_name string
	Environment    string
	Value          int32
	Scopes         []string
}

type MetadataAttrSum struct {
	ID             string
	Name           string
	Eq_type        string
	Attribute_name string
	ReferenceValue float64
	Scopes         []string
}

type MetadataACS struct {
	ID             string
	Name           string
	Eq_type        string
	Attribute_name string
	Value          string
	Scopes         []string
}
type MetadataMetrics struct {
	MetadataOPS       MetadataOPS          `json:"MetadataOPS,omitempty"`
	MetadataNUP       MetadataNUP          `json:"MetadataNUP,omitempty"`
	MetadataINM       MetadataINM          `json:"MetadataINM,omitempty"`
	MetadataUSS       MetadataUSS          `json:"MetadataUSS,omitempty"`
	MetadataSPS       MetadataSPS          `json:"MetadataSPS,omitempty"`
	MetadataUNS       MetadataUNS          `json:"MetadataUNS,omitempty"`
	MetadataSQL       MetadataSQL          `json:"MetadataSQL,omitempty"`
	MetadataEquipAttr []*MetadataEquipAttr `json:"MetadataEquipAttr,omitempty"`
	MetadataSS        MetadataSS           `json:"MetadataSS,omitempty"`
	MetadataAttrSum   MetadataAttrSum      `json:"MetadataAttrSum,omitempty"`
	MetadataACS       MetadataACS          `json:"MetadataACS,omitempty"`
	Scopes            []string
}

func (a *Attribute) ValidateAttrValFromString(val string) error {
	switch a.Type {
	case DataTypeInt:
		_, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid value type - type should be int")
		}
		return nil
	case DataTypeFloat:
		_, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return status.Error(codes.InvalidArgument, "invalid value type - type should be float")
		}
		return nil
	case DataTypeString:
		return nil
	default:
		return status.Error(codes.InvalidArgument, "invalid value type")
	}
}

func GlobalMetricMetadata(scope string) map[string]MetadataMetrics {
	var scopes []string
	scopes = append(scopes, scope)
	data := map[string]MetadataMetrics{
		"oracle.processor.standard": {
			Scopes: scopes,
			MetadataOPS: MetadataOPS{

				ID:                       "",
				Name:                     "oracle.processor",
				Num_core_attr_id:         "cores_per_processor",
				NumCPU_attr_id:           "server_processors_numbers",
				Core_factor_attr_id:      "oracle_core_factor",
				Start_eq_type_id:         "virtualmachine",
				Base_eq_type_id:          "server",
				AggerateLevel_eq_type_id: "cluster",
				End_eq_type_id:           "vcenter",
				Scopes:                   scopes,
			},
		},
		"oracle.nup.standard": {
			Scopes: scopes,
			MetadataNUP: MetadataNUP{

				ID:                       "",
				Name:                     "oracle.nup",
				Num_core_attr_id:         "cores_per_processor",
				NumCPU_attr_id:           "server_processors_numbers",
				Core_factor_attr_id:      "oracle_core_factor",
				Start_eq_type_id:         "virtualmachine",
				Base_eq_type_id:          "server",
				AggerateLevel_eq_type_id: "cluster",
				End_eq_type_id:           "vcenter",
				Number_of_users:          5,
				Transform:                false,
				Transform_metric_name:    "",
				Scopes:                   scopes,
			},
		},
		"instance.number.standard": {
			Scopes: scopes,
			MetadataINM: MetadataINM{
				ID:                 "",
				Name:               "one_instance",
				Num_Of_Deployments: 1,
				Scopes:             scopes,
			},
		},
		"user.sum.standard": {
			Scopes: scopes,
			MetadataUSS: MetadataUSS{
				ID:     "",
				Name:   "user",
				Scopes: scopes,
			},
		},
		"sag.processor.standard": {
			Scopes: scopes,
			MetadataSPS: MetadataSPS{

				ID:                  "",
				Name:                "sag.processor",
				Num_core_attr_id:    "cores_per_processor",
				NumCPU_attr_id:      "server_processors_numbers",
				Core_factor_attr_id: "sag_uvu",
				Base_eq_type_id:     "server",
				Scopes:              scopes,
			},
		},
		"ibm.pvu.standard": {
			Scopes: scopes,
			MetadataSPS: MetadataSPS{

				ID:                  "",
				Name:                "ibm.pvu",
				Num_core_attr_id:    "cores_per_processor",
				NumCPU_attr_id:      "server_processors_numbers",
				Core_factor_attr_id: "ibm_pvu",
				Base_eq_type_id:     "server",
				Scopes:              scopes,
			},
		},
		"microsoft.sql.standard": {
			Scopes: scopes,
			MetadataSQL: MetadataSQL{

				ID:         "",
				MetricName: "microsoft.sql.standard.2019",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Default:    true,
				Scopes:     scopes,
			},
		},
		"microsoft.sql.enterprise": {
			Scopes: scopes,
			MetadataSQL: MetadataSQL{

				ID:         "",
				MetricName: "microsoft.sql.enterprise.2019",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Default:    true,
				Scopes:     scopes,
			},
		},
		"windows.server.standard": {
			Scopes: scopes,
			MetadataSQL: MetadataSQL{

				ID:         "",
				MetricName: "windows.server.standard.2016",
				Reference:  "server",
				Core:       "cores_per_processor",
				CPU:        "server_processors_numbers",
				Default:    true,
				Scopes:     scopes,
			},
		},
		"user.nominative.standard": {
			Scopes: scopes,
			MetadataUNS: MetadataUNS{
				ID:      "",
				Name:    "user.nominative.standard",
				Profile: "All",
				Scopes:  scopes,
			},
		},
		"user.concurrent.standard": {
			Scopes: scopes,
			MetadataUNS: MetadataUNS{
				ID:      "",
				Name:    "user.concurrent.standard",
				Profile: "All",
				Scopes:  scopes,
			},
		},
		"equipment.attribute.standard": {
			Scopes: scopes,
			MetadataEquipAttr: []*MetadataEquipAttr{
				{
					ID:             "",
					Name:           "openshift.premium",
					Eq_type:        "virtualmachine",
					Environment:    "production",
					Attribute_name: "vcpu",
					Value:          4,
					Scopes:         scopes,
				},
				{
					ID:             "",
					Name:           "openshift.standard",
					Eq_type:        "virtualmachine",
					Environment:    "development,test,preproduction",
					Attribute_name: "vcpu",
					Value:          4,
					Scopes:         scopes,
				},
			},
		},
		"static.standard": {
			Scopes: scopes,
			MetadataSS: MetadataSS{
				ID:        "",
				Name:      "static.standard",
				Reference: 1,
				Scopes:    scopes,
			},
		},
		"attribute.sum.standard": {
			Scopes: scopes,
			MetadataAttrSum: MetadataAttrSum{
				ID:             "",
				Name:           "attribute.sum",
				Eq_type:        "server",
				Attribute_name: "cores_per_processor",
				ReferenceValue: 8,
				Scopes:         scopes,
			},
		},
		"attribute.counter.standard": {
			Scopes: scopes,
			MetadataACS: MetadataACS{
				ID:             "",
				Name:           "attribute.counter",
				Eq_type:        "server",
				Attribute_name: "hyperthreading",
				Value:          "5",
				Scopes:         scopes,
			},
		},
		"windows.server.datacenter": {},
	}
	return data
}
