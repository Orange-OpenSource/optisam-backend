package postgres

import (
	"database/sql"
	gendb "optisam-backend/dps-service/pkg/repository/v1/postgres/db"
)

var repoObj *DpsRepository

// DpsRepository is struct for service to repo
type DpsRepository struct {
	db *sql.DB
	*gendb.Queries
}

// GetDpsRepository give repo object
func GetDpsRepository() (obj *DpsRepository) {
	return repoObj
}

// SetDpsRepository creates new Repository
func SetDpsRepository(db *sql.DB) {
	if repoObj == nil {
		repoObj = &DpsRepository{
			db:      db,
			Queries: gendb.New(db)}
	}
}
