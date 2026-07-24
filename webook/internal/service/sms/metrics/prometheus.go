package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"webook/webook/internal/service/sms"
)

// PrometheusDecorator 装饰 SMS Service，统计发送耗时。
type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

// NewPrometheusDecorator 创建 SMS 性能监控装饰器。
func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

// Send 记录一次短信发送的耗时。
func (p *PrometheusDecorator) Send(ctx context.Context,
	biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
