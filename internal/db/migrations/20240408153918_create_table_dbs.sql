-- +goose NO TRANSACTION

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dbs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
    updated_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
    expired_at TIMESTAMP,

    migrate_from VARCHAR(255) NOT NULL DEFAULT '',
    migrate_to VARCHAR(255) NOT NULL DEFAULT '',

    status int4 NOT NULL DEFAULT 0,
    stage int4 NOT NULL DEFAULT 0,

    error_msg TEXT NOT NULL DEFAULT ''
  );

CREATE UNIQUE INDEX dbs_name_idx ON dbs (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS dbs_name_idx;
DROP TABLE IF EXISTS dbs;
-- +goose StatementEnd
