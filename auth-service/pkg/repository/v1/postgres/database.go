package postgres

import (
	"database/sql"

	"github.com/go-redis/redis/v8"
)

// Default implemets ../v1.Repository interface
type Default struct {
	db *sql.DB
	r  *redis.Client
}

// NewRepository returns an implementation of Repository interface.
func NewRepository(db *sql.DB, r *redis.Client) *Default {
	return &Default{
		db: db,
		r:  r,
	}
}
