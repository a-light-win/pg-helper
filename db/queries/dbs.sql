
-- name: CreateDb :one
INSERT INTO dbs (name, owner) VALUES ($1, $2) RETURNING *;

-- name: SetDbExpiredAt :exec
UPDATE dbs SET expired_at = @expired_at WHERE id = @id;

-- name: SetDbStatus :exec
UPDATE dbs SET status = @status, updated_at = timezone('utc', now()) WHERE id = @id;

-- name: GetDbByName :one
SELECT * FROM dbs WHERE name = @name;

-- name: GetDbByID :one
SELECT * FROM dbs WHERE id = @id;

-- name: ListDbs :many
SELECT * FROM dbs
ORDER BY status, name;
