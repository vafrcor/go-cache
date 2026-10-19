# go-cache

`go-cache` provides a common cache repository interface backed by existing Go
cache packages:

- [go-redis](https://github.com/redis/go-redis) for Redis
- [freecache](https://github.com/coocood/freecache) for bounded in-memory caching
- [go-cache](https://github.com/patrickmn/go-cache) for simple in-memory caching

All implementations support `Get`, `Set`, `Delete`, `Exists`, `Clear`, `TTL`,
and `Touch`. Values use JSON encoding by default, and MessagePack is also
available.

## Installation

```bash
go get github.com/vafrcor/go-cache
```

## Usage

### go-cache

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	cache "github.com/vafrcor/go-cache"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	ctx := context.Background()
	repository := cache.NewGoCacheRepository(5*time.Minute, 10*time.Minute, nil)

	if err := repository.Set(ctx, "user:1", User{ID: 1, Name: "John Doe"}, time.Minute); err != nil {
		panic(err)
	}

	var user User
	if err := repository.Get(ctx, "user:1", &user); err != nil {
		if errors.Is(err, cache.ErrCacheNotFound) {
			fmt.Println("cache miss")
			return
		}
		panic(err)
	}

	fmt.Println(user.Name)
}
```

### Freecache

The first argument is the cache size in bytes.

```go
repository := cache.NewFreecacheRepository(100*1024*1024, nil) // 100 MiB

err := repository.Set(ctx, "user:1", user, 5*time.Minute)
```

### Redis

```go
import "github.com/redis/go-redis/v9"

client := redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})
defer client.Close()

repository := cache.NewGoRedisRepository(client, nil)

err := repository.Set(ctx, "user:1", user, 5*time.Minute)
```

## Custom serialization

Pass a `codec.Codec` to a constructor to replace the default JSON codec. The
package includes a MessagePack implementation:

```go
import cachecodec "github.com/vafrcor/go-cache/codec"

repository := cache.NewGoCacheRepository(
	5*time.Minute,
	10*time.Minute,
	cachecodec.MsgpackCodec{},
)
```

## Logging

Repositories use Go's standard `log.Default()` logger. When a dedicated output
or prefix is needed, set the concrete repository's `Logger` field:

```go
import (
	"log"
	"os"
)

repository := cache.NewGoCacheRepository(5*time.Minute, 10*time.Minute, nil).(*cache.GoCacheRepository)
repository.Logger = log.New(os.Stderr, "cache: ", log.LstdFlags)
```

## Testing

```bash
go test ./...
```
