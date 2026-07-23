package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/ginx"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/service"
	jwtHandler "webook/webook/internal/web/jwt"
)

type ArticleHandler struct {
	svc     service.IArticleService
	l       logger.Logger
	intrSvc service.InteractiveService
	biz     string
}

func NewArticleHandler(
	svc service.IArticleService,
	intrSvc service.InteractiveService,
	l logger.Logger,
) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		l:       l,
		biz:     "article",
	}
}

func (u *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", u.Edit)
	g.POST("/publish", u.Publish)
	g.POST("/withdraw", u.Withdraw)
	// 创作者查询接口，遵循 RESTful 更宜用 GET
	g.POST("/list",
		ginx.WrapBodyAndToken[ListReq, *jwtHandler.UserClaims](u.List))
	g.GET("/detail/:id", ginx.WrapToken[*jwtHandler.UserClaims](u.Detail))

	pub := g.Group("/pub")
	pub.GET("/:id", u.PubDetail)
	// 点赞是这个接口，取消点赞也是这个接口
	// RESTful 风格
	//pub.POST("/like/:id", ginx.WrapBodyAndToken[LikeReq,
	//	*jwtHandler.UserClaims](u.Like))
	pub.POST("/like", ginx.WrapBodyAndToken[LikeReq,
		*jwtHandler.UserClaims](u.Like))
	//pub.POST("/cancel_like", ginx.WrapBodyAndToken[LikeReq,
	//	*jwtHandler.UserClaims](u.Like))
}

func (u *ArticleHandler) Like(
	ctx *gin.Context, req LikeReq, uc *jwtHandler.UserClaims,
) (ginx.Result, error) {
	var err error
	if req.Like {
		err = u.intrSvc.Like(ctx, u.biz, req.Id, uc.UserId)
	} else {
		err = u.intrSvc.CancelLike(ctx, u.biz, req.Id, uc.UserId)
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}

func (u *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "参数错误",
		})
		u.l.Error("前端输入的 ID 不对", logger.Error(err.Error()))
		return
	}
	art, err := u.svc.GetPublishedById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		u.l.Error("获得文章信息失败", logger.Error(err.Error()))
		return
	}

	// 增加阅读计数。
	go func() {
		// 开一个 goroutine，异步去执行
		er := u.intrSvc.IncrReadCnt(ctx, u.biz, art.Id)
		if er != nil {
			u.l.Error("增加阅读计数失败",
				logger.Int64("aid", art.Id),
				logger.Error(er.Error()))
		}
	}()

	// 这个功能是不是可以让前端，主动发一个 HTTP 请求，来增加一个计数？
	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			Author:  art.Author.Name,
			Ctime:   art.Ctime.Format(time.DateTime),
			Utime:   art.Utime.Format(time.DateTime),
		},
	})
}

func (u *ArticleHandler) Detail(
	ctx *gin.Context, usr *jwtHandler.UserClaims,
) (ginx.Result, error) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		return ginx.Result{
			Code: 4,
			Msg:  "参数错误",
		}, err
	}
	art, err := u.svc.GetById(ctx, id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	if art.Author.Id != usr.UserId {
		return ginx.Result{
			Code: 4,
			Msg:  "输入有误",
		}, fmt.Errorf("非法访问文章，创作者 ID 不匹配 %d", usr.UserId)
	}
	return ginx.Result{
		Data: ArticleVO{
			Id:      art.Id,
			Title:   art.Title,
			Status:  art.Status.ToUint8(),
			Content: art.Content,
			Ctime:   art.Ctime.Format(time.DateTime),
			Utime:   art.Utime.Format(time.DateTime),
		},
	}, nil
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

func (u *ArticleHandler) List(
	ctx *gin.Context, req ListReq, uc *jwtHandler.UserClaims,
) (ginx.Result, error) {
	res, err := u.svc.List(ctx, uc.UserId, req.Offset, req.Limit)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: lo.Map(res, func(src domain.Article, _ int) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				Status:   src.Status.ToUint8(),
				Ctime:    src.Ctime.Format(time.DateTime),
				Utime:    src.Utime.Format(time.DateTime),
			}
		}),
	}, nil
}
