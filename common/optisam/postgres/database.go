package postgres

import (
	"database/sql"
	"fmt"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/pgmigration"
	"strings"

	// pq driver
	"github.com/gobuffalo/packr/v2"
	_ "github.com/lib/pq"
	"github.com/opencensus-integrations/ocsql"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

// NewConnection returns a new database connection for the application.
func NewConnection(config Config) (*sql.DB, error) {
	driverName, err := ocsql.Register(
		"postgres",
		ocsql.WithAllTraceOptions(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to register ocsql driver")
	}

	db, err := sql.Open(driverName, config.DSN())

	// Max connections for DB to run multiple queries parellel/concurrent
	db.SetMaxOpenConns(1000)
	// max idle connections in pool time
	db.SetMaxIdleConns(0)
	return db, errors.WithStack(err)
}

func ConnectDBExecMig(dbcfg DBConfig) (db *sql.DB, err error) {
	var dba *sql.DB
	db, err = NewConnection(Config{
		Host: dbcfg.Host,
		Port: dbcfg.Port,
		Name: dbcfg.User.Name,
		User: dbcfg.User.User,
		Pass: dbcfg.User.Pass,
	})
	if err != nil {
		logger.Log.Error("failed to open connection with postgres: %v", zap.Error(err))
		return
	}
	defer func() {
		if err != nil {
			db.Close()
			if dba != nil {
				dba.Close()
			}
		}
	}()
	// Verify connection.
	if err = db.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		return
	}
	logger.Log.Info("Postgres connection verified to", zap.Any("", dbcfg.Host))
	// defer db.Close()

	dba, err = NewConnection(Config{
		Host: dbcfg.Host,
		Port: dbcfg.Port,
		Name: dbcfg.User.Name,
		User: dbcfg.Admin.User,
		Pass: dbcfg.Admin.Pass,
	})
	if err != nil {
		logger.Log.Error("failed to open connection with postgres: %v", zap.Error(err))
		return
	}
	defer func() {
		if r := recover(); r != nil {
			dba.Close()
			logger.Log.Error("Panic recovered from run server", zap.Any("recover", r))
		}
	}()
	// Verify connection.
	if err = dba.Ping(); err != nil {
		logger.Log.Error("failed to verify connection to PostgreSQL: %v", zap.Error(err))
		return
	}
	logger.Log.Info("Postgres connection verified to", zap.Any("", dbcfg.Host))

	// Run Migration
	migrations := &migrate.PackrMigrationSource{
		// "./../../pkg/repository/v1/postgres/schema"),
		Box: packr.New("migrations", dbcfg.Migration.MigrationPath),
	}
	version := strings.Split(dbcfg.Migration.Version, ";")
	direction := dbcfg.Migration.Direction
	logger.Log.Info("migration parameters dir: " + dbcfg.Migration.Direction + " versions: " + dbcfg.Migration.Version + " dir path: " + dbcfg.Migration.MigrationPath)
	n, err := pgmigration.ExecMigrations(dba, migrations, direction, version)
	dba.Close()
	if err != nil {
		logger.Log.Error("failed to execute database migration: ", zap.Error(err))
		return
	}

	logger.Log.Info("Migration", zap.Int("Migration Applied", n))
	if direction == "down" {
		logger.Log.Info("exiting due to rollback ", zap.Error(err))
		err = fmt.Errorf("exiting due to rollback ")
		return
	}
	return
}
