package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	AtKey = []byte("vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBa")
	RtKey = []byte("vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBr")
)

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明自己要放进 token 里的数据
	Uid       int64
	UserAgent string
	Ssid      string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Ssid string
	Uid  int64
}

func NewRedisJWTHandler(cmd redis.Cmdable) Handler {
	return &RedisJWTHandler{
		cmd: cmd,
	}
}

func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	// 方式二：使用 JWT 实现登陆状态的初始化
	// 这里不使用指针的原因是，不需要进行修改值，仅仅需要传递一下就可以
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 这个字符串 vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBr 是 JWT 的签名密钥(secret key)
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	// 学习的过程是探索的过程，可以先自己将 tokenStr 打印出来看看，然后再弄后面的东西
	// ctx.Header() 直接操控响应头的内容！ 所以说 ctx 是一次请求的上下文
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")

	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}

	return segs[1]
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	// 这里用长的 uuid
	ssid := uuid.New().String()
	err := h.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}

	err = h.setRefreshToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return err
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")

	c, ok := ctx.Get("claims") // 获取的是 any 类型
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}

	claims, ok := c.(*UserClaims)
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
	}

	return h.cmd.Set(ctx, fmt.Sprintf("users:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := h.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil
	case err == nil:
		if val == 0 {
			return nil
		}
		return errors.New("session 已经无效了")
	default:
		return err
	}
}

func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
