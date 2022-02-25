package postgres

import (
	"database/sql"
	gendb "optisam-backend/common/optisam/workerqueue/repository/postgres/db"
)

type Repository struct {
	*gendb.Queries
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Queries: gendb.New(db),
		db:      db,
	}
}
