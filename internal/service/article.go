package service

import (
	"context"

	"github.com/JaylanCharles/byline/internal/domain"
	events "github.com/JaylanCharles/byline/internal/events/article"
	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/JaylanCharles/byline/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer events.Producer
	l        logger.Logger
}

func NewArticleService(repo repository.ArticleRepository, producer events.Producer, l logger.Logger) ArticleService {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}

func (svc *articleService) Withdraw(ctx context.Context, art domain.Article) error {
	return svc.repo.SyncStatus(ctx, art.Id, art.Author.Id, domain.ArticleStatusPrivate)
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := svc.repo.Update(ctx, art)
		return art.Id, err
	}
	return svc.repo.Create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

func (svc *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, uid, offset, limit)
}

func (svc *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetByID(ctx, id)
}

func (svc *articleService) GetPublishedById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	// 另一个选项，在这里组装 Author，调用 UserService
	art, err := svc.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			er := svc.producer.ProduceReadEvent(ctx, events.ReadEvent{
				// 即便你的消费者要用 art 的里面的数据，
				// 让它去查询，你不要在 event 里面带
				Uid: uid,
				Aid: id,
			})
			if er != nil {
				svc.l.Error("发送读者阅读事件失败")
			}
		}()
	}

	return art, err
}
