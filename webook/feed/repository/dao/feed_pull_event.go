package dao

import (
	"context"

	"gorm.io/gorm"
)

// FeedPullEventDAO 拉模型
type FeedPullEventDAO interface {
	CreatePullEvent(ctx context.Context, event FeedPullEvent) error
	FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error)
}

// FeedPullEvent 拉模型
// 目前我们的业务里面没明显区别
// 在实践中很可能会有区别
type FeedPullEvent struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 发件人
	UID int64 `gorm:"index"`
	// Type 用来标记是什么类型的事件
	// 这边决定了 Content 怎么解读
	Type string
	// 大的 json 串
	Content string
	Ctime   int64 `gorm:"index"`
	// 这个表理论上来说，是没有 Update 操作的
	Utime int64
}

type feedPullEventDAO struct {
	db *gorm.DB
}

func NewFeedPullEventDAO(db *gorm.DB) FeedPullEventDAO {
	return &feedPullEventDAO{
		db: db,
	}
}

func (f *feedPullEventDAO) CreatePullEvent(ctx context.Context, event FeedPullEvent) error {
	return f.db.WithContext(ctx).Create(&event).Error
}

func (f *feedPullEventDAO) FindPullEventList(ctx context.Context, uids []int64, timestamp, limit int64) ([]FeedPullEvent, error) {
	panic("implement me")
}
