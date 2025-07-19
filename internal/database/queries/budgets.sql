-- name: GetBudget :one
SELECT * FROM budgets WHERE id = ? LIMIT 1;

-- name: CreateBudget :exec
INSERT INTO budgets (id, name) VALUES (?, ?);

-- name: UpdateBudget :exec
UPDATE budgets
SET name = ?
    ,updated_at = datetime('now')
WHERE id = ?;

-- name: GetBudgets :many
SELECT * FROM budgets;

