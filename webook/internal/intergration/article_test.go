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

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/web/jwt"
	"webook/webook/ioc"
	"webook/webook/startup"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.UserClaims{
			UserId: 123,
		})
	})
	artHdl := startup.InitArticleHandler()
	s.db = ioc.InitDB(logger.NewZapLogger(ioc.InitLogger()))
	artHdl.RegisterRoutes(s.server)
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("测试")
}

func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("truncate table articles")
}

func (s *ArticleTestSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)

		art      Article
		wantCode int // http resp code
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子--保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", "1").First(&art).Error
				require.Nil(t, err)
				art.Ctime = 0
				art.Utime = 0
				require.Equal(t, art, dao.Article{
					Id:       1,
					Title:    "测试帖子",
					Content:  "测试内容",
					AuthorId: 123,
				})
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
				s.db.Create(&dao.Article{
					Id:       2,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    100,
					Utime:    100,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", "2").First(&art).Error
				require.Nil(t, err)
				require.True(t, art.Utime > 100)
				art.Utime = 0
				require.Equal(t, art, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    100,
				})
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
				s.db.Create(&dao.Article{
					Id:       3,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 456,
					Ctime:    100,
					Utime:    100,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id=?", "3").First(&art).Error
				require.Nil(t, err)
				require.Equal(t, art, dao.Article{
					Id:       3,
					Title:    "修改帖子",
					Content:  "我的内容",
					AuthorId: 456,
					Ctime:    100,
					Utime:    100,
				})
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
			// 构造请求
			// 执行
			// 验证

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

func TestArticle(t *testing.T) {
	viper.SetConfigFile("../../config/config/dev.yaml")
	require.NoError(t, viper.ReadInConfig())
	suite.Run(t, new(ArticleTestSuite))
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}
