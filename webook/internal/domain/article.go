package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	// 摘要取前几句，需考虑中文按 rune 截取
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
}

type ArticleStatus uint8

const (
	ArticleStatusUnknow ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusUnpublished:
		return "未发布"
	case ArticleStatusPublished:
		return "已发布"
	case ArticleStatusPrivate:
		return "私密"
	default:
		return "未知"
	}

}

type Author struct {
	Id   int64
	Name string
}
