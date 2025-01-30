-- name: GetBudget :one
SELECT * FROM budgets WHERE id = ? LIMIT 1;

-- name: CreateBudget :exec
INSERT INTO budgets (id, name) VALUES (?, ?);

-- name: GetBudgets :many
SELECT * FROM budgets;

