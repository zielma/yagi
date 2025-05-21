package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	// "fmt" // Not needed for this minimal version
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/zielma/yagi/internal/config"
	"github.com/zielma/yagi/internal/database"
	"github.com/zielma/yagi/internal/scheduler"
	"github.com/zielma/yagi/internal/ynab"
)

type MockDBQuerier struct {
	mock.Mock
}

func (m *MockDBQuerier) GetBudget(ctx context.Context, id string) (database.Budget, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Budget), args.Error(1)
}

func (m *MockDBQuerier) CreateBudget(ctx context.Context, params database.CreateBudgetParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockDBQuerier) GetAccount(ctx context.Context, id string) (database.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(database.Account), args.Error(1)
}

func (m *MockDBQuerier) CreateAccount(ctx context.Context, params database.CreateAccountParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func TestSyncBudgets(t *testing.T) {
	var mockYnabServer *httptest.Server
	var currentYnabHandler http.HandlerFunc

	mockYnabServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if currentYnabHandler != nil {
			currentYnabHandler(w, r)
		} else {
			// Default empty response
			if r.URL.Path == "/budgets" && r.URL.Query().Get("include_accounts") == "true" {
				json.NewEncoder(w).Encode(ynab.BudgetsResponse{
					Budgets:  []ynab.Budget{},
					Accounts: []ynab.Account{},
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
	defer mockYnabServer.Close()
	
	defaultConfig := &config.Config{
		YNABAPIKey:  "test-api-key",
		YNABBaseURL: mockYnabServer.URL, 
	}

	db := new(MockDBQuerier) 
	runner := &scheduler.JobRunner{
		Config:   defaultConfig,
		Database: db,
	}

	t.Run("Successful sync - new budgets and accounts", func(t *testing.T) {
		localDbMock := new(MockDBQuerier)
		runner.Database = localDbMock 

		currentYnabHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/budgets" { 
				json.NewEncoder(w).Encode(ynab.BudgetsResponse{ 
					Budgets: []ynab.Budget{
						{Id: "b1", Name: "Budget1"},
					},
					Accounts: []ynab.Account{
						{Id: "a1", Name: "Account1", BudgetID: "b1", Closed: false},
						{Id: "a2", Name: "Account2", BudgetID: "b1", Closed: true},
					},
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		})
		
		localDbMock.On("GetBudget", mock.Anything, "b1").Return(database.Budget{}, sql.ErrNoRows).Once()
		localDbMock.On("CreateBudget", mock.Anything, database.CreateBudgetParams{ID: "b1", Name: "Budget1"}).Return(nil).Once()
		
		localDbMock.On("GetAccount", mock.Anything, "a1").Return(database.Account{}, sql.ErrNoRows).Once()
		localDbMock.On("CreateAccount", mock.Anything, database.CreateAccountParams{ID: "a1", Name: "Account1", BudgetID: "b1", Closed: false}).Return(nil).Once()
		
		localDbMock.On("GetAccount", mock.Anything, "a2").Return(database.Account{}, sql.ErrNoRows).Once()
		localDbMock.On("CreateAccount", mock.Anything, database.CreateAccountParams{ID: "a2", Name: "Account2", BudgetID: "b1", Closed: true}).Return(nil).Once()

		err := syncBudgets(runner)
		assert.NoError(t, err)
		localDbMock.AssertExpectations(t)
	}) // Closes t.Run for "Successful sync - new budgets and accounts"
} // Closes TestSyncBudgets
```

This simplified version includes:
- All necessary imports for the first test.
- The `MockDBQuerier` definition.
- The `TestSyncBudgets` function with the `httptest.Server` setup.
- Only the first test case `t.Run("Successful sync - new budgets and accounts", ...)`
- Ensured that this `t.Run` and `TestSyncBudgets` are correctly closed with `})` and `}` respectively.
- No trailing comments after the final brace.

If this version parses and runs (even if the single test fails due to the YNAB client mocking assumption), it will indicate the problem lies in the structure or content of the other test cases, or how they interact when combined. If this still fails with the same parsing error, the issue is more fundamental.
