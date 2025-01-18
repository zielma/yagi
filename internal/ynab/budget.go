package ynab

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/zielma/yagi/internal/config"
)

type Budget struct {
	Id   string `db:"id"`
	Name string `db:"name"`
}

type Account struct {
	Id       string `db:"id"`
	BudgetID string `db:"budget_id"`
	Name     string `db:"name"`
	Closed   bool   `db:"closed"`
	Balance  int64  `db:"balance"`
	Cleared  int64  `db:"cleared"`
}

const (
	root = "https://api.ynab.com/v1"
)

type budgetsResponse struct {
	Data budgetData `json:"data"`
}

type budgetData struct {
	Budgets []budgetDetails `json:"budgets"`
}

type budgetDetails struct {
	Id       string        `json:"id"`
	Name     string        `json:"name"`
	Accounts []accountData `json:"accounts"`
}

type accountData struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Closed  bool   `json:"closed"`
	Balance int64  `json:"balance"`
	Cleared int64  `json:"cleared"`
}

type BudgetsResponse struct {
	Budgets  []Budget
	Accounts []Account
}

type Client struct {
	token string
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		token: cfg.YNABAPIKey,
	}
}

func (c *Client) GetBudgets(includeAccounts bool) (BudgetsResponse, error) {
	url := root + "/budgets" + fmt.Sprintf("?include_accounts=%t", includeAccounts)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return BudgetsResponse{}, err
	}
	req.Header.Add("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return BudgetsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return BudgetsResponse{}, fmt.Errorf("failed to get budgets: status %d", resp.StatusCode)
	}

	budgetsResponse := budgetsResponse{}
	slog.Debug("got budgets response", "response", resp)
	if err := json.NewDecoder(resp.Body).Decode(&budgetsResponse); err != nil {
		return BudgetsResponse{}, err
	}

	budgets := []Budget{}
	accounts := []Account{}
	for _, budget := range budgetsResponse.Data.Budgets {
		budgets = append(budgets, Budget{
			Id:   budget.Id,
			Name: budget.Name,
		})

		for _, account := range budget.Accounts {
			accounts = append(accounts, Account{
				Id:       account.Id,
				BudgetID: budget.Id,
				Name:     account.Name,
				Closed:   account.Closed,
				Balance:  account.Balance,
				Cleared:  account.Cleared,
			})
		}
	}

	slog.Info("got budgets", "budgets", budgets)

	response := BudgetsResponse{
		Budgets:  budgets,
		Accounts: accounts,
	}

	return response, nil
}
