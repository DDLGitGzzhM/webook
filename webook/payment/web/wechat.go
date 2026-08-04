package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

	"webook/webook/internal/pkg/ginx"
	"webook/webook/internal/pkg/logger"
	"webook/webook/payment/service/wechat"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.Logger
	nativeSvc *wechat.NativePaymentService
}

func NewWechatHandler(
	handler *notify.Handler,
	nativeSvc *wechat.NativePaymentService,
	l logger.Logger,
) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
		l:         l,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("/hello", func(context *gin.Context) {
		context.String(http.StatusOK, "我进来了")
	})
	server.Any("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	transaction := &payments.Transaction{}
	// 第一个返回值里面的内容我们暂时用不上
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		return ginx.Result{}, err
	}
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	return ginx.Result{}, err
}
