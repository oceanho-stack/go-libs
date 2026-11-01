# Package xcache 缓存抽象

xcache 是 oceanho-stack 的缓存抽象层，接口化设计，默认 Redis 实现，可替换内存实现。

## 特性

- `Cache` 接口：`Get` / `Set` / `Del` / `Exists` / `TTL` / `Close`
- 默认 `Redis` 实现（`redis/go-redis/v9`）
- 内存实现 `NewMemoryCache`（进程内，适合单机/测试）
- 序列化：值存 `[]byte`，业务层自行编码（或配合 xresp 等）

## 快速开始

```go
// Redis
c, err := xcache.NewRedis(xcache.RedisOptions{
    Addr: "localhost:6379", Password: "", DB: 0,
})
if err != nil { panic(err) }

c.Set(ctx, "key", []byte("value"), 10*time.Minute)
val, _ := c.Get(ctx, "key")

// 内存缓存（测试/单机）
mc := xcache.NewMemoryCache()
```

## 设计

- 业务代码依赖 `Cache` 接口，便于测试用内存实现替换 Redis
- 不内置缓存 key 规范、业务策略（由业务层约定）
