// Package xcache 缓存抽象：接口化，默认 Redis，可选内存实现。
package xcache

import (
	"context"
	"time"
)

// Cache 缓存接口。
// 业务代码应依赖此接口，便于测试用内存实现替换 Redis。
type Cache interface {
	// Get 获取 key 的字节值；key 不存在返回 ErrNotFound
	Get(ctx context.Context, key string) ([]byte, error)
	// Set 设置 key，ttl<=0 表示永不过期
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Del 删除 key（可多个）
	Del(ctx context.Context, keys ...string) error
	// Exists 判断 key 是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// TTL 返回 key 剩余存活时间；不存在返回 -2，无过期返回 -1
	TTL(ctx context.Context, key string) (time.Duration, error)
	// Close 关闭缓存连接
	Close() error
}

// ErrNotFound key 不存在
var ErrNotFound = &notFoundError{}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "xcache: key not found" }
