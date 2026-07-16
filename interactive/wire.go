//go:build wireinject

package main

import (
	"github.com/JaylanCharles/byline/interactive/events"
	"github.com/JaylanCharles/byline/interactive/grpc"
	"github.com/JaylanCharles/byline/interactive/ioc"
	"github.com/JaylanCharles/byline/interactive/repository"
	"github.com/JaylanCharles/byline/interactive/repository/cache"
	"github.com/JaylanCharles/byline/interactive/repository/dao"
	"github.com/JaylanCharles/byline/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	// 暂时不理会 consumer 怎么启动
	ioc.InitRedis)

var interactiveServiceSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitAPP() *App {
	wire.Build(interactiveServiceSet,
		thirdPartySet,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
