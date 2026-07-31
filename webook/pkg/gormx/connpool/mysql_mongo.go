package connpool

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type MySQL2Mongo struct {
	db      gorm.ConnPool
	mdb     *mongo.Database
	pattern *atomicx.Value[string]
}

//func (d *MySQL2Mongo) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
//	switch d.pattern.Load() {
//	case patternSrcOnly, patternSrcFirst:
//		return d.db.QueryContext(ctx, query, args...)
//	case patternDstOnly, patternDstFirst:
//		return d.mdb.Collection("xxx").FindOne()
//	default:
//		panic("未知的双写模式")
//		//return nil, errors.New("未知的双写模式")
//	}
//}
