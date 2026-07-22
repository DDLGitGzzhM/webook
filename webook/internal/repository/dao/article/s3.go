package article

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"webook/webook/internal/domain"
)

var statusPrivate = domain.ArticleStatusPrivate.ToUint8()

type S3DAO struct {
	oss *s3.S3
	// 通过组合 ArticleGormDao 来简化操作
	// 当然在实践中，你是不太会有组合的机会
	// 你操作制作库总是一样的
	// 你就是操作线上库的时候不一样
	ArticleGormDao
	bucket *string
}

// NewOssDAO 因为组合 ArticleGormDao 是一个内部实现细节
// 所以这里要直接传入 DB
func NewOssDAO(oss *s3.S3, db *gorm.DB) ArticleDao {
	return &S3DAO{
		oss: oss,
		// 你也可以考虑利用依赖注入来传入。
		// 但是事实上这个很少变，所以你可以延迟到必要的时候再注入
		bucket: aws.String("webook-1314583317"),
		ArticleGormDao: ArticleGormDao{
			db: db,
		},
	}
}

func (o *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 保存制作库
	// 保存线上库，并且把 content 上传到 OSS
	var (
		id = art.Id
	)
	// 制作库流量不大，并发不高，你就保存到数据库就可以
	// 当然，有钱或者体量大，就还是考虑 OSS
	err := o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		now := time.Now().UnixMilli()
		// 制作库
		txDAO := NewArticleGormDao(tx)
		if id == 0 {
			id, err = txDAO.Insert(ctx, art)
		} else {
			err = txDAO.UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id
		publishArt := PublishArticleV1{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Status:   art.Status,
			Ctime:    now,
			Utime:    now,
		}
		// 线上库不保存 Content,要准备上传到 OSS 里面
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  art.Title,
				"utime":  now,
				"status": art.Status,
			}),
		}).Create(&publishArt).Error
	})
	// 说明保存到数据库的时候失败了
	if err != nil {
		return 0, err
	}
	// 接下来就是保存到 OSS 里面
	// 你要有监控，你要有重试，你要有补偿机制
	_, err = o.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      o.bucket,
		Key:         aws.String(strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: aws.String("text/plain;charset=utf-8"),
	})
	return id, err
}

func (o *S3DAO) SyncStatus(ctx context.Context, uid, id int64, status uint8) error {
	panic("implement me")
}
