package job

import (
	"context"
	"encoding/json"
	"time"

	"webook/webook/payment/events"
	"webook/webook/payment/repository"
)

// LocalMsgResendJob 本地消息表补偿
// 我们做一个简单的策略
//  1. 每次遍历所有处于初始状态的本地消息记录，并且发送
//  2. 如果一个本地消息记录的 Ctime 在十分钟之前，我们认为它已经重试了足够多的次数，
//     但是最终都没有成功，那么我们会把这个记录标记为失败
type LocalMsgResendJob struct {
	repo      repository.LocalMsgRepository
	producer  events.Producer
	threshold time.Duration
}

// NewLocalMsgResendJob 创建本地消息补偿任务
func NewLocalMsgResendJob(repo repository.LocalMsgRepository, producer events.Producer, threshold time.Duration) *LocalMsgResendJob {
	return &LocalMsgResendJob{repo: repo, producer: producer, threshold: threshold}
}

func (l LocalMsgResendJob) Name() string {
	return "LocalMsgResendJob"
}

func (l LocalMsgResendJob) Run() error {
	offset := 0
	const limit = 100
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		// 批量操作可以提高性能
		// 找出初始化状态的 msg
		msgs, err := l.repo.FindInitMsg(ctx, offset, limit)
		cancel()
		if err != nil {
			// 你也可以继续，但是没太大必要
			return err
		}
		for _, msg := range msgs {
			var evt events.PaymentEvent
			err = json.Unmarshal([]byte(msg.Content), &evt)
			if err != nil {
				continue
			}

			ctx, cancel = context.WithTimeout(context.Background(), time.Second*3)
			// 消费者那边一定要做到幂等
			err = l.producer.ProducePaymentEvent(ctx, evt)
			if err != nil {
				// 要判定一下是不是不值得重试了
				// 这个判定条件就是，过了好久，比如说 threshold 等于十分钟
				// 那么就是十分钟前的消息都还没成功，多半已经没法子成功了
				if msg.Ctime.Add(l.threshold).Before(time.Now()) {
					err = l.repo.MarkFailed(ctx, msg.Id)
					if err != nil {
						// 记录日志，多半是需要人手工处理的
					}
				}
			} else {
				err = l.repo.MarkSuccess(ctx, msg.Id)
				if err != nil {
					// 记录日志
				}
			}
			cancel()
		}
		if len(msgs) < limit {
			// 没了
			return nil
		}
	}
}
