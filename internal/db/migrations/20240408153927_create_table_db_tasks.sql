-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
CREATE TYPE DB_TASK_STATUS AS ENUM (
  'pending',
  'running',
  'cancelling',
  'completed',
  'failed',
  'cancelled'
);

CREATE TYPE DB_ACTION AS ENUM (
  'migrate_out',
  'create_user',
  'create',
  'backup',
  'daily_backup',
  'restore',
  'wait_ready',
  'drop'
);

CREATE TABLE IF NOT EXISTS db_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  db_id BIGINT NOT NULL,
  db_name TEXT NOT NULL,
  action DB_ACTION NOT NULL,
  reason TEXT NOT NULL,
  status DB_TASK_STATUS NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
  updated_at TIMESTAMP NOT NULL DEFAULT timezone('utc', now()),
  data JSONB,
  job_id UUID NOT NULL
);

CREATE INDEX db_tasks_db_id_idx ON db_tasks (db_id);
CREATE INDEX db_tasks_status_idx ON db_tasks (status);
CREATE INDEX db_tasks_job_id_idx ON db_tasks (job_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS db_tasks_job_id_idx;
DROP INDEX IF EXISTS db_tasks_status_idx;
DROP INDEX IF EXISTS db_tasks_db_id_idx;
DROP TABLE IF EXISTS db_tasks;
DROP TYPE IF EXISTS DB_ACTION;
DROP TYPE IF EXISTS DB_TASK_STATUS;
-- +goose StatementEnd
