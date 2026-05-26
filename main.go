package main

import (
	"strings"
	"time"

	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/internal/repository/dao"
	"github.com/JaylanCharles/byline/internal/service"
	"github.com/JaylanCharles/byline/internal/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db := InitDB()

	server := InitWebServer()
	u := InitUser(db)
	u.RegisterRoutes(server)

	server.Run(":8080")
}

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/byline"))
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
	// 作用于定义在server上的全部路由
	// 怎么改？就看前端页面的请求对照着改
	server.Use(cors.New(cors.Config{
		// 公司中最好直接指定网址
		//AllowOrigins:  []string{"http://localhost:3000/"},
		// 不写这行，表示全部都允许
		//AllowMethods:  []string{"POST", "PATCH"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这行配置不懂，后面jwt中能用起来
		//ExposeHeaders: []string{"Content-Length"},
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
