package pgmigration

import (
	"database/sql"
	"fmt"
	"optisam-backend/common/optisam/helper"
	"optisam-backend/common/optisam/logger"

	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

const lookupmigrations = `
SELECT id FROM gorp_migrations order by applied_at desc limit $1`

var ids []string

func ExecMigrations(dba *sql.DB, migrations *migrate.PackrMigrationSource, direction string, migFileVer []string) (n int, err error) {
	if direction == "down" && len(migFileVer) > 0 {
		dbversions, err := getdbversions(dba, migFileVer)
		if err != nil {
			logger.Log.Error("Migration Error: %v", zap.Error(err))
			return 0, err
		}
		if helper.ExactCompareSlices(dbversions, migFileVer) {
			n, err = migrate.ExecMax(dba, "postgres", migrations, migrate.Down, len(migFileVer))
			if err != nil {
				logger.Log.Error("Migration Error: %v", zap.Error(err))
				return n, err
			}
		} else {
			logger.Log.Error("Migration Error: Order of migration files did'nt matched")
			return n, fmt.Errorf("Order of migration files did'nt matched")
		}
	} else {
		n, err = migrate.Exec(dba, "postgres", migrations, migrate.Up)
		if err != nil {
			logger.Log.Error("Migration Error: %v", zap.Error(err))
			return
		}
	}
	return
}

func getdbversions(dba *sql.DB, migFileVer []string) ([]string, error) {
	rows, err := dba.Query(lookupmigrations, len(migFileVer))
	if err != nil {
		logger.Log.Error("Migration Error -getdbversions - %v", zap.Error(err))
		return nil, err
	}
	var ids []string
	var i string
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			logger.Log.Error("Migration Error -getdbversions - parsing error- %v", zap.Error(err))
			return nil, err
		}
		ids = append(ids, i)
	}
	return ids, nil
}
