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
	var l lock.Lock
	val := rand.Int()
	ctx := context.Background()

	err := locker.acquireLock(ctx, key, val)
	if err != nil {
		return l, err
	}
	l.Key = key
	l.Val = val
	l.RedisClient = locker.redisClient

	return l, nil
}

func (locker *DistrLocker) acquireLock(ctx context.Context, key string, val int) error {
	_, err := scripts.AcquireScript.Run(ctx, locker.redisClient, []string{key}, val, locker.timeout).Result()
	if err == redis.Nil {
		return errs.ErrLockCannotBeAcquired
	}
	return err
}
