package v1

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=dmock/mock.go -package=mock gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1 DgraphReport

// DgraphReport ...
type DgraphReport interface {
	// EquipmentTypeParents fetches the equipmenttype parents
	EquipmentTypeParents(ctx context.Context, equipType string, scope string) ([]string, error)
	EquipmentTypeAttrs(ctx context.Context, equipType string, scope string) ([]*EquipmentAttributes, error)
	ProductEquipments(ctx context.Context, editor string, scope string, eqType string) ([]*ProductEquipment, error)
	EquipmentParents(ctx context.Context, equipID, equipType string, scope string) ([]*Equipment, error)
	EquipmentAttributes(ctx context.Context, equipID, equipType string, attrs []*EquipmentAttributes, scope string) (json.RawMessage, error)
}
