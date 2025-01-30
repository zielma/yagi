package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/scheduler/jobs"
)

type dbQueries interface {
	GetJobs(ctx context.Context) ([]database.GetJobsRow, error)
}

type Scheduler struct {
	dbQueries dbQueries
	cfg       *config.Config
	scheduler gocron.Scheduler
	jobRunner *jobs.Runner
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

func (s *Scheduler) Reload() {
	s.scheduler.StopJobs()
	for _, j := range s.scheduler.Jobs() {
		s.scheduler.RemoveJob(j.ID())
	}

	s.Load()
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
		slog.Info("found job", "job_type", job.Type, "job_cron_expression", job.CronExpression)
		scheduledJob, err := s.scheduler.NewJob(
			gocron.CronJob(job.CronExpression, false),
			gocron.NewTask(s.jobRunner.GetJobFunc(job.Type)),
			gocron.WithEventListeners(
				gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, joberr error) {
					slog.Error("job failed", "job_id", jobID, "job_name", jobName, "error", joberr)
				}),
			),
		)
		if err != nil {
			return fmt.Errorf("failed to create new job: %w", err)
		}

		nextRuns, err := scheduledJob.NextRuns(3)
		if err != nil {
			return fmt.Errorf("failed to get next runs: %w", err)
		}

		for i, nextRun := range nextRuns {
			slog.Info("next run", "index", i, "next_run", nextRun)
		}
	}

	return nil
}
func New(db *sql.DB, cfg *config.Config) *Scheduler {
	dbQueries := database.New(db)
	s := Scheduler{
		dbQueries: dbQueries,
		cfg:       cfg,
		jobRunner: jobs.NewRunner(dbQueries, cfg)}

	var err error
	s.scheduler, err = gocron.NewScheduler(
		gocron.WithLocation(time.Now().Location()),
		gocron.WithLogger(slog.Default()),
	)

	if err != nil {
		slog.Error("failed to create scheduler", "error", err)
		return nil
	}

	s.Load()
	s.scheduler.Start()

	return &s
}
