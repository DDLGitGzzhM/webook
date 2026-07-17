package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"webook/webook/internal/service"
	"webook/webook/internal/service/oauth2/wechat"
)

type OAuth2WechatHandler struct {
	svc     wechat.Service
	userSvc service.IUserService
	jwtHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.IUserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthUrl)
	g.Any("/callback", h.CallBack)

}

func (h *OAuth2WechatHandler) AuthUrl(ctx *gin.Context) {
	url, err := h.svc.AuthURL(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
		return
	}
	ctx.JSON(http.StatusOK, url)
}

func (h *OAuth2WechatHandler) CallBack(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")
	info, err := h.svc.VerifyCode(ctx, code, state)
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
	}
	u, err := h.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
		return
	}
	err = h.setJWTToken(ctx, u.Id)
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
		return
	}
	ctx.JSON(http.StatusOK,
		Result{
			Code: 0,
			Msg:  "OK",
		})
}
