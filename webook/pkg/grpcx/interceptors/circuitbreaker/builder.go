package circuitbreaker

import (
	"context"
	rand2 "math/rand"

	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker

	// 设置标记位
	// 假如说我们考虑使用随机数 + 阈值的回复方式
	// 触发熔断的时候，直接将 threshold 置为0
	// 后续等一段时间，将 theshold 调整为 1，判定请求有没有问题
	threshold int
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp any, err error) {
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			// 借助这个区判定是不是业务错误
			//s, ok :=status.FromError(err)
			//if s != nil && s.Code() == codes.Unavailable {
			//	b.breaker.MarkFailed()
			//} else {
			//
			//}
			if err != nil {
				// 进一步区别是不是系统错
				// 我这边没有区别业务错误和系统错误
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
		}

		b.breaker.MarkFailed()
		// 触发了熔断器
		return nil, err
	}
}

func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !b.allow() {
			b.threshold = b.threshold / 2
			// 这里就是触发了熔断
			//b.threshold = 0
			//time.AfterFunc(time.Minute, func() {
			//	b.threshold = 1
			//})
		}
		// 下面就是随机数判定
		rand := rand2.Intn(100)
		if rand <= b.threshold {
			resp, err = handler(ctx, req)
			if err == nil && b.threshold != 0 {
				// 你要考虑调大 threshold
			} else if b.threshold != 0 {
				// 你要考虑调小 threshold
			}
		}
		return
	}
}

func (b *InterceptorBuilder) allow() bool {
	// 这边就套用我们之前在短信里面讲的，判定节点是否健康的各种做法
	// 从prometheus 里面拿数据判定
	// prometheus.DefaultGatherer.Gather()
	return false
}
