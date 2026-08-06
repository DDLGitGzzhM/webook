package dao

import (
	"context"

	"github.com/olivere/elastic/v7"
)

// AnyESDAO 任意索引写入。
type AnyESDAO struct {
	client *elastic.Client
}

// NewAnyESDAO 创建任意索引 DAO。
func NewAnyESDAO(client *elastic.Client) AnyDAO {
	return &AnyESDAO{client: client}
}

// Input 向指定索引写入文档。
func (a *AnyESDAO) Input(ctx context.Context, index, docId, data string) error {
	_, err := a.client.Index().
		Index(index).Id(docId).BodyString(data).Do(ctx)
	return err
}
