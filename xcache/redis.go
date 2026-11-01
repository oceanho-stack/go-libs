package xcache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions Redis 连接配置
type RedisOptions struct {
	Addr         string        // 地址 host:port
	Password     string        // 密码
	DB           int           // 数据库编号
	PoolSize     int           // 连接池大小（默认 20）
	MinIdleConns int           // 最小空闲连接
	DialTimeout  time.Duration // 连接超时（默认 5s）
	ReadTimeout  time.Duration // 读超时（默认 3s）
	WriteTimeout time.Duration // 写超时（默认 3s）
}

// Redis 基于 go-redis 的实现
type Redis struct {
	client *redis.Client
}

// NewRedis 创建 Redis 缓存
func NewRedis(opts RedisOptions) (*Redis, error) {
	if opts.Addr == "" {
		opts.Addr = "localhost:6379"
	}
	if opts.PoolSize == 0 {
		opts.PoolSize = 20
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 3 * time.Second
	}

	client := redis.NewClient(&redis.Options{
		Addr:         opts.Addr,
		Password:     opts.Password,
		DB:           opts.DB,
		PoolSize:     opts.PoolSize,
		MinIdleConns: opts.MinIdleConns,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	})

	// 连通性检查
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return &Redis{client: client}, nil
}

// Get 获取 key 字节值
func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	return val, err
}

// Set 设置 key
func (r *Redis) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		// 永不过期
		return r.client.Set(ctx, key, value, 0).Err()
	}
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Del 删除 key
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists key 是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

// TTL 剩余存活时间
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Close 关闭连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// compile-time 断言
var _ Cache = (*Redis)(nil)
