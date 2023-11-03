package postgres

import (
	"database/sql"

	gendb "gitlab.tech.orange/optisam/optisam-it/optisam-services/notification-service/pkg/repository/v1/postgres/db"
)

const (
	YYYYMMDD string = "2006-01-02"
	DDMMYYYY string = "02-01-2006"
)

var dateFormats = []string{YYYYMMDD, DDMMYYYY}

// NotificationRepository ...
type NotificationRepository struct {
	*gendb.Queries
	db *sql.DB
}

// NewNotificationRepository creates new Repository
func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{
		Queries: gendb.New(db),
		db:      db,
	}
}

// NotificationRepositoryTx ...
type NotificationRepositoryTx struct {
	*gendb.Queries
	db *sql.Tx
}

// NewNotificationRepositoryTx ...
func NewNotificationRepositoryTx(db *sql.Tx) *NotificationRepositoryTx {
	return &NotificationRepositoryTx{
		Queries: gendb.New(db),
		db:      db,
	}
}
