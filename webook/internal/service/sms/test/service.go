package test

import (
	"context"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	println("发送短信: ", args[0])
	return nil
}
