package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JaylanCharles/byline/internal/integration/startup"
	"github.com/JaylanCharles/byline/internal/repository/dao"
	ijwt "github.com/JaylanCharles/byline/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (s *ArticleHandlerSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.server = gin.Default()
	artHdl := startup.InitArticleHandler()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &ijwt.UserClaims{
			Uid: 123,
		})
	})
	artHdl.RegisterRoutes(s.server)
}

func (s *ArticleHandlerSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE articles")

}
func TestArticleHandlerSuite(t *testing.T) {
	suite.Run(t, new(ArticleHandlerSuite))
}

func (s *ArticleHandlerSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		// 前端传过来，肯定是一个 JSON
		// 预期中的输入
		art Article

		wantCode int
		// int64 表示希望结果里面带上文章的 id
		wantRes Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {
				s.TearDownTest()
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				art.Utime = 0
				art.Ctime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 我希望你的 ID 是 1
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "修改已有帖子，并保存",
			before: func(t *testing.T) {
				err := s.db.Create(dao.Article{
					// 为了测试方便，这个给一个指定的 Id，方便验证数据的时候使用这个 Id 进行查询
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Ctime:    123,
					Utime:    234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art dao.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 234)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:      2,
					Title:   "新的标题",
					Content: "新的内容",
					// 和上面保持一致
					AuthorId: 123,
					Ctime:    123,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 我希望你的 ID 是 1
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "修改帖子-别人的帖子",
			// 现在 uid 是 123 的是攻击者
			before: func(t *testing.T) {
				// 假装数据库已经有这个帖子
				err := s.db.Create(&dao.Article{
					Id:      3,
					Title:   "我的标题",
					Content: "我的内容",
					// 模拟别人
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 你要验证，保存到了数据库里面
				var art dao.Article
				err := s.db.Where("id=?", 3).
					First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Ctime:    123,
					Utime:    234,
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
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)

			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(reqBody))
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			// 执行
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
