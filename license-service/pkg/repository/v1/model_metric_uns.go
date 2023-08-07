package v1

// MetricSNS is a representation of saas.nominative.standard
type MetricUNS struct {
	ID      string
	Name    string
	Profile string
}

// MetricSNSComputed has all the information required to be computed
type MetricUNSComputed struct {
	Name    string
	Profile string
}
