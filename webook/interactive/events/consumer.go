package events

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"gorm.io/gorm"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/saramax"
	"webook/webook/pkg/canalx"
	"webook/webook/pkg/migrator"
	mevents "webook/webook/pkg/migrator/events"
	"webook/webook/pkg/migrator/validator"
)

type MySQLBinlogConsumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.Logger
	table    string
	srcToDst *validator.CanalIncrValidator[T]
	dstToSrc *validator.CanalIncrValidator[T]
	dstFirst *atomic.Bool
}

func NewMySQLBinlogConsumer[T migrator.Entity](
	client sarama.Client,
	l logger.Logger,
	table string,
	src *gorm.DB,
	dst *gorm.DB,
	p mevents.Producer,
) *MySQLBinlogConsumer[T] {
	srcToDst := validator.NewCanalIncrValidator[T](src, dst, "SRC", l, p)
	dstToSrc := validator.NewCanalIncrValidator[T](src, dst, "DST", l, p)
	return &MySQLBinlogConsumer[T]{
		client:   client,
		l:        l,
		dstFirst: &atomic.Bool{},
		srcToDst: srcToDst,
		dstToSrc: dstToSrc,
		table:    table,
	}
}

func (r *MySQLBinlogConsumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator_incr",
		r.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{"webook_binlog"},
			saramax.NewHandler[canalx.Message[T]](r.l, r.Consume))
		if err != nil {
			r.l.Error("退出了消费循环异常", logger.Error(err.Error()))
		}
	}()
	return err
}

func (r *MySQLBinlogConsumer[T]) Consume(msg *sarama.ConsumerMessage,
	val canalx.Message[T]) error {
	// 是不是源表为准
	dstFirst := r.dstFirst.Load()
	var v *validator.CanalIncrValidator[T]
	// db:
	//  src:
	//    dsn: "root:root@tcp(localhost:13316)/webook"
	//  dst:
	//    dsn: "root:root@tcp(localhost:13316)/webook_intr"
	if dstFirst && val.Database == "webook_intr" {
		// 目标表为准
		// 校验，用 dst 的来校验
		v = r.dstToSrc
	} else if !dstFirst && val.Database == "webook" {
		// 源表为准，并且消息恰好来自源表
		// 校验，用 src 来校验
		v = r.srcToDst
	}
	if v != nil {
		for _, data := range val.Data {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := v.Validate(ctx, data.ID())
			cancel()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *MySQLBinlogConsumer[T]) DstFirst() {
	r.dstFirst.Store(true)
}
