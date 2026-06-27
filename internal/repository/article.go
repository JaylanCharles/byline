package repository

import (
	"context"
	"time"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository/cache"
	dao "github.com/JaylanCharles/byline/internal/repository/dao/article"
	"github.com/JaylanCharles/byline/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error

	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error

	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)

	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type CachedArticleRepository struct {
	dao      dao.ArticleDAO
	cache    cache.ArticleCache
	userRepo UserRepository
	l        logger.Logger
}

func NewCachedArticleRepository(dao dao.ArticleDAO, l logger.Logger) ArticleRepository {
	return &CachedArticleRepository{
		dao: dao,
		l:   l,
	}
}

func (repo *CachedArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		_ = repo.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return repo.dao.Insert(ctx, repo.toEntity(art))
}

func (repo *CachedArticleRepository) Update(ctx context.Context, art domain.Article) error {
	defer func() {
		// 清空缓存
		_ = repo.cache.DelFirstPage(ctx, art.Author.Id)
	}()
	return repo.dao.UpdateById(ctx, repo.toEntity(art))
}

func (repo *CachedArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {

	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err == nil {
		// 清空缓存
		_ = repo.cache.DelFirstPage(ctx, art.Author.Id)
		err := repo.cache.SetPub(ctx, art)
		if err != nil {
			// 不需要特别关心
			// 比如说输出 WARN 日志
		}
	}
	return id, err
}

func (repo *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, author int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, id, author, status.ToUint8())
}

func (repo *CachedArticleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 你在这个地方，集成你的复杂的缓存方案
	// 你只缓存这一页
	if offset == 0 && limit <= 100 {
		data, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			go func() {
				repo.preCache(ctx, data)
			}()
			//return data[:limit], err
			return data, err
		}
	}
	res, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return repo.toDomain(src)
	})
	// 回写缓存的时候，可以同步，也可以异步
	go func() {
		err := repo.cache.SetFirstPage(ctx, uid, data)
		repo.l.Error("回写缓存失败", logger.Error(err))
		repo.preCache(ctx, data)
	}()
	return data, nil
}

func (repo *CachedArticleRepository) preCache(ctx context.Context, data []domain.Article) {
	//
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := repo.cache.Set(ctx, data[0])
		if err != nil {
			repo.l.Error("提前预加载缓存失败", logger.Error(err))
		}
	}
}

func (repo *CachedArticleRepository) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	data, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.toDomain(data), nil
}

func (repo *CachedArticleRepository) GetPublishedById(
	ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果你的 Content 被你放过去了 OSS 上，你就要让前端去读 Content 字段
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 你在这边要组装 user 了，适合单体应用
	usr, err := repo.userRepo.FindById(ctx, art.AuthorId)
	res := domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   usr.Id,
			Name: usr.Nickname,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
	return res, nil
}

func (repo *CachedArticleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *CachedArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			// 这里有一个错误
			Id: art.AuthorId,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}
