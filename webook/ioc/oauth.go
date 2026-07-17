package ioc

import "webook/webook/internal/service/oauth2/wechat"

func InitWeChatService() wechat.Service {
	return wechat.NewWeChatService("123", "secret")
}
