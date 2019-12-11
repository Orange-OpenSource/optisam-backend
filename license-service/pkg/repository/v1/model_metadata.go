// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
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
}
