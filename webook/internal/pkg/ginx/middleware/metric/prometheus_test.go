package metric

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddlewareBuilder_Build(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	oldRegisterer := prometheus.DefaultRegisterer
	oldGatherer := prometheus.DefaultGatherer
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = oldRegisterer
		prometheus.DefaultGatherer = oldGatherer
	})

	server := gin.New()
	server.Use((&MiddlewareBuilder{
		Namespace:  "test_ns",
		Subsystem:  "webook",
		Name:       "gin_http",
		Help:       "test help",
		InstanceID: "ut-1",
	}).Build())
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "ok", resp.Body.String())

	families, err := registry.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, families)

	names := make(map[string]struct{}, len(families))
	for _, family := range families {
		names[family.GetName()] = struct{}{}
	}
	assert.Contains(t, names, "test_ns_webook_gin_http_resp_time")
	assert.Contains(t, names, "test_ns_webook_gin_http_active_req")
}
