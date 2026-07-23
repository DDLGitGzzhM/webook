package service

import (
	"context"

	"webook/webook/internal/domain"
	events "webook/webook/internal/events/article"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/article"
)

type IArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type ArticleService struct {
	repo article.ArticleRepository

	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository

	l        logger.Logger
	producer events.Producer
}

func NewArticleService(
	repo article.ArticleRepository,
	l logger.Logger,
	producer events.Producer,
) IArticleService {
	return &ArticleService{
		repo:     repo,
		producer: producer,
		l:        l,
	}
}

func NewArticleServiceV1(
	auth article.ArticleAuthorRepository,
	reader article.ArticleReaderRepository,
	l logger.Logger,
) IArticleService {
	return &ArticleService{
		author: auth,
		reader: reader,
		l:      l,
	}
}

func (a ArticleService) GetPublishedById(
	ctx context.Context, id, uid int64,
) (domain.Article, error) {
	art, err := a.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			er := a.producer.ProduceReadEvent(
				// 脱离主链路超时控制，使用独立 context
				context.Background(),
				events.ReadEvent{
					// 即便你的消费者要用 art 里面的数据，
					// 让它去查询，你不要在 event 里面带
					Uid: uid,
					Aid: id,
				})
			if er != nil {
				a.l.Error("发送读者阅读事件失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(er.Error()))
			}
		}()
	}
	return art, err
}

func (a ArticleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetByID(ctx, id)
}

func (a ArticleService) List(
	ctx context.Context, uid int64, offset int, limit int,
) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a ArticleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func (a ArticleService) Withdraw(ctx context.Context, uid, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a ArticleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	art.Status = domain.ArticleStatusPublished
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
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.UpdateById(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}
