// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

// MetricACS is a representation of sag.processor.standard
type MetricACS struct {
	ID            string
	Name          string
	EqType        string
	AttributeName string
	Value         string
}

// MetricACSComputed has all the information required to be computed
type MetricACSComputed struct {
	Name      string
	BaseType  *EquipmentType
	Attribute *Attribute
	Value     string
}
