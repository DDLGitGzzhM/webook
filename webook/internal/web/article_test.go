package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"webook/webook/internal/domain"
	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/service"
	svcmocks "webook/webook/internal/service/mock"
	"webook/webook/internal/web/jwt"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) service.IArticleService
		reqBody  string
		wantCode int
		wantBody Result
	}{
		{
			name:    "正常发布",
			reqBody: `{"title": "标题", "content": "内容"}`,
			mock: func(ctrl *gomock.Controller) service.IArticleService {
				svc := svcmocks.NewMockIArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "标题",
					Content: "内容",
					Author: domain.Author{
						Id:   123,
						Name: "system",
					},
				}).Return(int64(1), nil)
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Data: float64(1),
				Msg:  "ok",
			},
		},
		{
			name:    "发布失败",
			reqBody: `{"title": "标题", "content": "内容"}`,
			mock: func(ctrl *gomock.Controller) service.IArticleService {
				svc := svcmocks.NewMockIArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "标题",
					Content: "内容",
					Author: domain.Author{
						Id:   123,
						Name: "system",
					},
				}).Return(int64(0), errors.New("publish error"))
				return svc
			},
			wantCode: http.StatusOK,
			wantBody: Result{
				Code: 5,
				Msg:  "保存失败",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			server.Use(func(c *gin.Context) {
				c.Set("claims", &jwt.UserClaims{
					UserId: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
			h.RegisterRoutes(server)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.Nil(t, err)

			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			server.ServeHTTP(resp, req)
			require.Equal(t, tc.wantCode, resp.Code)

			var got Result
			err = json.NewDecoder(resp.Body).Decode(&got)
			require.Nil(t, err)
			require.Equal(t, tc.wantBody, got)
		})
	}
}
