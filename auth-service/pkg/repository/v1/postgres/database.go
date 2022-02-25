package postgres

import "database/sql"

// Default implemets ../v1.Repository interface
type Default struct {
	db *sql.DB
}

// NewRepository returns an implementation of Repository interface.
func NewRepository(db *sql.DB) *Default {
	return &Default{
		db: db,
	}
}
