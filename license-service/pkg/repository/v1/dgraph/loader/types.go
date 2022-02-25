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
