package v1

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/equipment-service/pkg/repository/v1 Equipment

// Equipment interface
type Equipment interface {
	ListEquipmentsForProductAggregation(ctx context.Context, name string, eqType *EquipmentType, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)

	// ProductEquipments list all the equipments for a product for given equipment type
	ProductEquipments(ctx context.Context, swidTag string, eqType *EquipmentType, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)
	// MetadataAllWithType gets metadata for given metadata type
	MetadataAllWithType(ctx context.Context, typ MetadataType, scopes []string) ([]*Metadata, error)

	// MetadataWithID gets metadata for given id
	MetadataWithID(ctx context.Context, id string, scopes []string) (*Metadata, error)

	// CreateEquipmentType stores equipmentdata and creates schema with required primary key
	// and indexes.
	CreateEquipmentType(ctx context.Context, eqType *EquipmentType, scopes []string) (*EquipmentType, error)

	// EquipmentTypes fetches all equipment types from database
	EquipmentTypes(ctx context.Context, scopes []string) ([]*EquipmentType, error)

	// DeleteEquipmentType deletes equipment type from database
	DeleteEquipmentType(ctx context.Context, eqType, scope string) error

	// EquipmentTypeChildren fetches all equipment type children from database
	EquipmentTypeChildren(ctx context.Context, eqTypeID string, depth int, scopes []string) ([]*EquipmentType, error)

	// ParentHirearchy gives equipment along with parent hierarchy
	ParentsHirerachyForEquipment(ctx context.Context, equipID, equipType string, hirearchyLevel uint8, scopes ...string) ([]*EquipmentInfo, error)

	// GetAllEquipmentsInHierarchy gives all equipments fallen in ancestor
	GetAllEquipmentsInHierarchy(ctx context.Context, equipType string, endEquID string, scopes ...string) (*EquipmentHierarchy, error)

	GetAllEquipmentForSpecifiedProduct(ctx context.Context, swidtag string, scopes ...string) (*DeployedProducts, error)

	UpsertAllocateMetricInEquipmentHierarchy(ctx context.Context, mat *MetricAllocationRequest, scope string) error

	GetAllocatedMetrics(ctx context.Context, swidtag string, scopes ...string) (*AllocatedMetrics, error)

	UpsertAllocateMetricInProduct(ctx context.Context, swidTag string, metUids []string, scope string) error

	// UpsertMetaData stores metadata in dgrpah
	UpsertMetadata(ctx context.Context, metadata *Metadata) (string, error)

	EquipmentWithID(ctx context.Context, id string, scopes []string) (*EquipmentType, error)

	UpdateEquipmentType(ctx context.Context, id string, typ string, parentID string, req *UpdateEquipmentRequest, scopes []string) (retType []*Attribute, retErr error)
	Equipments(ctx context.Context, eqType *EquipmentType, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)
	DeleteEquipments(ctx context.Context, scope string) error

	// DropMetaData deletes metadata
	DropMetaData(ctx context.Context, scope string) error

	// Equipment gets equipmet for given type and id if exists,if not exist then ErrNotFound
	Equipment(ctx context.Context, eqType *EquipmentType, id string, scopes []string) (json.RawMessage, error)

	// EquipmentParents return parent of the given equipment
	EquipmentParents(ctx context.Context, eqType, parentEqType *EquipmentType, id string, scopes []string) (int32, json.RawMessage, error)

	// EquipmentChildren return children of the given equipment id for child type
	EquipmentChildren(ctx context.Context, eqType, childEqType *EquipmentType, id string, params *QueryEquipments, scopes []string) (int32, json.RawMessage, error)

	EquipmentTypeByType(ctx context.Context, typ string, scopes []string) (*EquipmentType, error)

	UpsertEquipment(ctx context.Context, scope string, eqType string, parentEqType string, eqData interface{}) error
}

// Queryable interface provide methods for something that can be queried
type Queryable interface {
	// Key that needed to be queried (coloumn name)
	Key() string
	// Value for key tha we need tio search
	Value() interface{}

	// Values for key tha we need tio search
	Values() []interface{}

	Priority() int32

	Type() Filtertype
}

// SortOrder - type defined for sorting parameters i.e ascending/descending
type SortOrder int32

const (
	// SortASC - sorting in ascending order
	SortASC SortOrder = 0
	// SortDESC - sorting in descending order
	SortDESC SortOrder = 1
)
