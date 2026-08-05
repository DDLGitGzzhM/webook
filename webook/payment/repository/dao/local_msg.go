package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Msg 本地消息表
type Msg struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 如果你要做一个通用的，适用于任何业务的，那么你在这里可以提供更加多的字段。
	Content string
	Status  uint8
	Ctime   int64
	Utime   int64 `gorm:"index"`
}

const (
	MsgStatusUnknown = 0
	MsgStatusInit    = 1
	MsgStatusSuccess = 2
	MsgStatusFailed  = 3
)

var _ LocalMsgDAO = (*LocalMsgGORMDAO)(nil)

// LocalMsgDAO 本地消息表 DAO
type LocalMsgDAO interface {
	AddMsg(ctx context.Context, msg Msg) (int64, error)
	UpdateStatus(ctx context.Context, id int64, status uint8) error
	FindInitMsg(ctx context.Context, offset int, limit int) ([]Msg, error)
}

// LocalMsgGORMDAO 基于 GORM 的本地消息表 DAO
type LocalMsgGORMDAO struct {
	db *gorm.DB
}

func (dao *LocalMsgGORMDAO) UpdateStatus(ctx context.Context, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Msg{}).Where("id = ?", id).Updates(map[string]any{
		"status": status,
		"utime":  now,
	}).Error
}

// NewLocalMsgGORMDAO 创建本地消息表 DAO
func NewLocalMsgGORMDAO(db *gorm.DB) *LocalMsgGORMDAO {
	return &LocalMsgGORMDAO{db: db}
}

func (dao *LocalMsgGORMDAO) AddMsg(ctx context.Context, msg Msg) (int64, error) {
	now := time.Now().UnixMilli()
	msg.Ctime = now
	msg.Utime = now
	err := dao.db.WithContext(ctx).Create(&msg).Error
	return msg.Id, err
}

func (dao *LocalMsgGORMDAO) FindInitMsg(ctx context.Context, offset int, limit int) ([]Msg, error) {
	var msgs []Msg
	err := dao.db.WithContext(ctx).
		Where("status = ?", MsgStatusInit).
		Offset(offset).Limit(limit).
		Find(&msgs).Error
	return msgs, err
}
