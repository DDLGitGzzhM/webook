package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"

	intrv1 "webook/webook/api/proto/gen/intr/v1"
	"webook/webook/internal/domain"
	svcmocks "webook/webook/internal/service/mock"
)

type stubIntrClient struct {
	intrv1.InteractiveServiceClient
	getByIds func(ctx context.Context, in *intrv1.GetByIdsRequest,
		opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error)
}

func (s *stubIntrClient) GetByIds(
	ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption,
) (*intrv1.GetByIdsResponse, error) {
	return s.getByIds(ctx, in, opts...)
}

func TestRankingTopN(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) (IArticleService,
			intrv1.InteractiveServiceClient)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "计算成功",
			// 怎么模拟我的数据？
			mock: func(ctrl *gomock.Controller) (IArticleService,
				intrv1.InteractiveServiceClient) {
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
				intrSvc := &stubIntrClient{
					getByIds: func(ctx context.Context, in *intrv1.GetByIdsRequest,
						opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
						return &intrv1.GetByIdsResponse{
							Intrs: map[int64]*intrv1.Interactive{
								1: {BizId: 1, LikeCnt: 1},
								2: {BizId: 2, LikeCnt: 2},
								3: {BizId: 3, LikeCnt: 3},
							},
						}, nil
					},
				}
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
			svc := NewBatchRankingService(artSvc, nil, intrSvc).(*BatchRankingService)
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

type stubRankingRepo struct {
	arts []domain.Article
	err  error
}

func (s *stubRankingRepo) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	s.arts = arts
	return nil
}

func (s *stubRankingRepo) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return s.arts, s.err
}

func TestRankingGetTopN(t *testing.T) {
	now := time.Now()
	want := []domain.Article{
		{Id: 3, Title: "t3", Utime: now, Ctime: now},
		{Id: 2, Title: "t2", Utime: now, Ctime: now},
	}
	repo := &stubRankingRepo{arts: want}
	svc := NewBatchRankingService(nil, repo, nil)

	arts, err := svc.GetTopN(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, want, arts)
}
