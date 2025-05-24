package scheduler

import (
	"errors"
	"fmt"
	"sync"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

var (
	ErrUnknownJobType = errors.New("unknown job type")
)

type RunnerJobFunc func(r *JobRunner) error

var jobRunners = make(map[string]any)
var jobsMutex sync.RWMutex

func RegisterJob(jobType string, function any) error {
	if function == nil {
		return fmt.Errorf("job function cannot be nil")
	}

	jobsMutex.Lock()
	defer jobsMutex.Unlock()

	if _, exist := jobRunners[jobType]; exist {
		return fmt.Errorf("job type already registered: %s", jobType)
	}

	jobRunners[jobType] = function
	return nil
}

func getJobFunc(jobType string) (any, error) {
	jobsMutex.RLock()
	defer jobsMutex.RUnlock()

	if jobFunc, ok := jobRunners[jobType]; ok {
		return jobFunc, nil
	}

	return nil, ErrUnknownJobType
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
