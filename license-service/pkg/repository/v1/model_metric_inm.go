// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

// MetricINM is a representation of instance.number.standard
type MetricINM struct {
	ID          string
	Name        string
	Coefficient float32
}

// MetricINMComputed has all the information required to be computed
type MetricINMComputed struct {
	Name        string
	Coefficient float32
}
