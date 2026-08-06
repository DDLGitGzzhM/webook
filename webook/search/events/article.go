package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
	"webook/webook/search/domain"
	"webook/webook/search/service"
)

const topicSyncArticle = "sync_article_event"

// ArticleConsumer 文章同步消费者。
type ArticleConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.Logger
}

// NewArticleConsumer 创建文章同步消费者。
func NewArticleConsumer(
	client sarama.Client,
	l logger.Logger,
	svc service.SyncService,
) *ArticleConsumer {
	return &ArticleConsumer{
		syncSvc: svc,
		client:  client,
		l:       l,
	}
}

// ArticleEvent 文章同步事件。
type ArticleEvent struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Status  int32  `json:"status"`
	Content string `json:"content"`
}

// Start 启动消费。
func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_article", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncArticle},
			saramax.NewHandler[ArticleEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

// Consume 处理单条文章同步消息。
func (a *ArticleConsumer) Consume(sg *sarama.ConsumerMessage, evt ArticleEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return a.syncSvc.InputArticle(ctx, a.toDomain(evt))
}

func (a *ArticleConsumer) toDomain(article ArticleEvent) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Status:  article.Status,
		Content: article.Content,
	}
}
