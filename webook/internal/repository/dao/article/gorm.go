package article

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleGormDao struct {
	db *gorm.DB
}

func NewArticleGormDao(db *gorm.DB) *ArticleGormDao {
	return &ArticleGormDao{
		db: db,
	}
}

func (a *ArticleGormDao) GetByAuthor(
	ctx context.Context, author int64, offset, limit int,
) ([]Article, error) {
	var arts []Article
	// SELECT * FROM XXX WHERE XX order by aaa
	// 设计 order by 时要注意让排序字段命中索引
	// author_id => author_id, utime 的联合索引
	err := a.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		// 升序：utime ASC；混合排序：ctime ASC, utime DESC
		Order("utime DESC").
		//Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		//	{Column: clause.Column{Name: "utime"}, Desc: true},
		//	{Column: clause.Column{Name: "ctime"}, Desc: false},
		//}}).
		Find(&arts).Error
	return arts, err
}

func (a *ArticleGormDao) GetPubById(ctx context.Context, id int64) (PublishArticle, error) {
	var pub PublishArticle
	err := a.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (a *ArticleGormDao) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := a.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (a *ArticleGormDao) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		art.Id = id
		// 操作线上库
		return txDao.Upsert(ctx, PublishArticle(art))
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SyncStatus 在同一事务中更新 articles 与 publish_articles 的状态
func (a *ArticleGormDao) SyncStatus(ctx context.Context, uid, id int64, status uint8) error {
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

func (a *ArticleGormDao) Upsert(ctx context.Context, art PublishArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Clauses(clause.OnConflict{
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

func (a *ArticleGormDao) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := a.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (a *ArticleGormDao) UpdateById(ctx context.Context, art Article) error {
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
