package web

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	// access_token_key
	atKey []byte
	// refresh_token_key
	rtKey []byte
}

type UserClaims struct {
	jwt.RegisteredClaims
	// 声明自己要放进 token 里的数据
	Uid       int64
	UserAgent string
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	Uid int64
}

// newJwtHandler 小写代表这是内部的，就是不希望外面人能调用
func newJwtHandler() JWTHandler {
	return JWTHandler{
		atKey: []byte("vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBa"),
		rtKey: []byte("vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBr"),
	}
}
func (j JWTHandler) setJWTToken(ctx *gin.Context, uid int64) error {
	// 方式二：使用 JWT 实现登陆状态的初始化
	// 这里不使用指针的原因是，不需要进行修改值，仅仅需要传递一下就可以
	claims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 这个字符串 vjYqKKBpPfsWGpfq1Ljo57BgjsMg9yBr 是 JWT 的签名密钥(secret key)
	tokenStr, err := token.SignedString(j.atKey)
	if err != nil {
		return err
	}
	// 学习的过程是探索的过程，可以先自己将 tokenStr 打印出来看看，然后再弄后面的东西
	// ctx.Header() 直接操控响应头的内容！ 所以说 ctx 是一次请求的上下文
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func ExtractToken(ctx *gin.Context) string {
	tokenHeader := ctx.GetHeader("Authorization")

	segs := strings.Split(tokenHeader, " ")
	if len(segs) != 2 {
		return ""
	}

	return segs[1]
}

func (j JWTHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	tokenStr, err := token.SignedString(j.rtKey)
	if err != nil {
		return err
	}

	ctx.Header("x-refresh-token", tokenStr)
	return nil
}
