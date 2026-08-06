package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"webook/webook/search/events"
)

// InitKafka 初始化 Kafka 客户端。
func InitKafka() sarama.Client {
	type Config struct {
		Addrs []string `yaml:"addrs"`
	}
	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addrs, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

// NewConsumers 注册所有 Kafka 消费者。
func NewConsumers(
	articleConsumer *events.ArticleConsumer,
	userConsumer *events.UserConsumer,
) []events.Consumer {
	return []events.Consumer{
		articleConsumer,
		userConsumer,
	}
}
