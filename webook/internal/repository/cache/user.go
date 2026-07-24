package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"webook/webook/internal/domain"
)

var ErrKeyNotExist = redis.Nil

type UserCache interface {
	SetUser(ctx context.Context, u domain.User) error
	GetUser(ctx context.Context, id int64) (domain.User, error)
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

/*
NewUserCache
1. redis 不要在内部初始化
2. 也不要定义在全局变量 让 main 函数初始化

A 用到了 B, B 一定是接口
A 用到了 B, B 一定是 A 的字段
A 用到了 B， B 绝对不能初始化 B, 而是外部注入
*/
func NewUserCache(client redis.Cmdable) *RedisUserCache {
	return &RedisUserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisUserCache) GetUser(ctx context.Context, id int64) (domain.User, error) {
	// ctx = context.WithValue(ctx, "biz", "user")
	// ctx = context.WithValue(ctx, "pattern", "user:info:%d")
	key := cache.key(id)
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) SetUser(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
