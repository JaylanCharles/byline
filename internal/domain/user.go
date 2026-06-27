package domain

import "time"

// User 领域对象
// BO
type User struct {
	Id         int64
	Email      string
	Password   string
	Phone      string
	Nickname   string
	WechatInfo WechatInfo
	// 不需要 comfirmPassword ，因为web层已经检验过了
	Ctime time.Time
}
