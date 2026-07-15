package repository

import (
	"context"

	"webook/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany = cache.ErrCodeSendTooMany
	ErrCodeSystem      = cache.ErrCodeSystem
	ErrorVerifyTooMany = cache.ErrorVerifyTooMany
)

type CodeRepository struct {
	codeCache *cache.CodeCache
}

func NewCodeRepository(codeCache *cache.CodeCache) *CodeRepository {
	return &CodeRepository{
		codeCache: codeCache,
	}
}

func (r *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return r.codeCache.Set(ctx, biz, phone, code)
}

func (r *CodeRepository) Verify(ctx context.Context, biz, phone, expectCode string) (bool, error) {
	return r.codeCache.Verify(ctx, biz, phone, expectCode)
}
