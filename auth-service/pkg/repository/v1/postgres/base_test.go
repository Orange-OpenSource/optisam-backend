// Copyright (C) 2019 Orange
//
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution
// or at 'http://www.apache.org/licenses/LICENSE-2.0'.
//
package postgres

import (
	"context"
	"database/sql"
	"io/ioutil"
	"optisam-backend/common/optisam/logger"
	"optisam-backend/common/optisam/postgres"
	"os"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var db *sql.DB

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	//	archive.TarWithOptions("src")
	//	archive.TarWithOptions("src")

	// cli.ImageBuild(context.Background(), bytes.NewBuffer(nil),
	// 	types.ImageBuildOptions{})

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "postgres",
		ExposedPorts: nat.PortSet{"5432": struct{}{}},
		Env: []string{
			"POSTGRES_DB=optisam",
			"POSTGRES_USER=optisam",
			"POSTGRES_PASSWORD=optisam",
		},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("5432"): {{HostIP: "127.0.0.1", HostPort: "5432"}}},
	}, nil, "optisam")
	if err != nil {
		panic(err)
	}

	cleanup := func() {
		//	return
		if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
			panic(err)
		}

		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
			panic(err)
		}
	}
	defer func() {
		cleanup()
	}()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	pgDB, err := postgres.NewConnection(postgres.Config{
		Host: "127.0.0.1",
		Port: 5432,
		User: "optisam",
		Pass: "optisam",
		Name: "optisam",
	})
	if err != nil {
		panic(err)
	}

	if err := pgDB.Ping(); err != nil {
		panic(err)
	}

	db = pgDB

	if err := loadData(); err != nil {
		panic(err)
	}
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func loadData() error {
	files := []string{"scripts/1_user_login.sql"}
	for _, file := range files {
		query, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(query)); err != nil {
			return err
		}

	}
	return nil
}
