package ratelimit

import (
	"context"
	"fmt"

	"github.com/JaylanCharles/byline/internal/service/sms"
	"github.com/JaylanCharles/byline/pkg/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type RatelimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRatelimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RatelimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RatelimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候
		// 可以不限流：你的下游很强，业务可用性要求很高，尽量容错策略
		return fmt.Errorf("短信服务限流出现问题，%w", err) // 包一下这个错误
	}
	if limited {
		return errLimited
	}
	err = s.svc.Send(ctx, tpl, args, numbers...)
	return err
}
