package startup

import (
	"github.com/JaylanCharles/byline/internal/service/oauth2/wechat"
	"github.com/JaylanCharles/byline/pkg/logger"
)

func InitOAuth2WechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "", l)
}
