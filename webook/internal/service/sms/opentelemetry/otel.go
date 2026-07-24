package opentelemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"webook/webook/internal/service/sms"
)

// Service 装饰 SMS Service，为发送过程创建 span。
type Service struct {
	svc    sms.Service
	tracer trace.Tracer
}

// NewService 创建带 OpenTelemetry tracing 的 SMS 装饰器。
func NewService(svc sms.Service) *Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("webook/webook/internal/service/sms/opentelemetry")
	return &Service{
		svc:    svc,
		tracer: tracer,
	}
}

// Send 记录一次短信发送的 span。
func (s *Service) Send(ctx context.Context,
	biz string,
	args []string,
	numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send_"+biz,
		// 因为我是一个调用短信服务商的客户端
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End(trace.WithStackTrace(true))

	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}

	return err
}
