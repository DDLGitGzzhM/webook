package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/ratelimit"
)

type LimiterUserServer struct {
	limiter ratelimit.Limiter
	l       logger.Logger
	UserServiceServer
}

func (s *LimiterUserServer) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	limited, err := s.limiter.Limited(ctx,

		//fmt.Sprintf("limiter:user:get_by_id:%d", req.Id))
		// limiter:user:456
		fmt.Sprintf("limiter:user:%d", req.Id))
	if err != nil {
		// err 不为nil，你要考虑你用保守的，还是用激进的策略
		// 这是保守的策略
		s.l.Error("判定限流出现问题", logger.Error(err.Error()))
		return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
		// 这是激进的策略
		// return handler(ctx, req)
	}
	if limited {
		return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
	}

	resp, err := s.UserServiceServer.GetById(ctx, req)
	return resp, err
}

//func (s *LimiterUserServer) UpdateById(ctx context.Context, req *UpdateByIdReq) (*UpdateByIdResp, error) {
//
//}
