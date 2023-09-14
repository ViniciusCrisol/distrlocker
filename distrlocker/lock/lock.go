package lock

import (
	"context"

	"github.com/redis/go-redis/v9"

	"distrlocker/distrlocker/errs"
	"distrlocker/distrlocker/scripts"
)

type Lock struct {
	Key         string
	Val         int
	RedisClient redis.Scripter
}

func (lock *Lock) Release() error {
	ks := []string{lock.Key}
	ctx := context.Background()

	r, err := scripts.ReleaseScript.Run(ctx, lock.RedisClient, ks, lock.Val).Result()
	if err == redis.Nil || !lock.isRGood(r) {
		return errs.ErrLockCannotBeReleased
	}
	return err
}

func (lock *Lock) isRGood(r interface{}) bool {
	i := r.(int64)
	return i == 1
}
