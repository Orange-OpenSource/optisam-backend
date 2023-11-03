package pgmigration

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/helper"
	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	lookupmigrations = `SELECT id FROM gorp_migrations order by applied_at desc limit $1`
	getAllmigrations = `SELECT id FROM gorp_migrations order by applied_at`
	deletemigrations = `DELETE FROM gorp_migrations WHERE id = ANY($1::TEXT[])`
)

var ids []string

func ExecMigrations(dba *sql.DB, migrations *migrate.PackrMigrationSource, direction string, migFileVer []string, migFilePath string, deletemig bool) (n int, err error) {
	if direction == "down" && len(migFileVer) > 0 {
		dbversions, err := getdbversions(dba, migFileVer)
		if err != nil {
			logger.Log.Sugar().Errorw("Migration Error:", "err", err.Error())
			return 0, err
		}
		if helper.ExactCompareSlices(dbversions, migFileVer) {
			n, err = migrate.ExecMax(dba, "postgres", migrations, migrate.Down, len(migFileVer))
			if err != nil {
				logger.Log.Sugar().Errorw("Migration Error:", "err", err.Error())
				return n, err
			}
		} else {
			logger.Log.Sugar().Errorw("Migration Error: Order of migration files did'nt matched")
			return n, fmt.Errorf("Order of migration files did'nt matched")
		}
	} else {
		if deletemig {
			err = delMigVersionDiffrence(dba, migFilePath)
			if err != nil {
				logger.Log.Sugar().Errorw("Migration Error:", "err", err.Error())
				return 0, err
			}
		}
		n, err = migrate.Exec(dba, "postgres", migrations, migrate.Up)
		if err != nil {
			logger.Log.Sugar().Errorw("Migration Error:", "err", err.Error())
			return 0, err
		}
		return n, nil
	}
	return
}

func delMigVersionDiffrence(dba *sql.DB, migrations string) error {
	// migSlice := []string{"entry4", "entry5", "entry1", "entry2", "entry3"}                    // Example migration slice
	// dbSlice := []string{"entry1", "entry2", "entry3", "entry4", "entry5", "entry6", "entry7"} // Example database slice
	files, err := ioutil.ReadDir(migrations)
	if err != nil {
		logger.Log.Sugar().Errorw("Migration Error - getting migration files from schema folder", "err", err.Error())
		return err
	}
	var migSlice []string
	for _, file := range files {
		migSlice = append(migSlice, file.Name())
	}

	dbSlice, err := getDBAllVersions(dba)
	if err != nil {
		logger.Log.Sugar().Errorw("Migration Error - getDBAllVersions ", "err", err.Error())
		return err
	}

	// Case 1: migSlice < dbSlice
	// migSlice should be an exact subset of dbSlice
	if len(migSlice) < len(dbSlice) && helper.IsSubsetSlice(migSlice, dbSlice) {
		diff := helper.DifferenceSlice(dbSlice, migSlice)
		logger.Log.Sugar().Infow("Difference:", "versions", diff) // delete query
		err := deleteMigrations(dba, diff)
		if err != nil {
			logger.Log.Sugar().Errorw("Migration Error - getDBAllVersions", "err", err.Error())
			return err
		}
	}

	// Case 2: migSlice == dbSlice
	if len(migSlice) == len(dbSlice) {
		if !helper.CompareSlices(migSlice, dbSlice) {
			diff := helper.DifferenceSlice(migSlice, dbSlice)
			logger.Log.Sugar().Errorw("Difference:", "versions", diff)
			return errors.New("migration version not matched")
		}
		diff := dbSlice[:1]
		logger.Log.Sugar().Infow("Difference:", "versions", diff) // delete query
		err := deleteMigrations(dba, diff)
		if err != nil {
			logger.Log.Sugar().Errorw("Migration Error - getDBAllVersions ", "err", err.Error())
			return err
		}
	}

	// Case 3: migSlice > dbSlice
	if len(migSlice) > len(dbSlice) {
		logger.Log.Sugar().Infow("New Migration")
		return nil
	}
	return nil
}

func getdbversions(dba *sql.DB, migFileVer []string) ([]string, error) {
	rows, err := dba.Query(lookupmigrations, len(migFileVer))
	if err != nil {
		logger.Log.Sugar().Errorw("Migration Error -getdbversions ", "err", err.Error())
		return nil, err
	}
	var ids []string
	var i string
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			logger.Log.Sugar().Errorw("Migration Error -getdbversions - parsing error", "err", err.Error())
			return nil, err
		}
		ids = append(ids, i)
	}
	return ids, nil
}

func getDBAllVersions(dba *sql.DB) ([]string, error) {
	rows, err := dba.Query(getAllmigrations)
	if err != nil {
		logger.Log.Sugar().Errorw("Migration Error -getAllmigrations ", "err", err.Error())
		return nil, err
	}
	var mids []string
	var i string
	for rows.Next() {
		if err := rows.Scan(&i); err != nil {
			logger.Log.Sugar().Errorw("Migration Error - getAllmigrations - parsing error ", "err", err.Error())
			return nil, err
		}
		mids = append(mids, i)
	}
	return mids, nil
}

func deleteMigrations(dba *sql.DB, migFileVer []string) error {
	res, err := dba.Exec(deletemigrations, pq.Array(migFileVer))
	if err != nil {
		logger.Log.Sugar().Error("Error executing deleteMigrations query: ", "err ", err.Error())
		return err
	}
	ra, _ := res.RowsAffected()
	logger.Log.Sugar().Infow("number of migrations entries deleted", "rows affected ", ra)
	return nil
}
