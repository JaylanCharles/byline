package ioc

import (
	"time"

	"github.com/JaylanCharles/byline/internal/job"
	"github.com/JaylanCharles/byline/internal/service"
	"github.com/JaylanCharles/byline/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
)

func InitRankingJob(svc service.RankingService, rlockClient *rlock.Client, l logger.Logger) *job.RankingJob {
	return job.NewRankingJob(svc, rlockClient, l, time.Second*30)
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 这里每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
