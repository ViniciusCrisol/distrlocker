package distrlocker

import (
	"context"
	"math/rand"

	"github.com/redis/go-redis/v9"

	"distrlocker/distrlocker/errs"
	"distrlocker/distrlocker/lock"
	"distrlocker/distrlocker/scripts"
)

type DistrLocker struct {
	timeout     int
	redisClient redis.Scripter
}

func NewDistrLocker(timeout int, redisClient redis.Scripter) *DistrLocker {
	return &DistrLocker{
		timeout:     timeout,
		redisClient: redisClient,
	}
}

func (locker *DistrLocker) Acquire(key string) (lock.Lock, error) {
	var (
		lck lock.Lock
		val = rand.Int()
		ctx = context.Background()
		err = locker.acquireLock(ctx, key, val)
	)
	if err != nil {
		return lck, err
	}
	lck.Key = key
	lck.Val = val
	lck.RedisClient = locker.redisClient

	return lck, nil
}

func (locker *DistrLocker) acquireLock(ctx context.Context, key string, val int) error {
	ks := []string{key}

	_, err := scripts.AcquireScript.Run(ctx, locker.redisClient, ks, val, locker.timeout).Result()
	if err == redis.Nil {
		return errs.ErrLockCannotBeAcquired
	}
	return err
}
