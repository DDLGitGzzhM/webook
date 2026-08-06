package service

import (
	"context"
	"strings"

	"golang.org/x/sync/errgroup"

	"webook/webook/search/domain"
	"webook/webook/search/repository"
)

// SearchService 搜索服务。
type SearchService interface {
	Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error)
}

type searchService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
}

// NewSearchService 创建搜索服务。
func NewSearchService(
	userRepo repository.UserRepository,
	articleRepo repository.ArticleRepository,
) SearchService {
	return &searchService{userRepo: userRepo, articleRepo: articleRepo}
}

func (s *searchService) Search(
	ctx context.Context,
	uid int64,
	expression string,
) (domain.SearchResult, error) {
	// 这边一般要对 expression 进行一些预处理
	keywords := strings.Split(expression, " ")
	var eg errgroup.Group
	var res domain.SearchResult
	eg.Go(func() error {
		users, err := s.userRepo.SearchUser(ctx, keywords)
		res.Users = users
		return err
	})
	eg.Go(func() error {
		arts, err := s.articleRepo.SearchArticle(ctx, uid, keywords)
		res.Articles = arts
		return err
	})
	return res, eg.Wait()
}
