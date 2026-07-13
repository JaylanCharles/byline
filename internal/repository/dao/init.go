package dao

import (
	dao2 "github.com/JaylanCharles/byline/interactive/repository/dao"
	"github.com/JaylanCharles/byline/internal/repository/dao/article"
	"gorm.io/gorm"
)

func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&article.Article{},
		&article.PublishedArticle{},
		&dao2.Interactive{},
		&dao2.UserLikeBiz{},
		&dao2.Collection{},
		&dao2.UserCollectionBiz{},
		&Job{},
	)
}
