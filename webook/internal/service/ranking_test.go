package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"webook/webook/internal/domain"
	svcmocks "webook/webook/internal/service/mock"
)

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (IArticleService,
			InteractiveService)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			// 怎么模拟我的数据？
			mock: func(ctrl *gomock.Controller) (IArticleService, InteractiveService) {
				artSvc := svcmocks.NewMockIArticleService(ctrl)
				// 最简单，一批就搞完
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 3).
					Return([]domain.Article{
						{Id: 1, Utime: now, Ctime: now},
						{Id: 2, Utime: now, Ctime: now},
						{Id: 3, Utime: now, Ctime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 3, 3).
					Return([]domain.Article{}, nil)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				intrSvc.EXPECT().GetByIds(gomock.Any(),
					"article", []int64{1, 2, 3}).
					Return(map[int64]domain.Interactive{
						1: {BizId: 1, LikeCnt: 1},
						2: {BizId: 2, LikeCnt: 2},
						3: {BizId: 3, LikeCnt: 3},
					}, nil)
				intrSvc.EXPECT().GetByIds(gomock.Any(),
					"article", []int64{}).
					Return(map[int64]domain.Interactive{}, nil)
				return artSvc, intrSvc
			},
			wantArts: []domain.Article{
				{Id: 3, Utime: now, Ctime: now},
				{Id: 2, Utime: now, Ctime: now},
				{Id: 1, Utime: now, Ctime: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(artSvc, intrSvc, nil).(*BatchRankingService)
			// 为了测试
			svc.batchSize = 3
			svc.n = 3
			svc.scoreFunc = func(t time.Time, likeCnt int64) float64 {
				return float64(likeCnt)
			}
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
