// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metadataType string

const (
	metadataTypeEquipment   metadataType = "equipment"
	metadataTypeUnsupported metadataType = "unsupported"
)

func (r *EquipmentRepository) UpsertMetadata(ctx context.Context, metadata *v1.Metadata) error {
	q := `query {
		var(func: eq(metadata.source,"` + metadata.Source + `")) @filter(eq(type_name,"metadata")){
			metadata as uid
			}
		}
		`
	set := `
		uid(metadata) <type_name> "metadata" .
		uid(metadata) <dgraph.type> "Metadata" .
		uid(metadata) <metadata.source> "` + metadata.Source + `" .
		uid(metadata) <metadata.type> "` + metadata.MetadataType + `" .
	`
	for _, attr := range metadata.Attributes {
		set += `
		uid(metadata) <metadata.attributes> "` + attr + `" .
		`
	}
	mu := &api.Mutation{
		SetNquads: []byte(set),
		//	CommitNow: true,
	}
	log.Printf("MU %+v", mu)
	_, err := r.dg.NewTxn().Do(ctx, &api.Request{
		CommitNow: true,
		Query:     q,
		Mutations: []*api.Mutation{mu}},
	)
	if err != nil {
		logger.Log.Error("dgraph/UpsertMetadata - failed to mutate", zap.String("reason", err.Error()))
		return fmt.Errorf("dgraph/UpsertMetadata - failed to mutuate")
	}
	return nil
}

// MetadataAllWithType implements Licence MetadataAllWithType function
func (lr *EquipmentRepository) MetadataAllWithType(ctx context.Context, typ v1.MetadataType, scopes []string) ([]*v1.Metadata, error) {
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
func (lr *EquipmentRepository) MetadataWithID(ctx context.Context, id string, scopes []string) (*v1.Metadata, error) {
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
