package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/article"
	repoMock "webook/webook/internal/repository/mock"
)

func Test_ArticleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository)
		art     domain.Article
		wantErr error
		wantId  int64
	}{
		{
			name: "新建发表成功",
			mock: func(ctrl *gomock.Controller) (article.ArticleAuthorRepository, article.ArticleReaderRepository) {
				author := repoMock.NewMockArticleRepository(ctrl)
				author.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id:   123,
						Name: "system",
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)
				reader := repoMock.NewMockArticleReaderRepository(ctrl)
				reader.EXPECT().Save(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id:   123,
						Name: "system",
					},
					Status: domain.ArticleStatusPublished,
				}).Return(int64(1), nil)
				return author, reader
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id:   123,
					Name: "system",
				},
			},
			wantId: 1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			author, reader := tt.mock(ctrl)
			svc := NewArticleServiceV1(author, reader, logger.NopLogger{})
			id, err := svc.PublishV1(context.Background(), tt.art)
			require.Equal(t, tt.wantErr, err)
			require.Equal(t, tt.wantId, id)
		})
	}
}
