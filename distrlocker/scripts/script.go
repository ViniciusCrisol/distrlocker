package scripts

import "github.com/redis/go-redis/v9"

var (
	AcquireScript = redis.NewScript(
		`
			if redis.call("GET", KEYS[1]) == ARGV[1]
			then
				redis.call("SET", KEYS[1], ARGV[1], "PX", ARGV[2])
				return 1
			else
				return redis.call("SET", KEYS[1], ARGV[1], "NX", "PX", ARGV[2])
			end
		`,
	)

	ReleaseScript = redis.NewScript(
		`
			if redis.call("GET", KEYS[1]) == ARGV[1]
			then
				return redis.call("DEL", KEYS[1])
			else
				return 0
			end
		`,
	)
)
