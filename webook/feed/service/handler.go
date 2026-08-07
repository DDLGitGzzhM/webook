package service

import (
	"context"

	"webook/webook/feed/domain"
)

type LikeEventHandlerV2 struct {
	LikeEventHandler
}

// CreateFeedEvent 重写发生了变化的方法
func (l *LikeEventHandlerV2) CreateFeedEvent(ctx context.Context, ext domain.ExtendFields) error {
	// 新版逻辑
	return nil
}
