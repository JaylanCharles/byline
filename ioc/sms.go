package ioc

import (
	"github.com/JaylanCharles/byline/internal/service/sms"
	"github.com/JaylanCharles/byline/internal/service/sms/memory"
)

func InitSMSService() sms.Service {
	// 在这里换 sms 的实现方法
	return memory.NewService()
}
