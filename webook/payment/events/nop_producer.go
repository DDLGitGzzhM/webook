package events

import "context"

type NopProducer struct{}

func (n NopProducer) ProducePaymentEvent(ctx context.Context, evt PaymentEvent) error {
	return nil
}
