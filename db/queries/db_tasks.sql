
-- name: CreateDbTask :one
INSERT INTO db_tasks (db_id, action, reason, status, data) VALUES (@db_id, @action, @reason, @status, @data) RETURNING *;

-- name: SetDbTaskStatus :exec
UPDATE db_tasks SET status = @status, data = jsonb_set(data, '{err_reason}', @err_reason::TEXT), updated_at = timezone('utc', now()) WHERE id = @id;

-- name: SetDbTaskData :exec
UPDATE db_tasks SET data = @data WHERE id = @id;

-- name: GetActiveDbTaskByDbID :one
SELECT * FROM db_tasks
WHERE db_id = @db_id
AND action = @action
AND status in ('pending', 'running')
LIMIT 1;

-- name: ListActiveDbTasksByDbID :many
SELECT * FROM db_tasks
WHERE db_id = @db_id
AND action = @action
AND status in ('pending', 'running');

-- name: GetTaskByID :one
SELECT * FROM db_tasks
WHERE id = @id;

-- name: ListActiveDbTasks :many
SELECT *
FROM db_tasks
WHERE db_tasks.status in ('pending', 'running');
