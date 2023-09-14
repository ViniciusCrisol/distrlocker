package distrlocker

import (
	"context"
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

const (
	containerURL     = "127.0.0.1:6379"
	containerImg     = "redis:latest"
	redisDefaultHost = "127.0.0.1"
	redisDefaultPort = "6379"
)

var (
	containerPort, _ = nat.NewPort("tcp", redisDefaultPort)

	hostBinding = nat.PortBinding{
		HostIP:   redisDefaultHost,
		HostPort: redisDefaultPort,
	}

	portBinding = nat.PortMap{
		containerPort: []nat.PortBinding{hostBinding},
	}
)

func createTestContainer() func() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ctr, err := cli.ContainerCreate(
		ctx,
		&container.Config{Image: containerImg},
		&container.HostConfig{PortBindings: portBinding}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	cli.ContainerStart(ctx, ctr.ID, types.ContainerStartOptions{})
	return func() {
		cli.ContainerStop(ctx, ctr.ID, container.StopOptions{})
		cli.ContainerRemove(ctx, ctr.ID, types.ContainerRemoveOptions{})
	}
}

func TestAcquire(ts *testing.T) {
	rm := createTestContainer()
	defer rm()

	dsl := NewDistrLocker(
		5000,
		redis.NewClient(&redis.Options{Addr: containerURL, WriteTimeout: time.Second * 3}))

	ts.Run("It should not be able to acquire a lock if it is not available", func(t *testing.T) {
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
	})

	ts.Run("It should be able to acquire the lock if it is available", func(t *testing.T) {
		l, err := dsl.Acquire("key")
		if err != nil {
			t.Errorf("Failed to acquire lock")
		}

		if l.Release() != nil {
			t.Errorf("Failed to release lock")
		}
	})

	ts.Run("It should be able to simulate a distributed lock system", func(t *testing.T) {
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

		for i := 1; i <= 40; i++ {
			wg.Add(1)
			go veryHardProcessing(i)
		}
		wg.Wait()

		for i, id := range execSequence {
			if i%2 != 0 && id != execSequence[i-1] {
				t.Errorf("The lock process failed")
			}
		}
	})
}
