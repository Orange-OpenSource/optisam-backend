package postgres

import (
	"database/sql"
	"optisam-backend/simulation-service/pkg/repository/v1/postgres/db"
)

// SimulationServiceRepo implements all the methods defined by repository of this interface
type SimulationServiceRepo struct {
	db *sql.DB
	*db.Queries
}

// NewSimulationServiceRepository returns an implementation of repository
func NewSimulationServiceRepository(d *sql.DB) *SimulationServiceRepo {
	return &SimulationServiceRepo{
		db:      d,
		Queries: db.New(d),
	}

}
