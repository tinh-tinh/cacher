# Cache for Tinh Tinh

<div align="center">
<img alt="GitHub Release" src="https://img.shields.io/github/v/release/tinh-tinh/cacher">
<img alt="GitHub License" src="https://img.shields.io/github/license/tinh-tinh/cacher">
<a href="https://codecov.io/gh/tinh-tinh/cacher" > 
 <img src="https://codecov.io/gh/tinh-tinh/cacher/graph/badge.svg?token=5P267CM3SA"/> 
 </a>
<a href="https://pkg.go.dev/github.com/tinh-tinh/cacher"><img src="https://pkg.go.dev/badge/github.com/tinh-tinh/cacher.svg" alt="Go Reference"></a>
</div>

<div align="center">
    <img src="https://avatars.githubusercontent.com/u/178628733?s=400&u=2a8230486a43595a03a6f9f204e54a0046ce0cc4&v=4" width="200" alt="Tinh Tinh Logo">
</div>

## Overview

The Cache Manager provides a unified API to manage caching in Tinh Tinh applications. It supports memory, Memcache, and Redis backends, and can be flexibly configured and injected into your modules and controllers.

## Features

- 🔌 **Pluggable Backends:** Supports in-memory, Memcache, and Redis stores
- 🛡️ **Type Safety:** Generic interface for strong typing
- 🏷️ **Namespace Support:** Isolate cache by logical namespace
- 📦 **Compression:** Optional data compression
- 🌐 **Context-aware:** Supports context propagation for advanced use cases
- 🎣 **Hooks:** Register hooks for cache lifecycle events

## Installation

```bash
go get -u github.com/tinh-tinh/cacher/v2
```

## Basic Usage

### Setting Up an In-Memory Cache

```go
import "github.com/tinh-tinh/cacher/v2"

cache := cacher.NewSchema[string](cacher.Config{
    Store: cacher.NewInMemory(cacher.StoreOptions{
        Ttl: 15 * time.Minute,
    }),
})

err := cache.Set("users", "John")
data, err := cache.Get("users")
```

### Using Namespaces

```go
cache1 := cacher.NewSchema[string](cacher.Config{
    Store:     store,
    Namespace: "cache1",
})
cache2 := cacher.NewSchema[string](cacher.Config{
    Store:     store,
    Namespace: "cache2",
})
```

### Memcache Example

```go
import "github.com/tinh-tinh/cacher/storage/memcache"

cache := memcache.New(memcache.Options{
    Addr: []string{"localhost:11211"},
    Ttl:  15 * time.Minute,
})
```

### Redis Example

```go
import (
    "github.com/tinh-tinh/cacher/storage/redis"
    redis_store "github.com/redis/go-redis/v9"
)

cache := redis.New(redis.Options{
    Connect: &redis_store.Options{
        Addr: "localhost:6379",
    },
    Ttl: 15 * time.Minute,
})
```

## API Overview

The main cache interface provides:

- `Set(key, value, opts...)`: Store a value
- `Get(key)`: Retrieve a value
- `Delete(key)`: Remove a value
- `Clear()`: Remove all values
- `MSet(...params)`: Batch set
- `MGet(...keys)`: Batch get

## Module Integration

You can register the cache as a provider in a Tinh Tinh module and inject it into controllers:

```go
import (
    "github.com/tinh-tinh/cacher/v2"
    "github.com/tinh-tinh/tinhtinh/v2/core"
)

func userController(module core.Module) core.Controller {
    cache := cacher.Inject[[]byte](module)
    ctrl := module.NewController("users")

    ctrl.Get("", func(ctx core.Ctx) error {
        data, err := cache.Get("users")
        // handle data
    })
    return ctrl
}
```

To register the cache provider:

```go
module := core.NewModule(core.NewModuleOptions{
    Imports: []core.Modules{
        cacher.Register(cacher.Config{ Store: cache }),
        userModule,
    },
})
```

## Advanced Features

### Compression
Set `CompressAlg` in `Config` to enable data compression:

```go
cache := cacher.NewSchema[string](cacher.Config{
    Store: store,
    CompressAlg: "gzip", // Enable compression
})
```

### Hooks
Use the `Hooks` field to register cache lifecycle hooks:

```go
cache := cacher.NewSchema[string](cacher.Config{
    Store: store,
    Hooks: []cacher.Hook{
        // Add your hooks here
    },
})
```

### Context Operations
Use `SetCtx` and `GetCtx` for context-aware operations:

```go
ctx := context.Background()
err := cache.SetCtx(ctx, "key", value)
data, err := cache.GetCtx(ctx, "key")
```

## Testing

The repository includes comprehensive tests for all stores and features. See:
- `cacher_test.go`
- `storage/memcache/memcache_test.go`

## Contributing

We welcome contributions! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or need help, you can:
- Open an issue in the GitHub repository
- Check our documentation
- Join our community discussions
