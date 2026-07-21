package article

import (
	"context"

	"gorm.io/gorm"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	// SyncStatus 同步制作库与线上库的文章状态
	SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
}

type CachedArticleRepository struct {
	dao article.ArticleDao

	auth   article.AuthorDao
	reader article.ReaderDao

	db *gorm.DB
}

func (c CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Sync(ctx, c.toEntity(art))
}

// SyncV2 同库不同表,使用事务操作
func (c CachedArticleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	//author := NewGorAuthor(ctx)
	//Reader := NewGorReader(ctx)

	/*
			auth : insert,update
			reader: upsert
		    if err ,tx.rollback()
			return tx.commit()
	*/
	return 0, nil

}
func (c CachedArticleRepository) SyncV1(ctx context.Context, art domain.Article) (int64, error) {
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

func (c CachedArticleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		Id:       art.Id,
		AuthorId: art.Author.Id,
		Title:    art.Title,
		Content:  art.Content,
		Status:   art.Status.ToUint8(),
	}
}

func NewCachedArticleRepository(dao article.ArticleDao) *CachedArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, c.toEntity(art))
}

func (c CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, c.toEntity(art))
}

func (c CachedArticleRepository) SyncStatus(
	ctx context.Context, uid, id int64, status domain.ArticleStatus,
) error {
	return c.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}
