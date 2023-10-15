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
	ctx := context.Background()

	r, err := scripts.ReleaseScript.Run(ctx, lock.RedisClient, []string{lock.Key}, lock.Val).Result()
	if err == redis.Nil || !lock.isResponseGood(r) {
		return errs.ErrLockCannotBeReleased
	}
	return err
}

func (lock *Lock) isResponseGood(r interface{}) bool {
	i := r.(int64)
	return i == 1
}
