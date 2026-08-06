package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"

	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

type RankingService interface {
	// TopN 计算热榜并写入缓存
	TopN(ctx context.Context) error
	// GetTopN 从缓存读取热榜
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	artSvc    IArticleService
	intrSvc   intrv1.InteractiveServiceClient
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64

	// 负载
	load int64
}

func NewBatchRankingService(artSvc IArticleService,
	repo repository.RankingRepository,
	intrSvc intrv1.InteractiveServiceClient) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		repo:      repo,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(float64(sec+2), 1.5)
		},
	}
}

// TopN 计算热榜并写入缓存。
func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	return svc.repo.ReplaceTopN(ctx, arts)
}

// GetTopN 从缓存读取热榜。
func (svc *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return svc.repo.GetTopN(ctx)
}

// topN 分批计算热榜。
func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// 我只取七天内的数据
	now := time.Now()
	// 先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}
	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			} else {
				return -1
			}
		})

	for {
		// 这里拿了一批
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}
		if len(arts) == 0 {
			break
		}
		ids := slice.Map[domain.Article, int64](arts,
			func(idx int, src domain.Article) int64 {
				return src.Id
			})
		// 要去找到对应的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			Biz: "article", Ids: ids,
		})
		if err != nil {
			return nil, err
		}
		if len(intrs.Intrs) == 0 {
			return nil, errors.New("没有数据")
		}
		// 合并计算 score
		// 排序
		for _, art := range arts {
			intr := intrs.Intrs[art.Id]
			//if !ok {
			//	// 你都没有，肯定不可能是热榜
			//	continue
			//}
			score := svc.scoreFunc(art.Utime, intr.GetLikeCnt())
			// 我要考虑，我这个 score 在不在前一百名
			// 拿到热度最低的
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			// 这种写法，要求 topN 已经满了
			if err == queue.ErrOutOfCapacity {
				val, _ := topN.Dequeue()
				if val.score < score {
					_ = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}

		// 一批已经处理完了，问题来了，我要不要进入下一批？我怎么知道还有没有？
		if len(arts) < svc.batchSize ||
			now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			// 我这一批都没取够，我当然可以肯定没有下一批了
			// 又或者已经取到了七天之前的数据了，说明可以中断了
			break
		}
		// 这边要更新 offset
		offset = offset + len(arts)
	}
	// 最后得出结果；队列按分数升序出队，逆序写入即为从高到低
	res := make([]domain.Article, svc.n)
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 取完了，不够 n，裁掉前面的空槽
			return res[i+1:], nil
		}
		res[i] = val.art
	}
	return res, nil
}
