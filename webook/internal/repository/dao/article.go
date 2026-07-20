package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ArtNotFound = errors.New("文章不存在")
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}

type ArticleGormDao struct {
	db *gorm.DB
}

func NewArticleGormDao(db *gorm.DB) *ArticleGormDao {
	return &ArticleGormDao{
		db: db,
	}
}

func (a ArticleGormDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (a ArticleGormDao) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	res := a.db.WithContext(ctx).Model(&art).
		Where("id=? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"utime":   art.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ArtNotFound
	}
	return nil
}

type Article struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(1024)"`
	Content  string `gorm:"type=BLOB"`
	AuthorId int64  `gorm:"index=aid_ctime"`
	Ctime    int64  `gorm:"index=aid_ctime"`
	Utime    int64
}

/*
设计索引 :
1. Where 条件
2. Select * from articles where author_id = my_id
3. Select * from articles where id = 1
3. sort by createTime

因此考虑 创建 author_id + ctime 的联合索引

// todo

仅在 author_id 创建索引
和 author_id 同 ctime 创建联合索引
*/

// todo 在 articles 准备 一百万条数据 ,auth_id 各不相同，准备 auth_id = 123 的 插入两百条数据
// 执行 select * from articles where author_id = 123 order by ctime desc 比较两种索引的性能
