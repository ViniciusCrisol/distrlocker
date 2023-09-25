package mocks

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func NewRedisContainer() (string, func()) {
	rContainer := redisContainer{
		dockerClient: newDockerClient(),
	}
	return rContainer.create()
}

func newDockerClient() *client.Client {
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
	return c
}

const (
	containerImg     = "redis"
	redisDefaultPort = "6379"
	redisDefaultHost = "127.0.0.1"
)

type redisContainer struct {
	dockerClient *client.Client
}

func (rContainer *redisContainer) create() (string, func()) {
	ctx := context.Background()
	containerURL, portBindings := rContainer.getConnConfigs()

	cont, err := rContainer.dockerClient.ContainerCreate(
		ctx,
		&container.Config{Image: containerImg},
		&container.HostConfig{PortBindings: portBindings}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	err = rContainer.dockerClient.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	killContainer := func() {
		err = rContainer.dockerClient.ContainerStop(ctx, cont.ID, container.StopOptions{})
		if err != nil {
			panic(err)
		}
		err = rContainer.dockerClient.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
		if err != nil {
			panic(err)
		}
	}

	return containerURL, killContainer
}

func (rContainer *redisContainer) getConnConfigs() (string, nat.PortMap) {
	containerPort, err := nat.NewPort("tcp", redisDefaultPort)
	if err != nil {
		panic(err)
	}

	containerURL := fmt.Sprintf("%s:%s", redisDefaultHost, redisDefaultPort)
	portBindings := nat.PortMap{
		containerPort: []nat.PortBinding{{HostIP: redisDefaultHost, HostPort: redisDefaultPort}},
	}
	return containerURL, portBindings
}
