package ratelimit

import "context"

type Limiter interface {
	Limited(ctx context.Context, key string) (bool, error)
}
