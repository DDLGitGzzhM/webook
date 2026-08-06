package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGORMTagDAO_GetTagsByBiz(t *testing.T) {
	// 这里你可以通过查看 SQL 来确定自己写的 JOIN 查询对不对
	db, err := gorm.Open(sqlite.Open("gorm.db?mode=memory"), &gorm.Config{
		// 只输出 SQL，不执行查询
		DryRun: true,
	})
	require.NoError(t, err)
	db = db.Debug()
	dao := NewGORMTagDAO(db)
	res, err := dao.GetTagsByBiz(context.Background(), 123, "test", 456)
	if err != nil {
		return
	}
	t.Log(res)
}
