package events

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/JaylanCharles/byline/interactive/repository"
	"github.com/JaylanCharles/byline/pkg/logger"
	"github.com/JaylanCharles/byline/pkg/saramax"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(client sarama.Client, l logger.Logger, repo repository.InteractiveRepository) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		l:      l,
		repo:   repo,
	}
}

func (r *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", r.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{"article_read"}, saramax.NewHandler[ReadEvent](r.l, r.Consume))
		if er != nil {
			r.l.Error("退出了消费循环异常", logger.Error(er))
		}
	}()
	return err
}

// Consume 这个不是幂等的
func (r *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return r.repo.IncrReadCnt(ctx, "article", t.Aid)
}
