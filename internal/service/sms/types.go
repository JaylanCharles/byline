package sms

import "context"

// sms.Service 是发短信的服务
type Service interface {
	// Send biz 是一个很含糊的参数
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}
