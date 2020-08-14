// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package v1


import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=dmock/mock.go -package=mock optisam-backend/report-service/pkg/repository/v1 DgraphReport

//DgraphReport ...
type DgraphReport interface {
	//EquipmentTypeParents fetches the equipmenttype parents
	EquipmentTypeParents(ctx context.Context, equipType string) ([]string, error)
	EquipmentTypeAttrs(ctx context.Context, equipType string) ([]*EquipmentAttributes, error)
	ProductEquipments(ctx context.Context, swidTag string, scope string, eqType string) ([]*ProductEquipment, error)
	EquipmentParents(ctx context.Context, equipID, equipType string, scope string) ([]*ProductEquipment, error)
	EquipmentAttributes(ctx context.Context, equipID, equipType string, attrs []*EquipmentAttributes) (json.RawMessage, error)
}
