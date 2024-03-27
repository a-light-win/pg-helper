
-- name: CreateDbTask :exec
INSERT INTO db_tasks (db_id, action, reason, status) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: SetDbTaskStatus :exec
UPDATE db_tasks SET status = $2, updated_at = timezone('utc', now()) WHERE id = $1;

-- name: SetDbTaskData :exec
UPDATE db_tasks SET data = $2 WHERE id = $1;
