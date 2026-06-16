package ratelimit

import "context"

type Limiter interface {
	// key 是限流对象
	// bool 代表是否触发了限流，error 表示限流器本身有没有错误
	Limit(ctx context.Context, key string) (bool, error)
}
