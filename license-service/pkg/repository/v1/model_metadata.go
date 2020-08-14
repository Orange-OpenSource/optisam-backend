// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
	ID     string
	Type   MetadataType
	Source string
	// Attributes
	//  example: headers of csv files
	Attributes []string
}

// Attribute for attribute of data
type Attribute struct {
	ID                 string
	Name               string
	Type               DataType
	IsIdentifier       bool
	IsDisplayed        bool
	IsSearchable       bool
	IsParentIdentifier bool
	MappedTo           string
	IsSimulated        bool
	IntVal             int
	StringVal          string
	FloatVal           float32
	IntValOld          int
	StringValOld       string
	FloatValOld        float32
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

func (a *Attribute) ValFloat() float64 {
	if !a.IsSimulated {
		return a.ValFloatOld()
	}
	switch a.Type {
	case DataTypeInt:
		return float64(a.IntVal)
	case DataTypeFloat:
		return float64(a.FloatVal)
	default:
		return 0
	}
}

func (a *Attribute) ValFloatOld() float64 {
	switch a.Type {
	case DataTypeInt:
		return float64(a.IntValOld)
	case DataTypeFloat:
		return float64(a.FloatValOld)
	default:
		return 0
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
