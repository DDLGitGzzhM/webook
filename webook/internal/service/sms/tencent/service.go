package tencent

import (
	"context"
	"errors"
	"fmt"

	"github.com/samber/lo"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

var ErrorNumerEmpty = errors.New("手机号不能为空")

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
}

func NewService(client *sms.Client, appId, signName string) *Service {
	return &Service{
		appId:    &appId,
		signName: &signName,
		client:   client,
	}
}

func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	if numbers == nil {
		return ErrorNumerEmpty
	}
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = &tpl

	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.TemplateParamSet = s.toStringPtrSlice(args)

	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *status.Code != "Ok" {
			return fmt.Errorf("短信发送失败: code:%s, 原因:%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s Service) toStringPtrSlice(src []string) []*string {
	return lo.Map[string, *string](src, func(src string, idx int) *string {
		return &src
	})
}
