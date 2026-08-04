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
	// 这地方不能 wrap
	server.Any("/pay/callback", ginx.Wrap(h.HandleNative))
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) (ginx.Result, error) {
	txn := new(payments.Transaction)
	_, err := h.handler.ParseNotifyRequest(ctx.Request.Context(), ctx.Request, txn)
	if err != nil {
		// 这里不可能触发对账，你解密都出错了，你拿不到 BizTradeNO
		// 返回非 2xx 的响应
		// 就一个原因：有人伪造请求，有人在伪造微信支付的回调
		// 做好监控和告警
		// 大量进来这个分支，就说明有人搞你
		return ginx.Result{}, err
	}
	// 当你下来这里的时候，交易信息已经被解密好了，放到了 txn 里面
	// 也就是说，我们现在就是要处理一下 txn 就可以
	err = h.nativeSvc.HandleCallback(ctx, txn)
	if err != nil {
		// 这里处理失败了，你可以再次触发对账
		// 返回非 2xx 的响应
		return ginx.Result{}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
