package fixer

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"gorm.io/gorm"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
	"webook/webook/pkg/migrator"
	"webook/webook/pkg/migrator/events"
	"webook/webook/pkg/migrator/fixer"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.Logger
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](
	client sarama.Client,
	l logger.Logger,
	topic string,
	src *gorm.DB,
	dst *gorm.DB) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{
		client:   client,
		l:        l,
		srcFirst: srcFirst,
		dstFirst: dstFirst,
		topic:    topic,
	}, nil
}

// Start 这边就是自己启动 goroutine 了
func (r *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{r.topic},
			saramax.NewHandler[events.InconsistentEvent](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

func (r *Consumer[T]) Consume(msg *sarama.ConsumerMessage, t events.InconsistentEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	switch t.Direction {
	case "SRC":
		return r.srcFirst.Fix(ctx, t.ID)
	case "DST":
		return r.dstFirst.Fix(ctx, t.ID)
	}
	return errors.New("未知的校验方向")
}
