package v1

// MetricSS is a representation of static.standard
type MetricSS struct {
	ID             string
	Name           string
	ReferenceValue int32
}

// MetricSSComputed has all the information required to be computed
type MetricSSComputed struct {
	Name           string
	ReferenceValue int32
}
