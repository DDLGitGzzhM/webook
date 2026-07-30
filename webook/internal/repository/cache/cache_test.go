package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"webook/webook/internal/domain"
)

func TestCache_Set(t *testing.T) {
	// 测试set
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 服务器地址
		Password: "",               // Redis 访问密码，如果没有则留空
		DB:       0,                // 选择的数据库
	})
	cacheService := NewRedisInteractiveCache(client)
	biz := "article"
	// 测试没有该元素会自动创建
	cacheService.SetRankingScore(context.Background(), biz, 101, 10)
	cacheService.SetRankingScore(context.Background(), biz, 99, 12)
	cacheService.SetRankingScore(context.Background(), biz, 100, 11)
	cacheService.SetRankingScore(context.Background(), biz, 89, 13)
	//  如果排名数据不存在更新到缓存，如果更新过就+1
	cacheService.SetRankingScore(context.Background(), biz, 89, 13)
	//  测试排名数据+1
	cacheService.IncrRankingIfPresent(context.Background(), biz, 89)
	err := cacheService.IncrRankingIfPresent(context.Background(), biz, 11)
	assert.Equal(t, RankingUpdateErr, err)
	// 获取排名数据
	val, err := cacheService.LikeTop(context.Background(), biz)
	assert.Equal(t, []domain.Interactive{
		{
			Biz:     biz,
			BizId:   89,
			LikeCnt: 15,
		},
		{
			Biz:     biz,
			BizId:   99,
			LikeCnt: 12,
		},
		{
			Biz:     biz,
			BizId:   100,
			LikeCnt: 11,
		},
		{
			Biz:     biz,
			BizId:   101,
			LikeCnt: 10,
		},
	}, val)
	key := fmt.Sprintf("top_100_%s", biz)
	// 删除
	client.Del(context.Background(), key).Result()
}
