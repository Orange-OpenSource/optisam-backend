package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Config contains container info
type Config struct {

	// Name of the container
	Name string

	// Host Ip to bind with container
	Host string

	// Bindings of container port to system ports{ "containerPOrt:sysport1,sysport2"}
	Bindings []string

	// Image name that will be run in container
	Image string

	// Env are environment variables, if need to set in container{eg: PWD="abc1123"}
	Env []string

	// Cmd are commands need to execute in container before performing any task in container
	Cmd []string

	// Wait before executing any command
	Wait time.Duration

	Tty bool
}

// DockerInfo contains docker info
type DockerInfo struct { // nolint: golint
	cli  *client.Client
	name string
	ctx  context.Context
	body container.ContainerCreateCreatedBody
}
