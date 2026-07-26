package ioc

import (
	"context"
	"time"

	"webook/webook/internal/domain"
	"webook/webook/internal/job"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/service"
)

func InitScheduler(l logger.Logger,
	local *job.LocalFuncExecutor,
	svc service.JobService) *job.Scheduler {
	res := job.NewScheduler(svc, l)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor(svc service.RankingService) *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	// 要在数据库里面插入一条记录。
	// ranking job 的记录，通过管理任务接口来插入
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}
