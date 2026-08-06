package repository

import (
	"context"

	"webook/webook/search/domain"
)

// UserRepository 用户搜索仓储。
type UserRepository interface {
	InputUser(ctx context.Context, msg domain.User) error
	SearchUser(ctx context.Context, keywords []string) ([]domain.User, error)
}

// ArticleRepository 文章搜索仓储。
type ArticleRepository interface {
	InputArticle(ctx context.Context, msg domain.Article) error
	SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error)
}
