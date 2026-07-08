package article

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{
		db: db,
	}
}

func (dao *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := dao.db.WithContext(ctx).
		Where("utime<?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, author int64, status uint8) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, author).
			Updates(map[string]any{
				"status": status,
				"Utime":  now,
			})

		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}

		if res.RowsAffected != 1 {
			// 要么 ID 是错的，要么是作者不对
			// 后者情况下，可能有人在搞你的系统
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&Article{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"Utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})

}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		art.Id = id

		now := time.Now().UnixMilli()
		art.Ctime = now
		art.Utime = now

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   art.Title,
				"content": art.Content,
				"status":  art.Status,
				"utime":   now,
			}),
		}).Create(&art).Error
	})
	return id, err
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).
		Updates(map[string]any{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   art.Utime,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是创作者非法 id %d, author_id %d", art.Id, art.AuthorId)
	}

	return res.Error
}
