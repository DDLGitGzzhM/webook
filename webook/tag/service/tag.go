package service

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"webook/webook/internal/pkg/logger"
	"webook/webook/tag/domain"
	"webook/webook/tag/events"
	"webook/webook/tag/repository"
)

// TagService 标签服务。
type TagService interface {
	CreateTag(ctx context.Context, uid int64, name string) (int64, error)
	AttachTags(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
}

type tagService struct {
	repo     repository.TagRepository
	logger   logger.Logger
	producer events.Producer
}

// NewTagService 创建标签服务。
func NewTagService(
	repo repository.TagRepository,
	producer events.Producer,
	l logger.Logger,
) TagService {
	return &tagService{
		producer: producer,
		repo:     repo,
		logger:   l,
	}
}

// AttachTags 覆盖式给业务资源打标签，并异步同步到搜索。
func (svc *tagService) AttachTags(
	ctx context.Context,
	uid int64,
	biz string,
	bizId int64,
	tags []int64,
) error {
	err := svc.repo.BindTagToBiz(ctx, uid, biz, bizId, tags)
	if err != nil {
		return err
	}
	// 异步发送
	go func() {
		ts, err := svc.repo.GetTagsById(ctx, tags)
		if err != nil {
			svc.logger.Error("查询标签失败", logger.Error(err.Error()))
			return
		}
		// 这里要根据 tag_index 的结构来定义
		// 同样要注意顺序，即同一个用户对同一个资源打标签的顺序，
		// 是不能乱的
		pctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = svc.producer.ProduceSyncEvent(pctx, events.BizTags{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Tags: slice.Map(ts, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		if err != nil {
			svc.logger.Error("发送标签搜索事件失败", logger.Error(err.Error()))
		}
	}()
	return nil
}

// GetBizTags 获取业务资源上的标签。
func (svc *tagService) GetBizTags(
	ctx context.Context,
	uid int64,
	biz string,
	bizId int64,
) ([]domain.Tag, error) {
	return svc.repo.GetBizTags(ctx, uid, biz, bizId)
}

// CreateTag 创建标签。
func (svc *tagService) CreateTag(ctx context.Context, uid int64, name string) (int64, error) {
	return svc.repo.CreateTag(ctx, domain.Tag{
		Uid:  uid,
		Name: name,
	})
}

// GetTags 获取用户标签。
func (svc *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return svc.repo.GetTags(ctx, uid)
}
