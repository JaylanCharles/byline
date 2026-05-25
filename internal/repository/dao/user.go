package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// User 直接对应数据库表
// 也叫做 entity、PO、model
type User struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Email    string `gorm:"unique"`
	Password string

	// 毫秒数，因为 time.Time 跟时区有关，很麻烦
	Ctime int64
	Utime int64
}

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 推荐存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	return dao.db.WithContext(ctx).Create(&u).Error
}
