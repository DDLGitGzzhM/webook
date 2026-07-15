package retryable

//type Service struct {
//	svc sms.Service
//	// 重试次数
//	retryCount int
//}
//
//func NewService(svc sms.Service, retryCount int) *Service {
//	return &Service{
//		svc:        svc,
//		retryCount: retryCount,
//	}
//}
//
//func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
//	err := s.svc.Send(ctx, tpl, args, numbers...)
//	for err != nil && s.retryCount < 3 {
//		s.retryCount++
//		err = s.svc.Send(ctx, tpl, args, numbers...)
//	}
//	if err != nil {
//		return err
//	}
//	return nil
//}
