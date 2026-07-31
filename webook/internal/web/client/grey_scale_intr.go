package client

import (
	"context"
	"math/rand"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"

	intrv1 "webook/webook/api/proto/gen/intr/v1"
)

type GreyScaleInteractiveServiceClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient
	// 用随机数 + 阈值的小技巧控制流量
	threshold *atomicx.Value[int32]
}

func NewGreyScaleInteractiveServiceClient(
	remote intrv1.InteractiveServiceClient,
	local intrv1.InteractiveServiceClient,
) *GreyScaleInteractiveServiceClient {
	return &GreyScaleInteractiveServiceClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValue[int32](),
	}
}

func (g *GreyScaleInteractiveServiceClient) OnChange(ch <-chan int32) {
	go func() {
		for newTh := range ch {
			g.threshold.Store(newTh)
		}
	}()
}

func (g *GreyScaleInteractiveServiceClient) OnChangeV1() chan<- int32 {
	ch := make(chan int32, 100)
	go func() {
		for newTh := range ch {
			g.threshold.Store(newTh)
		}
	}()
	return ch
}

func (g *GreyScaleInteractiveServiceClient) IncrReadCnt(
	ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption,
) (*intrv1.IncrReadCntResponse, error) {
	return g.client().IncrReadCnt(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Like(
	ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption,
) (*intrv1.LikeResponse, error) {
	return g.client().Like(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) CancelLike(
	ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption,
) (*intrv1.CancelLikeResponse, error) {
	return g.client().CancelLike(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Collect(
	ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption,
) (*intrv1.CollectResponse, error) {
	return g.client().Collect(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Get(
	ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption,
) (*intrv1.GetResponse, error) {
	return g.client().Get(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) GetByIds(
	ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption,
) (*intrv1.GetByIdsResponse, error) {
	return g.client().GetByIds(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) UpdateThreshold(newThreshold int32) {
	g.threshold.Store(newThreshold)
}

func (g *GreyScaleInteractiveServiceClient) client() intrv1.InteractiveServiceClient {
	threshold := g.threshold.Load()
	// [0-100) 的随机数
	num := rand.Int31n(100)
	// 举例来说，如果要是 threshold 是 100，
	// 可以预见的是，所有的 num 都会进去，返回 remote
	if num < threshold {
		return g.remote
	}
	// 假如说我的 threshold 是 0，那么就会永远用本地的
	return g.local
}
