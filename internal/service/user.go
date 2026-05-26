package service

import (
	"context"
	"errors"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// 无论是什么函数，好习惯是不知道返回什么或者没有具体的要返回，都要返回 error
// 为什么没有传递 *domain.User ？因为，1. 内容少 2. 传指针的话，还要判断==nil 3. 很大可能分配到栈上，没有逃逸问题
func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	// 要考虑加密问题
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 存起来
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	// 这里就要考虑，返回错误，要让前端知道是系统错误还是账号密码错误，所以要定义新的错误类型
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// 需要日志 可以使用 DEBUG 类型，或者使用 INFO 类型，因为密码输入错误常见
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}
