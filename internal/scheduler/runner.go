package scheduler

import (
	"log/slog"
	"sync"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

type RunnerJobFunc func(r *JobRunner) error

var jobRunners = make(map[string]RunnerJobFunc)
var jobsMutex sync.RWMutex

func RegisterJob(jobType string, jobFunc RunnerJobFunc) {
	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	if jobFunc == nil {
		panic("job function cannot be nil")
	}

	if _, dup := jobRunners[jobType]; dup {
		panic("job type already registered: " + jobType)
	}

	jobRunners[jobType] = jobFunc
}

func getJobFunc(jobType string) RunnerJobFunc {
	jobsMutex.RLock()
	defer jobsMutex.RUnlock()

	if jobFunc, ok := jobRunners[jobType]; ok {
		return jobFunc
	}

	slog.Error("unknown job type", "job_type", jobType)
	return func(_ *JobRunner) error { return nil }
}

type JobRunner struct {
	Database *database.Queries
	Config   *config.Config
}

func NewJobRunner(dbQueries *database.Queries, cfg *config.Config) *JobRunner {
	return &JobRunner{
		Database: dbQueries,
		Config:   cfg,
	}
}
