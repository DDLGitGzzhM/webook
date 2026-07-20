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
}

func (u *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
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
	save, err := u.svc.Save(ctx, domain.Article{
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id:   claims.UserId,
			Name: "system",
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result[any]{
			Code: 5,
			Msg:  "保存失败",
		})
		u.l.Error("保存文章失败", logger.Error(err.Error()))
		return
	}
	ctx.JSON(http.StatusOK, Result[int64]{
		Code: 0,
		Msg:  "ok",
		Data: save,
	})
}
