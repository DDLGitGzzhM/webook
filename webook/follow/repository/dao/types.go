package dao

import "context"

// FollowRelation 这个是类似于点赞的表设计
// 取消关注，不是真的删除了数据，而是更新了状态
type FollowRelation struct {
	ID int64 `gorm:"primaryKey,autoIncrement,column:id"`

	// 在这里建一个联合唯一索引
	// 注意列的顺序
	// 考虑到我们的典型场景是，我关注了多少人 where follower = ? (传入 uid = 123)
	// 所以应该是 <follower, followee>

	// 如果我的典型场景是，我有多少粉丝 WHERE followee = ? （传入 uid = 123)
	// 这种情况下 <followee, follower> 在后
	Follower int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`
	Followee int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`

	// 对应于关注来说，就是插入或者将这个状态更新为可用状态
	// 对于取消关注来说，就是将这个状态更新为不可用状态
	Status uint8

	// 这里你可以根据自己的业务来增加字段，比如说
	// 关系类型，可以搞些什么普通关注，特殊关注
	// Type int64 `gorm:"column:type;type:int(11);comment:关注类型 0-普通关注"`
	// 备注
	// Note string `gorm:"column:remark;type:varchar(255);"`
	// 创建时间
	Ctime int64
	Utime int64
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

// FollowRelationDao 关注关系 DAO
type FollowRelationDao interface {
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	// FollowerRelationList 获取某人的粉丝列表
	FollowerRelationList(ctx context.Context, followee, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}

// UserRelation 另外一种设计方案，但是不要这么做
type UserRelation struct {
	ID     int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid1   int64 `gorm:"column:uid1;type:int(11);not null;uniqueIndex:user_contact_index"`
	Uid2   int64 `gorm:"column:uid2;type:int(11);not null;uniqueIndex:user_contact_index"`
	Block  bool  // 拉黑
	Mute   bool  // 屏蔽
	Follow bool  // 关注
}

// FollowStatics 关注统计表
type FollowStatics struct {
	ID  int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid int64 `gorm:"unique"`
	// 有多少粉丝
	Followers int64
	// 关注了多少人
	Followees int64

	Utime int64
	Ctime int64
}
