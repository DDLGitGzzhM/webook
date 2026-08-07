package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"webook/webook/internal/events"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/article"
)

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

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

func InitMySQLBinlogConsumer(
	client sarama.Client,
	l logger.Logger,
	repo *article.CachedArticleRepository,
) *events.MySQLBinlogConsumer {
	return events.NewMySQLBinlogConsumer(client, l, repo)
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(binlog *events.MySQLBinlogConsumer) []events.Consumer {
	return []events.Consumer{
		binlog,
	}
}
