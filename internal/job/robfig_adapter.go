package job

import (
	"time"

	"github.com/JaylanCharles/byline/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type RankingJobAdapter struct {
	j Job
	l logger.Logger
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.Logger) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "cron_job",
		ConstLabels: map[string]string{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{}
}
func (r *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		r.p.Observe(float64(duration))
	}()

	err := r.j.Run()
	if err != nil {
		r.l.Error("运行任务失败", logger.Error(err), logger.String("job", r.j.Name()))
	}
}
