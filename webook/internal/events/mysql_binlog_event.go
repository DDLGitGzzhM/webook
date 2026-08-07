package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
	"webook/webook/internal/repository/article"
	dao "webook/webook/internal/repository/dao/article"
	"webook/webook/pkg/canalx"
)

type MySQLBinlogConsumer struct {
	client sarama.Client
	l      logger.Logger
	// 耦合到实现，而不是耦合到接口，除非你把操作缓存的方法也定义到 repository 接口上。
	repo *article.CachedArticleRepository
}

func NewMySQLBinlogConsumer(
	client sarama.Client,
	l logger.Logger,
	repo *article.CachedArticleRepository,
) *MySQLBinlogConsumer {
	return &MySQLBinlogConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (r *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("pub_articles_cache",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			// 这里逼不得已和 DAO 耦合在了一起
			saramax.NewHandler[canalx.Message[dao.PublishArticle]](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

func (r *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage,
	cmsg canalx.Message[dao.PublishArticle]) error {
	// 别的表的 binlog，你不关心
	// 可以考虑，不同的表用不同的 topic，那么你这里就不需要判定了
	// GORM 对 PublishArticle 的默认表名是 publish_articles
	if cmsg.Table != "publish_articles" {
		return nil
	}
	// 要在这里更新缓存了
	// 增删改的消息，实际上在 publish article 里面是没有删的消息的
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for _, data := range cmsg.Data {
		var err error
		switch data.Status {
		case domain.ArticleStatusPublished.ToUint8():
			// 发表，要写入缓存
			err = r.repo.Cache().SetPub(ctx, r.repo.ToDomain(dao.Article(data)))
		case domain.ArticleStatusPrivate.ToUint8():
			err = r.repo.Cache().DelPub(ctx, data.Id)
		}
		if err != nil {
			// 正常记录一下日志就行
			r.l.Error("根据 binlog 更新文章缓存失败",
				logger.Error(err.Error()),
				logger.Int64("id", data.Id))
		}
	}
	return nil
}
