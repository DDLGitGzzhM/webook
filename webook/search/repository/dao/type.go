package dao

import "context"

// UserDAO 用户索引读写。
type UserDAO interface {
	InputUser(ctx context.Context, user User) error
	Search(ctx context.Context, keywords []string) ([]User, error)
}

// ArticleDAO 文章索引读写。
type ArticleDAO interface {
	InputArticle(ctx context.Context, article Article) error
	Search(ctx context.Context, tagArtIds []int64, keywords []string) ([]Article, error)
}

// TagDAO 标签索引查询。
type TagDAO interface {
	Search(ctx context.Context, uid int64, biz string, keywords []string) ([]int64, error)
}

// AnyDAO 任意索引写入。
type AnyDAO interface {
	Input(ctx context.Context, index, docID, data string) error
}
