package repository

import (
	"context"

	"github.com/ecodeclub/ekit/slice"

	"webook/webook/search/domain"
	"webook/webook/search/repository/dao"
)

type articleRepository struct {
	dao  dao.ArticleDAO
	tags dao.TagDAO
}

// NewArticleRepository 创建文章仓储。
func NewArticleRepository(d dao.ArticleDAO, td dao.TagDAO) ArticleRepository {
	return &articleRepository{
		dao:  d,
		tags: td,
	}
}

func (a *articleRepository) SearchArticle(
	ctx context.Context,
	uid int64,
	keywords []string,
) ([]domain.Article, error) {
	// 标签命中了的
	ids, err := a.tags.Search(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	// 加一个 bizids 的输入，这个 bizid 是标签含有关键字的 biz_id
	arts, err := a.dao.Search(ctx, ids, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Status:  src.Status,
			Content: src.Content,
			Tags:    src.Tags,
		}
	}), nil
}

func (a *articleRepository) InputArticle(ctx context.Context, msg domain.Article) error {
	return a.dao.InputArticle(ctx, dao.Article{
		Id:      msg.Id,
		Title:   msg.Title,
		Status:  msg.Status,
		Content: msg.Content,
	})
}
