package main

import (
	"strings"
	"time"

	"github.com/JaylanCharles/byline/config"
	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/internal/repository/dao"
	"github.com/JaylanCharles/byline/internal/service"
	"github.com/JaylanCharles/byline/internal/web"
	"github.com/JaylanCharles/byline/internal/web/middleware"
	"github.com/JaylanCharles/byline/pkg/ginx/middlewares/ratelimit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := InitDB()

	server := InitWebServer()
	u := InitUser(db)
	u.RegisterRoutes(server)

	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	// 只会在初始化过程中使用 panic
	// panic 相当于整个 goroutine 结束
	// 一旦初始化过程出错，应用就不要启动了
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func InitWebServer() *gin.Engine {
	server := gin.Default()

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	// 作用于定义在server上的全部路由
	// 怎么改？就看前端页面的请求对照着改
	server.Use(cors.New(cors.Config{
		// 公司中最好直接指定网址
		//AllowOrigins:  []string{"http://localhost:3000/"},
		// 不写这行，表示全部都允许
		//AllowMethods:  []string{"POST", "PATCH"},
		// 表示允许客户端哪些可以过来
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这行配置不懂，后面jwt中能用起来
		// 这行配置的意思是：将服务器端的 header 暴露给前端，允许前端得到这个
		ExposeHeaders: []string{"x-jwt-token"},
		// 是否允许带上 cookie 之类的
		AllowCredentials: true,
		// 这种方式推荐推荐
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				// 开发环境
				return true
			}
			return strings.Contains(origin, "company.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	// 方式一：使用 session 方式进行登录验证
	// cookie 是可以换成 redis 的
	//store, err := redis.NewStore(16,
	//	"tcp", "localhost:6379",
	//	"", "",
	//	[]byte("vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBr"),
	//	[]byte("WgK8Rmob6C5b2PCixdcERXXAzj4wAw7Y"))
	//if err != nil {
	//	panic(err)
	//}
	//server.Use(sessions.Sessions("mysession", store))
	//server.Use(middleware.NewLoginMiddlewareBuilder().
	//	IgorePaths("/users/signup").
	//	IgorePaths("/users/login").
	//	Build())

	// 方式二：使用 JWT 方式进行登陆验证
	server.Use(middleware.NewLoginJWTMiddlewareBuilder().
		IgorePaths("/users/signup").
		IgorePaths("/users/login").
		Build())

	return server
}

func InitUser(db *gorm.DB) *web.UserHandler {
	ud := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(ud)
	svc := service.NewUserService(repo)
	//u := &web.UserHandler{}
	u := web.NewUserHandler(svc)
	return u
}
