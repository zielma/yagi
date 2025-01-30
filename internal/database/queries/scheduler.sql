-- name: GetJobs :many
SELECT id
 ,type
 ,status
 ,cron_expression
FROM jobs;

-- name: UpdateJobCronExpression :exec
UPDATE jobs 
SET cron_expression = ?
    ,updated_at = datetime('now')
WHERE id = ?;