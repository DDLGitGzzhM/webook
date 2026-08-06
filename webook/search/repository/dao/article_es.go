package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ecodeclub/ekit/slice"
	"github.com/olivere/elastic/v7"
)

const ArticleIndexName = "article_index"
const TagIndexName = "tags_index"

// Article ES 文章文档。
type Article struct {
	Id      int64    `json:"id"`
	Title   string   `json:"title"`
	Status  int32    `json:"status"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// ArticleElasticDAO 基于 ES 的文章 DAO。
type ArticleElasticDAO struct {
	client *elastic.Client
}

// NewArticleElasticDAO 创建文章 ES DAO。
func NewArticleElasticDAO(client *elastic.Client) ArticleDAO {
	return &ArticleElasticDAO{client: client}
}

// Search 按标签命中 ID、标题、内容搜索已发布文章。
func (h *ArticleElasticDAO) Search(
	ctx context.Context,
	tagArtIds []int64,
	keywords []string,
) ([]Article, error) {
	queryString := strings.Join(keywords, " ")
	// 标签命中
	tagArtIdAnys := slice.Map(tagArtIds, func(idx int, src int64) any {
		return src
	})
	title := elastic.NewMatchQuery("title", queryString)
	content := elastic.NewMatchQuery("content", queryString)
	or := elastic.NewBoolQuery().Should(title, content)
	if len(tagArtIds) > 0 {
		tag := elastic.NewTermsQuery("id", tagArtIdAnys...).Boost(2.0)
		or = or.Should(tag)
	}
	query := elastic.NewBoolQuery().Must(
		or,
		elastic.NewTermQuery("status", 2),
	)
	resp, err := h.client.Search(ArticleIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]Article, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele Article
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele)
	}
	return res, nil
}

// InputArticle 写入文章文档。
func (h *ArticleElasticDAO) InputArticle(ctx context.Context, art Article) error {
	_, err := h.client.Index().
		Index(ArticleIndexName).
		Id(strconv.FormatInt(art.Id, 10)).
		BodyJson(art).Do(ctx)
	return err
}
