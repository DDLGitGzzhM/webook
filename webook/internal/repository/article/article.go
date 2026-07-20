package article

import (
	"context"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
}

type CachedArticleRepository struct {
	dao dao.ArticleDao
}

func NewCachedArticleRepository(dao dao.ArticleDao) *CachedArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
	}
}

func (c CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return c.dao.Insert(ctx, dao.Article{
		AuthorId: art.Author.Id,
		Content:  art.Content,
		Title:    art.Title,
	})
}

func (c CachedArticleRepository) UpdateById(ctx context.Context, art domain.Article) error {
	return c.dao.UpdateById(ctx, dao.Article{
		Id:       art.Id,
		AuthorId: art.Author.Id,
		Content:  art.Content,
		Title:    art.Title,
	})
}
