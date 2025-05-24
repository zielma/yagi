-- name: GetJobs :many
SELECT type
 ,cron_expression
 ,params
FROM jobs;

-- name: UpdateJobCronExpression :exec
UPDATE jobs 
SET cron_expression = ?
    ,updated_at = datetime('now')
WHERE type = ?;