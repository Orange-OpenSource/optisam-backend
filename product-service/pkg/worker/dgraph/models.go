package dgraph

// nolint: maligned
type UpsertAcqRightsRequest struct {
	Sku                       string  `json:"sku,omitempty"`
	Swidtag                   string  `json:"swidtag,omitempty"`
	ProductName               string  `json:"product_name,omitempty"`
	ProductEditor             string  `json:"product_editor,omitempty"`
	MetricType                string  `json:"metric_type,omitempty"`
	StartOfMaintenance        string  `json:"start_of_maintenance,omitempty"`
	EndOfMaintenance          string  `json:"end_of_maintenance,omitempty"`
	Version                   string  `json:"version,omitempty"`
	NumLicensesAcquired       int32   `json:"num_licenses_acquired,omitempty"`
	AvgUnitPrice              float64 `json:"avg_unit_price,omitempty"`
	AvgMaintenanceUnitPrice   float64 `json:"avg_maintenance_unit_price,omitempty"`
	TotalPurchaseCost         float64 `json:"total_purchase_cost,omitempty"`
	TotalMaintenanceCost      float64 `json:"total_maintenance_cost,omitempty"`
	TotalCost                 float64 `json:"total_cost,omitempty"`
	Scope                     string  `json:"scope,omitempty"`
	NumLicencesMaintenance    int32   `json:"num_licences_maintainance,omitempty"` //nolint
	IsSwidtagModified         bool    `json:"isSwidtagModified"`
	IsMetricModifed           bool    `json:"isMetricModified"`
	OrderingDate              string  `json:"orderingDate"`
	CorporateSourcingContract string  `json:"corporateSourcingContract"`
	SoftwareProvider          string  `json:"softwareProvider"`
	LastPurchasedOrder        string  `json:"lastPurchasedOrder"`
	SupportNumber             string  `json:"supportNumber"`
	MaintenanceProvider       string  `json:"maintenanceProvider"`
	Repartition               bool    `json:"repartition"`
}

type DeleteAcqRightRequest struct {
	Sku   string `json:"sku"`
	Scope string `json:"scope"`
}

// nolint
type UpsertAggregationRequest struct {
	ID            int32    `json:"id,omitempty"`
	Name          string   `json:"name,omitempty"`
	Swidtags      []string `json:"swidtags,omitempty"`
	Products      []string `json:"product_names,omitempty"`
	ProductEditor string   `json:"product_editor,omitempty"`
	Scope         string   `json:"scope,omitempty"`
}

// nolint: maligned
type UpsertAggregatedRight struct {
	Sku                       string  `json:"sku,omitempty"`
	AggregationID             int32   `json:"aggregationID,omitempty"`
	Metric                    string  `json:"metric,omitempty"`
	StartOfMaintenance        string  `json:"start_of_maintenance,omitempty"`
	EndOfMaintenance          string  `json:"end_of_maintenance,omitempty"`
	NumLicensesAcquired       int32   `json:"num_licenses_acquired,omitempty"`
	AvgUnitPrice              float64 `json:"avg_unit_price,omitempty"`
	AvgMaintenanceUnitPrice   float64 `json:"avg_maintenance_unit_price,omitempty"`
	TotalPurchaseCost         float64 `json:"total_purchase_cost,omitempty"`
	TotalMaintenanceCost      float64 `json:"total_maintenance_cost,omitempty"`
	TotalCost                 float64 `json:"total_cost,omitempty"`
	Scope                     string  `json:"scope,omitempty"`
	NumLicencesMaintenance    int32   `json:"num_licences_maintenance,omitempty"`
	OrderingDate              string  `json:"orderingDate"`
	CorporateSourcingContract string  `json:"corporateSourcingContract"`
	SoftwareProvider          string  `json:"softwareProvider"`
	LastPurchasedOrder        string  `json:"lastPurchasedOrder"`
	SupportNumber             string  `json:"supportNumber"`
	MaintenanceProvider       string  `json:"maintenanceProvider"`
	Repartition               bool    `json:"repartition"`
}

type DeleteAggregatedRightRequest struct {
	Sku   string `json:"sku"`
	Scope string `json:"scope"`
}
