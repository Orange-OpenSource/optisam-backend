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
	INST_ID                   string = "idinstance"
	OWNER                     string = "owner"
	APP_ID                    string = "idapplication"
	NBUSERS                   string = "nbusers"
	EQUIP_ID                  string = "idequipment"
	NAME                      string = "name"
	VERSION                   string = "version"
	IS_OPTION_OF              string = "isoptionof"
	CATEGORY                  string = "category"
	SWIDTAG                   string = "swidtag"
	SKU                       string = "sku"
	ENTITY                    string = "entity"
	PRODUCT_NAME              string = "product name"
	EDITOR                    string = "editor"
	METRIC                    string = "metric"
	ACQ_LIC_NO                string = "acquired licenses number"
	LIC_UNDER_MAINTENANCE_NO  string = "licenses under maintenance number"
	AVG_UNIT_PRICE            string = "avg unit price"
	AVG_MAINENANCE_UNIT_PRICE string = "avg maintenant unit price"
	TOTAL_PURCHASE_COST       string = "total purchase cost"
	TOTAL_MAINENANCE_COST     string = "total maintenance cost"
	TOTAL_COST                string = "total cost"
	FLAG                      string = "flag"
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
	EQUIPMENTS              string = "EQUIPMENTS"
)

//SERVICES
const (
	APP_SERVICE   = "application"
	PROD_SERVICE  = "product"
	ACQ_SERVICE   = "acqright"
	EQUIP_SERVICE = "equipment"
)

// general constants
const (
	DELIMETER       string = ";"
	FILE_EXTENSION  string = ".CSV"
	SCOPE_DELIMETER string = "_"
	FILEWORKER      string = "FILE_WORKER"
	APIWORKER       string = "API_WORKER"
	DPSQUEUE        string = "DPS_QUEUE"
	UPSERT          string = "UPSERT"
	DELETE          string = "DELETE"
)

//Services
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
		PRODUCTS_ACQUIREDRIGHTS: []string{ACQ_SERVICE},
		METADATA:                []string{EQUIP_SERVICE},
	}
)

//These are constants, please don't mutate it
var (
	FILETYPE    = sql.NullString{String: FILEWORKER, Valid: true}
	APITYPE     = sql.NullString{String: APIWORKER, Valid: true}
	ACTION_TYPE = map[string]string{"1": UPSERT, "0": DELETE}
	API_ACTION  = map[string]string{UPSERT: "add", DELETE: "delete"}
)
