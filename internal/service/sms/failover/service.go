package failover

import (
	"context"
	"errors"
	"log"

	"github.com/JaylanCharles/byline/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
}

func NewFailoverSMSService(svcs []sms.Service) sms.Service {
	return &FailoverSMSService{
		svcs: svcs,
	}
}

func (f *FailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range f.svcs {
		err := svc.Send(ctx, tpl, args, numbers...)

		//发送成功
		if err == nil {
			return nil
		}
		// 输出日志
		// 做好监控，因为正常来说走不到这里的
		log.Println(err)
	}
	return errors.New("全部服务商都失败了")
}
