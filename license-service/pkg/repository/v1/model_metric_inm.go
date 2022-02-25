package v1

// MetricINM is a representation of instance.number.standard
type MetricINM struct {
	ID          string
	Name        string
	Coefficient int32
}

// MetricINMComputed has all the information required to be computed
type MetricINMComputed struct {
	Name        string
	Coefficient int32
}
