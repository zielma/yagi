package jobs

import (
	"log/slog"

	"github.com/zielma/yagi/internal/scheduler"
)

func RegisterJobs() {
	slog.Info("initializing jobs")
	jobs := []struct {
		jobType string
		jobFunc func(r *scheduler.JobRunner) error
	}{
		{
			jobType: "syncBudgets",
			jobFunc: syncBudgets,
		},
	}

	for _, job := range jobs {
		if err := scheduler.RegisterJob(job.jobType, job.jobFunc); err != nil {
			panic(err)
		}
	}
}
