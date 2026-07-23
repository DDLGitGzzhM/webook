package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// MiddlewareBuilder 构建 Prometheus HTTP 指标中间件。
type MiddlewareBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

// Build 返回统计请求耗时与活跃请求数的 Gin 中间件。
func (m *MiddlewareBuilder) Build() gin.HandlerFunc {
	// pattern 是指你命中的路由
	// status 是指你的 HTTP 的 status
	labels := []string{"method", "pattern", "status"}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_resp_time",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(summary)
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_active_req",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()
		defer func() {
			duration := time.Since(start)
			gauge.Dec()
			pattern := ctx.FullPath()
			if pattern == "" {
				pattern = "unknown"
			}
			summary.WithLabelValues(
				ctx.Request.Method,
				pattern,
				strconv.Itoa(ctx.Writer.Status()),
			).Observe(float64(duration.Milliseconds()))
		}()
		ctx.Next()
	}
}
