package service

import (
	"context"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

type IArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type ArticleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) IArticleService {
	return &ArticleService{
		repo: repo,
	}
}

func (a ArticleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Create(ctx, art)
}
