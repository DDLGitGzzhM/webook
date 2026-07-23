package article

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

const TopicReadEvent = "article_read"

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

type ReadEvent struct {
	Uid int64
	Aid int64
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

// ProduceReadEvent 如果你有复杂的重试逻辑，就用装饰器
// 你认为你的重试逻辑很简单，你就放这里
func (k *KafkaProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	_, _, err = k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicReadEvent,
		Value: sarama.ByteEncoder(data),
	})
	return err
}

func NewKafkaProducer(pc sarama.SyncProducer) Producer {
	return &KafkaProducer{
		producer: pc,
	}
}

// NoOpProducer 用于不需要真正发消息的场景（例如集成测试）
type NoOpProducer struct{}

func NewNoOpProducer() Producer {
	return &NoOpProducer{}
}

func (n *NoOpProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	return nil
}
