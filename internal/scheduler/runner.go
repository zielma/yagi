package scheduler

import (
	"log/slog"
	"sync"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

type RunnerJobFunc func(r *JobRunner) error

var jobRunners = make(map[string]any)
var jobsMutex sync.RWMutex

func RegisterJob(jobType string, function any) {
	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	if function == nil {
		panic("job function cannot be nil")
	}

	if _, dup := jobRunners[jobType]; dup {
		panic("job type already registered: " + jobType)
	}

	jobRunners[jobType] = function
}

func getJobFunc(jobType string) any {
	jobsMutex.RLock()
	defer jobsMutex.RUnlock()

	if jobFunc, ok := jobRunners[jobType]; ok {
		return jobFunc
	}

	slog.Error("unknown job type", "job_type", jobType)
	return func() error { return nil }
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
