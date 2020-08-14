// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1

import (
	"encoding/json"
	"time"
)

// Metadata is the metadata for the given equipment type and attribute name
type Metadata struct {
	ID             int32
	AttributeName  string
	ConfigFileName string
}

// ConfigValue is the struct which contains config values in key value pair.
type ConfigValue struct {
	Key   string
	Value json.RawMessage
}

// ConfigData is the data for one file
type ConfigData struct {
	ConfigMetadata *Metadata
	ConfigValues   []*ConfigValue
}

//MasterData is the data for the master table
type MasterData struct {
	ID            int32
	Name          string
	Status        int
	EquipmentType string
	CreatedBy     string
	CreatedOn     time.Time
	UpdatedBy     string
	UpdatedOn     time.Time
}
