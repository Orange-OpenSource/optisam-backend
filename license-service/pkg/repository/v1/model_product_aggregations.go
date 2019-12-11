// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// ProductAggregation is the logical grouping of products
type ProductAggregation struct {
	ID       string
	Name     string
	Editor   string
	Product  string
	Metric   string
	Products []string // list of ids of the prioduct which  are in aggregations
}
