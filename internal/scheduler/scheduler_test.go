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
			Type:           "testFunc",
			CronExpression: "5 4 * * *",
			Status:         "active",
			Params:         "[\"test\"]",
		},
	}, nil
}

func TestScheduler(t *testing.T) {
	RegisterJob("testFunc", func(r *JobRunner, a string) error {
		return nil
	})

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
			t.Fatalf("load should not return error, err: %s", err)
		}

		jobs := s.scheduler.Jobs()
		for _, job := range jobs {
			if job.Name() != "testFunc" {
				t.Fatalf("job name expected to be testFunc, got: %s", job.Name())
			}
		}
	})
}
