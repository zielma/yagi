CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
); 

CREATE TABLE IF NOT EXISTS budgets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
); 

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    budget_id TEXT NOT NULL,
    closed BOOLEAN NOT NULL
); 

CREATE TABLE IF NOT EXISTS jobs (
    type TEXT NOT NULL,
    params TEXT,
    cron_expression TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
); 

INSERT INTO jobs (type, params, cron_expression)
VALUES ('fetchBudgets', null, '0 * * * *');
