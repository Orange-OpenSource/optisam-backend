package v1

type ProductEquipment struct {
	Swidtag     string
	ProductName string
	Equipments  []*Equipment
}

// ProductEquipment ...
type Equipment struct {
	EquipmentID   string
	EquipmentType string
}

// EquipmentAttributes ...
type EquipmentAttributes struct {
	AttributeName       string
	AttributeValue      string
	ParentIdentifier    bool
	AttributeIdentifier bool
}

// // MetadataType of metadata
// type MetadataType uint8
// type DataType uint8

// const (
// 	// MetadataTypeProduct is for product
// 	MetadataTypeProduct MetadataType = 0
// 	// MetadataTypeApplication is for application
// 	MetadataTypeApplication MetadataType = 1
// 	// MetadataTypeInstance is for instance
// 	MetadataTypeInstance MetadataType = 2
// 	// MetadataTypeEquipment is for equipment
// 	MetadataTypeEquipment MetadataType = 3
// 	// MetadataTypeMetadata is for metadata
// 	MetadataTypeMetadata MetadataType = 4
// )

// type EquipmentType struct {
// 	ID         string
// 	Type       string
// 	SourceID   string
// 	SourceName string
// 	ParentID   string
// 	ParentType string
// 	Scopes     []string
// 	Attributes []*Attribute
// }

// // Attribute for attribute of data
// type Attribute struct {
// 	Type               DataType
// 	IsIdentifier       bool
// 	IsDisplayed        bool
// 	IsSearchable       bool
// 	IsParentIdentifier bool
// 	IsSimulated        bool
// 	IntVal             int
// 	IntValOld          int
// 	FloatVal           float32
// 	FloatValOld        float32
// 	ID                 string
// 	Name               string
// 	MappedTo           string
// 	StringVal          string
// 	StringValOld       string
// }
