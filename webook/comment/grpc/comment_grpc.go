package grpc

import (
	"context"

	commentv1 "webook/webook/api/proto/gen/comment/v1"
)

type CommentServiceAsyncServer struct {
	CommentServiceServer
}

func (c *CommentServiceAsyncServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	if ctx.Value("limited") == "true" || ctx.Value("downgrad") == "true" {
		// 在这里发送到 Kafka 里面
		return &commentv1.CreateCommentResponse{}, nil
	} else {
		err := c.svc.CreateComment(ctx, convertToDomain(request.GetComment()))
		return &commentv1.CreateCommentResponse{}, err
	}
}

func (c *CommentServiceAsyncServer) CreateCommentV1(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	// 在这里发送到 Kafka 里面
	return &commentv1.CreateCommentResponse{}, nil
}
