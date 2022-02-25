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
