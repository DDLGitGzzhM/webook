package dao

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

// ErrDataNotFound 通用的数据没找到
var ErrDataNotFound = gorm.ErrRecordNotFound

//go:generate mockgen -source=./comment.go -package=daomocks -destination=mocks/comment.mock.go CommentDAO
type CommentDAO interface {
	Insert(ctx context.Context, u Comment) error
	// FindByBiz 只查找一级评论
	FindByBiz(ctx context.Context, biz string,
		bizID, minID, limit int64) ([]Comment, error)
	// FindCommentList Comment的ID为0 获取一级评论，如果不为0获取对应的评论，和其评论的所有回复
	FindCommentList(ctx context.Context, u Comment) ([]Comment, error)
	FindRepliesByPID(ctx context.Context, pID int64, offset, limit int) ([]Comment, error)
	// Delete 删除本节点和其对应的子节点
	Delete(ctx context.Context, u Comment) error
	FindOneByIDs(ctx context.Context, ID []int64) ([]Comment, error)
	FindRepliesByRID(ctx context.Context, rID int64, maxID int64, limit int64) ([]Comment, error)
}

type GORMCommentDAO struct {
	db *gorm.DB
}

func (c *GORMCommentDAO) FindRepliesByRID(ctx context.Context,
	rID int64, maxId int64, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("root_ID = ? AND ID > ?", rID, maxId).
		Order("ID ASC").
		Limit(int(limit)).Find(&res).Error
	return res, err
}

func NewCommentDAO(db *gorm.DB) CommentDAO {
	return &GORMCommentDAO{
		db: db,
	}
}

func (c *GORMCommentDAO) FindOneByIDs(ctx context.Context, IDs []int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		Where("ID in ?", IDs).
		Find(&res).
		Error
	return res, err
}

func (c *GORMCommentDAO) FindByBiz(ctx context.Context, biz string,
	bizID, minID, limit int64) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).
		// 我只要顶级评论
		Where("biz = ? AND biz_ID = ? AND id < ? AND pid IS NULL", biz, bizID, minID).
		Order("id DESC").
		Limit(int(limit)).
		Find(&res).Error
	return res, err
}

// FindRepliesByPID 查找评论的直接评论
func (c *GORMCommentDAO) FindRepliesByPID(ctx context.Context,
	pid int64,
	offset,
	limit int) ([]Comment, error) {
	var res []Comment
	err := c.db.WithContext(ctx).Where("pid = ?", pid).
		Order("ID DESC").
		Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (c *GORMCommentDAO) Insert(ctx context.Context, u Comment) error {
	return c.db.
		WithContext(ctx).
		Create(&u).
		Error
}

func (c *GORMCommentDAO) FindCommentList(ctx context.Context, u Comment) ([]Comment, error) {
	var res []Comment
	builder := c.db.WithContext(ctx)
	if u.ID == 0 {
		builder = builder.
			Where("biz=?", u.Biz).
			Where("biz_ID=?", u.BizID).
			Where("root_ID is null")
	} else {
		builder = builder.Where("root_ID=? or id =?", u.ID, u.ID)
	}
	err := builder.Find(&res).Error
	return res, err

}

func (c *GORMCommentDAO) Delete(ctx context.Context, u Comment) error {
	// 数据库帮你级联删除了，不需要担忧并发问题
	// 假如 4 已经删了，按照外键的约束，如果你插入一个 pid=4 的行，你是插不进去的
	return c.db.WithContext(ctx).Delete(&Comment{
		ID: u.ID,
	}).Error
}

// Comment 总结：所有的索引设计，都是针对 WHERE，ORDER BY，SELECT xxx 来进行的
// 如果有 JOIN，那么还要考虑 ON
// 永远考虑最频繁的查询
// 在没有遇到更新、查询性能瓶颈之前，不需要太过于担忧维护索引的开销
// 有一些时候，随着业务发展，有一些索引用不上了，要及时删除
type Comment struct {
	// 代表你评论本体
	ID int64
	// 发表评论的人
	// 要不要在这个列创建索引？
	// 取决于有没有 WHERE uID = ? 的查询
	Uid int64
	// 这个代表的是你评论的对象是什么？
	// 比如说代表某个帖子，代表某个视频，代表某个图片
	Biz   string `gorm:"index:biz_type_id"`
	BizID int64  `gorm:"index:biz_type_id"`

	// 用 NULL 来表达没有父亲
	// 你可以考虑用 -1 来代表没有父亲
	// 索引是如何处理 NULL 的？？？
	// NULL 的取值非常多

	PID sql.NullInt64 `gorm:"index"`
	// 外键指向的也是同一张表
	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`

	// 引入 RootID 这个设计
	// 顶级评论的 ID
	// 主要是为了加载整棵评论的回复组成树
	RootID sql.NullInt64 `gorm:"index:root_ID_ctime"`
	Ctime  int64         `gorm:"index:root_ID_ctime"`

	// 评论的内容
	Content string

	Utime int64
}

func (*Comment) TableName() string {
	return "comments"
}
