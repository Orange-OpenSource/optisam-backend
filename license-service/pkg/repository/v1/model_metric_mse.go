package v1

// MetricMSE is a representation of microsoft.sql.enterprise
type MetricMSE struct {
	ID        string
	Name      string
	Reference string
	Core      string
	CPU       string
	Scope     string
	Default   bool
}

// MetricMSEComputed has all the information required to be computed
type MetricMSEComputed struct {
	Name      string
	Reference string
	Core      string
	CPU       string
	Scope     string
	Default   bool
}
