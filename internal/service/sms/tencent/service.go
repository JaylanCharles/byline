package tencent

import (
	"context"
	"fmt"

	"github.com/JaylanCharles/byline/pkg/ratelimit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string // 传入指针的原因是，设计的参数要传入指针，垃圾设计
	signName *string
	client   *sms.Client
	limiter  ratelimit.Limiter
}

func NewService(client *sms.Client, signName string, appId string, limiter ratelimit.Limiter) *Service {
	return &Service{
		appId:    new(appId),
		signName: new(signName),
		client:   client,
		limiter:  limiter,
	}
}
func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		// 系统错误
		// 可以限流：保守策略，你的下游很坑的时候
		// 可以不限流：你的下游很强，业务可用性要求很高，尽量容错策略
		return fmt.Errorf("短信服务限流出现问题，%w", err) // 包一下这个错误
	}
	if limited {
		return fmt.Errorf("触发了限流")
	}

	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	req.TemplateId = new(tplId)
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	req.TemplateParamSet = s.toStringPtrSlice(args)
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败，code：%s，原因：%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (s *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
