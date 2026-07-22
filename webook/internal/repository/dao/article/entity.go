package article

type Article struct {
	Id int64 `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	// 标题的长度
	// 正常都不会超过这个长度
	Title   string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content string `gorm:"type=BLOB" bson:"content,omitempty"`
	// 作者
	AuthorId int64 `gorm:"index=aid_ctime" bson:"author_id,omitempty"`
	Status   uint8 `bson:"status,omitempty"`
	Ctime    int64 `gorm:"index=aid_ctime" bson:"ctime,omitempty"`
	Utime    int64 `bson:"utime,omitempty"`
}

// PublishArticle 线上库，衍生类型
type PublishArticle Article
