package web

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	rewardv1 "webook/webook/api/proto/gen/reward/v1"
	"webook/webook/internal/pkg/ginx"
	jwtHandler "webook/webook/internal/web/jwt"
)

type RewardHandler struct {
	client rewardv1.RewardServiceClient
}

func NewRewardHandler(client rewardv1.RewardServiceClient) *RewardHandler {
	return &RewardHandler{client: client}
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("/reward")
	rg.POST("/detail",
		ginx.WrapBodyAndToken[GetRewardReq, *jwtHandler.UserClaims](h.GetReward))
}

type GetRewardReq struct {
	Rid int64 `json:"rid"`
}

// GetReward 前端传过来一个超长的超时时间，例如说 10s
// 后端去轮询
// 可能引来巨大的性能问题
// 真正优雅的还是前端来轮询
func (h *RewardHandler) GetReward(
	ctx *gin.Context,
	req GetRewardReq,
	claims *jwtHandler.UserClaims,
) (ginx.Result, error) {
	for {
		newCtx, cancel := context.WithTimeout(ctx, time.Second)
		resp, err := h.client.GetReward(newCtx, &rewardv1.GetRewardRequest{
			Rid: req.Rid,
			Uid: claims.UserId,
		})
		cancel()
		if err != nil {
			return ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			}, err
		}
		if resp.Status == rewardv1.RewardStatus_RewardStatusInit {
			continue
		}
		return ginx.Result{
			Data: resp.Status.String(),
		}, nil
	}
}
