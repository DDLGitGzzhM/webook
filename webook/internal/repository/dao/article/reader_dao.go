package article

import "context"

type ReaderDao interface {
	Upsert(ctx context.Context, art Article) error
}
