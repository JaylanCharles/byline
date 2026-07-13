//go:build wireinject

package startup

import (
	"github.com/JaylanCharles/byline/interactive/repository"
	"github.com/JaylanCharles/byline/interactive/repository/cache"
	"github.com/JaylanCharles/byline/interactive/repository/dao"
	"github.com/JaylanCharles/byline/interactive/service"
	"github.com/google/wire"
)

func InitInteractiveService() service2.InteractiveService {
	wire.Build(thirdPartySet, interactiveServiceSet)
	return service2.NewInteractiveService(nil, nil)
}

var thirdPartySet = wire.NewSet(
	InitLogger,
	InitRedis,
	InitDB,
)

var interactiveServiceSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)
