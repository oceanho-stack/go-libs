package xcache

import (
	"context"
	"sync"
	"time"
)

// memoryItem 内存缓存项
type memoryItem struct {
	value     []byte
	expiresAt time.Time // 零值表示永不过期
}

// MemoryCache 进程内缓存实现（单机 / 测试用）
type MemoryCache struct {
	mu    sync.Mutex
	items map[string]memoryItem
	ttl   time.Duration // 默认 TTL，0 表示永不过期
}

// NewMemoryCache 创建内存缓存。
// defaultTTL<=0 时 Set 不传 ttl 则永不过期（仍可显式传 ttl）。
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: make(map[string]memoryItem)}
}

func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.items[key]
	if !ok {
		return nil, ErrNotFound
	}
	// 检查过期
	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(m.items, key)
		return nil, ErrNotFound
	}
	// 返回副本，防止外部修改内部数据
	out := make([]byte, len(item.value))
	copy(out, item.value)
	return out, nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	v := make([]byte, len(value))
	copy(v, value)

	item := memoryItem{value: v}
	if ttl > 0 {
		item.expiresAt = time.Now().Add(ttl)
	}
	m.items[key] = item
	return nil
}

func (m *MemoryCache) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.items, k)
	}
	return nil
}

func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.items[key]
	return ok, nil
}

func (m *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, ok := m.items[key]
	if !ok {
		return -2, nil
	}
	if item.expiresAt.IsZero() {
		return -1, nil
	}
	return time.Until(item.expiresAt), nil
}

func (m *MemoryCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = make(map[string]memoryItem)
	return nil
}

// compile-time 断言
var _ Cache = (*MemoryCache)(nil)
