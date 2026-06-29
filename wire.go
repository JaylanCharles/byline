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
		// 最基础的第三方依赖
		ioc.InitDB, ioc.InitRedis, ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		// consumer
		article.NewInteractiveReadEventConsumer,
		article.NewKafkaProducer,

		//初始化 DAO
		dao.NewUserDAO,
		articleDAO.NewGORMArticleDAO,
		dao.NewGORMInteractiveDAO,

		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisInteractiveCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		service.NewUserService,
		service.NewArticleService,
		service.NewCodeService,
		service.NewInteractiveService,

		ioc.InitSMSService,
		ioc.InitOAuth2WechatService,

		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,

		ioc.InitMiddlewares,
		ioc.InitWebServer,

		// 组装我这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
