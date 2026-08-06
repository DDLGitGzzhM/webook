package ioc

import (
	"context"
	"time"

	"webook/webook/internal/pkg/logger"
	"webook/webook/tag/repository"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
)

// InitRepository 初始化仓储并异步预加载缓存。
func InitRepository(d dao.TagDAO, c cache.TagCache, l logger.Logger) repository.TagRepository {
	repo := repository.NewTagRepository(d, c, l)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		// 也可以同步执行。但是在一些场景下，同步执行会占用很长的时间，所以可以考虑异步执行。
		_ = repo.PreloadUserTags(ctx)
	}()
	return repo
}
