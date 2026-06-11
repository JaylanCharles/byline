package ioc

import (
	"github.com/JaylanCharles/byline/config"
	"github.com/JaylanCharles/byline/internal/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	// 只会在初始化过程中使用 panic
	// panic 相当于整个 goroutine 结束
	// 一旦初始化过程出错，应用就不要启动了
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
