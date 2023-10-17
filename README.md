# Distrlocker

O README está disponível também em [português](README-pt.md).

---

The Distrlocker library, written in Golang, provides a simple and effective way to manage distributed locks through
Redis. Distributed locks are a valuable primitive in environments where multiple processes need to operate on shared
resources exclusively.

## Installation

To facilitate integration with Redis, this library uses the official Redis client library as a dependency 
([github.com/redis/go-redis](https://github.com/redis/go-redis)). To incorporate the packages into your
project, execute the following commands:

```bash
go get -u github.com/viniciuscrisol/distrlocker
go get -u github.com/redis/go-redis/v9
```

## Example

The following example illustrates a simple use case for acquiring and releasing a lock:

```go
func main() {
    // Creating a Redis client
    redisClient := redis.NewClient(
        &redis.Options{Addr: "127.0.0.1:6379", WriteTimeout: time.Second * 3},
    )

    // Creating a distributed locker with a timeout of 5000 ms
    locker := distrlocker.NewDistrLocker(5000, redisClient)

    // Acquiring the lock
    myLock, err := locker.Acquire("my-key")
    if err != nil {
        fmt.Println("Failed to acquire the lock:", err)
        return
    }
    // Releasing the lock. If the specified 5000 ms expire,
    // the lock will be released automatically
    defer myLock.Release()

    // Intensive processing...
}
```

## Tests

The library includes a set of tests that simulate scenarios of acquiring and releasing locks. To run the tests, Docker
must be installed. Additionally, ensure that port 6379 is available. By executing the following command, coverage will
be obtained:

```bash
go test ./... -cover -coverprofile=coverage
```
