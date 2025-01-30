package jobs

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/ynab"
)

func (s *Runner) syncBudgets() error {
	slog.Debug("syncing budgets")
	client := ynab.NewClient(s.cfg)
	response, err := client.GetBudgets(true)
	if err != nil {
		slog.Debug("failed to get budgets", "error", err)
		return err
	}

	for _, budget := range response.Budgets {
		existing, err := s.dbQueries.GetBudget(context.Background(), budget.Id)
		if err != nil && err != sql.ErrNoRows {
			slog.Debug("failed to get budget", "error", err)
			return err
		}

		if existing.ID != "" {
			continue
		}

		if err := s.dbQueries.CreateBudget(context.Background(), database.CreateBudgetParams{
			ID:   budget.Id,
			Name: budget.Name,
		}); err != nil {
			slog.Debug("failed to create budget", "error", err)
			return err
		}
	}

	for _, account := range response.Accounts {
		existing, err := s.dbQueries.GetAccount(context.Background(), account.Id)
		if err != nil && err != sql.ErrNoRows {
			slog.Debug("failed to get account", "error", err)
			return err
		}

		if existing.ID != "" {
			continue
		}

		if err := s.dbQueries.CreateAccount(context.Background(), database.CreateAccountParams{
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
