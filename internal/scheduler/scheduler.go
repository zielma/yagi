package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/to"
)

type JobDefinition struct {
	Type           string
	CronExpression string
	Params         *string
}

func (jd *JobDefinition) HasParams() bool {
	return jd.Params != nil && *jd.Params != ""
}

func (jd *JobDefinition) GetParams() string {
	if jd.Params != nil {
		return *jd.Params
	}
	return ""
}

// Scheduler manages the scheduling and execution of jobs
type Scheduler struct {
	store    SchedulerStore
	config   *config.Config
	cron     gocron.Scheduler
	handlers map[string]JobHandler // jobName->jobHandler
	mu       sync.Mutex            // to protect concurrent access to jobs map
}

type SchedulerStore interface {
	GetJobs(ctx context.Context) ([]JobDefinition, error)
}

// RegisterJob registers a new job with the scheduler
func (s *Scheduler) AddJobHandler(handler JobHandler) error {
	if handler == nil {
		return fmt.Errorf("job cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.handlers[handler.Name()]; exists {
		return fmt.Errorf("job %s already registered", handler.Name())
	}

	s.handlers[handler.Name()] = handler
	return nil
}

func (s *Scheduler) Shutdown() {
	if err := s.cron.Shutdown(); err != nil {
		slog.Error("failed to shutdown scheduler", "error", err)
	} else {
		slog.Info("scheduler shutdown successfully")
	}
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	slog.Info("loading jobs execution plan from database")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := s.store.GetJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	for _, job := range jobs {
		slog.Debug("found job", "job_type", job.Type, "job_cron_expression", job.CronExpression, "job_params", job.Params)

		// Get the job implementation
		handler, exists := s.handlers[job.Type]
		if !exists {
			return fmt.Errorf("no implementation found for job type: %s", job.Type)
		}

		// Parse job parameters
		var params []any
		if job.HasParams() {
			p := job.GetParams()
			slog.Debug("decoding job params", "params", p)
			if err := json.NewDecoder(strings.NewReader(p)).Decode(&params); err != nil {
				return fmt.Errorf("failed to decode job params[%s]: %w", p, err)
			}
		}

		// Validate parameters before scheduling
		slog.Debug("validating job parameters", "job_type", job.Type, "params", params)
		if err := handler.ValidateParams(params); err != nil {
			return fmt.Errorf("invalid parameters for job %s: %w", job.Type, err)
		}

		// Create job definition with gocron
		slog.Debug("creating job", "job_type", job.Type, "job_cron_expression", job.CronExpression, "job_params", params)
		scheduledJob, err := s.cron.NewJob(
			gocron.CronJob(job.CronExpression, false),
			gocron.NewTask(
				func() error {
					return handler.Execute(context.Background(), params...)
				},
			),
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

		// Retrieve 3 next run times and log them
		nextRuns, err := scheduledJob.NextRuns(3)
		if err != nil {
			return fmt.Errorf("failed to get next runs: %w", err)
		}
		slog.Info("next runs", "next_run", nextRuns)
	}

	return nil
}

func New(store SchedulerStore,
	config *config.Config,
	handlers ...JobHandler) (*Scheduler, error) {
	if store == nil {
		return nil, fmt.Errorf("must specify a store for scheduler")
	}
	if config == nil {
		return nil, fmt.Errorf("must specify a config")
	}

	cron, err := gocron.NewScheduler(
		gocron.WithLocation(time.Now().Location()),
		gocron.WithLogger(slog.Default()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Create a map of job handlers (name, handler) used by the scheduler and load it
	// with supplied by the client handlers.
	hm := make(map[string]JobHandler)
	for _, h := range handlers {
		if h == nil {
			return nil, fmt.Errorf("job handler cannot be nil")
		}
		if _, exists := hm[h.Name()]; exists {
			return nil, fmt.Errorf("job handler %s already registered", h.Name())
		}
		hm[h.Name()] = h
	}

	return &Scheduler{
		store:    store,
		config:   config,
		cron:     cron,
		handlers: hm,
	}, nil
}

type schedulerStore struct {
	db *database.Queries
}

// GetJobs retrieves all job definitions from the database and converts database model to JobDefinition.
func (s *schedulerStore) GetJobs(ctx context.Context) ([]JobDefinition, error) {
	rows, err := s.db.GetJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs from database: %w", err)
	}

	var jobs []JobDefinition
	for _, row := range rows {
		var params *string
		if row.Params.Valid {
			params = to.Ptr(row.Params.String)
		}

		jobs = append(jobs, JobDefinition{
			Type:           row.Type,
			CronExpression: row.CronExpression,
			Params:         params,
		})
	}

	return jobs, nil
}

// NewStore creates a new SchedulerStore instance.
func NewStore(db *database.Queries) SchedulerStore {
	return &schedulerStore{db: db}
}
