package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/redis/go-redis/v9"
)

// 规范的写法是这样，但是我们这个系统以后的很长时间内都是用 redis,所以可以偷懒
// var ErrKeyNotExist = errors.New("key 不存在")
// var ErrKeyNotExist = redis.Nil

type UserCache struct {
	client redis.Cmdable
	// 过期时间
	expiration time.Duration
}

// A 用到了 B，B 一定是接口
// A 用到了 B，B 一定是字段
// A 用到了 B，A 绝对不初始化 B，而是外面注入
func NewUserCache(cmd redis.Cmdable) *UserCache {
	return &UserCache{
		client:     cmd,
		expiration: time.Minute * 15,
	}
}

// 如果没有数据，返回一个特定的 error
func (cache *UserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.Key(id)
	// 数据不存在，err = redis.Nil
	val, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.Key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) Key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
