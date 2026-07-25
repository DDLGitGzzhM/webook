package ioc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	trace2 "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/embedded"
	"go.opentelemetry.io/otel/trace/noop"
)

// InitOTEL 初始化 OpenTelemetry，返回关闭函数。
func InitOTEL() func(ctx context.Context) {
	res, err := newResource("webook", "v0.0.1")
	if err != nil {
		panic(err)
	}
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tp)
	newTp := &MyTracerProvider{
		Enable:      true,
		nopProvider: noop.NewTracerProvider(),
		provider:    tp,
	}
	// 监听配置变更就可以了
	otel.SetTracerProvider(newTp)

	return func(ctx context.Context) {
		_ = tp.Shutdown(ctx)
	}
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New(
		"http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

type MyTracerProvider struct {
	embedded.TracerProvider
	// 改原子操作
	Enable      bool
	nopProvider trace2.TracerProvider
	provider    trace2.TracerProvider
}

func (m *MyTracerProvider) Tracer(name string, options ...trace2.TracerOption) trace2.Tracer {
	if m.Enable {
		return m.provider.Tracer(name, options...)
	}
	return m.nopProvider.Tracer(name, options...)
}

func (m *MyTracerProvider) TracerV1(name string, options ...trace2.TracerOption) trace2.Tracer {
	return &MyTracer{
		nopTracer: m.nopProvider.Tracer(name, options...),
		tracer:    m.provider.Tracer(name, options...),
	}
}

type MyTracer struct {
	embedded.Tracer
	Enable    bool
	nopTracer trace2.Tracer
	tracer    trace2.Tracer
}

func (m *MyTracer) Start(ctx context.Context, spanName string, opts ...trace2.SpanStartOption) (context.Context, trace2.Span) {
	if m.Enable {
		return m.tracer.Start(ctx, spanName, opts...)
	}
	return m.nopTracer.Start(ctx, spanName, opts...)
}
