package jobs

import (
	"log/slog"

	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

type Runner struct {
	dbQueries *database.Queries
	cfg       *config.Config
}

func NewRunner(dbQueries *database.Queries, cfg *config.Config) *Runner {
	return &Runner{
		dbQueries: dbQueries,
		cfg:       cfg,
	}
}

func (r *Runner) GetJobFunc(jobType string) func() error {
	switch jobType {
	case "syncBudgets":
		return r.syncBudgets
	default:
		slog.Error("unknown job type", "job_type", jobType)
		return func() error { return nil }
	}
}
