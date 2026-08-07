package service

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"

	"webook/webook/feed/domain"
	"webook/webook/feed/repository"
)

type feedService struct {
	repo       repository.FeedEventRepo
	handlerMap map[string]Handler
	//followClient followv1.FollowServiceClient
}

func NewFeedService(repo repository.FeedEventRepo,
	//client followv1.FollowServiceClient,
	handlerMap map[string]Handler) FeedService {
	return &feedService{
		repo: repo,
		// 你可以注入那个 health client
		//followClient: client,
		handlerMap: handlerMap,
	}
}

func (f *feedService) registerService(typ string, handler Handler) {
	f.handlerMap[typ] = handler
}

func (f *feedService) CreateFeedEvent(ctx context.Context, feed domain.FeedEvent) error {
	handler, ok := f.handlerMap[feed.Type]
	if !ok {
		// 这里你可以考虑引入一个兜底的处理机制。
		// 例如说在找不到的时候就默认丢过去 PushEvent 里面
		// 对于大部分业务来说，都是合适的
		return fmt.Errorf("未找到具体的业务处理逻辑 %s", feed.Type)
	}
	return handler.CreateFeedEvent(ctx, feed.Ext)
}

// GetFeedEventListV1 不依赖于 Handler 的直接查询
// service 层面上的统一实现
// 基本思路就是，收件箱查一下，发件箱查一下，合并结果（排序，分页），返回结果。
// 按照时间戳倒序排序
// 查询的时候，业务上不要做特殊处理
//func (f *feedService) GetFeedEventListV1(ctx context.Context,
//	uid int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
//	var (
//		eg         errgroup.Group
//		pushEvents []domain.FeedEvent
//		pullEvents []domain.FeedEvent
//	)
//	eg.Go(func() error {
//
//		// 性能瓶颈大概率出现在这里
//		// 你可以考虑说，在触发了降级的时候，或者 follow 本身触发了降级的时候
//		// 不走这个分支
//		// 我怎么知道 follow 降级了呢？
//
//		// 在这边，pull event 你要获得你关注的所有人的 id
//		resp, err := f.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
//			// 你的 ID，为了获得你关注的所有人
//			Follower: uid,
//			// 可以把全部取过来
//			Limit: 100000,
//			// 你把时间戳过去，只查询[时间戳 - 1 天，时间戳]
//		})
//		if err != nil {
//			return err
//		}
//		uids := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
//			return src.Followee
//		})
//		pullEvents, err = f.repo.FindPullEvents(ctx, uids, timestamp, limit)
//		return err
//	})
//	eg.Go(func() error {
//		var err error
//		// 只有一次本地数据库查询，非常快
//		pushEvents, err = f.repo.FindPushEvents(ctx, uid, timestamp, limit)
//		return err
//	})
//	err := eg.Wait()
//	if err != nil {
//		return nil, err
//	}
//	events := append(pushEvents, pullEvents...)
//	// 这边你要再次排序
//	sort.Slice(events, func(i, j int) bool {
//		return events[i].Ctime.After(events[j].Ctime)
//	})
//	// 要小心不够数量。就是你想取10 条。结果总共才查到了 8 条
//	// min 这个方法在高版本 GO 里面才有
//	// slice.Min
//	return events[:min[int](len(events), int(limit))], nil
//}

func (f *feedService) GetFeedEventList(ctx context.Context, uid int64, timestamp, limit int64) ([]domain.FeedEvent, error) {
	// 万一，我有一部分业务有自己的查询逻辑；我另外一些业务没有特殊的查询逻辑
	// 怎么写代码？
	// 要注意尽可能减少数据库查询次数，和 follow client 的调用次数
	var eg errgroup.Group
	res := make([]domain.FeedEvent, 0, limit*int64(len(f.handlerMap)))
	var mu sync.RWMutex
	for _, handler := range f.handlerMap {
		// 每一个业务方你都查一遍
		// 要小心这一步，不要忘了
		h := handler
		eg.Go(func() error {
			events, err := h.FindFeedEvents(ctx, uid, timestamp, limit)
			if err != nil {
				return err
			}
			mu.Lock()
			res = append(res, events...)
			mu.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// 聚合排序，难免的
	sort.Slice(res, func(i, j int) bool {
		return res[i].Ctime.Unix() > res[j].Ctime.Unix()
	})
	return res[:slice.Min[int]([]int{int(limit), len(res)})], nil
}
