package dao

import (
	"context"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
)

// Tag 标签表。
// ID=1 => uid = 123
// ID = 1 => uid =234
type Tag struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 我要不要在这里创建一个唯一索引<uid, name>
	Name string `gorm:"type=varchar(4096)"`
	// 要在 uid 上创建一个索引
	// 因为你有一个典型的根据 uid 来查询的场景
	Uid   int64 `gorm:"index"`
	Ctime int64
	Utime int64
}

// TagBiz 某个人对某个资源打了标签。
type TagBiz struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	BizId int64  `gorm:"index:biz_type_id"`
	Biz   string `gorm:"index:biz_type_id"`
	// 冗余字段，加快查询和删除
	// 这个字段可以删除的
	Uid int64 `gorm:"index"`
	//TagName string
	Tid   int64
	Tag   *Tag  `gorm:"ForeignKey:Tid;AssociationForeignKey:Id;constraint:OnDelete:CASCADE"`
	Ctime int64 `bson:"ctime,omitempty"`
	Utime int64 `bson:"utime,omitempty"`
}

// TagDAO 标签数据访问接口。
type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, ids []int64) ([]Tag, error)
}

// GORMTagDAO 基于 GORM 的标签 DAO。
type GORMTagDAO struct {
	db *gorm.DB
}

// NewGORMTagDAO 创建 GORM 标签 DAO。
func NewGORMTagDAO(db *gorm.DB) TagDAO {
	return &GORMTagDAO{
		db: db,
	}
}

// GetTagsById 按 ID 批量查询标签。
func (dao *GORMTagDAO) GetTagsById(ctx context.Context, ids []int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("id IN ?", ids).Find(&res).Error
	return res, err
}

// CreateTag 创建标签。
func (dao *GORMTagDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.Ctime = now
	tag.Utime = now
	err := dao.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

// CreateTagBiz 覆盖式绑定标签到业务资源。
func (dao *GORMTagDAO) CreateTagBiz(ctx context.Context, tagBiz []TagBiz) error {
	if len(tagBiz) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	for i := range tagBiz {
		tagBiz[i].Ctime = now
		tagBiz[i].Utime = now
	}
	first := tagBiz[0]
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 完成了覆盖式的操作
		// 如果 tag_biz 里面没有 uid 字段。你的删除就很麻烦
		// delete from tag_biz where tid IN
		// (select distinct id from tag where uid = ?) AND biz = ? AND biz_id = ?
		err := tx.Model(&TagBiz{}).
			Where("uid = ? AND biz = ? AND biz_id = ?", first.Uid, first.Biz, first.BizId).
			Delete(&TagBiz{}).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBiz).Error
	})
}

// GetTagsByUid 按用户查询标签。
func (dao *GORMTagDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Where("uid = ?", uid).Find(&res).Error
	return res, err
}

// GetTagsByBiz 查询用户给某资源打的标签。
func (dao *GORMTagDAO) GetTagsByBiz(
	ctx context.Context,
	uid int64,
	biz string,
	bizId int64,
) ([]Tag, error) {
	// 这边使用 JOIN 查询，如果你不想使用 JOIN 查询，
	// 你就在 repository 里面分成两次查询
	var res []TagBiz
	err := dao.db.WithContext(ctx).Model(&TagBiz{}).
		InnerJoins("Tag", dao.db.Model(&Tag{})).
		Where("Tag.uid = ? AND biz = ? AND biz_id = ?", uid, biz, bizId).
		Find(&res).Error
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src TagBiz) Tag {
		return *src.Tag
	}), nil
}

// GetTags 分页查询全部标签。
func (dao *GORMTagDAO) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := dao.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}
