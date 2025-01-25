package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/ynab"
)

type Scheduler struct {
	query     *database.Queries
	cfg       *config.Config
	scheduler gocron.Scheduler
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

func New(db *sql.DB, cfg *config.Config) *Scheduler {
	s := Scheduler{query: database.New(db), cfg: cfg}

	var err error
	s.scheduler, err = gocron.NewScheduler()
	if err != nil {
		slog.Error("failed to create scheduler", "error", err)
		return nil
	}

	_, err = s.scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(2*time.Second)),
		),
		gocron.NewTask(s.syncBudgets),
		gocron.WithEventListeners(
			gocron.AfterJobRunsWithError(func(jobID uuid.UUID, jobName string, err error) {
				slog.Error("job failed", "job_id", jobID, "job_name", jobName, "error", err)
			}),
		),
	)
	if err != nil {
		slog.Error("failed to create job", "error", err)
	}

	s.scheduler.Start()

	return &s
}

func (s *Scheduler) syncBudgets() error {

	slog.Debug("syncing budgets")

	client := ynab.NewClient(s.cfg)
	response, err := client.GetBudgets(true)
	if err != nil {
		slog.Debug("failed to get budgets", "error", err)
		return err
	}

	for _, budget := range response.Budgets {
		existing, err := s.query.GetBudget(context.Background(), budget.Id)
		if err != nil && err != sql.ErrNoRows {
			slog.Debug("failed to get budget", "error", err)
			return err
		}

		if existing.ID != "" {
			continue
		}

		if err := s.query.CreateBudget(context.Background(), database.CreateBudgetParams{
			ID:   budget.Id,
			Name: budget.Name,
		}); err != nil {
			slog.Debug("failed to create budget", "error", err)
			return err
		}
	}

	for _, account := range response.Accounts {
		existing, err := s.query.GetAccount(context.Background(), account.Id)
		if err != nil && err != sql.ErrNoRows {
			slog.Debug("failed to get account", "error", err)
			return err
		}

		if existing.ID != "" {
			continue
		}

		if err := s.query.CreateAccount(context.Background(), database.CreateAccountParams{
			ID:       account.Id,
			Name:     account.Name,
			BudgetID: account.BudgetID,
			Closed:   account.Closed,
			Balance:  account.Balance,
			Cleared:  account.Cleared,
		}); err != nil {
			slog.Debug("failed to create account", "error", err)
			return err
		}
	}

	return nil
}
