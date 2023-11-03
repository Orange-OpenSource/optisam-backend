package docker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitlab.tech.orange/optisam/optisam-it/optisam-services/common/optisam/logger"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
)

var cli *client.Client

func initClient() error {
	var err error
	if cli == nil {
		cli, err = client.NewEnvClient()
		if err != nil {
			logger.Log.Error("Failed to get client, err ", zap.String("reaons", err.Error()))
			return err
		}
	}
	return nil
}

func getPortBindingMap(binds []string, host string) (nat.PortMap, error) {
	var val nat.PortMap
	if len(binds) == 0 || host == "" {
		err := fmt.Errorf("bindings/host is missing in config")
		return val, err
	}
	val = make(nat.PortMap)
	var containerPort nat.Port
	for _, v := range binds {
		bindings := []nat.PortBinding{}
		ports := strings.Split(v, ":")
		if len(ports) == 2 {
			containerPort = nat.Port(strings.TrimSpace(ports[0]))
			systemPorts := strings.Split(ports[1], ",")
			for _, port := range systemPorts {
				bindings = append(bindings, nat.PortBinding{HostIP: host, HostPort: strings.TrimSpace(port)})
			}
		} else {
			continue
		}
		val[containerPort] = bindings
	}

	return val, nil
}

func getHostConfig(cfg Config) (*container.HostConfig, error) {
	bindings, err := getPortBindingMap(cfg.Bindings, cfg.Host)
	if err != nil {
		return nil, err
	}

	return &container.HostConfig{PortBindings: bindings}, nil
}

func getContainerConfig(cfg Config) (*container.Config, error) {
	type temp struct{}

	if len(cfg.Bindings) == 0 {
		return nil, fmt.Errorf("bindings are missing in config")
	}
	if cfg.Image == "" {
		return nil, fmt.Errorf("image is missing in config")
	}
	mappings := make(nat.PortSet)
	for _, v := range cfg.Bindings {
		exposedPort := strings.Split(v, ":")[0]
		mappings[nat.Port(exposedPort)] = temp{}
	}
	return &container.Config{
		Image:        cfg.Image,
		ExposedPorts: mappings,
		Env:          cfg.Env,
		Cmd:          cfg.Cmd,
		Tty:          cfg.Tty,
	}, nil
}

// Start function starts the docker basis on configuration
func Start(cfg []Config) (dockers []*DockerInfo, err error) {
	err = initClient()
	if err != nil {
		return dockers, err
	}

	dockers = make([]*DockerInfo, 0)
	prerequisites(cfg)
	for _, data := range cfg {
		docker := &DockerInfo{}
		var containerConfig *container.Config
		var hostConfig *container.HostConfig
		if data.Name == "" {
			return dockers, fmt.Errorf("docker/container name is missing , mandatory")
		}
		docker.name = data.Name
		hostConfig, err = getHostConfig(data)
		if err != nil {
			logger.Log.Error("failed to get host config ,err :", zap.String("reasong", err.Error()))
			return dockers, err
		}
		containerConfig, err = getContainerConfig(data)
		if err != nil {
			logger.Log.Error("Failed to get container config, err: ", zap.String("resson", err.Error()))
			return dockers, err
		}
		docker.cli = cli
		docker.ctx = context.Background()
		docker.body, err = docker.cli.ContainerCreate(docker.ctx, containerConfig, hostConfig, nil, data.Name)
		if err != nil {
			logger.Log.Error("Failed to bind ports and create docker, err: ", zap.String("resson", err.Error()))
			return dockers, err
		}

		err = docker.cli.ContainerStart(docker.ctx, docker.body.ID, types.ContainerStartOptions{})
		if err != nil {
			logger.Log.Error("Failed to start docker container , err : ", zap.String("resson", err.Error()))
			docker.cli.ContainerRemove(docker.ctx, docker.body.ID, types.ContainerRemoveOptions{})
			return dockers, err
		}
		if len(containerConfig.Cmd) > 0 {
			time.Sleep(data.Wait * time.Second)
			for _, cmd := range containerConfig.Cmd {
				var execute types.IDResponse
				execConf := types.ExecConfig{
					Tty: containerConfig.Tty,
					Cmd: []string{cmd},
				}
				execute, err = docker.cli.ContainerExecCreate(docker.ctx, docker.body.ID, execConf)
				if err != nil {
					logger.Log.Error("Failed to create docker cmd executer in container for cmd : ", zap.String("reason", err.Error()))
					return dockers, err
				}
				if err = docker.cli.ContainerExecStart(docker.ctx, execute.ID, types.ExecStartCheck{}); err != nil {
					logger.Log.Error("Failed to execute cmd in docker : ", zap.String("reason", err.Error()))
					panic(err)
				}
			}
		}
		dockers = append(dockers, docker)
	}
	return dockers, nil
}

func prerequisites(cfg []Config) {
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}
	for _, c := range containers {
		if isContainerRunning(cfg, c.Names) {
			if err := cli.ContainerStop(context.Background(), c.ID, nil); err != nil {
				panic(err)
			}

			if err := cli.ContainerRemove(context.Background(), c.ID, types.ContainerRemoveOptions{}); err != nil {
				panic(err)
			}
			logger.Log.Error("Container " + strings.Join(c.Names, ",") + " removed")
		}
	}
}

func isContainerRunning(cfg []Config, names []string) bool {
	temp := make(map[string]bool)
	for _, key := range names {
		temp[key] = true
	}

	for _, data := range cfg {
		str := "/" + data.Name
		if temp[str] {
			return true
		}
	}
	return false

}

// Stop func stop the docker image
func Stop(dockers []*DockerInfo) error {

	if dockers == nil {
		logger.Log.Error("Dockers are not created ....")
		return nil
	}
	for _, dock := range dockers {

		if dock.cli == nil {
			continue
		}
		if err := dock.cli.ContainerStop(dock.ctx, dock.body.ID, nil); err != nil {
			logger.Log.Error("ERROR : ", zap.String("reason", err.Error()))
			continue
		}

		if err := dock.cli.ContainerRemove(dock.ctx, dock.body.ID, types.ContainerRemoveOptions{}); err != nil {
			logger.Log.Error("ERROR : ", zap.String("reason", err.Error()))
			continue
		}
	}
	return nil
}
