package dao

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic/v7"
)

// TagESDAO 基于 ES 的标签 DAO。
type TagESDAO struct {
	client *elastic.Client
}

// NewTagESDAO 创建标签 ES DAO。
func NewTagESDAO(client *elastic.Client) TagDAO {
	return &TagESDAO{client: client}
}

// BizTags 标签文档。
type BizTags struct {
	Uid   int64    `json:"uid"`
	Biz   string   `json:"biz"`
	BizId int64    `json:"biz_id"`
	Tags  []string `json:"tags"`
}

// Search 按用户、业务、标签关键字查询 biz_id。
func (t *TagESDAO) Search(
	ctx context.Context,
	uid int64,
	biz string,
	keywords []string,
) ([]int64, error) {
	query := elastic.NewBoolQuery().Must(
		// 第一个条件，一定有 uid
		elastic.NewTermQuery("uid", uid),
		// 第二个条件，biz 一定符合预期
		elastic.NewTermQuery("biz", biz),
		// 第三个条件，关键字命中了标签
		elastic.NewTermsQueryFromStrings("tags", keywords...),
	)
	resp, err := t.client.Search(TagIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]int64, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele BizTags
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		// 把 biz_id 拿出来了
		res = append(res, ele.BizId)
	}
	return res, nil
}
