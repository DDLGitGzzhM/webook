package article

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ArtNotFound = errors.New("文章不存在")
)

type ArticleDao interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	// SyncStatus 同库不同表，事务同步更新制作库与线上库状态
	SyncStatus(ctx context.Context, uid, id int64, status uint8) error
	Upsert(ctx context.Context, art PublishArticle) error
}

type ArticleGormDao struct {
	db *gorm.DB
}

func (a ArticleGormDao) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.Transaction(func(tx *gorm.DB) error {
		var err error
		txDao := NewArticleGormDao(tx)
		if id > 0 {
			err = txDao.UpdateById(ctx, art)
		} else {
			id, err = txDao.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 操作线上库
		return txDao.Upsert(ctx, PublishArticle{art})

	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NewArticleGormDao(db *gorm.DB) *ArticleGormDao {
	return &ArticleGormDao{
		db: db,
	}
}

// SyncStatus 在同一事务中更新 articles 与 publish_articles 的状态
func (a ArticleGormDao) SyncStatus(ctx context.Context, uid, id int64, status uint8) error {
	now := time.Now().UnixMilli()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).
			Where("id = ? AND author_id = ?", id, uid).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ArtNotFound
		}
		res = tx.Model(&PublishArticle{}).
			Where("id = ? AND author_id = ?", id, uid).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		return nil
	})
}

func (a ArticleGormDao) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now()
	art.Ctime = now.UnixMilli()
	art.Utime = now.UnixMilli()
	err := a.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		}),
	}).Create(&art).Error
	// insert xxx on duplicate key update xxx
	return err
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
			"status":  art.Status,
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
	Status   uint8
	Ctime    int64 `gorm:"index=aid_ctime"`
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
