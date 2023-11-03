package v1

// Metadata for injectors
type MetricWSD struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
}

// MetricWSDComputed has all the information required to be computed
type MetricWSDComputed struct {
	Name          string
	BaseType      []string
	ReferenceType string
	NumCoresAttr  string
	NumCPUAttr    string
	IsSA          bool
}
