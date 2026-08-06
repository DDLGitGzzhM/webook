package web

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/ginx"
	"webook/webook/internal/service"
)

type RankingHandler struct {
	svc service.RankingService
}

func NewRankingHandler(svc service.RankingService) *RankingHandler {
	return &RankingHandler{svc: svc}
}

func (h *RankingHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/ranking")
	g.GET("/top", ginx.Wrap(h.TopN))
}

// TopN 获取热榜文章列表。
func (h *RankingHandler) TopN(ctx *gin.Context) (ginx.Result, error) {
	arts, err := h.svc.GetTopN(ctx)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: lo.Map(arts, func(src domain.Article, _ int) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Author:   src.Author.Name,
				Status:   src.Status.ToUint8(),
				Ctime:    src.Ctime.Format(time.DateTime),
				Utime:    src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}
