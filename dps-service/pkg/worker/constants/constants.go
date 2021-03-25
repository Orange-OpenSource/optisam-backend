// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package constants

import (
	"database/sql"
)

//File Fields
const (
	INST_ID                   string = "instance_id"
	ENVIRONMENT               string = "environment"
	OWNER                     string = "owner"
	APP_ID                    string = "application_id"
	NBUSERS                   string = "nbusers"
	EQUIP_ID                  string = "equipment_id"
	NAME                      string = "name"
	VERSION                   string = "version"
	PRODUCT_VERSION           string = "product_version"
	DOMAIN                    string = "domain"
	IS_OPTION_OF              string = "isoptionof"
	CATEGORY                  string = "category"
	SWIDTAG                   string = "swidtag"
	SKU                       string = "sku"
	ENTITY                    string = "entity"
	PRODUCT_NAME              string = "product_name"
	EDITOR                    string = "editor"
	METRIC                    string = "metric"
	ACQ_LIC_NO                string = "acquired_licenses"
	LIC_UNDER_MAINTENANCE_NO  string = "maintenance_licenses"
	AVG_UNIT_PRICE            string = "unit_price"
	AVG_MAINENANCE_UNIT_PRICE string = "maintenance_unit_price"
	TOTAL_PURCHASE_COST       string = "total_license_cost"
	TOTAL_MAINENANCE_COST     string = "total_maintenance_cost"
	TOTAL_COST                string = "total_cost"
	FLAG                      string = "flag"
	START_OF_MAINTENANCE      string = "maintenance_start"
	END_OF_MAINTENANCE        string = "maintenance_end"
)

//FILETYPES
const (
	APPLICATIONS            string = "APPLICATIONS"
	PRODUCTS                string = "PRODUCTS"
	PRODUCTS_EQUIPMENTS     string = "PRODUCTS_EQUIPMENTS"
	APPLICATIONS_PRODUCTS   string = "APPLICATIONS_PRODUCTS"
	APPLICATIONS_INSTANCES  string = "APPLICATIONS_INSTANCES"
	INSTANCES_EQUIPMENTS    string = "INSTANCES_EQUIPMENTS"
	INSTANCES_PRODUCTS      string = "INSTANCES_PRODUCTS"
	PRODUCTS_ACQUIREDRIGHTS string = "PRODUCTS_ACQUIREDRIGHTS"
	METADATA                string = "METADATA"
	GLOBALDATA              string = "GLOBALDATA"
	EQUIPMENTS              string = "EQUIPMENTS"
)

//SERVICES
const (
	APP_SERVICE  = "application"
	PROD_SERVICE = "product"
	//ACQ_SERVICE   = "acqright"
	EQUIP_SERVICE = "equipment"
)

// general constants
const (
	DELIMETER       string = ";"
	FILE_EXTENSION  string = ".CSV"
	SCOPE_DELIMETER string = "_"
	FILEWORKER      string = "FILE_WORKER"
	APIWORKER       string = "API_WORKER"
	DEFERWORKER     string = "DEFER_WORKER"
	DPSQUEUE        string = "DPS_QUEUE"
	UPSERT          string = "UPSERT"
	DELETE          string = "DELETE"
	DROP            string = "DROP"
)

//fileName to Services mapping
var (
	SERVICES = map[string][]string{
		PRODUCTS:                []string{PROD_SERVICE},
		EQUIPMENTS:              []string{EQUIP_SERVICE},
		APPLICATIONS:            []string{APP_SERVICE},
		PRODUCTS_EQUIPMENTS:     []string{PROD_SERVICE},
		APPLICATIONS_INSTANCES:  []string{APP_SERVICE},
		APPLICATIONS_PRODUCTS:   []string{PROD_SERVICE},
		INSTANCES_EQUIPMENTS:    []string{APP_SERVICE},
		INSTANCES_PRODUCTS:      []string{APP_SERVICE},
		PRODUCTS_ACQUIREDRIGHTS: []string{PROD_SERVICE}, // change to product-service in OPTISAM-1708
		METADATA:                []string{EQUIP_SERVICE},
	}
)

//These are constants, please don't mutate it
var (
	FILETYPE    = sql.NullString{String: FILEWORKER, Valid: true}
	APITYPE     = sql.NullString{String: APIWORKER, Valid: true}
	DEFERTYPE   = sql.NullString{String: DEFERWORKER, Valid: true}
	ACTION_TYPE = map[string]string{"1": UPSERT, "0": DELETE}
	API_ACTION  = map[string]string{UPSERT: "add", DELETE: "delete"}
)
