// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

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
	return nil, errors.New("Primary key attribute is not found")
}

// ParentKeyAttribute returns primary key attribute of equipment type
func (e *EquipmentType) ParentKeyAttribute() (*Attribute, error) {
	for _, attr := range e.Attributes {
		if attr.IsParentIdentifier {
			return attr, nil
		}
	}
	return nil, errors.New("Primary key attribute is not found")
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
