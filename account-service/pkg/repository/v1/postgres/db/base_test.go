package db_test

import (
	"database/sql"
	"os"
	"testing"

	base "gitlab.tech.orange/optisam/optisam-it/optisam-services/account-service/pkg/repository/v1/postgres/common"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/docker"
)

// nolint: gochecknoglobals
var sqldb *sql.DB

const (
	serverPath = "../../../../../cmd/server/"
	dropQuery  = `
	DROP TABLE IF EXISTS group_ownership;
	DROP TABLE IF EXISTS groups;
	DROP TABLE IF EXISTS scopes;
	DROP TABLE IF EXISTS users;
	DROP TABLE IF EXISTS users_audit;
	DROP TYPE  IF EXISTS audit_status;
	DROP TYPE  IF EXISTS scope_types;
	DROP TABLE IF EXISTS roles;
	`
)

func TestMain(m *testing.M) {
	var err error
	var dockers []*docker.DockerInfo
	cleanup := func() {
		if sqldb == nil {
			return
		}
		if _, err := sqldb.Exec(string(dropQuery)); err != nil {
			panic(err)
		}
	}
	defer func() {
		cleanup()
		docker.Stop(dockers)
	}()
	files := []string{"../scripts/1_user_login.sql", "../schema/2_add_users_audit_table.sql"}
	sqldb, dockers, err = base.Testdata(serverPath, files)
	if err != nil {
		panic(err)
	}
	code := m.Run()
	cleanup()
	docker.Stop(dockers)
	os.Exit(code)
}
