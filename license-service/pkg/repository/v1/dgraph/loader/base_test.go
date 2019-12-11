// Copyright (C) 2019 Orange
// 
// This software is distributed under the terms and conditions of the 'Apache License 2.0'
// license which can be found in the file 'License.txt' in this package distribution 
// or at 'http://www.apache.org/licenses/LICENSE-2.0'. 
//
package loader

import (
	"context"
	"optisam-backend/common/optisam/dgraph"
	"optisam-backend/common/optisam/logger"

	"go.uber.org/zap"

	"os"
	"testing"

	"github.com/dgraph-io/dgo"

	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

var dgClient *dgo.Dgraph

func TestMain(m *testing.M) {
	logger.Init(-1, "")
	badgerDir := "badger"
	if err := os.RemoveAll(badgerDir); err != nil {
		panic(err)
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "dgraph/dgraph",
		Tty:   true,
		ExposedPorts: nat.PortSet{
			"5080": struct{}{},
			"6080": struct{}{},
			"8080": struct{}{},
			"9080": struct{}{},
			"8000": struct{}{},
		},
		Cmd: []string{
			"dgraph",
			"zero",
		},
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			// TODO: all host ports should be customised for running the tests
			nat.Port("5080"): {{HostIP: "127.0.0.1", HostPort: "5080"}},
			nat.Port("6080"): {{HostIP: "127.0.0.1", HostPort: "6080"}},
			nat.Port("8080"): {{HostIP: "127.0.0.1", HostPort: "8080"}},
			nat.Port("9080"): {{HostIP: "127.0.0.1", HostPort: "9080"}},
			nat.Port("8000"): {{HostIP: "127.0.0.1", HostPort: "8000"}},
		},
	}, nil, "optisam-dgraph")
	if err != nil {
		panic(err)
	}
	cleanup := func() {

		if err := cli.ContainerStop(ctx, resp.ID, nil); err != nil {
			panic(err)
		}

		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
			panic(err)
		}

		if err := os.RemoveAll(badgerDir); err != nil {
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

	excResp, err := cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
		Tty: true,
		Cmd: []string{"dgraph", "alpha", "--lru_mb", "2048"},
	})
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerExecStart(ctx, excResp.ID, types.ExecStartCheck{}); err != nil {
		panic(err)
	}

	exc1Resp, err := cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
		Tty: true,
		Cmd: []string{"dgraph-ratel"},
	})
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerExecStart(ctx, exc1Resp.ID, types.ExecStartCheck{}); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)
	conn, err := dgraph.NewDgraphConnection(&dgraph.Config{
		Hosts: []string{":9080"},
	})
	if err != nil {
		logger.Log.Error("test main cannot connect to alpha", zap.String("reason", err.Error()))
	}

	dgClient = conn
	code := m.Run()
	cleanup()
	os.Exit(code)
}
