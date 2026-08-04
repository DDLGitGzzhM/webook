package ioc

import (
	"github.com/IBM/sarama"
	"github.com/spf13/viper"

	"webook/webook/internal/pkg/logger"
	"webook/webook/reward/events"
	"webook/webook/reward/service"
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

func InitPaymentEventConsumer(
	client sarama.Client,
	l logger.Logger,
	svc service.RewardService,
) *events.PaymentEventConsumer {
	return events.NewPaymentEventConsumer(client, l, svc)
}
