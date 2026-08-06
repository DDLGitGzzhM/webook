package service

import (
	"context"

	"webook/webook/search/domain"
	"webook/webook/search/repository"
)

// SyncService 搜索数据同步服务。
type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	InputAny(ctx context.Context, index, docID, data string) error
}

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	anyRepo     repository.AnyRepository
}

// NewSyncService 创建同步服务。
func NewSyncService(
	anyRepo repository.AnyRepository,
	userRepo repository.UserRepository,
	articleRepo repository.ArticleRepository,
) SyncService {
	return &syncService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		anyRepo:     anyRepo,
	}
}

func (s *syncService) InputAny(ctx context.Context, index, docID, data string) error {
	//cvt := s.converter(index)
	//data = cvt.Convert(data)
	return s.anyRepo.Input(ctx, index, docID, data)
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}
