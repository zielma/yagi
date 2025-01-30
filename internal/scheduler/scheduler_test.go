package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

type mockDBQueries struct{}

func (m mockDBQueries) GetJobs(ctx context.Context) ([]database.GetJobsRow, error) {
	return []database.GetJobsRow{
		{
			ID:             "1",
			Type:           "syncBudgets",
			CronExpression: "5 4 * * *",
			Status:         "active",
		},
	}, nil
}

func TestScheduler(t *testing.T) {
	s, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
		gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
	)
	if err != nil {
		t.Fatal("could not create scheduler", err)
	}

	t.Run("load from database", func(t *testing.T) {
		s := Scheduler{
			dbQueries: mockDBQueries{},
			cfg:       &config.Config{},
			scheduler: s,
		}

		err := s.Load()
		if err != nil {
			t.Fatal("could not load jobs from database", err)
		}
	})
}
