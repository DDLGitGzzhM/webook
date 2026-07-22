package article

import (
	"context"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
	// Sync 存储并同步数据到线上库
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	// SyncStatus 同步制作库与线上库的文章状态
	SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao      article.ArticleDao
	userRepo repository.UserRepository

	auth   article.AuthorDao
	reader article.ReaderDao

	db *gorm.DB

	cache cache.ArticleCache
	l     logger.Logger
}

func (c *CachedArticleRepository) GetPublishedById(
	ctx context.Context, id int64,
) (domain.Article, error) {
	art, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	usr, err := c.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		return domain.Article{}, err
	}
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}, nil
}

func (c *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := c.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return c.toDomain(data), nil
}

func (c *CachedArticleRepository) List(
	ctx context.Context, uid int64, offset int, limit int,
) ([]domain.Article, error) {
	if offset == 0 && limit <= 100 {
		data, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				c.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := lo.Map(res, func(src article.Article, _ int) domain.Article {
		return c.toDomain(src)
	})
	go func() {
		err := c.cache.SetFirstPage(ctx, uid, data)
		if err != nil {
			c.l.Error("回写缓存失败", logger.Error(err.Error()))
		}
		c.preCache(ctx, data)
	}()
	return data, nil
}

func (c *CachedArticleRepository) toDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}

func (c *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(art))
	if err == nil {
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
		_ = c.cache.SetPub(ctx, art)
	}
	return id, err
}

// SyncV2 同库不同表,使用事务操作
func (c *CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	/*
			auth : insert,update
			reader: upsert
		    if err ,tx.rollback()
			return tx.commit()
	*/
	return 0, nil
}

func (c *CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = c.auth.Insert(ctx, c.toEntity(art))
	} else {
		err = c.auth.UpdateById(ctx, c.toEntity(art))
	}

	if err != nil {
		return 0, err
	}
	err = c.reader.Upsert(ctx, c.toEntity(art))
	return id, err
}

func (c *CachedArticleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		AuthorId: art.Author.Id,
		Title:    art.Title,
		Content:  art.Content,
		Status:   art.Status.ToUint8(),
	}
}

func (c *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := c.cache.Set(ctx, data[0])
		if err != nil {
			c.l.Error("提前预加载缓存失败", logger.Error(err.Error()))
		}
	}
}

func NewCachedArticleRepository(
	dao article.ArticleDao,
	userRepo repository.UserRepository,
	cache cache.ArticleCache,
	l logger.Logger,
) *CachedArticleRepository {
	return &CachedArticleRepository{
		dao:      dao,
		userRepo: userRepo,
		cache:    cache,
		l:        l,
	}
}

func (c *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.Insert(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	defer func() {
		_ = c.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c *CachedArticleRepository) SyncStatus(
	ctx context.Context, uid, id int64, status domain.ArticleStatus,
) error {
	return c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}
