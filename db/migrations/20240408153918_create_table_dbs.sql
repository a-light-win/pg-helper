-- +goose NO TRANSACTION

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS dbs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    owner VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
    expires_at TIMESTAMP,
    disabled_at TIMESTAMP,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE
  );

CREATE UNIQUE INDEX dbs_name_idx ON dbs (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS dbs_name_idx;
DROP TABLE IF EXISTS dbs;
-- +goose StatementEnd
