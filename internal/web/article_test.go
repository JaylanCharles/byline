package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/service"
	svcmocks "github.com/JaylanCharles/byline/internal/service/mocks"
	ijwt "github.com/JaylanCharles/byline/internal/web/jwt"
	"github.com/JaylanCharles/byline/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody string

		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123, // 这要模拟登录态，直接薅过来
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
 "title": "我的标题",
 "content": "我的内容"
}
`,
			wantCode: 200,
			wantRes: Result{
				Data: float64(1),
				Msg:  "OK",
			},
		},
		{
			name: "publish错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      1,
					Title:   "新的标题",
					Content: "新的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("mock error"))
				return svc
			},
			reqBody: `
{
"id": 1,
"title": "新的标题",
"content": "新的内容"
}
`,
			wantCode: 200,
			wantRes: Result{
				// 原本是 int64的，但是因为 Data 是any，所以在反序列化的时候，
				// 用的 float64
				Msg:  "系统错误",
				Code: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("claims", &ijwt.UserClaims{
					Uid: 123,
				})
			})
			h := NewArticleHandler(tc.mock(ctrl), &logger.NopLogger{})
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")

			require.NoError(t, err)

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result
			err = json.NewDecoder(resp.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
