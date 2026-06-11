package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	// 正常来说，需要自己写错误信息，但是 gorm 提供了这个错误信息，就不需要自己写了
	ErrUserNotFound = gorm.ErrRecordNotFound
)

// User 直接对应数据库表
// 也叫做 entity、PO、model
type User struct {
	Id       int64          `gorm:"primaryKey,autoIncrement"`
	Email    sql.NullString `gorm:"unique"`
	Password string

	// 唯一索引允许有多个空值，但是不能有多个 "" 空字符串，同样的，Email 也是
	// Phone *string 这种方式也可以，但是不推荐，因为需要解引用，需要判空
	Phone sql.NullString `gorm:"unique"`

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
func (dao *UserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}
func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}
func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	return u, err
}
func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	// 推荐存毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	// 这段代码是跟底层强耦合的代码，因为如果底层数据库不适用 mysql 的话，这段代码就 gg 了
	// 虽然是强耦合，但是问题不大，因为正常人不会动不动换数据库
	if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		const uniqueConfictsErrNo uint16 = 1062
		if mysqlErr.Number == uniqueConfictsErrNo {
			// 邮箱冲突 or 手机号码冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}
