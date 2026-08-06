package dao

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/olivere/elastic/v7"
)

const UserIndexName = "user_index"

// User ES 用户文档。
type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

// UserElasticDAO 基于 ES 的用户 DAO。
type UserElasticDAO struct {
	client *elastic.Client
}

// NewUserElasticDAO 创建用户 ES DAO。
func NewUserElasticDAO(client *elastic.Client) UserDAO {
	return &UserElasticDAO{client: client}
}

// Search 按昵称搜索用户。
func (h *UserElasticDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	queryString := strings.Join(keywords, " ")
	query := elastic.NewBoolQuery().Must(
		elastic.NewMatchQuery("nickname", queryString),
	)
	resp, err := h.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele User
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele)
	}
	return res, nil
}

// InputUser 写入用户文档。
func (h *UserElasticDAO) InputUser(ctx context.Context, user User) error {
	_, err := h.client.Index().
		Index(UserIndexName).
		Id(strconv.FormatInt(user.Id, 10)).
		BodyJson(user).Do(ctx)
	return err
}
