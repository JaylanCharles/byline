package main

import (
	"github.com/JaylanCharles/byline/internal/web"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	//u := &web.UserHandler{}
	u := web.NewUserHandler()
	u.RegisterRoutes(server)

	server.Run(":8080")
}
