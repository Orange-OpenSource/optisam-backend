package constants

import (
	"database/sql"
)

// File Fields
const (
	INSTID                    string = "instance_id"
	ENVIRONMENT               string = "environment"
	OWNER                     string = "owner"
	APPID                     string = "application_id"
	NBUSERS                   string = "nbusers"
	EQUIPID                   string = "equipment_id"
	NAME                      string = "name"
	VERSION                   string = "version"
	PRODUCTVERSION            string = "product_version"
	DOMAIN                    string = "domain"
	ISOPTIONOF                string = "isoptionof"
	CATEGORY                  string = "category"
	SWIDTAG                   string = "swidtag"
	SKU                       string = "sku"
	PRODUCTNAME               string = "product_name"
	EDITOR                    string = "editor"
	METRIC                    string = "metric"
	ACQLICNO                  string = "acquired_licenses"
	LICUNDERMAINTENANCENO     string = "maintenance_licenses"
	AVGUNITPRICE              string = "unit_price"
	AVGMAINENANCEUNITPRICE    string = "maintenance_unit_price"
	TOTALPURCHASECOST         string = "total_license_cost"
	TOTALMAINENANCECOST       string = "total_maintenance_cost"
	TOTALCOST                 string = "total_cost"
	FLAG                      string = "flag"
	StartOfMaintenance        string = "maintenance_start"
	EndOfMaintenance          string = "maintenance_end"
	BadFile                   string = "BadFile"
	SoftwareProvider          string = "software_provider"
	MaintenanceProvider       string = "maintenance_provider"
	OrderingDate              string = "ordering_date"
	CorporateSourcingContract string = "corporate_sourcing_contract"
	SupportNumber             string = "support_number"
	LastPurchasedOrder        string = "last_purchased_order"
	AllocatedMetric           string = "allocated_metric"
	AllocatedUsers            string = "allocated_users"
)

// FILETYPES
const (
	APPLICATIONS           string = "APPLICATIONS"
	PRODUCTS               string = "PRODUCTS"
	ProductsEquipments     string = "PRODUCTS_EQUIPMENTS"
	ApplicationsProducts   string = "APPLICATIONS_PRODUCTS"
	ApplicationsInstances  string = "APPLICATIONS_INSTANCES"
	ApplicationEquipments  string = "APPLICATION_EQUIPMENTS"
	InstancesProducts      string = "INSTANCES_PRODUCTS"
	ProductsAcquiredRights string = "PRODUCTS_ACQUIREDRIGHTS"
	METADATA               string = "METADATA"
	GLOBALDATA             string = "GLOBALDATA"
	EQUIPMENTS             string = "EQUIPMENTS"
)

// SERVICES
const (
	AppService   = "application"
	ProdService  = "product"
	EquipService = "equipment"
)

// general constants
const (
	DELIMETER         string = ";"
	FileExtension     string = ".CSV"
	ScopeDelimeter    string = "_"
	NifiFileDelimeter string = "#"
	NIFI              string = "NIFI"
	FILEWORKER        string = "FILE_WORKER"
	APIWORKER         string = "API_WORKER"
	DEFERWORKER       string = "DEFER_WORKER"
	DPSQUEUE          string = "DPS_QUEUE"
	UPSERT            string = "UPSERT"
	DELETE            string = "DELETE"
	DROP              string = "DROP"
	PROCESSING        string = "PROCESSING"
	FailedData        string = "FAILED_DATA"
	SuccessData       string = "SUCCESS_DATA"
	FAILED            string = "FAILED"
	SUCCESS           string = "SUCCESS"
	PARTIAL           string = "PARTIAL"
)

// fileName to Services mapping
var (
	SERVICES = map[string][]string{
		PRODUCTS:               {ProdService},
		EQUIPMENTS:             {EquipService},
		APPLICATIONS:           {AppService},
		ProductsEquipments:     {ProdService},
		ApplicationsInstances:  {AppService},
		ApplicationsProducts:   {ProdService},
		ApplicationEquipments:  {AppService},
		InstancesProducts:      {AppService},
		ProductsAcquiredRights: {ProdService}, // change to product-service in OPTISAM-1708
		METADATA:               {EquipService},
	}
)

// These are constants, please don't mutate it
var (
	FILETYPE   = sql.NullString{String: FILEWORKER, Valid: true}
	APITYPE    = sql.NullString{String: APIWORKER, Valid: true}
	DEFERTYPE  = sql.NullString{String: DEFERWORKER, Valid: true}
	ActionType = map[string]string{"1": UPSERT, "0": DELETE}
	APIAction  = map[string]string{UPSERT: "add", DELETE: "delete"}
)
