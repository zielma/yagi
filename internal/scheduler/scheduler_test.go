package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

func setupLoad(t *testing.T) (*Scheduler, *dbStub) {
	t.Helper()

	gs, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
		gocron.WithLogger(gocron.NewLogger(gocron.LogLevelDebug)),
	)
	if err != nil {
		t.Fatal("could not create scheduler", err)
	}

	db := dbStub{}
	s := Scheduler{
		dbQueries: &db,
		cfg:       &config.Config{},
		scheduler: gs,
	}

	return &s, &db
}

func TestLoadWithRegisteredJob(t *testing.T) {
	RegisterJob("testFunc", func(r *JobRunner, a string) error {
		return nil
	})

	s, db := setupLoad(t)
	db.GetJobsFunc = func(context.Context) ([]database.GetJobsRow, error) {
		return []database.GetJobsRow{
			{
				Type:           "testFunc",
				CronExpression: "5 4 * * *",
				Params:         sql.NullString{String: "[\"test\"]", Valid: true},
			},
		}, nil
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
}

func TestJobReturningAnError(t *testing.T) {
	ch := make(chan bool)
	if err := RegisterJob("errorFunc", func(r *JobRunner, a string) error {
		ch <- true
		return errors.New("error from the job")
	}); err != nil {
		t.Fatal("register job should not return error")
	}

	s, db := setupLoad(t)
	db.GetJobsFunc = func(context.Context) ([]database.GetJobsRow, error) {
		return []database.GetJobsRow{
			{
				Type:           "errorFunc",
				CronExpression: "5 4 * * *",
				Params:         sql.NullString{String: "[\"test\"]", Valid: true},
			},
		}, nil
	}

	err := s.Load()
	if err != nil {
		t.Fatalf("load should not return error, err: %s", err)
	}

	jobs := s.scheduler.Jobs()
	s.scheduler.Start()
	defer func() { _ = s.scheduler.Shutdown() }()

	if !slices.ContainsFunc(jobs, func(j gocron.Job) bool {
		return j.Name() == "errorFunc"
	}) {
		t.Fatal("job should be registered")
	}

	for _, job := range jobs {
		if job.Name() == "errorFunc" {
			if err := job.RunNow(); err != nil {
				t.Fatalf("job should not return error, err: %s", err)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Fatal("job should have finished")
	case jobRun := <-ch:
		if !jobRun {
			t.Fatalf("job should run and return bool through channel, want: %t, got: %t", true, jobRun)
		}
	}
}

func TestLoadWithoutRegisteredJob(t *testing.T) {
	s, db := setupLoad(t)
	db.GetJobsFunc = func(context.Context) ([]database.GetJobsRow, error) {
		return []database.GetJobsRow{
			{
				Type:           "unknownFunc",
				CronExpression: "5 4 * * *",
				Params:         sql.NullString{String: "[\"test\"]", Valid: true},
			},
		}, nil
	}

	err := s.Load()
	if err == nil {
		t.Fatal("load should return error when there's no handler for a job registered")
	}
}

func TestLoadWithInvalidCronExpression(t *testing.T) {
	s, db := setupLoad(t)
	db.GetJobsFunc = func(context.Context) ([]database.GetJobsRow, error) {
		return []database.GetJobsRow{
			{
				Type:           "testFunc",
				CronExpression: "invalid",
				Params:         sql.NullString{String: "[\"test\"]", Valid: true},
			},
		}, nil
	}

	err := s.Load()
	if err == nil {
		t.Fatal("load should return error when cron expression is invalid")
	}

	if !errors.Is(err, gocron.ErrCronJobParse) {
		t.Fatalf("load should return ErrCronJobParse when cron expression is invalid, got: %s", err)
	}
}

func TestNewWithoutDb(t *testing.T) {
	_, err := New(nil, &config.Config{})
	if err == nil {
		t.Fatal("new should return error when db is nil")
	}
}

func TestNewWithoutConfig(t *testing.T) {
	_, err := New(&sql.DB{}, nil)
	if err == nil {
		t.Fatal("new should return error when config is nil")
	}
}

func TestNew(t *testing.T) {
	s, err := New(&sql.DB{}, &config.Config{})

	if s == nil {
		t.Fatal("new should return a scheduler")
	}
	if err != nil {
		t.Fatal("new should not return error")
	}

	if s.scheduler == nil {
		t.Fatal("new should return a scheduler with a gocron scheduler")
	}
	if s.dbQueries == nil {
		t.Fatal("new should return a scheduler with a dbQueries")
	}
	if s.jobRunner == nil {
		t.Fatal("new should return a scheduler with a jobRunner")
	}
	if s.cfg == nil {
		t.Fatal("new should return a scheduler with a config")
	}
}

type dbStub struct {
	GetJobsFunc func(context.Context) ([]database.GetJobsRow, error)
}

func (m *dbStub) GetJobs(ctx context.Context) ([]database.GetJobsRow, error) {
	if m.GetJobsFunc == nil {
		return nil, errors.New("implement GetJobsFunc")
	}

	return m.GetJobsFunc(ctx)
}
