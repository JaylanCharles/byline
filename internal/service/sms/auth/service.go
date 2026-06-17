package auth

import (
	"context"
	"errors"

	"github.com/JaylanCharles/byline/internal/service/sms"
	"github.com/golang-jwt/jwt/v5"
)

type SMSService struct {
	svc sms.Service
	key []byte
}

// ← 缺了这个:构造时把 key 传进来
func NewSMSService(svc sms.Service, key []byte) sms.Service {
	return &SMSService{
		svc: svc,
		key: key, // ← key 从外部参数注入
	}
}

func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	var claims Claims
	token, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token 不合法")
	}
	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}

type Claims struct {
	jwt.RegisteredClaims
	Tpl string
}
