package domain

// Article 搜索侧文章领域模型。
type Article struct {
	Id      int64
	Title   string
	Status  int32
	Content string
	Tags    []string
}
