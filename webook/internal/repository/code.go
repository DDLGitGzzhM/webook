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

type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CacheCodeRepository struct {
	codeCache cache.CodeCache
}

func NewCodeRepository(codeCache cache.CodeCache) *CacheCodeRepository {
	return &CacheCodeRepository{
		codeCache: codeCache,
	}
}

func (r *CacheCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return r.codeCache.Set(ctx, biz, phone, code)
}

func (r *CacheCodeRepository) Verify(ctx context.Context, biz, phone, expectCode string) (bool, error) {
	return r.codeCache.Verify(ctx, biz, phone, expectCode)
}
