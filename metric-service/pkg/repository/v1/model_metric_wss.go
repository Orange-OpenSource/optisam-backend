package v1

// MetricWSS is a representation of metric window.server.standard
type MetricWSS struct {
	ID         string
	MetricType string
	MetricName string
	Reference  string
	Core       string
	CPU        string
	Default    bool
	Scope      string
}
