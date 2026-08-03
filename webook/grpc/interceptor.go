package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/pkg/ratelimit"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.Logger
}

// NewInterceptorBuilder key: user-service
// 整个应用、集群限流
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

func (b *InterceptorBuilder) BuildServerInterceptorServiceBiz() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if idReq, ok := req.(*GetByIdReq); ok {
			limited, err := b.limiter.Limited(ctx,
				// limiter:user:456
				fmt.Sprintf("limiter:user:%s:%d", info.FullMethod, idReq.Id))
			if err != nil {
				// err 不为nil，你要考虑你用保守的，还是用激进的策略
				// 这是保守的策略
				b.l.Error("判定限流出现问题", logger.Error(err.Error()))
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
				// 这是激进的策略
				// return handler(ctx, req)
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
		}
		return handler(ctx, req)
	}
}
