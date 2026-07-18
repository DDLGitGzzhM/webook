package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	jwtHandler "webook/webook/internal/web/jwt"
)

type LoginMiddlewareBuilder struct {
	jwtHandler jwtHandler.Handler
}

func NewLoginMiddlewareBuilder(jwtHandler jwtHandler.Handler) *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{
		jwtHandler: jwtHandler,
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
		tokenHeader := l.jwtHandler.ExtractToken(ctx)
		claims := &jwtHandler.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenHeader, claims, func(token *jwt.Token) (any, error) {
			return []byte(jwtHandler.AtKey), nil
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
		err = l.jwtHandler.CheckSession(ctx, claims.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("claims", claims)
	}
}
