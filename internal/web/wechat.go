package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JaylanCharles/byline/internal/service"
	"github.com/JaylanCharles/byline/internal/service/oauth2/wechat"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
)

type OAuth2WechatHandler struct {
	svc             wechat.Service
	userSvc         service.UserService
	stateKey        []byte
	stateCookieName string
	JWTHandler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:     svc,
		userSvc: userSvc,
		// 这些数据我们写死就可以，不需要进行依赖注入
		// 不同的地方换一换 key, 防止人家攻破一个地方，所有地方都被攻破
		stateKey:        []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgB"),
		stateCookieName: "jwt-state",
		JWTHandler:      newJwtHandler(),
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.Auth2URL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) Auth2URL(ctx *gin.Context) {
	// uuid 使用短的，为什么不用长的，因为太长了
	state := uuid.New()

	url, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "构造扫码登录 URL 失败",
		})
		return
	}

	err = h.setStateCookie(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}

	ctx.JSON(http.StatusOK, Result{
		Data: url,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}

	code := ctx.Query("code")

	wechatInfo, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	// 设置 jwttoken 需要拿到 uid
	// 从 UserService 中拿，没有合适的方法，自己重新创建一个方法
	user, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}

	err = h.setLoginToken(ctx, user.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	claims := StateClaims{
		State: state,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenStr, err := token.SignedString(h.stateKey)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return err
	}

	ctx.SetCookie(h.stateCookieName, tokenStr, 600, "/oauth2/wechat/callback", "", false, true)

	return nil
}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")

	ck, err := ctx.Cookie(h.stateCookieName)
	if err != nil {
		// 妥妥的有人搞你
		// 做好监控
		return fmt.Errorf("无法获得 cookie %w", err)
	}

	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.stateKey, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}

	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}

	return nil
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}
