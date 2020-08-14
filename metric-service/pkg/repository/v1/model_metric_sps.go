// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

// MetricSPS is a representation of sag.processor.standard
type MetricSPS struct {
	ID               string
	Name             string
	NumCoreAttrID    string
	CoreFactorAttrID string
	BaseEqTypeID     string
}

// MetricSPSComputed has all the information required to be computed
type MetricSPSComputed struct {
	Name           string
	BaseType       *EquipmentType
	CoreFactorAttr *Attribute
	NumCoresAttr   *Attribute
}

//MetricSPSConfig is a representation of sag.processor.standard metric configuration
type MetricSPSConfig struct {
	ID             string
	Name           string
	NumCoreAttr    string
	CoreFactorAttr string
	BaseEqType     string
}
