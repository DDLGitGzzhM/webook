package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type MiddleWareBuilder struct {
	loggerFunc    func(ctx context.Context, al *Accesslog)
	allowReqBody  bool
	allowRespBody bool
}

func NewMiddleWareBuilder(loggerFunc func(ctx context.Context, al *Accesslog)) *MiddleWareBuilder {
	return &MiddleWareBuilder{
		loggerFunc: loggerFunc,
	}
}
func (b *MiddleWareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		al := Accesslog{
			Method: ctx.Request.Method,
			Url:    ctx.Request.URL.String(),
		}
		if ctx.Request.Body != nil && b.allowReqBody {
			body, _ := io.ReadAll(ctx.Request.Body)
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			al.Req = string(body)
		}
		if b.allowRespBody {
			w := responseWriter{
				al:             &al,
				ResponseWriter: ctx.Writer,
			}
			ctx.Writer = w
		}
		defer func() {
			duration := time.Since(start)
			al.Duration = duration
			b.loggerFunc(ctx, &al)
		}()
		ctx.Next()
	}
}

func (b *MiddleWareBuilder) AllowReqBody() *MiddleWareBuilder {
	b.allowReqBody = true
	return b
}
func (b *MiddleWareBuilder) AllowRespBody() *MiddleWareBuilder {
	b.allowRespBody = true
	return b
}

type responseWriter struct {
	al *Accesslog
	gin.ResponseWriter
}

func (w responseWriter) Write(data []byte) (int, error) {
	w.al.Resp = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.al.Resp = data
	return w.ResponseWriter.WriteString(data)
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)

}

type Accesslog struct {
	Method   string
	Url      string
	Req      string
	Resp     string
	Duration time.Duration
	status   int
}
