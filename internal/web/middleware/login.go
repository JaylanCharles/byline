package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	paths []string
}

func NewLoginMiddlewareBuilder() *LoginMiddlewareBuilder {
	return &LoginMiddlewareBuilder{}
}

func (l *LoginMiddlewareBuilder) IgorePaths(path string) *LoginMiddlewareBuilder {
	l.paths = append(l.paths, path)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	// 用 Go 的方式编码解码
	gob.Register(time.Now())

	// 每个请求进来，先看是不是登录/注册接口，是的话直接放行；否则检查 session 里有没有 userId，没有就返回 401
	return func(ctx *gin.Context) {
		// 不需要登录校验的
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(ctx)
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		// 取出来的是 interface{} 类型（any）,所以需要进行断言
		updateTime := sess.Get("update_time")
		// 刚登陆，说明还没有刷新过
		if updateTime == nil {
			sess.Set("update_time", now)
			sess.Options(sessions.Options{
				MaxAge: 300,
			})
			if err := sess.Save(); err != nil {
				panic(err)
			}
			return
		}

		// updateTime 有的话
		updateTimeVal, _ := updateTime.(time.Time)
		if now.Sub(updateTimeVal) > time.Second*10 {
			sess.Set("update_time", now)
			sess.Options(sessions.Options{
				MaxAge: 300,
			})
			if err := sess.Save(); err != nil {
				panic(err)
			}
		}
	}
}
