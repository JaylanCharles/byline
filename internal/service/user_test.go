package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository"
	repomocks "github.com/JaylanCharles/byline/internal/repository/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestEUserService_Login(t *testing.T) {
	// 做成一个测试用例都用到的时间
	// 以后写测试会经常遇到关于时间的问题，可以这么处理
	now := time.Now()
	// 不使用自动生成的，太难用了
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) repository.UserRepository

		//输入
		//ctx      context.Context  // 在 login 中没有用到，这个 ctx 主要是想往下传递的，所以，可以使用固定入参
		email    string
		password string

		//输出
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Phone:    "",
					Password: "$2a$10$70UR.SvCd1f6Kfpszs/i3.8wxg1BF7XbYArHLx6RPE.nGwFBsLL1m",
					Ctime:    now,
				}, nil)
				return repo
			},

			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{
				Email:    "123@qq.com",
				Phone:    "",
				Password: "$2a$10$70UR.SvCd1f6Kfpszs/i3.8wxg1BF7XbYArHLx6RPE.nGwFBsLL1m", // 这里返回的密码不是 hello#world123，返回的是加密后的密码
				Ctime:    now,
			},
			wantErr: nil,
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},

			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name: "DB错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{}, errors.New("mock db 错误"))
				return repo
			},

			email:    "123@qq.com",
			password: "hello#world123",

			wantUser: domain.User{},
			wantErr:  errors.New("mock db 错误"),
		},
		{
			name: "密码不对",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := repomocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(domain.User{
					Email:    "123@qq.com",
					Phone:    "",
					Password: "$2a$10$70UR.SvCd1f6Kfpszs/i3.8wxg1BF7XbYArHLx6RPE.nGwFBsLL1m",
					Ctime:    now,
				}, nil)
				return repo
			},

			email:    "123@qq.com",
			password: "xxxello#world123",

			wantUser: domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := NewUserService(tc.mock(ctrl))
			u, err := svc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err) // 先判断 error
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestEncrypted(t *testing.T) {
	res, err := bcrypt.GenerateFromPassword([]byte("hello#world123"), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(res))
	}
}
