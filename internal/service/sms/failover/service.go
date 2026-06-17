package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/JaylanCharles/byline/internal/service/sms"
)

type FailoverSMSService struct {
	svcs []sms.Service
	// 理论上会用完，但是实际上根本用不完
	idx uint64 // ← 这个字段一直存在,跨多次请求 所以会一直往上加
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
func (f *FailoverSMSService) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&f.idx, 1)
	length := uint64(len(f.svcs))
	// 我要迭代 length
	for i := idx; i < idx+length; i++ {
		// 取余数来计算下标
		svc := f.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			// 前者是被取消，后者是超时
			return err
		}
		log.Println(err)
	}
	return errors.New("轮询了所有的服务商，但是发送都失败了")
}
