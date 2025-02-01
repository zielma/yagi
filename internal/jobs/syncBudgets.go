package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/scheduler"
	"github.com/zielma/yagi/internal/ynab"
)

func syncBudgets(r *scheduler.JobRunner) error {
	slog.Debug("syncing budgets")
	client := ynab.NewClient(r.Config)
	response, err := client.GetBudgets(true)
	if err != nil {
		return fmt.Errorf("failed to get budgets from YNAB: %w", err)
	}

	for _, budget := range response.Budgets {
		existing, err := r.Database.GetBudget(context.Background(), budget.Id)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get budget[id:%s] from database: %w", budget.Id, err)
		}

		if existing.ID != "" {
			continue
		}

		if err := r.Database.CreateBudget(context.Background(), database.CreateBudgetParams{
			ID:   budget.Id,
			Name: budget.Name,
		}); err != nil {
			return fmt.Errorf("failed to create budget[id:%s][name:%s]: %w", budget.Id, budget.Name, err)
		}
	}

	for _, account := range response.Accounts {
		existing, err := r.Database.GetAccount(context.Background(), account.Id)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get account[id:%s] from database: %w", account.Id, err)
		}

		if existing.ID != "" {
			continue
		}

		if err := r.Database.CreateAccount(context.Background(), database.CreateAccountParams{
			ID:       account.Id,
			Name:     account.Name,
			BudgetID: account.BudgetID,
			Closed:   account.Closed,
			Balance:  account.Balance,
			Cleared:  account.Cleared,
		}); err != nil {
			return fmt.Errorf("failed to create account[id:%s][name:%s]: %w", account.Id, account.Name, err)
		}
	}

	return nil
}
