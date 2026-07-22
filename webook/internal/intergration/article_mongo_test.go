//go:build e2e

package intergration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"webook/webook/internal/domain"
	"webook/webook/internal/repository/dao/article"
	"webook/webook/internal/web/jwt"
	"webook/webook/startup"
)

type ArticleMongoTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (s *ArticleMongoTestSuite) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(c *gin.Context) {
		c.Set("claims", &jwt.UserClaims{
			UserId: 123,
		})
		c.Next()
	})
	s.mdb = startup.InitMongoDB()
	node, err := snowflake.NewNode(1)
	require.NoError(s.T(), err)
	err = article.InitCollections(s.mdb)
	if err != nil {
		panic(err)
	}
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	hdl := startup.InitArticleHandler(article.NewMongoDBDAO(s.mdb, node))
	hdl.RegisterRoutes(s.server)
}

func (s *ArticleMongoTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.mdb.Collection("articles").DeleteMany(ctx, bson.D{})
	require.NoError(s.T(), err)
	_, err = s.mdb.Collection("published_articles").DeleteMany(ctx, bson.D{})
	require.NoError(s.T(), err)
}

func (s *ArticleMongoTestSuite) TestEdit() {
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
			name: "新建帖子",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "author_id", Value: 123}}).
					Decode(&art)
				require.NoError(t, err)
				require.True(t, art.Ctime > 0)
				require.True(t, art.Utime > 0)
				require.True(t, art.Id > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				require.Equal(t, article.Article{
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
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
			name: "更新帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				require.NoError(t, err)
				require.True(t, art.Utime > 234)
				art.Utime = 0
				require.Equal(t, article.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
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
			name: "更新别人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				require.NoError(t, err)
				require.Equal(t, article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
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
			data, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewReader(data))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			require.NoError(t, err)
			require.Equal(t, tc.wantRes.Code, result.Code)
			require.Equal(t, tc.wantRes.Msg, result.Msg)
			// 雪花算法生成的 ID 无法确定具体值
			if tc.wantRes.Data > 0 {
				require.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func (s *ArticleMongoTestSuite) TestPublish() {
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&art)
				require.NoError(t, err)
				require.True(t, art.Id > 0)
				require.Equal(t, "hello，你好", art.Title)
				require.Equal(t, "随便试试", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				require.True(t, art.Ctime > 0)
				require.True(t, art.Utime > 0)

				var publishedArt article.PublishArticle
				err = s.liveCol.FindOne(ctx,
					bson.D{bson.E{Key: "author_id", Value: 123}}).Decode(&publishedArt)
				require.NoError(t, err)
				require.True(t, publishedArt.Id > 0)
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &article.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).Decode(&art)
				require.NoError(t, err)
				require.Equal(t, int64(2), art.Id)
				require.Equal(t, "新的标题", art.Title)
				require.Equal(t, "新的内容", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, int64(456), art.Ctime)
				require.True(t, art.Utime > 234)

				var publishedArt article.PublishArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 2}}).
					Decode(&publishedArt)
				require.NoError(t, err)
				require.Equal(t, int64(2), publishedArt.Id)
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 123,
				}
				_, err := s.col.InsertOne(ctx, &art)
				require.NoError(t, err)
				part := article.PublishArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&art)
				require.NoError(t, err)
				require.Equal(t, int64(3), art.Id)
				require.Equal(t, "新的标题", art.Title)
				require.Equal(t, "新的内容", art.Content)
				require.Equal(t, int64(123), art.AuthorId)
				require.Equal(t, int64(456), art.Ctime)
				require.True(t, art.Utime > 234)

				var part article.PublishArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 3}}).Decode(&part)
				require.NoError(t, err)
				require.Equal(t, int64(3), part.Id)
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
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				art := article.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorId: 789,
				}
				_, err := s.col.InsertOne(ctx, &art)
				require.NoError(t, err)
				part := article.PublishArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
				defer cancel()
				var art article.Article
				err := s.col.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&art)
				require.NoError(t, err)
				require.Equal(t, int64(4), art.Id)
				require.Equal(t, "我的标题", art.Title)
				require.Equal(t, "我的内容", art.Content)
				require.Equal(t, int64(456), art.Ctime)
				require.Equal(t, int64(234), art.Utime)
				require.Equal(t, int64(789), art.AuthorId)

				var part article.PublishArticle
				err = s.liveCol.FindOne(ctx, bson.D{bson.E{Key: "id", Value: 4}}).Decode(&part)
				require.NoError(t, err)
				require.Equal(t, int64(4), part.Id)
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
			data, err := json.Marshal(tc.art)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			require.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			require.NoError(t, err)
			require.Equal(t, tc.wantRes.Code, result.Code)
			require.Equal(t, tc.wantRes.Msg, result.Msg)
			if tc.wantRes.Data > 0 {
				require.True(t, result.Data > 0)
			}
			tc.after(t)
		})
	}
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, new(ArticleMongoTestSuite))
}
