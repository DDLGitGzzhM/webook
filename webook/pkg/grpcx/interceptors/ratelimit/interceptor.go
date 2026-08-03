package ratelimit

import (
	"context"
	"strings"

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

	// key 是 FullMethod, value 是默认值的 json
	//defaultValueMap map[string]string
	//defaultValueFuncMap map[string]func() any
}

// NewInterceptorBuilder key: user-service
// "limiter:service:user" 整个应用、集群限流
// "limiter:service:user:UserService" user 里面的 UserService 限流
func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limited(ctx, b.key)
		if err != nil {
			// err 不为nil，你要考虑你用保守的，还是用激进的策略
			// 这是保守的策略
			b.l.Error("判定限流出现问题", logger.Error(err.Error()))
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")

			// 这是激进的策略
			// return handler(ctx, req)
		}
		if limited {
			//defVal, ok := b.defaultValueMap[info.FullMethod]
			//if ok {
			//	err = json.Unmarshal([]byte(defVal), &resp)
			//	return defVal, err
			//}
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerInterceptorV1 用来配合后面业务的
func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limited(ctx, b.key)
		if err != nil || limited {
			ctx = context.WithValue(ctx, "limited", "true")
		}

		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limited, err := b.limiter.Limited(ctx, b.key)
		if err != nil {
			// err 不为nil，你要考虑你用保守的，还是用激进的策略
			// 这是保守的策略
			b.l.Error("判定限流出现问题", logger.Error(err.Error()))
			return status.Errorf(codes.ResourceExhausted, "触发限流")
			// 这是激进的策略
			// return handler(ctx, req)
		}
		if limited {
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// BuildServerInterceptorService 服务级别限流
func (b *InterceptorBuilder) BuildServerInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := b.limiter.Limited(ctx, "limiter:service:user:UserService")
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
