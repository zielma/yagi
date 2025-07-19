package scheduler

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/to"
)

func setupLoad(t *testing.T) (*Scheduler, *mockSchedulerStore) {
	t.Helper()

	store := mockSchedulerStore{}
	s, err := New(&store, &config.Config{})
	if err != nil {
		t.Fatalf("new should not return error, got: %s", err)
	}
	return s, &store
}

func TestLoadWithRegisteredJob(t *testing.T) {
	s, store := setupLoad(t)
	err := s.AddJobHandler(&mockJobHandler{
		name: "testFunc",
		run: func(ctx context.Context, params ...any) error {
			if len(params) != 1 {
				data := params[0].([]interface{})
				str := data[0].(string)
				if str != "test" {
					return errors.New("invalid parameter")
				}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("register job should not return error, got: %s", err)
	}

	store.getJobsFunc = func(context.Context) ([]JobDefinition, error) {
		return []JobDefinition{
			{
				Type:           "testFunc",
				CronExpression: "5 4 * * *",
				Params:         to.Ptr("[\"test\"]"),
			},
		}, nil
	}

	err = s.Load()
	if err != nil {
		t.Fatalf("load should not return error, err: %s", err)
	}

	jobs := s.cron.Jobs()
	for _, job := range jobs {
		if job.Name() != "testFunc" {
			t.Fatalf("job name expected to be testFunc, got: %s", job.Name())
		}
	}
}

func TestLoadWithoutRegisteredJob(t *testing.T) {
	s, db := setupLoad(t)
	db.getJobsFunc = func(context.Context) ([]JobDefinition, error) {
		return []JobDefinition{
			{
				Type:           "unknownFunc",
				CronExpression: "5 4 * * *",
				Params:         to.Ptr("[\"test\"]"),
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
	err := s.AddJobHandler(&mockJobHandler{
		name: "testFunc",
		run: func(ctx context.Context, params ...any) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("register job should not return error, got: %s", err)
	}

	db.getJobsFunc = func(context.Context) ([]JobDefinition, error) {
		return []JobDefinition{
			{
				Type:           "testFunc",
				CronExpression: "invalid",
				Params:         to.Ptr("[\"test\"]"),
			},
		}, nil
	}

	err = s.Load()
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

type mockSchedulerStore struct {
	getJobsFunc func(ctx context.Context) ([]JobDefinition, error)
}

func (m *mockSchedulerStore) GetJobs(ctx context.Context) ([]JobDefinition, error) {
	if m.getJobsFunc == nil {
		return nil, errors.New("GetJobsFunc not implemented")
	}
	return m.getJobsFunc(ctx)
}

func TestNewWithoutConfig(t *testing.T) {
	_, err := New(&mockSchedulerStore{}, nil)
	if err == nil {
		t.Fatal("new should return error when config is nil")
	}
}
func TestNew(t *testing.T) {
	s, err := New(&mockSchedulerStore{
		getJobsFunc: func(ctx context.Context) ([]JobDefinition, error) {
			return []JobDefinition{}, nil
		},
	}, &config.Config{})

	if s == nil {
		t.Fatal("new should return a scheduler")
	}
	if err != nil {
		t.Fatal("new should not return error")
	}

	if s.cron == nil {
		t.Fatal("new should return a scheduler with a gocron scheduler")
	}
	if s.store == nil {
		t.Fatal("new should return a scheduler with a dbQueries")
	}
	if s.config == nil {
		t.Fatal("new should return a scheduler with a config")
	}
}

type mockJobHandler struct {
	name         string
	run          func(ctx context.Context, params ...any) error
	validateFunc func(params []any) error
}

func (m *mockJobHandler) Name() string {
	return m.name
}

func (m *mockJobHandler) Execute(ctx context.Context, params ...any) error {
	if m.run == nil {
		return errors.New("run function not implemented")
	}
	return m.run(ctx, params)
}

func (m *mockJobHandler) ValidateParams(params []any) error {
	if m.validateFunc != nil {
		return m.validateFunc(params)
	}
	if m.run == nil {
		return errors.New("run function not implemented")
	}
	// Run the function to validate parameters
	return m.run(context.Background(), params)
}

func TestJobReturningAnError(t *testing.T) {
	s, db := setupLoad(t)

	ch := make(chan bool)
	jobHandler := &mockJobHandler{
		name: "errorFunc",
		run: func(ctx context.Context, params ...any) error {
			ch <- true
			return errors.New("error from the job")
		},
	}
	// Override ValidateParams to not use the run function during validation
	jobHandler.validateFunc = func(params []any) error {
		return nil // Just validate without sending to channel
	}

	if err := s.AddJobHandler(jobHandler); err != nil {
		t.Fatal("register job should not return error")
	}

	db.getJobsFunc = func(context.Context) ([]JobDefinition, error) {
		return []JobDefinition{
			{
				Type:           "errorFunc",
				CronExpression: "5 4 * * *",
				Params:         to.Ptr("[\"test\"]"),
			},
		}, nil
	}

	if err := s.Load(); err != nil {
		t.Fatalf("load should not return error, err: %s", err)
	}

	jobs := s.cron.Jobs()
	s.cron.Start()
	defer func() { _ = s.cron.Shutdown() }()

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
