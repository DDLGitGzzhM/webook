package article

import "context"

type ReaderDao interface {
	Upsert(ctx context.Context, art Article) error
}

// PublishArticle 不同表
type PublishArticle struct {
	Article
}
