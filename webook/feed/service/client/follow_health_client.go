package client

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	followv1 "webook/webook/api/proto/gen/follow/v1"
)

type FollowClient struct {
	// 这个是真实的 RPC 客户端
	followv1.FollowServiceClient

	downgrade *atomic.Bool
}

func (f *FollowClient) GetFollowee(ctx context.Context, in *followv1.GetFolloweeRequest, opts ...grpc.CallOption) (resp *followv1.GetFolloweeResponse, err error) {
	if f.downgrade.Load() {
		// 或者返回特定的 error
		return nil, nil
	}
	defer func() {
		// 比如说这个，限流
		if status.Code(err) == codes.Unavailable {
			f.downgrade.Store(true)
			// 这边呢？
			go func() {
				// 发心跳给 follow 检测，尝试退出 downgrade 状态
			}()
		}
	}()
	resp, err = f.FollowServiceClient.GetFollowee(ctx, in)
	return
}
