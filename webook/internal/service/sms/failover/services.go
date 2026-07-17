package failover

import (
	"context"
	"errors"
	"sync/atomic"

	"webook/webook/internal/service/sms"
)

type FailOverSMSService struct {
	svcs []sms.Service
	idx  uint64
}

func NewFailOverSMSService(svcs ...sms.Service) *FailOverSMSService {
	return &FailOverSMSService{
		svcs: svcs,
	}
}

func (f *FailOverSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	for _, svc := range f.svcs { // 绝大多熟请求都会在 svcs[0] , 负载不均衡
		err := svc.Send(ctx, biz, args, numbers...)
		if err == nil {
			return nil
		}
		// 输出日志
		// 做监控
	}
	return errors.New("所有服务商都失败了")
}

func (f *FailOverSMSService) SendV1(ctx context.Context, biz string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	for i := idx; i < idx+length; i++ {
		svc := f.svcs[int(i%length)]
		err := svc.Send(ctx, biz, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		default:
			return nil
		}
	}
	return errors.New("所有服务商都失败了")
}
