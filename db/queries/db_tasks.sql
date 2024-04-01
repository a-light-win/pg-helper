
-- name: CreateDbTask :one
INSERT INTO db_tasks (db_id, action, reason, status, data) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: SetDbTaskStatus :exec
UPDATE db_tasks SET status = $2, updated_at = timezone('utc', now()) WHERE id = $1;

-- name: SetDbTaskData :exec
UPDATE db_tasks SET data = $2 WHERE id = $1;

-- name: GetActiveTaskByDbID :one
SELECT * FROM db_tasks
WHERE db_id = $1
AND action = $2 
AND status in ('pending', 'running')
LIMIT 1;

-- name: GetTaskByID :one
SELECT * FROM db_tasks
WHERE id = $1;
