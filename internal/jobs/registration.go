package jobs

import (
	"log/slog"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/scheduler"
	"github.com/zielma/yagi/internal/ynab"
)

func RegisterJobs(cfg *config.Config) {
	slog.Info("registering jobs")
	client := ynab.NewClient(cfg)

	jobs := []struct {
		jobType string
		jobFunc func(r *scheduler.JobRunner) error
	}{
		{
			jobType: "fetchBudgets",
			jobFunc: func(r *scheduler.JobRunner) error {
				job := newFetchBudgetsJob(r.Database, client)
				return job.Run()
			},
		},
	}

	for _, job := range jobs {
		if err := scheduler.RegisterJob(job.jobType, job.jobFunc); err != nil {
			panic(err)
		}
	}
}
