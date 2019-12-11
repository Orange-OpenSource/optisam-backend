// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package v1

// AcquiredRightsSearchKey is type to represent search keys string
type AcquiredRightsSearchKey string

const (
	// AcquiredRightsSearchKeySKU ...
	AcquiredRightsSearchKeySKU AcquiredRightsSearchKey = "SKU"
	// AcquiredRightsSearchKeySwidTag ...
	AcquiredRightsSearchKeySwidTag AcquiredRightsSearchKey = "swidtag"
	// AcquiredRightsSearchKeyProductName ...
	AcquiredRightsSearchKeyProductName AcquiredRightsSearchKey = "productName"
	// AcquiredRightsSearchKeyEditor ...
	AcquiredRightsSearchKeyEditor AcquiredRightsSearchKey = "editor"
	// AcquiredRightsSearchKeyMetric ...
	AcquiredRightsSearchKeyMetric AcquiredRightsSearchKey = "metric"
)

// String implemetd stringer interface
func (a AcquiredRightsSearchKey) String() string {
	return string(a)
}

// AcquiredRightsSortBy identifies sorting key of acquired rights
type AcquiredRightsSortBy int32

const (
	// AcquiredRightsSortByEntity ...
	AcquiredRightsSortByEntity AcquiredRightsSortBy = 0
	// AcquiredRightsSortBySKU ...
	AcquiredRightsSortBySKU AcquiredRightsSortBy = 1
	// AcquiredRightsSortBySwidTag ...
	AcquiredRightsSortBySwidTag AcquiredRightsSortBy = 2
	// AcquiredRightsSortByProductName ...
	AcquiredRightsSortByProductName AcquiredRightsSortBy = 3
	// AcquiredRightsSortByEditor ...
	AcquiredRightsSortByEditor AcquiredRightsSortBy = 4
	// AcquiredRightsSortByMetric ...
	AcquiredRightsSortByMetric AcquiredRightsSortBy = 5
	// AcquiredRightsSortByAcquiredLicensesNumber ...
	AcquiredRightsSortByAcquiredLicensesNumber AcquiredRightsSortBy = 6
	// AcquiredRightsSortByLicensesUnderMaintenanceNumber ...
	AcquiredRightsSortByLicensesUnderMaintenanceNumber AcquiredRightsSortBy = 7
	// AcquiredRightsSortByAvgLicenseUnitPrice ...
	AcquiredRightsSortByAvgLicenseUnitPrice AcquiredRightsSortBy = 8
	// AcquiredRightsSortByAvgMaintenanceUnitPrice ...
	AcquiredRightsSortByAvgMaintenanceUnitPrice AcquiredRightsSortBy = 9
	// AcquiredRightsSortByTotalPurchaseCost ...
	AcquiredRightsSortByTotalPurchaseCost AcquiredRightsSortBy = 10
	// AcquiredRightsSortByTotalMaintenanceCost ...
	AcquiredRightsSortByTotalMaintenanceCost AcquiredRightsSortBy = 11
	// AcquiredRightsSortByTotalCost ...
	AcquiredRightsSortByTotalCost AcquiredRightsSortBy = 12
)

// AcquiredRights represent Acquired Rights of a product
type AcquiredRights struct {
	Entity                         string
	SKU                            string
	SwidTag                        string
	ProductName                    string
	Editor                         string
	Metric                         string
	AcquiredLicensesNumber         int64
	LicensesUnderMaintenanceNumber int64
	AvgLicenesUnitPrice            float32
	AvgMaintenanceUnitPrice        float32
	TotalPurchaseCost              float32
	TotalMaintenanceCost           float32
	TotalCost                      float32
}

// QueryAcquiredRights represents query rights.
type QueryAcquiredRights struct { //
	PageSize  int32
	Offset    int32
	SortBy    AcquiredRightsSortBy
	SortOrder SortOrder
	Filter    *AggregateFilter
}
