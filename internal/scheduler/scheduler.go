package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
)

type dbQueries interface {
	GetJobs(ctx context.Context) ([]database.GetJobsRow, error)
}

type Scheduler struct {
	dbQueries dbQueries
	cfg       *config.Config
	scheduler gocron.Scheduler
	jobRunner *JobRunner
}

type _ struct {
	Id        string
	Type      string
	NextRunAt time.Time
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Scheduler) Shutdown() {
	_ = s.scheduler.Shutdown()
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
}

func (s *Scheduler) Load() error {
	slog.Info("loading jobs from database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := s.dbQueries.GetJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for _, job := range jobs {
		slog.Debug("found job", "job_type", job.Type, "job_cron_expression", job.CronExpression, "job_params", job.Params)
		// decode job parameters to slice, position of arguments matters
		jsonParams := []any{}
		if job.Params.Valid {
			if err := json.NewDecoder(strings.NewReader(job.Params.String)).Decode(&jsonParams); err != nil {
				return fmt.Errorf("failed to decode job params[%s]: %w", job.Params.String, err)
			}
		}

		jobFunc, err := getJobFunc(job.Type)
		if err != nil && errors.Is(err, ErrUnknownJobType) {
			return fmt.Errorf("getJobFunc[%s]: %w", job.Type, err)
		}

		// add job to the scheduler
		// we always want to pass a value of jobRunner first and then optional parameters
		taskParams := append([]any{s.jobRunner}, jsonParams...)

		slog.Debug("creating job", "job_type", job.Type, "job_cron_expression", job.CronExpression, "job_params", job.Params)
		scheduledJob, err := s.scheduler.NewJob(
			gocron.CronJob(job.CronExpression, false),
			gocron.NewTask(jobFunc, taskParams...),
			gocron.WithName(job.Type),
			gocron.WithEventListeners(
				gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, joberr error) {
					slog.Error("job failed", "job_id", jobID, "job_name", jobName, "error", joberr)
				}),
			),
		)
		if err != nil {
			return fmt.Errorf("failed to create new job: %w", err)
		}

		// get next 3 run times and log them
		nextRuns, err := scheduledJob.NextRuns(3)
		if err != nil {
			return fmt.Errorf("failed to get next runs: %w", err)
		}

		slog.Info("next runs", "next_run", nextRuns)
	}

	return nil
}

func New(db *sql.DB, cfg *config.Config) (*Scheduler, error) {
	if db == nil {
		return nil, fmt.Errorf("must specify a database connection")
	}
	if cfg == nil {
		return nil, fmt.Errorf("must specify a config")
	}

	dbQueries := database.New(db)
	gcs, err := gocron.NewScheduler(
		gocron.WithLocation(time.Now().Location()),
		gocron.WithLogger(slog.Default()),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &Scheduler{
		dbQueries: dbQueries,
		cfg:       cfg,
		jobRunner: NewJobRunner(dbQueries, cfg),
		scheduler: gcs,
	}, nil
}
