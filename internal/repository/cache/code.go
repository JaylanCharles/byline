package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany        = errors.New("发送太频繁")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
	ErrUnkownForCode          = errors.New("我也不知道发什么了，反正跟 code 有关")
)

// 编译器会在编译的时候，把 set_code 的代码放进 luaSetCode 变量中
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

// interface 重构 -> 第二步：提出 interface
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
	// 注意 key 是小写的，其他方法自己调用的，不要干成接口了
}

// interface 重构 -> 第一步：这种结构体就是一种实现，改一个名字，因为基于 redis 改成 RedisCodeCache
type RedisCodeCache struct {
	client redis.Cmdable
}

// interface 重构 -> 第三步：将 NewXxx 入参和返回值改成 接口类型
func NewCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		//没毛病
		return nil
	case -1:
		// 发送频繁
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")
	}
}
func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		//没毛病
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooManyTimes
	case -2:
		return false, nil
	}
	return false, ErrUnkownForCode
}

func (c *RedisCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
