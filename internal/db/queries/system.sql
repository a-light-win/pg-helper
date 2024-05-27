
-- name: IsUserExists :one
SELECT true FROM pg_catalog.pg_user WHERE usename = $1;

-- name: GetDbOwner :one
SELECT r.rolname as owner FROM pg_catalog.pg_database dbs
JOIN pg_catalog.pg_roles r ON r.oid = dbs.datdba
WHERE dbs.datname = @dbName;

-- name: CountDbTables :one
SELECT COUNT(*) FROM pg_catalog.pg_tables
WHERE schemaname not in ('pg_catalog', 'information_schema', 'pg_toast');
