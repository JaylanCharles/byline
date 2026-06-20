package service

import (
	"context"
	"errors"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email, password string) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

type userService struct {
	repo repository.UserRepository
	l    logger.Logger
}

func NewUserService(repo repository.UserRepository, l logger.Logger) UserService {
	return &userService{
		repo: repo,
		l:    l,
	}
}

// 无论是什么函数，好习惯是不知道返回什么或者没有具体的要返回，都要返回 error
// 为什么没有传递 *domain.User ？因为，1. 内容少 2. 传指针的话，还要判断==nil 3. 很大可能分配到栈上，没有逃逸问题
func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 要考虑加密问题
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 存起来
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, email, password string) (domain.User, error) {
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

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	// 有人认为可以不用写这行以及“判断有没有这个用户”逻辑。
	// 但是如果你不写，所有的请求都到下面了，database 受不了；
	// 如果你写了，可能 10w 请求，只有 1w 到 database
	// 业务研发的时候，经常遇到，这个部分就是快路径
	u, err := svc.repo.FindByPhone(ctx, phone)

	// 判断有没有这个用户
	if !errors.Is(err, repository.ErrUserNotFound) {
		// nil 会进来这里
		// 不为 ErrUserNotFound 的也会进来这里
		return u, err
	}

	// 这就是 慢路径
	// 在系统资源不足，触发降级之后，不执行慢路经
	u = domain.User{
		Phone: phone,
	}
	// 没有这个用户的话
	err = svc.repo.Create(ctx, u)

	if err != nil && !errors.Is(err, repository.ErrUserDuplicateEmail) {
		return u, err
	}

	// 这里会遇到主从延迟的问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return u, err
	}
	// 这边就是意味着是一个新用户
	// JSON 格式的 wechatInfo
	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: wechatInfo,
	})
	if err != nil && !errors.Is(err, repository.ErrDuplicateUser) {
		return domain.User{}, err
	}
	return svc.repo.FindByWechat(ctx, wechatInfo.OpenId)
}
