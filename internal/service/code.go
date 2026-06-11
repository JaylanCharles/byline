package service

// 这个就没有写成接口，为什么？
// 因为你发短信验证码这个操作就这一个！用不着抽象
import (
	"context"
	"fmt"
	"math/rand"

	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/internal/service/sms"
)

// 可以做成可配置的，但是没意义，因为好几年都不变
const codeTplId = "xxxxx" // 自己的 id

var (
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
)

type CodeService struct {
	repo *repository.CodeRepository // 这是结构体，所以指针
	//codeTplId string
	smsSvc sms.Service // 这是接口，所以不用指针
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.Service) *CodeService {
	return &CodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// biz 用于区分业务场景
	// 生成验证码
	code := svc.generateCode()
	// 塞进 redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送出去
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	//if err != nil {
	// Redis 有这个验证码，但是没有发送成功
	// 能不能删掉这个验证码？不能，因为有可能是因为超时所以没有发成功
	// 处理方式：在这里重试
	// 但是你不要在这里自己实现重试的方法，应该让调用方传入一个重试的服务
	//}
	return err
}

func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	num := rand.Intn(1000000) // 包含 0，不包含 1000000
	// 不够六位补前导 0
	return fmt.Sprintf("%06d", num)
}
