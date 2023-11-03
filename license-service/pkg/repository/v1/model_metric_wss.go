package v1

// Metadata for injectors
type MetricWSS struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
}

// MetricWSSComputed has all the information required to be computed
type MetricWSSComputed struct {
	Name          string
	BaseType      []string
	ReferenceType string
	NumCoresAttr  string
	NumCPUAttr    string
	IsSA          bool
}
