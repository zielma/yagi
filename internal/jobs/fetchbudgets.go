package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/ynab"
)

// Client represents a client for interacting with the YNAB API
type YNABClient interface {
	GetBudgets(includeAccounts bool) (ynab.BudgetsResponse, error)
}

// Ensure the concrete Client implements the interface
var _ YNABClient = (*ynab.Client)(nil)

// BudgetStore defines the database operations needed by the fetchBudgets job
type BudgetStore interface {
	GetBudget(ctx context.Context, id string) (database.Budget, error)
	CreateBudget(ctx context.Context, arg database.CreateBudgetParams) error
	GetAccount(ctx context.Context, id string) (database.Account, error)
	CreateAccount(ctx context.Context, arg database.CreateAccountParams) error
}

// Ensure the concrete Queries type implements the BudgetStore interface
var _ BudgetStore = (*database.Queries)(nil)

// This task fetches budgets from the YNAB API and stores them in database
type fetchBudgetsJob struct {
	store  BudgetStore
	client YNABClient
}

func newFetchBudgetsJob(store BudgetStore, client YNABClient) *fetchBudgetsJob {
	return &fetchBudgetsJob{
		store:  store,
		client: client,
	}
}

func (j *fetchBudgetsJob) Run() error {
	slog.Debug("starting fetch budgets job...")
	response, err := j.client.GetBudgets(true)
	if err != nil {
		return fmt.Errorf("failed to get budgets from YNAB: %w", err)
	}

	for _, budget := range response.Budgets {
		existing, err := j.store.GetBudget(context.Background(), budget.Id)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get budget[id:%s] from database: %w", budget.Id, err)
		}

		if existing.ID != "" {
			continue
		}

		if err := j.store.CreateBudget(context.Background(), database.CreateBudgetParams{
			ID:   budget.Id,
			Name: budget.Name,
		}); err != nil {
			return fmt.Errorf("failed to create budget[id:%s][name:%s]: %w", budget.Id, budget.Name, err)
		}
	}

	for _, account := range response.Accounts {
		existing, err := j.store.GetAccount(context.Background(), account.Id)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get account[id:%s] from database: %w", account.Id, err)
		}

		if existing.ID != "" {
			continue
		}

		if err := j.store.CreateAccount(context.Background(), database.CreateAccountParams{
			ID:       account.Id,
			Name:     account.Name,
			BudgetID: account.BudgetID,
			Closed:   account.Closed,
		}); err != nil {
			return fmt.Errorf("failed to create account[id:%s][name:%s]: %w", account.Id, account.Name, err)
		}
	}

	return nil
}
