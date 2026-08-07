package dao

import "gorm.io/gorm"

type FeedEvent struct {
	Id   int64
	Type string
	// 公共字段。你可以继续加

	// 指向具体的事件的结构体
	Ext any `gorm:"-"`
}

type ArticleEvent struct {
	Id int64
	// FeedEvent 的 id
	Fid int64

	// 个性化数据
	// 文章的 ID
	Aid int64

	// 你的冗余字段
	AuthorID   int64
	AuthorName string
}

type FeedEventDAO struct {
	db *gorm.DB
}

func (f *FeedEventDAO) Find() ([]FeedEvent, error) {
	var res []FeedEvent
	// 查询主表
	f.db.Where("").Find(&res)
	// 在这边分组
	// 比如说查询 article event

	//var aevents []ArticleEvent
	//f.db.Where("fid IN ?", ids).Find(&aevents)
	// 放进去 FeedEvent.Ext 里面

	return res, nil
}
