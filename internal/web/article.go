package web

import "github.com/gin-gonic/gin"

//var _ handler = (*ArticleHandler)(nil)

type ArticleHandler struct {
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles") // 就是习惯使用复数了，也算不上好的实践
	g.POST("/edit", h.Edit)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {

}
