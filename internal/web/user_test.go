package web

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/service"
	svcmocks "github.com/JaylanCharles/byline/internal/service/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string
		// 本来应该写 mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService) 但是，code 根本用不上啊，所以可以简化
		// 这里我依旧是返回 nil，而不是直接不写 service.CodeService ，我感觉还是要保持一致，用不用自己调用的时候决定就好
		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)

		reqBody  string
		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					// 这里表示我心情好，我就是愿意写一下，期望传入的数据必须符合这些
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(nil) // name : "注册成功"   返回的是 nil,这个测试用例都叫做注册成功了，一定是没有错误返回啊！
				return usersvc, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : "hello#world123",
	"confirmPassword" : "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "注册成功",
		},
		{
			//switch {
			//		case errors.As(err, &maxBytesErr):
			//			c.AbortWithError(http.StatusRequestEntityTooLarge, err).SetType(ErrorTypeBind) //nolint: errcheck
			//		default:
			//			c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind) //nolint: errcheck
			//		}
			// 现在来说，这个测试用例，是有点过时了，因为还有一个错误 http.StatusRequestEntityTooLarge 我们不处理了，知道就好
			name: "参数不对， bind 失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				// 因为参数不对，所以根本没有走到调用 svc
				return nil, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : 
}
`, // 这是一段错误 json 格式，故意的
			wantCode: http.StatusBadRequest,
		},
		{
			name: "邮箱格式不对",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				// 依旧是走不到 signup 调用
				return nil, nil
			},
			reqBody: `
{	
	"email" : "123@qq",
	"password" : "hello#world123",
	"confirmPassword" : "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "你的邮箱格式不对",
		},
		{
			name: "两次输入的密码不匹配",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : "hello#world123",
	"confirmPassword" : "hello#world12"
}
`,
			wantCode: http.StatusOK,
			wantBody: "两次输入的密码不一致",
		},
		{
			name: "密码格式有问题",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : "hello",
	"confirmPassword" : "hello"
}
`,
			wantCode: http.StatusOK,
			wantBody: "密码必须大于8位，包含数字、特殊字符",
		},
		{
			name: "邮箱冲突",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(service.ErrUserDuplicateEmail)
				return usersvc, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : "hello#world123",
	"confirmPassword" : "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "邮箱冲突",
		},
		{
			name: "系统异常",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				usersvc := svcmocks.NewMockUserService(ctrl)
				usersvc.EXPECT().SignUp(gomock.Any(), domain.User{
					Email:    "123@qq.com",
					Password: "hello#world123",
				}).Return(errors.New("随便")) // 这个“随便”就可以充当系统异常
				return usersvc, nil
			},
			reqBody: `
{	
	"email" : "123@qq.com",
	"password" : "hello#world123",
	"confirmPassword" : "hello#world123"
}
`,
			wantCode: http.StatusOK,
			wantBody: "系统异常", // 这里就要注意了，有的是系统异常，有的是系统错误，在你第一遍编写 user.go 的时候就需要自己想好！
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.Default()
			//h := NewUserHandler(tc.mock(ctrl))
			userSvc, codeSvc := tc.mock(ctrl)
			h := NewUserHandler(userSvc, codeSvc, nil)
			h.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/users/signup", bytes.NewBuffer([]byte(tc.reqBody)))
			req.Header.Set("Content-Type", "application/json")
			// 这行代表这是我自己构造的，不可能出问题
			require.NoError(t, err)

			resp := httptest.NewRecorder()

			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			assert.Equal(t, tc.wantBody, resp.Body.String())
		})
	}
}

func TestMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// svcmocks 就是 -package=svcmocks 指定的
	usersvc := svcmocks.NewMockUserService(ctrl)

	// gomock.Any() 的个数就是原函数形参个数
	// Return 返回的是要看函数签名 SignUp(ctx context.Context, u domain.User) error 返回的是 error
	usersvc.EXPECT().SignUp(gomock.Any(), gomock.Any()).Return(errors.New("mock error"))
	err := usersvc.SignUp(context.Background(), domain.User{
		Email: "123@qq.com",
	})
	t.Log(err)
}
