package v1

import "errors"

// DataType for database
type DataType uint8

func (d DataType) String() string {
	switch d {
	case DataTypeInt:
		return "int"
	case DataTypeFloat:
		return "float"
	case DataTypeString:
		return "string"
	default:
		return "unsupported"
	}
}

const (
	// DataTypeString for string
	DataTypeString DataType = 1
	// DataTypeInt for int
	DataTypeInt DataType = 2
	// DataTypeFloat for float
	DataTypeFloat DataType = 3
)

// EquipmentProductSortBy - type defined for sorting fields of equipment products
type EquipmentProductSortBy uint8

const (
	// EquipmentProductSortBySwidTag - sorting by product swid tag
	EquipmentProductSortBySwidTag EquipmentProductSortBy = 0
	// EquipmentProductSortByName - sorting by product name
	EquipmentProductSortByName EquipmentProductSortBy = 1
	// EquipmentProductSortByEditor - sorting by product editor
	EquipmentProductSortByEditor EquipmentProductSortBy = 2
	// EquipmentProductSortByVersion - sorting by product version
	EquipmentProductSortByVersion EquipmentProductSortBy = 3
)

// EquipmentProductSearchKey - type defined for searching fields of equipment products
type EquipmentProductSearchKey string

func (e EquipmentProductSearchKey) String() string {
	return string(e)
}

const (
	// EquipmentProductSearchKeySwidTag - searching by product swid tag
	EquipmentProductSearchKeySwidTag EquipmentProductSearchKey = "swidtag"
	// EquipmentProductSearchKeyName - searching by product name
	EquipmentProductSearchKeyName EquipmentProductSearchKey = "name"
	// EquipmentProductSearchKeyEditor - searching by product editor
	EquipmentProductSearchKeyEditor EquipmentProductSearchKey = "editor"
	// EquipmentProductSearchKeyVersion - searching by product version
	EquipmentProductSearchKeyVersion EquipmentProductSearchKey = "release"
)

// EquipmentType for creating equipment type
type EquipmentType struct {
	ID         string
	Type       string
	SourceID   string
	SourceName string
	ParentID   string
	ParentType string
	Attributes []*Attribute
}

// QueryEquipmentProduct has params to query products of an equipment
type QueryEquipmentProduct struct {
	PageSize  int32
	Offset    int32
	SortBy    EquipmentProductSortBy
	SortOrder SortOrder
	Filter    *AggregateFilter
}

// EquipmentProduct represents fields required for equipment
type EquipmentProduct struct {
	SwidTag string
	Name    string
	Editor  string
	Version string
}

// Equipment has generic information about equipment an ancestors
type Equipment struct {
	ID      string
	EquipID string
	Type    string
	Parent  *Equipment
}

// PrimaryKeyAttribute returns primary key attribute of equipment type
func (e *EquipmentType) PrimaryKeyAttribute() (*Attribute, error) {
	for _, attr := range e.Attributes {
		if attr.IsIdentifier {
			return attr, nil
		}
	}
	return nil, errors.New("primary key attribute is not found")
}

// ParentKeyAttribute returns primary key attribute of equipment type
func (e *EquipmentType) ParentKeyAttribute() (*Attribute, error) {
	for _, attr := range e.Attributes {
		if attr.IsParentIdentifier {
			return attr, nil
		}
	}
	return nil, errors.New("primary key attribute is not found")
}

// QueryEquipments has parameters for query
type QueryEquipments struct {
	PageSize          int32
	Offset            int32
	SortBy            string
	SortOrder         SortOrder
	Filter            *AggregateFilter
	ProductFilter     *AggregateFilter
	ApplicationFilter *AggregateFilter
}

// UpdateEquipmentRequest ...
type UpdateEquipmentRequest struct {
	ParentID string
	Attr     []*Attribute
}
