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

const topicSyncUser = "sync_user_event"

// UserConsumer 用户同步消费者。
type UserConsumer struct {
	syncSvc service.SyncService
	client  sarama.Client
	l       logger.Logger
}

// UserEvent 用户同步事件。
type UserEvent struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

// NewUserConsumer 创建用户同步消费者。
func NewUserConsumer(
	client sarama.Client,
	l logger.Logger,
	svc service.SyncService,
) *UserConsumer {
	return &UserConsumer{
		syncSvc: svc,
		client:  client,
		l:       l,
	}
}

// Start 启动消费。
func (u *UserConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("sync_user", u.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicSyncUser},
			saramax.NewHandler[UserEvent](u.l, u.Consume))
		if err != nil {
			u.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

// Consume 处理单条用户同步消息。
func (u *UserConsumer) Consume(sg *sarama.ConsumerMessage, evt UserEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return u.syncSvc.InputUser(ctx, u.toDomain(evt))
}

func (u *UserConsumer) toDomain(evt UserEvent) domain.User {
	return domain.User{
		Id:       evt.Id,
		Email:    evt.Email,
		Nickname: evt.Nickname,
		Phone:    evt.Phone,
	}
}
