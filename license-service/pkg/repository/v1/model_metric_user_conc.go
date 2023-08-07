package v1

// MetricSCS is a representation of saas.concurrent.standard
type MetricUCS struct {
	ID      string
	Name    string
	Profile string
}

// MetricSCSComputed has all the information required to be computed
type MetricUCSComputed struct {
	Name    string
	Profile string
}
