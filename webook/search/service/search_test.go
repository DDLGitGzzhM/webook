package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/search/domain"
)

type stubUserRepo struct {
	users []domain.User
	err   error
}

func (s *stubUserRepo) InputUser(ctx context.Context, msg domain.User) error {
	return nil
}

func (s *stubUserRepo) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	return s.users, s.err
}

type stubArticleRepo struct {
	arts []domain.Article
	err  error
}

func (s *stubArticleRepo) InputArticle(ctx context.Context, msg domain.Article) error {
	return nil
}

func (s *stubArticleRepo) SearchArticle(
	ctx context.Context,
	uid int64,
	keywords []string,
) ([]domain.Article, error) {
	return s.arts, s.err
}

func TestSearchService_Search(t *testing.T) {
	userRepo := &stubUserRepo{
		users: []domain.User{{Id: 1, Nickname: "Tom"}},
	}
	artRepo := &stubArticleRepo{
		arts: []domain.Article{{Id: 2, Title: "Tom 的小秘密", Status: 2}},
	}
	svc := NewSearchService(userRepo, artRepo)

	res, err := svc.Search(context.Background(), 1001, "Tom 内容")
	require.NoError(t, err)
	assert.Equal(t, userRepo.users, res.Users)
	assert.Equal(t, artRepo.arts, res.Articles)
}

type stubAnyRepo struct {
	called bool
}

func (s *stubAnyRepo) Input(ctx context.Context, index, docID, data string) error {
	s.called = true
	return nil
}

func TestSyncService_Input(t *testing.T) {
	userRepo := &stubUserRepo{}
	artRepo := &stubArticleRepo{}
	anyRepo := &stubAnyRepo{}
	svc := NewSyncService(anyRepo, userRepo, artRepo)

	require.NoError(t, svc.InputUser(context.Background(), domain.User{Id: 1}))
	require.NoError(t, svc.InputArticle(context.Background(), domain.Article{Id: 2}))
	require.NoError(t, svc.InputAny(context.Background(), "tags_index", "1", "{}"))
	assert.True(t, anyRepo.called)
}
