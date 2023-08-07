package dgraph

import "time"

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

type UpserNominativeUserRequest struct {
	Editor         string                   `json:"editor,omitempty"`
	Scope          string                   `json:"scope,omitempty"`
	ProductName    string                   `json:"product_name,omitempty"`
	ProductVersion string                   `json:"product_version,omitempty"`
	AggregationId  int32                    `json:"aggregation_id,omitempty"`
	SwidTag        string                   `json:"swid_tag,omitempty"`
	CreatedBy      string                   `json:"created_by,omitempty"`
	UserDetails    []*NominativeUserDetails `json:"user_details,omitempty"`
}

type NominativeUserDetails struct {
	UserName       string    `json:"user_name,omitempty"`
	FirstName      string    `json:"first_name,omitempty"`
	Email          string    `json:"email,omitempty"`
	Profile        string    `json:"profile,omitempty"`
	ActivationDate time.Time `json:"activation_date,omitempty"`
}

type UpserConcurrentUserRequest struct {
	IsAggregations bool   `json:"is_aggregations,omitempty"`
	AggregationID  int32  `json:"aggregation_id,omitempty"`
	Editor         string `json:"product_editor,omitempty"`
	ProductName    string `json:"product_name,omitempty"`
	ProductVersion string `json:"product_version,omitempty"`
	SwidTag        string `json:"swidtag,omitempty"`
	NumberOfUsers  int32  `json:"number_of_users,omitempty"`
	ProfileUser    string `json:"profile_user,omitempty"`
	Team           string `json:"team,omitempty"`
	Scope          string `json:"scope,omitempty"`
	CreatedBy      string `json:"created_by,omitempty"`
	PurchaseDate   string `json:"purchase_date,omitempty"`
}

type DeleteProductRequest struct {
	SwidTag string `json:"swidtag"`
	Scope   string `json:"scope"`
}
