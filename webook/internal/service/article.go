package service

import (
	"context"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/article"
)

type IArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
}

type ArticleService struct {
	repo article.ArticleRepository

	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository

	l logger.Logger
}

func NewArticleService(repo article.ArticleRepository) IArticleService {
	return &ArticleService{
		repo: repo,
	}
}

func NewArticleServiceV1(auth article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger.Logger) IArticleService {
	return &ArticleService{
		author: auth,
		reader: reader,
		l:      l,
	}
}

func (a ArticleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	id, err := a.repo.Create(ctx, art)
	if err != nil {
		return 0, err
	}
	return id, err
}

func (a ArticleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	if art.Id > 0 {
		err = a.author.UpdateById(ctx, art)
	} else {
		id, err = a.author.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		/*
				service 并不能很好的管理事务
				1. 同库不同表
				2. 不同库
				因此只能尝试重试

			   1. 接入告警系统手动处理
			   2. 走异步
			   3. canal 监听 binlog
			   4. 打 MQ
		*/
		id, err = a.reader.Save(ctx, art)
		if err == nil {
			return id, nil
		}
		a.l.Error("保存文章失败", logger.Error(err.Error()))
	}
	return 0, err
}

func (a ArticleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.UpdateById(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}
