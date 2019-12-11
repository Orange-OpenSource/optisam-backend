// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// ApplicationData gives the details os an application
type ApplicationData struct {
	Name             string
	ApplicationID    string
	ApplicationOwner string
	NumOfInstances   int32
	NumOfProducts    int32
}

// ApplicationInfo ...
type ApplicationInfo struct {
	NumOfRecords []TotalRecords
	Applications []ApplicationData
}

// QueryApplications ...
type QueryApplications struct { //
	PageSize  int32
	Offset    int32
	SortBy    string
	SortOrder string
	Filter    *AggregateFilter
}

// ProductsForApplicationData ...
type ProductsForApplicationData struct {
	Name            string
	Version         string
	Editor          string
	Edition         string
	SwidTag         string
	NumOfEquipments int32
	NumOfInstances  int32
}

// ApplicationDetails gives the details os an application
type ApplicationDetails struct {
	Name             string
	ApplicationID    string
	ApplicationOwner string
	NumOfInstances   int32
	NumOfProducts    int32
}

// ProductsForApplication ...
type ProductsForApplication struct {
	NumOfRecords []TotalRecords
	Products     []ProductsForApplicationData
}
