package grpc

import (
	"context"
	"errors"

	commentv1 "webook/webook/api/proto/gen/comment/v1"
)

type RateLimitCommentService struct {
	CommentServiceServer
}

func (c *RateLimitCommentService) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	if ctx.Value("downgrade") == "true" && !c.hotBiz(request.Biz, request.GetBizid()) {
		return nil, errors.New("触发了降级，非热门资源")
	}
	return c.CommentServiceServer.GetCommentList(ctx, request)
}

func (c *RateLimitCommentService) GetMoreReplies(ctx context.Context, req *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	if ctx.Value("downgrade") == "true" {
		return nil, errors.New("触发限流")
	}
	return c.CommentServiceServer.GetMoreReplies(ctx, req)
}

func (c *RateLimitCommentService) hotBiz(biz string, bizId int64) bool {
	// 这个热门资源怎么判定
	// 一般是借助周期性的任务来计算一个白名单，放进去 redis 里面。
	return true
}
