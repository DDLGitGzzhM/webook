package ioc

import (
	"webook/webook/internal/service/sms"
	smsTest "webook/webook/internal/service/sms/test"
)

func InitSMSService() sms.Service {
	return smsTest.NewService()
}
