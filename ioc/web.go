package ioc

import (
	"strings"
	"time"

	"github.com/JaylanCharles/byline/internal/web"
	ijwt "github.com/JaylanCharles/byline/internal/web/jwt"
	"github.com/JaylanCharles/byline/internal/web/middleware"
	"github.com/JaylanCharles/byline/pkg/ginx"
	"github.com/JaylanCharles/byline/pkg/ginx/middlewares/metric"
	loggerMy "github.com/JaylanCharles/byline/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

// 这个方法一定是不稳定的，意思就是以后可能经常改，这是不可避免的
// 为了运行项目，oauth2WechatHdl *web.OAuth2WechatHandler 先不注入
func InitWebServer(mdls []gin.HandlerFunc, hdl *web.UserHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	hdl.RegisterRoutes(server)
	//oauth2WechatHdl.RegisterRoutes(server)
	articleHdl.RegisterRoutes(server)
	return server
}

// 这里是最能体现依赖注入的，redisClient redis.Cmdable 我只需要这个，具体怎么来的我不管
func InitMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l loggerMy.Logger) []gin.HandlerFunc {
	//bd := logger.NewBuilder(func(ctx context.Context, al *logger.AccessLog) {
	//	l.Debug("HTTP 请求", loggerMy.Field{
	//		Key:   "al",
	//		Value: al,
	//	})
	//}).AllowReqBody(true).AllowRespBody()
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.logreq")
	//	bd.AllowReqBody(ok)
	//})
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "dev",
		Subsystem: "byline",
		Name:      "http_biz_code",
		Help:      "HTTP 的业务错误码",
	})
	return []gin.HandlerFunc{
		corsHdl(),
		(&metric.MiddlewareBuilder{
			Namespace:  "dev",
			Subsystem:  "byline",
			Name:       "gin_http", // 千万不要使用连字符
			Help:       "统计 GIN 的 HTTP 接口",
			InstanceID: "my-instance-1", // 这个可以使用连字符
		}).Build(),
		//bd.Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
			IgorePaths("/users/signup").
			IgorePaths("/users/login_sms/code/send").
			IgorePaths("/users/login_sms").
			IgorePaths("/users/refresh_token").
			IgorePaths("/oauth2/wechat/authurl").
			IgorePaths("/oauth2/wechat/callback").
			IgorePaths("/users/login").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 公司中最好直接指定网址
		//AllowOrigins:  []string{"http://localhost:3000/"},
		// 不写这行，表示全部都允许
		//AllowMethods:  []string{"POST", "PATCH"},
		// 表示允许客户端哪些可以过来
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 这行配置不懂，后面jwt中能用起来
		// 这行配置的意思是：将服务器端的 header 暴露给前端，允许前端得到这个
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
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
	})
}
