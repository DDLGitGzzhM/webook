package intergration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/web/jwt"
	"webook/webook/ioc"
	"webook/webook/startup"
)

// ArticleGORMTestSuite GORM 文章集成测试套件
type ArticleGORMTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleGORMTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.UserClaims{
			UserId: 123,
		})
	})
	s.db = ioc.InitDB(logger.NewZapLogger(ioc.InitLogger()))
	artHdl := startup.InitArticleHandler(article.NewArticleGormDao(s.db))
	artHdl.RegisterRoutes(s.server)
}

func (s *ArticleGORMTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE `articles`")
	s.db.Exec("TRUNCATE TABLE `publish_articles`")
}

func (s *ArticleGORMTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子--保存成功",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id=?", "1").First(&art).Error
				require.Nil(t, err)
				require.True(t, art.Ctime > 0)
				require.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				require.Equal(t, article.Article{
					Id:       1,
					Title:    "测试帖子",
					Content:  "测试内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Title:   "测试帖子",
				Content: "测试内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "ok",
				Data: 1,
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func(t *testing.T) {
				s.db.Create(&article.Article{
					Id:       2,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				})
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id=?", "2").First(&art).Error
				require.Nil(t, err)
				require.True(t, art.Utime > 100)
				art.Utime = 0
				require.Equal(t, article.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Ctime:    100,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "ok",
				Data: 2,
			},
		},
		{
			name: "修改别人的文章",
			before: func(t *testing.T) {
				s.db.Create(&article.Article{
					Id:       3,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				})
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id=?", "3").First(&art).Error
				require.Nil(t, err)
				require.Equal(t, article.Article{
					Id:       3,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "保存失败",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit",
				bytes.NewReader(reqBody))

			require.Nil(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			require.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var got Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.Nil(t, err)
			require.Equal(t, tc.wantRes, got)
			tc.after(t)
		})
	}
}

func (s *ArticleGORMTestSuite) TestPublish() {
	t := s.T()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)
		art    Article

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				require.NoError(t, err)
				require.Equal(t, "hello，你好", art.Title)
				require.Equal(t, "随便试试", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				require.True(t, art.Ctime > 0)
				require.True(t, art.Utime > 0)

				var publishedArt article.PublishArticle
				err = s.db.Where("author_id = ?", 123).First(&publishedArt).Error
				require.NoError(t, err)
				require.Equal(t, "hello，你好", publishedArt.Title)
				require.Equal(t, "随便试试", publishedArt.Content)
				require.Equal(t, int64(123), publishedArt.AuthorId)
				require.Equal(t, domain.ArticleStatusPublished.ToUint8(), publishedArt.Status)
				require.True(t, publishedArt.Ctime > 0)
				require.True(t, publishedArt.Utime > 0)
			},
			art: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "ok",
				Data: 1,
			},
		},
		{
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				s.db.Create(&article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				})
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				require.NoError(t, err)
				require.Equal(t, "新的标题", art.Title)
				require.Equal(t, "新的内容", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, int64(456), art.Ctime)
				require.True(t, art.Utime > 234)
				require.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)

				var publishedArt article.PublishArticle
				err = s.db.Where("id = ?", 2).First(&publishedArt).Error
				require.NoError(t, err)
				require.Equal(t, "新的标题", publishedArt.Title)
				require.Equal(t, "新的内容", publishedArt.Content)
				require.Equal(t, int64(123), publishedArt.AuthorId)
				require.True(t, publishedArt.Ctime > 0)
				require.True(t, publishedArt.Utime > 0)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "ok",
				Data: 2,
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				}
				s.db.Create(&art)
				part := article.PublishArticle(art)
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				require.NoError(t, err)
				require.Equal(t, "新的标题", art.Title)
				require.Equal(t, "新的内容", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, int64(456), art.Ctime)
				require.True(t, art.Utime > 234)

				var part article.PublishArticle
				err = s.db.Where("id = ?", 3).First(&part).Error
				require.NoError(t, err)
				require.Equal(t, "新的标题", part.Title)
				require.Equal(t, "新的内容", part.Content)
				require.Equal(t, int64(123), part.AuthorId)
				require.Equal(t, int64(456), part.Ctime)
				require.True(t, part.Utime > 234)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 0,
				Msg:  "ok",
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := article.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
				}
				s.db.Create(&art)
				part := article.PublishArticle(art)
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 4).First(&art).Error
				require.NoError(t, err)
				require.Equal(t, "我的标题", art.Title)
				require.Equal(t, "我的内容", art.Content)
				require.Equal(t, int64(456), art.Ctime)
				require.Equal(t, int64(234), art.Utime)
				require.Equal(t, int64(789), art.AuthorId)

				var part article.PublishArticle
				err = s.db.Where("id = ?", 4).First(&part).Error
				require.NoError(t, err)
				require.Equal(t, "我的标题", part.Title)
				require.Equal(t, "我的内容", part.Content)
				require.Equal(t, int64(789), part.AuthorId)
				require.Equal(t, int64(456), part.Ctime)
				require.Equal(t, int64(234), part.Utime)
			},
			art: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "保存失败",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish",
				bytes.NewReader(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			s.server.ServeHTTP(resp, req)
			require.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != http.StatusOK {
				return
			}
			var got Result[int64]
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.NoError(t, err)
			require.Equal(t, tc.wantRes, got)
			tc.after(t)
		})
	}
}

func TestGORMArticle(t *testing.T) {
	viper.SetConfigFile("../../config/config/dev.yaml")
	require.NoError(t, viper.ReadInConfig())
	suite.Run(t, new(ArticleGORMTestSuite))
}
