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
func (l *LicenseRepository) MetadataAllWithType(ctx context.Context, typ v1.MetadataType, scopes ...string) ([]*v1.Metadata, error) {
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

	resp, err := l.dg.NewTxn().Query(ctx, q)
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

func convertMetadataTypeDGType(typ v1.MetadataType) (metadataType, error) {
	switch typ {
	case v1.MetadataTypeEquipment:
		return metadataTypeEquipment, nil
	default:
		return metadataTypeUnsupported, fmt.Errorf("dgraph/metadataID - is not supported for MetadataType: %v", typ)
	}
}
