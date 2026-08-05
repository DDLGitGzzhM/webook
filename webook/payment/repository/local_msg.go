package repository

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"

	"webook/webook/payment/domain"
	"webook/webook/payment/repository/dao"
)

var _ LocalMsgRepository = (*LocalMsgGORMRepository)(nil)

// LocalMsgRepository 本地消息表仓储
type LocalMsgRepository interface {
	AddMsg(ctx context.Context, content string) (int64, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
	FindInitMsg(ctx context.Context, offset, limit int) ([]domain.Msg, error)
}

// LocalMsgGORMRepository 基于 GORM 的本地消息表仓储
type LocalMsgGORMRepository struct {
	dao *dao.LocalMsgGORMDAO
}

func (l *LocalMsgGORMRepository) FindInitMsg(ctx context.Context, offset, limit int) ([]domain.Msg, error) {
	msgs, err := l.dao.FindInitMsg(ctx, offset, limit)
	return slice.Map(msgs, func(idx int, src dao.Msg) domain.Msg {
		return domain.Msg{
			Id:      src.Id,
			Content: src.Content,
			Ctime:   time.UnixMilli(src.Ctime),
		}
	}), err
}

func (l *LocalMsgGORMRepository) MarkFailed(ctx context.Context, id int64) error {
	return l.dao.UpdateStatus(ctx, id, dao.MsgStatusFailed)
}

func (l *LocalMsgGORMRepository) MarkSuccess(ctx context.Context, id int64) error {
	return l.dao.UpdateStatus(ctx, id, dao.MsgStatusSuccess)
}

// NewLocalMsgGORMRepository 创建本地消息表仓储
func NewLocalMsgGORMRepository(db *gorm.DB) *LocalMsgGORMRepository {
	return &LocalMsgGORMRepository{dao: dao.NewLocalMsgGORMDAO(db)}
}

func (l *LocalMsgGORMRepository) AddMsg(ctx context.Context, content string) (int64, error) {
	return l.dao.AddMsg(ctx, dao.Msg{
		Content: content,
		Status:  dao.MsgStatusInit,
	})
}
