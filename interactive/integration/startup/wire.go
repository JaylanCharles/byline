//go:build wireinject

package startup

import (
	"github.com/JaylanCharles/byline/interactive/grpc"
	"github.com/JaylanCharles/byline/interactive/repository"
	"github.com/JaylanCharles/byline/interactive/repository/cache"
	"github.com/JaylanCharles/byline/interactive/repository/dao"
	"github.com/JaylanCharles/byline/interactive/service"
	"github.com/google/wire"
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdPartySet, interactiveServiceSet)
	return service.NewInteractiveService(nil, nil)
}

func InitInteractiveGRPCServer() *grpc.InteractiveServiceServer {
	wire.Build(thirdPartySet, interactiveServiceSet, grpc.NewInteractiveServiceServer)
	return new(grpc.InteractiveServiceServer)
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
