package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"webook/webook/internal/domain"
)

var (
	//go:embed lua/interative_incr_cnt.lua
	luaIncrCnt string
	//go:embed lua/interactive_ranking_incr.lua
	luaRankingIncr string
	//go:embed lua/interactive_ranking_set.lua
	luaRankingSet string
)

// RankingUpdateErr 指定的排行榜元素不存在
var RankingUpdateErr = errors.New("指定的元素不存在")

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

//go:generate mockgen -source=./interactive.go -package=cachemocks -destination=mocks/interactive.mock.go InteractiveCache
type InteractiveCache interface {
	// IncrReadCntIfPresent 如果在缓存中有对应的数据，就 +1
	IncrReadCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizId int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	// Get 查询缓存中数据
	// 事实上，这里 liked 和 collected 是不需要缓存的
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizId int64, intr domain.Interactive) error
	// IncrRankingIfPresent 如果排名数据存在就+1
	IncrRankingIfPresent(ctx context.Context, biz string, bizId int64) error
	// SetRankingScore 如果排名数据不存在就把数据库中读取到的更新到缓存，如果更新过就+1
	SetRankingScore(ctx context.Context, biz string, bizId int64, count int64) error
	// LikeTop 基本实现，是借助 zset
	// 1. 前 100 名是一个高频数据，你可以结合本地缓存。
	//    你可以定时刷新本地缓存，比如说每 5s 调用 LikeTop，放进去本地缓存
	// 2. 如果你有一亿的数据，你怎么实时维护？zset 放 一亿个元素，你的 Redis 撑不住
	// 		2.1 不是真的维护一亿，而是维护近期的数据的点赞数，比如说三天内的
	//      2.2 你要分 key。这是 Redis 解决大数据结构常见的方案
	// 3. 借助定时任务，我每分钟计算一次。如果你有很多数据，一分钟不够你遍历一遍
	// 4. 我定时计算，算 1000 名；而后我借助 zset 来实时维护者 1000 名的分数
	LikeTop(ctx context.Context, biz string) ([]domain.Interactive, error)
}

type RedisInteractiveCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (r *RedisInteractiveCache) IncrCollectCntIfPresent(
	ctx context.Context, biz string, bizId int64,
) error {
	return r.client.Eval(ctx, luaIncrCnt, []string{r.key(biz, bizId)},
		fieldCollectCnt, 1).Err()
}

func (r *RedisInteractiveCache) IncrReadCntIfPresent(
	ctx context.Context, biz string, bizId int64,
) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldReadCnt, 1).Err()
}

func (r *RedisInteractiveCache) IncrLikeCntIfPresent(
	ctx context.Context, biz string, bizId int64,
) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, 1).Err()
}

func (r *RedisInteractiveCache) DecrLikeCntIfPresent(
	ctx context.Context, biz string, bizId int64,
) error {
	return r.client.Eval(ctx, luaIncrCnt,
		[]string{r.key(biz, bizId)},
		fieldLikeCnt, -1).Err()
}

// IncrRankingIfPresentV1 分 key 的写入
func (r *RedisInteractiveCache) IncrRankingIfPresentV1(
	ctx context.Context, biz string, bizId int64,
) error {
	h := fnv.New32a()
	_, _ = h.Write([]byte(biz))
	key := fmt.Sprintf("top_100_%d_%s", h.Sum32()%100, biz)
	res, err := r.client.Eval(ctx, luaRankingIncr, []string{key}, bizId).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return RankingUpdateErr
	}
	return nil
}

func (r *RedisInteractiveCache) IncrRankingIfPresent(
	ctx context.Context, biz string, bizId int64,
) error {
	res, err := r.client.Eval(ctx, luaRankingIncr,
		[]string{r.rankingKey(biz)}, bizId).Result()
	if err != nil {
		return err
	}
	if res.(int64) == 0 {
		return RankingUpdateErr
	}
	return nil
}

func (r *RedisInteractiveCache) SetRankingScore(
	ctx context.Context, biz string, bizId int64, count int64,
) error {
	return r.client.Eval(ctx, luaRankingSet,
		[]string{r.rankingKey(biz)}, bizId, count).Err()
}

// BatchSetRankingScore 设置整个数据
func (r *RedisInteractiveCache) BatchSetRankingScore(
	ctx context.Context, biz string, interactives []domain.Interactive,
) error {
	members := make([]redis.Z, 0, len(interactives))
	for _, interactive := range interactives {
		members = append(members, redis.Z{
			Score:  float64(interactive.LikeCnt),
			Member: interactive.BizId,
		})
	}
	return r.client.ZAdd(ctx, r.rankingKey(biz), members...).Err()
}

func (r *RedisInteractiveCache) Get(
	ctx context.Context, biz string, bizId int64,
) (domain.Interactive, error) {
	// 直接使用 HMGet，即便缓存中没有对应的 key，也不会返回 error
	data, err := r.client.HGetAll(ctx, r.key(biz, bizId)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		// 缓存不存在
		return domain.Interactive{}, ErrKeyNotExist
	}

	// 理论上来说，这里不可能有 error
	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	likeCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)

	return domain.Interactive{
		BizId:      bizId,
		CollectCnt: collectCnt,
		LikeCnt:    likeCnt,
		ReadCnt:    readCnt,
	}, err
}

func (r *RedisInteractiveCache) Set(
	ctx context.Context, biz string, bizId int64, intr domain.Interactive,
) error {
	key := r.key(biz, bizId)
	err := r.client.HMSet(ctx, key,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
		fieldReadCnt, intr.ReadCnt).Err()
	if err != nil {
		return err
	}
	return r.client.Expire(ctx, key, time.Minute*15).Err()
}

// LikeTopV1 分 key 版本
func (r *RedisInteractiveCache) LikeTopV1(
	ctx context.Context, biz string,
) ([]domain.Interactive, error) {
	// 我从 100 个 key 里面，各取前 100
	// 然后，合并再取前 100
	interactives := make([]domain.Interactive, 0, 100*100)
	for i := 0; i < 100; i++ {
		var start int64 = 0
		var end int64 = 99
		key := fmt.Sprintf("top_100_%d_%s", i, biz)
		res, err := r.client.ZRevRangeWithScores(ctx, key, start, end).Result()
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(res); j++ {
			id, _ := strconv.ParseInt(res[j].Member.(string), 10, 64)
			interactives = append(interactives, domain.Interactive{
				Biz:     biz,
				BizId:   id,
				LikeCnt: int64(res[j].Score),
			})
		}
	}

	// 进一步排序，然后取前 100
	return interactives, nil
}

func (r *RedisInteractiveCache) LikeTop(
	ctx context.Context, biz string,
) ([]domain.Interactive, error) {
	var start int64 = 0
	var end int64 = 99
	key := r.rankingKey(biz)
	res, err := r.client.ZRevRangeWithScores(ctx, key, start, end).Result()
	if err != nil {
		return nil, err
	}
	interactives := make([]domain.Interactive, 0, 100)
	for i := 0; i < len(res); i++ {
		id, _ := strconv.ParseInt(res[i].Member.(string), 10, 64)
		interactives = append(interactives, domain.Interactive{
			Biz:     biz,
			BizId:   id,
			LikeCnt: int64(res[i].Score),
		})
	}
	return interactives, nil
}

func (r *RedisInteractiveCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizId)
}

func (r *RedisInteractiveCache) rankingKey(biz string) string {
	return fmt.Sprintf("top_100_%s", biz)
}

func NewRedisInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &RedisInteractiveCache{
		client: client,
	}
}
