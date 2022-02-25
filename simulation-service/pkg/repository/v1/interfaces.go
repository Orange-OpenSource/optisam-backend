package v1

import (
	"context"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
)

//go:generate mockgen -destination=mock/mock.go -package=mock optisam-backend/simulation-service/pkg/repository/v1 Repository

// Repository interface defines all the methods defined by this interface
type Repository interface {
	db.Querier

	// CreateConfig will insert the config data into the system
	CreateConfig(ctx context.Context, masterData *MasterData, configData []*ConfigData) error

	// UpdateConfig updates the configuration
	UpdateConfig(ctx context.Context, configID int32, eqType string, metadataIDs []int32, data []*ConfigData) error
}
