package startup

import (
	"github.com/JaylanCharles/byline/pkg/logger"
)

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
