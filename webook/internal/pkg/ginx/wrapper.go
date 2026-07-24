package ginx

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"

	"webook/webook/internal/pkg/logger"
)

// L 使用包变量
var L logger.Logger

var vector *prometheus.CounterVec

// InitCounter 初始化业务错误码计数器。
func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt,
		[]string{"code"})
	prometheus.MustRegister(vector)
	// 你可以考虑使用 code, method, 命中路由，HTTP 状态码
}

func reportCode(code int) {
	if vector == nil {
		return
	}
	vector.WithLabelValues(strconv.Itoa(code)).Inc()
}

func WrapToken[C jwt.Claims](fn func(ctx *gin.Context, uc C) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, c)
		if err != nil {
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err.Error()))
		}
		reportCode(res.Code)
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyAndToken[Req any, C jwt.Claims](
	fn func(ctx *gin.Context, req Req, uc C) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			return
		}

		val, ok := ctx.Get("claims")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c, ok := val.(C)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		res, err := fn(ctx, req, c)
		if err != nil {
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err.Error()))
		}
		reportCode(res.Code)
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBodyV1[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err.Error()))
		}
		reportCode(res.Code)
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapBody[T any](
	l logger.Logger, fn func(ctx *gin.Context, req T) (Result, error),
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			l.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err.Error()))
		}
		reportCode(res.Code)
		ctx.JSON(http.StatusOK, res)
	}
}
