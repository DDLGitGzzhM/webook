package article

import (
	"context"

	"webook/webook/internal/domain"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	UpdateById(ctx context.Context, art domain.Article) error
}
