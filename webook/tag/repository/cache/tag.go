package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"webook/webook/tag/domain"
)

var ErrKeyNotExist = redis.Nil

// TagCache 标签缓存。
type TagCache interface {
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	Append(ctx context.Context, uid int64, tags ...domain.Tag) error
	DelTags(ctx context.Context, uid int64) error
}

// RedisTagCache 基于 Redis List 的标签缓存。
type RedisTagCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// NewRedisTagCache 创建 Redis 标签缓存。
func NewRedisTagCache(client redis.Cmdable) TagCache {
	return &RedisTagCache{
		client: client,
	}
}

// DelTags 删除用户标签缓存。
func (r *RedisTagCache) DelTags(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.userTagsKey(uid)).Err()
}

// Append 追加用户标签到缓存。
func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags ...domain.Tag) error {
	data := make([]any, 0, len(tags))
	for _, tag := range tags {
		val, err := json.Marshal(tag)
		if err != nil {
			return err
		}
		data = append(data, val)
	}
	key := r.userTagsKey(uid)
	// 利用 pipeline 来执行，性能好一点
	pip := r.client.Pipeline()
	pip.RPush(ctx, key, data...)
	if r.expiration > 0 {
		pip.Expire(ctx, key, r.expiration)
	}
	_, err := pip.Exec(ctx)
	return err
}

// GetTags 读取用户标签缓存。
func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	key := r.userTagsKey(uid)
	data, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, ErrKeyNotExist
	}
	res := make([]domain.Tag, 0, len(data))
	for _, ele := range data {
		var t domain.Tag
		err = json.Unmarshal([]byte(ele), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *RedisTagCache) userTagsKey(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}
