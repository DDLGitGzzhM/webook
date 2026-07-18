package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"

	"webook/webook/internal/web"
)

type LoginMiddlewareBuilder struct {
	cmd redis.Cmdable
}

func NewLoginMiddlewareBuilder(cmd redis.Cmdable) *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{
		cmd: cmd,
	}
}
func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" ||
			ctx.Request.URL.Path == "/users/refresh_token" ||
			ctx.Request.URL.Path == "/users/login_sms/code/send" ||
			ctx.Request.URL.Path == "/users/login_sms" ||
			ctx.Request.URL.Path == "/oauth2/wechat/authurl" ||
			ctx.Request.URL.Path == "/oauth2/wechat/callback" {
			ctx.Next()
			return
		}
		tokenHeader := web.ExtractToken(ctx)
		claims := &web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenHeader, claims, func(token *jwt.Token) (any, error) {
			return []byte("secret"), nil
		})
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if claims.UserAgents != ctx.Request.UserAgent() {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		/*
			如果 redis 崩了 return 不去考虑是否登陆
		*/
		cnt, err := l.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid)).Result()
		if err != nil || cnt > 0 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("claims", claims)
	}
}
