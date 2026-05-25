package main

import (
	"strings"
	"time"

	"github.com/JaylanCharles/byline/internal/web"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
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

	//u := &web.UserHandler{}
	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
