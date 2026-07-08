//go:build wireinject

package main

import (
	"github.com/JaylanCharles/byline/internal/events/article"
	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/internal/repository/cache"
	"github.com/JaylanCharles/byline/internal/repository/dao"
	articleDAO "github.com/JaylanCharles/byline/internal/repository/dao/article"

	"github.com/JaylanCharles/byline/internal/service"
	"github.com/JaylanCharles/byline/internal/web"
	ijwt "github.com/JaylanCharles/byline/internal/web/jwt"
	"github.com/JaylanCharles/byline/ioc"
	"github.com/google/wire"
)

// func 名字随便
func InitWebServer() *App {
	wire.Build(
		thirdPartySet,
		interactiveServiceSet,
		rankingServiceSet,
		userServiceSet,
		articlServiceSet,
		codeServiceSet,

		// consumer
		article.NewInteractiveReadEventConsumer,
		article.NewKafkaProducer,

		ioc.InitSMSService,
		//ioc.InitOAuth2WechatService,
		ioc.InitJobs,
		ioc.InitRankingJob,

		web.NewUserHandler,
		web.NewArticleHandler,
		//web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitMiddlewares,
		ioc.InitWebServer,

		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}

var thirdPartySet = wire.NewSet(
	// 最基础的第三方依赖
	ioc.InitLogger,
	ioc.InitRedis,
	ioc.InitRLockClient,
	ioc.InitDB,
	ioc.InitKafka,
	ioc.NewConsumers,
	ioc.NewSyncProducer,
)

var interactiveServiceSet = wire.NewSet(
	service.NewInteractiveService,
	repository.NewCachedInteractiveRepository,
	dao.NewGORMInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

var rankingServiceSet = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCachedRankingRepository,
	cache.NewRankingRedisCache,
)

var userServiceSet = wire.NewSet(
	service.NewUserService,
	repository.NewUserRepository,
	cache.NewUserCache,
	dao.NewUserDAO,
)

var articlServiceSet = wire.NewSet(
	service.NewArticleService,
	repository.NewCachedArticleRepository,
	articleDAO.NewGORMArticleDAO,
)

var codeServiceSet = wire.NewSet(
	service.NewCodeService,
	repository.NewCodeRepository,
	cache.NewCodeCache,
)
