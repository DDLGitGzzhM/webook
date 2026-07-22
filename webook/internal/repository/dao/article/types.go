package article

import (
	"context"
	"errors"
)

var (
	ArtNotFound                = errors.New("文章不存在")
	ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishArticle, error)
	Sync(ctx context.Context, art Article) (int64, error)
	// SyncStatus 同库不同表，事务同步更新制作库与线上库状态
	SyncStatus(ctx context.Context, uid, id int64, status uint8) error
}
