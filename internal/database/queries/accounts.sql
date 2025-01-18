-- name: GetAccount :one
SELECT * FROM accounts WHERE id = ? LIMIT 1;

-- name: GetAccounts :many
SELECT * FROM accounts;

-- name: CreateAccount :exec
INSERT INTO accounts (id, name, budget_id, closed, balance, cleared) 
VALUES (?, ?, ?, ?, ?, ?);