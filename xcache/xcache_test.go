package xcache

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryCacheBasic(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	// 不存在
	if _, err := c.Get(ctx, "k1"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get 不存在应返回 ErrNotFound, got %v", err)
	}
	if ok, _ := c.Exists(ctx, "k1"); ok {
		t.Error("k1 不应存在")
	}

	// Set + Get
	if err := c.Set(ctx, "k1", []byte("v1"), time.Minute); err != nil {
		t.Fatal(err)
	}
	val, err := c.Get(ctx, "k1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(val, []byte("v1")) {
		t.Errorf("Get = %q, want v1", val)
	}
	if ok, _ := c.Exists(ctx, "k1"); !ok {
		t.Error("k1 应存在")
	}

	// TTL
	ttl, err := c.TTL(ctx, "k1")
	if err != nil || ttl <= 0 {
		t.Errorf("TTL = %v, want >0", ttl)
	}

	// Del
	if err := c.Del(ctx, "k1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "k1"); !errors.Is(err, ErrNotFound) {
		t.Error("删除后应不存在")
	}
}

func TestMemoryCacheExpire(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	if err := c.Set(ctx, "tmp", []byte("x"), 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(80 * time.Millisecond)

	if _, err := c.Get(ctx, "tmp"); !errors.Is(err, ErrNotFound) {
		t.Errorf("过期后应不存在, got %v", err)
	}
}

func TestMemoryCacheNoExpire(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	// 不传 ttl（<=0）→ 永不过期
	if err := c.Set(ctx, "forever", []byte("x"), 0); err != nil {
		t.Fatal(err)
	}
	ttl, _ := c.TTL(ctx, "forever")
	if ttl != -1 {
		t.Errorf("TTL = %v, want -1（永不过期）", ttl)
	}
}

func TestMemoryCacheIsolation(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	orig := []byte("data")
	_ = c.Set(ctx, "k", orig, 0)
	// 外部修改原切片，不应影响缓存内容
	orig[0] = 'X'

	val, _ := c.Get(ctx, "k")
	if val[0] == 'X' {
		t.Error("缓存应隔离外部修改（值拷贝）")
	}
}
