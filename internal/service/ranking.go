package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/JaylanCharles/byline/internal/domain"
	"github.com/JaylanCharles/byline/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
)

type RankingService interface {
	TopN(ctx context.Context) error
}

type BatchRankingService struct {
	artSvc    ArticleService
	intrSvc   InteractiveService
	repo      repository.RankingRepository
	batchSize int
	n         int
	// scoreFunc 不能返回负数
	scoreFunc func(t time.Time, likeCnt int64) float64
}

func NewBatchRankingService(artSvc ArticleService, intrSvc InteractiveService) RankingService {
	return &BatchRankingService{
		artSvc:    artSvc,
		intrSvc:   intrSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(t time.Time, likeCnt int64) float64 {
			sec := time.Since(t).Seconds()
			return float64(likeCnt-1) / math.Pow(sec+2, 1.5)
		},
	}
}

func (svc *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := svc.topN(ctx)
	if err != nil {
		return err
	}
	// 在这里，存起来
	return svc.repo.ReplaceTopN(ctx, arts)
}

func (svc *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	// now 的作用是：往下传入时间参数，dao 中处理时间，实现仅仅处理一段时间的数据
	now := time.Now()

	// 先拿一批数据
	offset := 0
	type Score struct {
		art   domain.Article
		score float64
	}

	// 这里可以用非并发安全
	topN := queue.NewConcurrentPriorityQueue[Score](svc.n,
		func(src Score, dst Score) int {
			if src.score > dst.score {
				return 1
			} else if src.score == dst.score {
				return 0
			}
			return -1
		})

	for {
		// arts => ids => intrs
		arts, err := svc.artSvc.ListPub(ctx, now, offset, svc.batchSize)
		if err != nil {
			return nil, err
		}

		ids := slice.Map[domain.Article, int64](arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})

		// 要去找到对应的点赞数据
		intrs, err := svc.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}

		// 计算 score
		for _, art := range arts {
			intr := intrs[art.Id]
			score := svc.scoreFunc(art.Utime, intr.LikeCnt)

			// 若空，则先直接入队
			err = topN.Enqueue(Score{
				art:   art,
				score: score,
			})
			// 若溢出，则比较、出队
			if errors.Is(err, queue.ErrOutOfCapacity) {
				val, _ := topN.Dequeue()
				if val.score < score {
					err = topN.Enqueue(Score{
						art:   art,
						score: score,
					})
				} else {
					_ = topN.Enqueue(val)
				}
			}
		}

		// 当前一批都没取够，可以肯定没有下一批了
		// 又或者已经取到了七天之前的数据了，说明可以中断了
		if len(arts) < svc.batchSize || now.Sub(arts[len(arts)-1].Utime).Hours() > 7*24 {
			break
		}

		// 更新 offset
		offset = offset + len(arts)
	}

	// 拿到结果，注意每次出队的 score 是最小的
	res := make([]domain.Article, svc.n)
	for i := svc.n - 1; i >= 0; i-- {
		val, err := topN.Dequeue()
		if err != nil {
			// 说明取完了，不够 n
			break
		}

		res[i] = val.art
	}
	return res, nil
}
