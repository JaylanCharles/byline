package service

import (
	"context"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository"
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
	// 存起来
	return svc.repo.Create(ctx, u)
}
