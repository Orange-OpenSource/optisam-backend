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

// MasterData is the data for the master table
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
