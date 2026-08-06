package events

// Consumer Kafka 消费者。
type Consumer interface {
	Start() error
}
