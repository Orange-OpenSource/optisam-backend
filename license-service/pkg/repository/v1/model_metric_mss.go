package v1

// Metadata for injectors
type MetricMSS struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
}

// MetricMSSComputed has all the information required to be computed
type MetricMSSComputed struct {
	Name          string
	BaseType      []string
	ReferenceType string
	NumCoresAttr  string
	NumCPUAttr    string
	IsSA          bool
}
