package startup

import (
	"webook/webook/internal/pkg/logger"
	"webook/webook/tag/repository"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
)

// InitRepository 创建测试仓储。
func InitRepository(d dao.TagDAO, c cache.TagCache, l logger.Logger) repository.TagRepository {
	return repository.NewTagRepository(d, c, l)
}
