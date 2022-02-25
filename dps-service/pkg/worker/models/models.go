package models

import (
	"encoding/json"
)

// ProductInfo will carry prod.csv file content
type ProductInfo struct {
	Name       string
	Version    string
	Editor     string
	IsOptionOf string
	Category   string
	SwidTag    string
	Action     string
}

type ProdEquipemtInfo struct {
	EquipID string
	NbUsers string
}

// ApplicationInfo will carry application.csv file data
type ApplicationInfo struct {
	ID      string
	Name    string
	Version string
	Owner   string
	Domain  string
	Action  string // This tells whether this info need to upsert or delete
}

// AppInstance for App-instance relation
type AppInstance struct {
	ID     string
	Env    string
	Action string
}

// Acqright
type AcqRightsInfo struct {
	Version              string
	SwidTag              string
	Sku                  string
	ProductName          string
	Editor               string
	Metric               string
	NumOfMaintenanceLic  int
	NumOfAcqLic          int
	AvgPrice             float64
	AvgMaintenantPrice   float64
	TotalPurchasedCost   float64
	TotalMaintenanceCost float64
	TotalCost            float64
	StartOfMaintenance   string
	EndOfMaintenance     string
	Action               string
}

// FileData will carry combine information of whole file scope
type FileData struct {
	Products            map[string]ProductInfo
	Equipments          map[string][]map[string]interface{}
	Applications        map[string]ApplicationInfo
	AppInstances        map[string][]AppInstance
	ProdInstances       map[string]map[string][]string
	EquipInstances      map[string]map[string][]string
	AppProducts         map[string]map[string][]string
	ProdEquipments      map[string]map[string][]ProdEquipemtInfo
	AcqRights           map[string]AcqRightsInfo
	Schema              []string // map[type]{schema names}, eg: [cluster]{name, parent}
	TotalCount          int32
	InvalidCount        int32
	TargetServices      []string
	FileType            string
	Scope               string
	FileName            string
	TransfromedFileName string
	UploadID            int32
	GlobalID            int32 `json:"gid"`
	FileFailureReason   string
	InvalidDataRowNum   []int
	DuplicateRecords    []interface{}
}

type HeadersInfo struct {
	IndexesOfHeaders map[string]int // mapping of header's location in file
	MaxIndexVal      int            // max val of index
}

type Envlope struct {
	Data                json.RawMessage // tells data will be sent to that target
	TargetService       string          // tells target service
	TargetAction        string          // tells what action to do on service
	TargetRPC           string          // tell this action to do on which rpc
	FileName            string
	TransfromedFileName string
	UploadID            int32
	GlobalFileID        int32
}

type EquipmentRequest struct {
	Scope  string          `json:"scope,omitempty"`
	EqType string          `json:"eq_type,omitempty"`
	EqData json.RawMessage `json:"eq_data,omitempty"`
}

type ProdApplink struct {
	ProdID string `json:"productId"`
	AppID  string `json:"applicationID"`
	Action string `json:"action"`
}

type ProdInstancelink struct {
	ProdID     string `json:"productId"`
	InstanceID string `json:"instanceId"`
	Action     string `json:"action"`
}

type AppInstanceLink struct {
	AppID      string `json:"applicationId"`
	InstanceID string `json:"instanceId"`
	Env        string `json:"env"`
	Action     string `json:"action"`
}

type ProductEquipmentLink struct {
	ProdID  string `json:"prodId"`
	EquipID string `json:"equipmentId"`
	NbUser  string `json:"nbUser"`
	Action  string `json:"action"`
}

type EquipmentInstanceLink struct {
	InstanceID string `json:"intanceId"`
	EquipID    string `json:"equipmentId"`
	Action     string `json:"action"`
}
