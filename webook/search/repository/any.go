package repository

import (
	"context"

	"webook/webook/search/repository/dao"
)

// AnyRepository 任意索引写入仓储。
type AnyRepository interface {
	Input(ctx context.Context, index string, docID string, data string) error
}

type anyRepository struct {
	dao dao.AnyDAO
}

// NewAnyRepository 创建任意索引仓储。
func NewAnyRepository(d dao.AnyDAO) AnyRepository {
	return &anyRepository{dao: d}
}

func (repo *anyRepository) Input(
	ctx context.Context,
	index string,
	docID string,
	data string,
) error {
	return repo.dao.Input(ctx, index, docID, data)
}
