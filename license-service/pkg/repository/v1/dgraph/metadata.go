// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/license-service/pkg/repository/v1"

	"go.uber.org/zap"
)

type metadataType string

const (
	metadataTypeEquipment   metadataType = "equipment"
	metadataTypeUnsupported metadataType = "unsupported"
)

// MetadataAllWithType implements Licence MetadataAllWithType function
func (lr *LicenseRepository) MetadataAllWithType(ctx context.Context, typ v1.MetadataType, scopes []string) ([]*v1.Metadata, error) {
	id, err := convertMetadataTypeDGType(typ)
	if err != nil {
		return nil, err
	}
	q := `{
		Metadatas(func: eq(metadata.type,` + string(id) + `),orderasc: metadata.source ) {
		   ID:         uid
		   Source:     metadata.source
		   Attributes: metadata.attributes
		}
	  }`

	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/Metadata - cannot complete query")
	}

	type data struct {
		Metadatas []*v1.Metadata
	}

	metadata := data{}

	if err := json.Unmarshal(resp.GetJson(), &metadata); err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/Metadata - cannot unmarshal Json object")
	}
	return metadata.Metadatas, nil
}

// MetadataWithID implements Licence MetadataWithID function
func (lr *LicenseRepository) MetadataWithID(ctx context.Context, id string, scopes []string) (*v1.Metadata, error) {
	q := `{
		Metadatas(func: uid(` + id + `)) @cascade{
		   ID:         uid
		   Source:     metadata.source
		   Attributes: metadata.attributes
		}
	  }`

	resp, err := lr.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/Metadata - cannot complete query")
	}

	type data struct {
		Metadatas []*v1.Metadata
	}

	metadata := data{}

	if err := json.Unmarshal(resp.GetJson(), &metadata); err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/Metadata - cannot unmarshal Json object")
	}
	if len(metadata.Metadatas) == 0 {
		// TODO: Add unit test case for this
		return nil, v1.ErrNoData
	}
	return metadata.Metadatas[0], nil
}

func convertMetadataTypeDGType(typ v1.MetadataType) (metadataType, error) {
	switch typ {
	case v1.MetadataTypeEquipment:
		return metadataTypeEquipment, nil
	default:
		return metadataTypeUnsupported, fmt.Errorf("dgraph/metadataID - is not supported for MetadataType: %v", typ)
	}
}
