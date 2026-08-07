package service

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"

	followv1 "webook/webook/api/proto/gen/follow/v1"
	"webook/webook/feed/domain"
	"webook/webook/feed/repository"
)

type ArticleEventHandler struct {
	repo         repository.FeedEventRepo
	followClient followv1.FollowServiceClient
}

const (
	ArticleEventName = "article_event"
	threshold        = 4
	//threshold        = 32
)

func NewArticleEventHandler(repo repository.FeedEventRepo, client followv1.FollowServiceClient) Handler {
	return &ArticleEventHandler{
		repo:         repo,
		followClient: client,
	}
}

func (a *ArticleEventHandler) FindFeedEvents(ctx context.Context, uid, timestamp, limit int64) ([]domain.FeedEvent, error) {
	var (
		eg errgroup.Group
		mu sync.Mutex
	)
	events := make([]domain.FeedEvent, 0, limit*2)
	// Push Event
	eg.Go(func() error {
		pushEvents, err := a.repo.FindPushEventsWithTyp(ctx, ArticleEventName, uid, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pushEvents...)
		mu.Unlock()
		return nil
	})

	// Pull Event
	eg.Go(func() error {
		resp, rerr := a.followClient.GetFollowee(ctx, &followv1.GetFolloweeRequest{
			Follower: uid,
			Offset:   0,
			Limit:    200,
		})
		if rerr != nil {
			return rerr
		}
		followeeIds := slice.Map(resp.FollowRelations, func(idx int, src *followv1.FollowRelation) int64 {
			return src.Followee
		})
		pullEvents, err := a.repo.FindPullEventsWithTyp(ctx, ArticleEventName, followeeIds, timestamp, limit)
		if err != nil {
			return err
		}
		mu.Lock()
		events = append(events, pullEvents...)
		mu.Unlock()
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Ctime.Unix() > events[j].Ctime.Unix()
	})

	return events[:slice.Min[int]([]int{int(limit), len(events)})], nil
}

func (a *ArticleEventHandler) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	uid, err := ext.Get("uid").AsInt64()
	if err != nil {
		return err
	}
	// 根据粉丝数判断使用推模型还是拉模型
	resp, err := a.followClient.GetFollowStatic(ctx, &followv1.GetFollowStaticRequest{
		Followee: uid,
	})
	if err != nil {
		return err
	}
	// 粉丝数超出阈值使用拉模型
	if resp.FollowStatic.Followers > threshold {
		return a.repo.CreatePullEvent(ctx, domain.FeedEvent{
			Uid:   uid,
			Type:  ArticleEventName,
			Ctime: time.Now(),
			Ext:   ext,
		})
	}
	// 使用推模型
	fresp, err := a.followClient.GetFollower(ctx, &followv1.GetFollowerRequest{
		Followee: uid,
	})
	if err != nil {
		return err
	}
	events := make([]domain.FeedEvent, 0, len(fresp.FollowRelations))
	for _, r := range fresp.GetFollowRelations() {
		events = append(events, domain.FeedEvent{
			Uid:   r.Follower,
			Type:  ArticleEventName,
			Ctime: time.Now(),
			Ext:   ext,
		})
	}
	return a.repo.CreatePushEvents(ctx, events)
}
