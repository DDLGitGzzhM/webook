package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
	"webook/webook/search/service"
)

const topicSyncData = "search_sync_data"

// SyncDataEvent 通用搜索同步事件。
type SyncDataEvent struct {
	IndexName string
	DocID     string
	Data      string
	// 假如说用于同步 user
	// IndexName = user_index
	// DocID = "123"
	// Data = {"id": 123, "email":xx, nickname: ""}
}

// SyncDataEventConsumer 通用数据同步消费者。
type SyncDataEventConsumer struct {
	svc    service.SyncService
	client sarama.Client
	l      logger.Logger
}

// NewSyncDataEventConsumer 创建通用数据同步消费者。
func NewSyncDataEventConsumer(
	client sarama.Client,
	l logger.Logger,
	svc service.SyncService,
) *SyncDataEventConsumer {
	return &SyncDataEventConsumer{
		svc:    svc,
		client: client,
		l:      l,
	}
}

// Start 启动消费。
func (a *SyncDataEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("search_sync_data", a.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncData},
			saramax.NewHandler[SyncDataEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

// Consume 将任意索引文档写入 ES。
func (a *SyncDataEventConsumer) Consume(
	sg *sarama.ConsumerMessage,
	evt SyncDataEvent,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 在这里执行转化
	return a.svc.InputAny(ctx, evt.IndexName, evt.DocID, evt.Data)
}
