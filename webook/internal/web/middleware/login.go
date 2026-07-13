package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}
func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" {
			ctx.Next()
			return
		}
		sess := sessions.Default(ctx)
		if sess == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		id := sess.Get("userId")
		if id == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		sess.Set("userId", id)
		sess.Options(sessions.Options{
			MaxAge: 20 * 60,
		})
		updateTime := sess.Get("update_time")
		if updateTime == nil {
			sess.Set("update_time", time.Now().UnixMilli())
			sess.Save()
			return
		}
		updateTimeVal, _ := updateTime.(int64)
		if time.Now().UnixMilli()-updateTimeVal > 10*1000 {
			sess.Set("update_time", time.Now().UnixMilli())
			sess.Save()
			return
		}
	}
}
