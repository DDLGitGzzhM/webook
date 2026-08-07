package events

import (
	"context"
	"strconv"
	"time"

	"github.com/IBM/sarama"

	"webook/webook/feed/domain"
	"webook/webook/feed/service"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
)

const topicArticleEvent = "article_publish_event"

// ArticlePublishedEvent 由业务方定义，本服务做适配
// 监听业务方的事件
// 业务方强势，我们自己适配
type ArticlePublishedEvent struct {
	Uid int64
	Aid int64
}

type ArticleEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	svc    service.FeedService
}

func NewArticleEventConsumer(
	client sarama.Client,
	l logger.Logger,
	svc service.FeedService) *ArticleEventConsumer {
	ac := &ArticleEventConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
	return ac
}

// Start 这边就是自己启动 goroutine 了
func (r *ArticleEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("articleFeed",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicArticleEvent},
			saramax.NewHandler[ArticlePublishedEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}
func (r *ArticleEventConsumer) Consume(msg *sarama.ConsumerMessage,
	evt ArticlePublishedEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.svc.CreateFeedEvent(ctx, domain.FeedEvent{
		Type: service.ArticleEventName,
		Ext: map[string]string{
			"uid": strconv.FormatInt(evt.Uid, 10),
			"aid": strconv.FormatInt(evt.Aid, 10),
		},
	})
}

// json 序列化与反序列化的问题
// uid: 123
// 反序列化 map[string]any => uid: 123, 但是 123 是 float64
