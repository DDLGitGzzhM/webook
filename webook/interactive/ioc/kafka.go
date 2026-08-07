package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	ievents "webook/webook/interactive/events"
	"webook/webook/internal/events"
	"webook/webook/internal/events/article"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/migrator/events/fixer"
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

func InitSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return res
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(
	intr *article.InteractiveReadEventBatchConsumer,
	fix *fixer.Consumer[dao.Interactive],
	binlog *ievents.MySQLBinlogConsumer[dao.Interactive],
) []events.Consumer {
	return []events.Consumer{
		intr,
		fix,
		binlog,
	}
}
