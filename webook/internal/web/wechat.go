package web

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"

	"webook/webook/internal/service"
	"webook/webook/internal/service/oauth2/wechat"
	jwtHandler "webook/webook/internal/web/jwt"
)

type OAuth2WechatHandler struct {
	svc        wechat.Service
	userSvc    service.IUserService
	jwtHandler jwtHandler.Handler
	stateKey   []byte
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.IUserService, handler jwtHandler.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:        svc,
		userSvc:    userSvc,
		stateKey:   []byte("stateKey"),
		jwtHandler: handler,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.AuthUrl)
	g.Any("/callback", h.CallBack)

}

func (h *OAuth2WechatHandler) AuthUrl(ctx *gin.Context) {
	state := uuid.New()
	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	})
	tokenStr, err := token.SignedString([]byte("secret"))
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
		return
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/oauth2/wechat/callback", "", true, true)
	ctx.JSON(http.StatusOK, url)
}

func (h *OAuth2WechatHandler) CallBack(ctx *gin.Context) {
	code := ctx.Query("code")
	state := ctx.Query("state")

	ck, err := ctx.Cookie("jwt-state")
	if err != nil {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  err.Error(),
			})
	}
	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil || !token.Valid {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  "token 无效",
			})
		return
	}

	if sc.State != state {
		ctx.JSON(http.StatusOK,
			Result{
				Code: 5,
				Msg:  "state 无效",
			})
		return
	}

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
	if err = h.jwtHandler.SetLoginToken(ctx, u.Id); err != nil {
		return
	}
	ctx.JSON(http.StatusOK,
		Result{
			Code: 0,
			Msg:  "OK",
		})
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
