package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/service"
	jwtHandler "webook/webook/internal/web/jwt"
)

type ArticleHandler struct {
	svc service.IArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.IArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{
		svc: svc,
		l:   l,
	}
}
func (u *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", u.Edit)
	g.POST("/publish", u.Publish)
	g.POST("/withdraw", u.Withdraw)
}

func (u *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwtHandler.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err := u.svc.Withdraw(ctx, claims.UserId, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		u.l.Error("撤回文章失败", logger.Error(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "ok",
	})
}

func (u *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// todo 检测输入
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwtHandler.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	save, err := u.svc.Publish(ctx, req.toDomain(claims.UserId))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "保存失败",
		})
		u.l.Error("发布文章失败", logger.Error(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "ok",
		Data: save,
	})
}

func (u *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// todo 检测输入
	c := ctx.MustGet("claims")
	claims, ok := c.(*jwtHandler.UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	save, err := u.svc.Save(ctx, req.toDomain(claims.UserId))
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "保存失败",
		})
		u.l.Error("保存文章失败", logger.Error(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Code: 0,
		Msg:  "ok",
		Data: save,
	})
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id:   uid,
			Name: "system",
		},
	}
}
