// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package loader

type dgraphType string

func (d dgraphType) String() string {
	return string(d)
}

const (
	dgraphTypeAcquiredRights     dgraphType = "AcquiredRights"
	dgraphTypeApplication        dgraphType = "Application"
	dgraphTypeEditor             dgraphType = "Editor"
	dgraphTypeEquipment          dgraphType = "Equipment"
	dgraphTypeInstance           dgraphType = "Instance"
	dgraphTypeMetadata           dgraphType = "Metadata"
	dgraphTypeAttribute          dgraphType = "Attribute"
	dgraphTypeMetricIPS          dgraphType = "MetricIPS"
	dgraphTypeMetricOPS          dgraphType = "MetricOPS"
	dgraphTypeMetricOracleNUP    dgraphType = "MetricOracleNUP"
	dgraphTypeMetricSPS          dgraphType = "MetricSPS"
	dgraphTypeProductAggregation dgraphType = "ProductAggregation"
	dgraphTypeProduct            dgraphType = "Product"
	dgraphTypeUser               dgraphType = "User"
)
