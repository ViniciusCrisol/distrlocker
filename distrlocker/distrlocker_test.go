package distrlocker

import (
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"distrlocker/distrlocker/errs"
	"distrlocker/distrlocker/mocks"
)

func TestAcquire(ts *testing.T) {
	url, rm := mocks.NewRedisContainer()
	defer rm()

	dsl := NewDistrLocker(
		5000,
		redis.NewClient(&redis.Options{Addr: url, WriteTimeout: time.Second * 3}))

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
	})
}
