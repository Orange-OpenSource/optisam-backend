package dgraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"optisam-backend/common/optisam/logger"
	v1 "optisam-backend/equipment-service/pkg/repository/v1"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"go.uber.org/zap"
)

type metadataType string
type metadata struct {
	ID         string
	Source     string
	Attributes []string
	Scopes     []string
}

const (
	metadataTypeEquipment   metadataType = "equipment"
	metadataTypeUnsupported metadataType = "unsupported"
)

// UpsertMetadata ...
func (r *EquipmentRepository) UpsertMetadata(ctx context.Context, metadata *v1.Metadata) (string, error) {

	q := `query {
		var(func: eq(metadata.source,"` + metadata.Source + `"))  @filter(eq(type_name, "metadata") AND eq(scopes,"` + metadata.Scope + `")){
			metadata as uid
			}
		}
		`
	set := `
		uid(metadata) <type_name> "metadata" .
		uid(metadata) <dgraph.type> "Metadata" .
		uid(metadata) <metadata.source> "` + metadata.Source + `" .
		uid(metadata) <metadata.type> "` + metadata.MetadataType + `" .
		uid(metadata) <scopes> "` + metadata.Scope + `" .
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
	r.mu.Lock()
	defer r.mu.Unlock()
	resp, err := r.dg.NewTxn().Do(ctx, &api.Request{
		CommitNow: true,
		Query:     q,
		Mutations: []*api.Mutation{mu}},
	)
	if err != nil {
		logger.Log.Error("dgraph/UpsertMetadata - failed to mutate", zap.String("reason", err.Error()))
		return "", fmt.Errorf("dgraph/UpsertMetadata - failed to mutuate")
	}
	return resp.Uids["uid(metadata)"], nil
}

// MetadataAllWithType implements Licence MetadataAllWithType function
func (r *EquipmentRepository) MetadataAllWithType(ctx context.Context, typ v1.MetadataType, scopes []string) ([]*v1.Metadata, error) {
	id, err := convertMetadataTypeDGType(typ)
	if err != nil {
		return nil, err
	}
	q := `{
		Metadatas(func: eq(metadata.type,` + string(id) + `),orderasc: metadata.source )  ` + agregateFilters(scopeFilters(scopes)) + ` {
		   ID:         uid
		   Source:     metadata.source
		   Attributes: metadata.attributes
		   Scopes: 	   scopes
		}
	  }`
	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/Metadata - cannot complete query")
	}

	type data struct {
		Metadatas []*metadata
	}

	metadata := data{}

	if err := json.Unmarshal(resp.GetJson(), &metadata); err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()))
		return nil, fmt.Errorf("dgraph/Metadata - cannot unmarshal Json object")
	}
	return convertMetadataAll(metadata.Metadatas), nil
}

// MetadataWithID implements Licence MetadataWithID function
func (r *EquipmentRepository) MetadataWithID(ctx context.Context, id string, scopes []string) (*v1.Metadata, error) {
	q := `{
		Metadatas(func: uid(` + id + `))  ` + agregateFilters(scopeFilters(scopes)) + `@cascade{
		   ID:         uid
		   Source:     metadata.source
		   Attributes: metadata.attributes
		   Scopes: 	   scopes
		}
	  }`

	resp, err := r.dg.NewTxn().Query(ctx, q)
	if err != nil {
		logger.Log.Error("dgraph/Metadata - ", zap.String("reason", err.Error()), zap.String("query", q))
		return nil, errors.New("dgraph/Metadata - cannot complete query")
	}

	type data struct {
		Metadatas []*metadata
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
	return convertMetadata(metadata.Metadatas[0]), nil
}

func convertMetadataTypeDGType(typ v1.MetadataType) (metadataType, error) {
	switch typ {
	case v1.MetadataTypeEquipment:
		return metadataTypeEquipment, nil
	default:
		return metadataTypeUnsupported, fmt.Errorf("dgraph/metadataID - is not supported for MetadataType: %v", typ)
	}
}

func convertMetadataAll(dbData []*metadata) []*v1.Metadata {
	srvData := make([]*v1.Metadata, len(dbData))
	for i := range dbData {
		srvData[i] = convertMetadata(dbData[i])
	}
	return srvData
}

func convertMetadata(dbData *metadata) *v1.Metadata {
	return &v1.Metadata{
		ID:         dbData.ID,
		Source:     dbData.Source,
		Attributes: dbData.Attributes,
		Scope:      dbData.Scopes[0],
	}
}
