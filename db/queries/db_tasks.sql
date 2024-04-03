
-- name: CreateDbTask :one
INSERT INTO db_tasks (db_id, action, reason, status, data) VALUES (@db_id, @action, @reason, @status, @data) RETURNING *;

-- name: SetDbTaskStatus :exec
UPDATE db_tasks SET status = @status, updated_at = timezone('utc', now()) WHERE id = @id;

-- name: SetDbTaskData :exec
UPDATE db_tasks SET data = $2 WHERE id = $1;

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
SELECT sqlc.embed(db_tasks),
ARRAY_AGG(db_task_depends.depends_on_task_id)::UUID[] AS depends_on
FROM db_tasks
LEFT JOIN db_task_depends ON db_tasks.id = db_task_depends.task_id
WHERE db_tasks.status in ('pending', 'running')
GROUP BY db_tasks.id;
