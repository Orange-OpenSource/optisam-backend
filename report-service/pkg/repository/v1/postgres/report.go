package postgres

import (
	"database/sql"

	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/report-service/pkg/repository/v1/postgres/db"
)

// ReportRepository
type ReportRepository struct {
	*gendb.Queries
	db *sql.DB
}

// NewReportRepository creates new Repository
func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}
