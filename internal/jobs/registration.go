package jobs

import (
	"log/slog"

	"github.com/zielma/yagi/internal/scheduler"
)

func RegisterJobs() {
	slog.Info("registering jobs")
	jobs := []struct {
		jobType string
		jobFunc func(r *scheduler.JobRunner) error
	}{
		{
			jobType: "fetchBudgets",
			jobFunc: fetchBudgets,
		},
	}

	for _, job := range jobs {
		if err := scheduler.RegisterJob(job.jobType, job.jobFunc); err != nil {
			panic(err)
		}
	}
}
