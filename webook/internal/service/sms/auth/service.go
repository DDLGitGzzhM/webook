package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"

	"webook/webook/internal/service/sms"
)

type SMSService struct {
	svc sms.Service
	key string
}

func (s *SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	var tc Claims

	token, err := jwt.ParseWithClaims(biz, &tc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token is invalid")
	}
	return s.svc.Send(ctx, tc.TplId, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	TplId string
}
