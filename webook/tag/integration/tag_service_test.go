package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	tagv1 "webook/webook/api/proto/gen/tag/v1"
	evtmocks "webook/webook/tag/events/mocks"
	"webook/webook/tag/integration/startup"
	"webook/webook/tag/repository"
	"webook/webook/tag/repository/cache"
	"webook/webook/tag/repository/dao"
)

type TagServiceTestSuite struct {
	suite.Suite
	db  *gorm.DB
	rdb redis.Cmdable
}

func (s *TagServiceTestSuite) SetupSuite() {
	s.db = startup.InitTestDB()
	s.rdb = startup.InitRedis()
}

func (s *TagServiceTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `tag_bizs`").Error
	require.NoError(s.T(), err)
	// 在有外键约束的情况下，不能用 TRUNCATE
	err = s.db.Exec("DELETE FROM `tags`").Error
	require.NoError(s.T(), err)
}

func TestTagService(t *testing.T) {
	suite.Run(t, new(TagServiceTestSuite))
}

func (s *TagServiceTestSuite) TestPreload() {
	data := make([]dao.Tag, 0, 200)
	for i := 0; i < 200; i++ {
		data = append(data, dao.Tag{
			Id:   int64(i + 1),
			Name: fmt.Sprintf("tag_%d", i),
			Uid:  int64(i+1) % 3,
		})
	}
	err := s.db.Create(&data).Error
	require.NoError(s.T(), err)
	d := dao.NewGORMTagDAO(s.db)
	c := cache.NewRedisTagCache(s.rdb)
	l := startup.InitLog()
	repo := repository.NewTagRepository(d, c, l)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	err = repo.PreloadUserTags(ctx)
	require.NoError(s.T(), err)
}

func (s *TagServiceTestSuite) TestFullFlow() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var uid int64 = 123
	var bizId int64 = 456
	ctrl := gomock.NewController(s.T())
	p := evtmocks.NewMockProducer(ctrl)
	p.EXPECT().ProduceSyncEvent(gomock.Any(), gomock.Any()).
		AnyTimes().Return(nil)
	svc := startup.InitGRPCService(p)
	resp, err := svc.CreateTag(ctx, &tagv1.CreateTagRequest{
		Uid:  123,
		Name: "tag1",
	})
	require.NoError(s.T(), err)
	tid0 := resp.Tag.Id
	_, err = svc.AttachTags(ctx, &tagv1.AttachTagsRequest{
		Tids:  []int64{tid0},
		Uid:   uid,
		Biz:   "test",
		BizId: bizId,
	})
	require.NoError(s.T(), err)
	tagsResp, err := svc.GetBizTags(ctx, &tagv1.GetBizTagsRequest{
		Uid:   123,
		Biz:   "test",
		BizId: bizId,
	})
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(tagsResp.Tags))
	assert.Equal(s.T(), &tagv1.Tag{
		Id:   tid0,
		Name: "tag1",
		Uid:  uid,
	}, tagsResp.Tags[0])
}
