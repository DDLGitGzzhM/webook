package ginx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"webook/webook/internal/pkg/logger"
)

type testClaims struct {
	jwt.RegisteredClaims
}

func TestInitCounter_WrapToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := prometheus.NewRegistry()
	oldReg := prometheus.DefaultRegisterer
	prometheus.DefaultRegisterer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = oldReg
		vector = nil
	})

	InitCounter(prometheus.CounterOpts{
		Namespace: "test",
		Subsystem: "webook",
		Name:      "http_biz_code",
		Help:      "test",
	})
	L = logger.NopLogger{}

	server := gin.New()
	server.GET("/demo", func(ctx *gin.Context) {
		ctx.Set("claims", testClaims{})
		WrapToken[testClaims](func(ctx *gin.Context, uc testClaims) (Result, error) {
			return Result{Code: 401002, Msg: "bad"}, nil
		})(ctx)
	})

	req := httptest.NewRequest(http.MethodGet, "/demo", nil)
	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	families, err := registry.Gather()
	require.NoError(t, err)
	require.NotEmpty(t, families)
	assert.Equal(t, "test_webook_http_biz_code", families[0].GetName())
}
