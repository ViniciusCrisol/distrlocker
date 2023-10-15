package distrlocker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/redis/go-redis/v9"

	"distrlocker/distrlocker/errs"
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
		&container.HostConfig{PortBindings: portBindings}, nil, nil, "",
	)
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

func TestAcquire(ts *testing.T) {
	url, rm := NewRedisContainer()
	defer rm()

	dsl := NewDistrLocker(
		5000,
		redis.NewClient(&redis.Options{Addr: url, WriteTimeout: time.Second * 3}),
	)

	ts.Run(
		"It should acquire the lock if it is available", func(t *testing.T) {
			l, err := dsl.Acquire("key")
			if err != nil {
				t.Errorf("Failed to acquire lock")
			}

			if l.Release() != nil {
				t.Errorf("Failed to release lock")
			}
		},
	)

	ts.Run(
		"It should not acquire a lock if it is not available", func(t *testing.T) {
			l, err := dsl.Acquire("key")
			if err != nil {
				t.Errorf("Failed to acquire lock")
			}

			_, err = dsl.Acquire("key")
			if err != errs.ErrLockCannotBeAcquired {
				t.Errorf("Failed to lock resource")
			}

			if l.Release() != nil {
				t.Errorf("Failed to release lock")
			}
		},
	)

	ts.Run(
		"It should simulate a distributed lock system", func(t *testing.T) {
			wg := sync.WaitGroup{}
			execSequence := []int{}
			veryHardProcessing := func(id int) {
				defer wg.Done()

				for {
					l, err := dsl.Acquire("key")
					if err == errs.ErrLockCannotBeAcquired {
						continue
					}
					if err != nil {
						t.Errorf("Failed to acquire lock")
					}

					execSequence = append(execSequence, id)
					time.Sleep(time.Millisecond * 50)
					execSequence = append(execSequence, id)

					if l.Release() != nil {
						t.Errorf("Failed to release lock")
					}
					return
				}
			}

			for i := 1; i <= 80; i++ {
				wg.Add(1)
				go veryHardProcessing(i)
			}
			wg.Wait()

			for i, id := range execSequence {
				if i%2 != 0 && id != execSequence[i-1] {
					t.Errorf("The lock process failed")
				}
			}
		},
	)
}
