package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/ynab"
)

// mockYNABClient is a mock implementation of YNABClient for testing
type mockYNABClient struct {
	getBudgetsFunc func(includeAccounts bool) (ynab.BudgetsResponse, error)
}

func (m *mockYNABClient) GetBudgets(includeAccounts bool) (ynab.BudgetsResponse, error) {
	return m.getBudgetsFunc(includeAccounts)
}

func TestFetchBudgets(t *testing.T) {
	// Create local store and mocks
	budgets := []database.Budget{}
	accounts := []database.Account{}

	mockStore := &mockBudgetStore{
		getBudget: func(ctx context.Context, id string) (database.Budget, error) {
			return database.Budget{}, sql.ErrNoRows
		},
		createBudget: func(ctx context.Context, arg database.CreateBudgetParams) error {
			budgets = append(budgets, database.Budget{
				ID:   arg.ID,
				Name: arg.Name,
			})
			return nil
		},
		updateBudget: func(ctx context.Context, arg database.UpdateBudgetParams) error {
			for i, b := range budgets {
				if b.ID == arg.ID {
					budgets[i].Name = arg.Name
					return nil
				}
			}
			return fmt.Errorf("budget with ID %s not found", arg.ID)
		},
		getAccount: func(ctx context.Context, id string) (database.Account, error) {
			return database.Account{}, sql.ErrNoRows
		},
		createAccount: func(ctx context.Context, arg database.CreateAccountParams) error {
			accounts = append(accounts, database.Account{
				ID:       arg.ID,
				BudgetID: arg.BudgetID,
				Name:     arg.Name,
				Closed:   arg.Closed,
			})

			return nil
		},
	}

	mockClient := &mockYNABClient{
		getBudgetsFunc: func(includeAccounts bool) (ynab.BudgetsResponse, error) {
			return ynab.BudgetsResponse{
				Budgets: []ynab.Budget{
					{
						Id:   "test-budget-id",
						Name: "Test Budget",
					},
				},
				Accounts: []ynab.Account{
					{
						Id:       "test-account-id",
						BudgetID: "test-budget-id",
						Name:     "Test Account",
						Closed:   false,
					},
				},
			}, nil
		},
	}

	// Create and run the job
	job := NewFetchBudgetsJob(mockStore, mockClient)
	if err := job.Run(); err != nil {
		t.Errorf("fetchBudgets returned an error: %v", err)
	}

	// Validate budgets
	if len(budgets) != 1 {
		t.Errorf("expected 1 budget, got %d", len(budgets))
	}
	if budgets[0].ID != "test-budget-id" || budgets[0].Name != "Test Budget" {
		t.Errorf("expected budget ID 'test-budget-id' and name 'Test Budget', got ID '%s' and name '%s'", budgets[0].ID, budgets[0].Name)
	}

	// Validate accounts
	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].ID != "test-account-id" || accounts[0].BudgetID != "test-budget-id" || accounts[0].Name != "Test Account" || accounts[0].Closed {
		t.Errorf("expected account ID 'test-account-id', budget ID 'test-budget-id', name 'Test Account', closed false, got ID '%s', budget ID '%s', name '%s', closed %t", accounts[0].ID, accounts[0].BudgetID, accounts[0].Name, accounts[0].Closed)
	}
}

// mockBudgetStore implements BudgetStore interface for testing
type mockBudgetStore struct {
	getBudget     func(ctx context.Context, id string) (database.Budget, error)
	createBudget  func(ctx context.Context, arg database.CreateBudgetParams) error
	updateBudget  func(ctx context.Context, arg database.UpdateBudgetParams) error
	getAccount    func(ctx context.Context, id string) (database.Account, error)
	createAccount func(ctx context.Context, arg database.CreateAccountParams) error
}

func (m *mockBudgetStore) GetBudget(ctx context.Context, id string) (database.Budget, error) {
	return m.getBudget(ctx, id)
}

func (m *mockBudgetStore) CreateBudget(ctx context.Context, arg database.CreateBudgetParams) error {
	return m.createBudget(ctx, arg)
}

func (m *mockBudgetStore) UpdateBudget(ctx context.Context, arg database.UpdateBudgetParams) error {
	return m.updateBudget(ctx, arg)
}

func (m *mockBudgetStore) GetAccount(ctx context.Context, id string) (database.Account, error) {
	return m.getAccount(ctx, id)
}

func (m *mockBudgetStore) CreateAccount(ctx context.Context, arg database.CreateAccountParams) error {
	return m.createAccount(ctx, arg)
}
