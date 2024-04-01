
-- name: CreateDb :exec
INSERT INTO dbs (name, owner) VALUES ($1, $2);

-- name: SetDbExpiresAt :exec
UPDATE dbs SET expires_at = $2 WHERE id = $1;

-- name: DisableDb :exec
UPDATE dbs SET is_enabled = FALSE, disabled_at = timezone('utc', now()) WHERE id = $1;

-- name: GetDbByName :one
SELECT * FROM dbs WHERE name = $1;

-- name: GetDbByID :one
SELECT * FROM dbs WHERE id = $1;
