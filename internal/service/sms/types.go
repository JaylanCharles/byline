package sms

import "context"

// sms.Service 是发短信的服务
type Service interface {
	Send(ctx context.Context, tpl string, args []string, numbers ...string) error
}
