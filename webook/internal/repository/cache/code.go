package cache

import (
	"context"
	_ "embed"
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany = errors.New("code:send:too:many")
	ErrCodeSystem      = errors.New("code:system:error")
	ErrorVerifyTooMany = errors.New("code:verify:too:many")
	ErrorUnknown       = errors.New("code:unknown")
)

// 编译器会在编译的时候，放入
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}
type RedisCodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *RedisCodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.Key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		return nil
	case -1: // 发送频繁
		return ErrCodeSendTooMany
	default: // 系统错误
		return ErrCodeSystem
	}
}

func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, expectCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.Key(biz, phone)}, expectCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrorVerifyTooMany
	case -2:
		return false, nil
	default:
		return false, ErrorUnknown
	}
}

func (c *RedisCodeCache) Key(biz, phone string) string {
	return "phone_code:" + biz + ":" + phone
}
