package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"

	"webook/webook/internal/pkg/logger"
	"webook/webook/tag/domain"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
)

// TagRepository 标签仓储。
type TagRepository interface {
	CreateTag(ctx context.Context, tag domain.Tag) (int64, error)
	BindTagToBiz(ctx context.Context, uid int64, biz string, bizId int64, tags []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error)
	PreloadUserTags(ctx context.Context) error
}

// CachedTagRepository 带缓存的标签仓储。
type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.Logger
}

// NewTagRepository 创建标签仓储。
func NewTagRepository(tagDAO dao.TagDAO, c cache.TagCache, l logger.Logger) *CachedTagRepository {
	return &CachedTagRepository{
		dao:   tagDAO,
		l:     l,
		cache: c,
	}
}

// PreloadUserTags 在 toB 的场景下，你可以提前预加载缓存。
func (repo *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	offset := 0
	const batch = 100
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		// 在这里还有一点点的优化手段，就是 GetTags 的时候，order by uid
		tags, err := repo.dao.GetTags(dbCtx, offset, batch)
		cancel()
		if err != nil {
			return err
		}
		for _, tag := range tags {
			rctx, ccancel := context.WithTimeout(ctx, time.Second)
			err = repo.cache.Append(rctx, tag.Uid, repo.toDomain(tag))
			ccancel()
			if err != nil {
				continue
			}
		}
		if len(tags) < batch {
			return nil
		}
		offset += batch
	}
}

// GetTagsById 按 ID 查询标签。
func (repo *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

// BindTagToBiz 覆盖式绑定标签到业务资源。
func (repo *CachedTagRepository) BindTagToBiz(
	ctx context.Context,
	uid int64,
	biz string,
	bizId int64,
	tags []int64,
) error {
	return repo.dao.CreateTagBiz(ctx, slice.Map(tags, func(idx int, src int64) dao.TagBiz {
		return dao.TagBiz{
			Tid:   src,
			BizId: bizId,
			Biz:   biz,
			Uid:   uid,
		}
	}))
}

// GetTags 获取用户标签。
func (repo *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	res, err := repo.cache.GetTags(ctx, uid)
	if err == nil {
		return res, nil
	}
	tags, err := repo.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	res = slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	})
	_ = repo.cache.Append(ctx, uid, res...)
	return res, nil
}

// GetBizTags 获取业务资源上的标签。
func (repo *CachedTagRepository) GetBizTags(
	ctx context.Context,
	uid int64,
	biz string,
	bizId int64,
) ([]domain.Tag, error) {
	tags, err := repo.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return slice.Map(tags, func(idx int, src dao.Tag) domain.Tag {
		return repo.toDomain(src)
	}), nil
}

// CreateTag 创建标签。
func (repo *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	id, err := repo.dao.CreateTag(ctx, repo.toEntity(tag))
	if err != nil {
		return 0, err
	}
	tag.Id = id
	_ = repo.cache.Append(ctx, tag.Uid, tag)
	return id, nil
}

func (repo *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (repo *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
