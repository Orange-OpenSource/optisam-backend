package v1

// MetricSPS is a representation of sag.processor.standard
type MetricSQLStand struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
	Scope      string
	Default    bool
}
