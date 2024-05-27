
-- name: CreateDb :one
INSERT INTO dbs (name, owner) VALUES ($1, $2) RETURNING *;

-- name: SetDbExpiredAt :exec
UPDATE dbs SET expired_at = @expired_at WHERE id = @id;

-- name: SetDbStatus :one
UPDATE dbs SET status = @status, stage = @stage, updated_at = timezone('utc', now())
WHERE id = @id AND updated_at = @updated_at
RETURNING *;

-- name: GetDbByName :one
SELECT * FROM dbs WHERE name = @name;

-- name: GetDbByID :one
SELECT * FROM dbs WHERE id = @id;

-- name: ListDbs :many
SELECT * FROM dbs
ORDER BY status, name;
