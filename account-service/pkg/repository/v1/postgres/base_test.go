// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 

package postgres

import (
	"database/sql"
	base "optisam-backend/account-service/pkg/repository/v1/postgres/common"
	"optisam-backend/common/optisam/docker"
	"os"
	"testing"
)

// nolint: gochecknoglobals
var db *sql.DB

const (
	serverPath = "../../../../cmd/server/"
	dropQuery  = `
	DROP TABLE IF EXISTS group_ownership;
	DROP TABLE IF EXISTS groups;
	DROP TABLE IF EXISTS scopes;
	DROP TABLE IF EXISTS users;
	DROP TABLE IF EXISTS users_audit;
	DROP TYPE  IF EXISTS audit_status;
	DROP TABLE IF EXISTS roles;
	`
)

func TestMain(m *testing.M) {
	var err error
	var dockers []*docker.DockerInfo
	cleanup := func() {
		if db == nil {
			return
		}
		if _, err := db.Exec(string(dropQuery)); err != nil {
			panic(err)
		}
	}
	defer func() {
		cleanup()
		docker.Stop(dockers)
	}()
	files := []string{"scripts/1_user_login.sql", "schema/2_add_users_audit_table.sql"}
	db, dockers, err = base.Testdata(serverPath, files)
	if err != nil {
		panic(err)
	}
	code := m.Run()
	cleanup()
	docker.Stop(dockers)
	os.Exit(code)
}
