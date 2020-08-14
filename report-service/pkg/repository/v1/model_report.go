// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

//ProductEquipment ...
type ProductEquipment struct {
	EquipmentID   string
	EquipmentType string
}

//EquipmentAttributes ...
type EquipmentAttributes struct {
	AttributeName       string
	AttributeValue      string
	ParentIdentifier    bool
	AttributeIdentifier bool
}
